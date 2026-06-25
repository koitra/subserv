package subscriptions

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stretchr/testify/require"

	"github.com/koitra/subserv/internal/dba/factory"
	"github.com/koitra/subserv/internal/dba/models"
	"github.com/koitra/subserv/internal/testdb"
	"github.com/koitra/subserv/internal/uuidext"
)

func TestCreate(t *testing.T) {
	t.Parallel()
	r, db, ctx := testEnv(t)

	s1T := time.Now().UTC()
	s2T := s1T.Add(time.Millisecond * 10)

	subs := []Subscription{
		newSubscription(uuidext.NewV7(),
			"s1",
			10,
			uuidext.NewV7(),
			time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			nil,
			s1T,
			s1T,
		),
		newSubscription(uuidext.NewV7(),
			"s2",
			100,
			uuidext.NewV7(),
			time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			nil,
			s2T,
			s2T,
		),
	}

	for _, sub := range subs {
		err := r.Create(ctx, sub)
		require.NoError(t, err)
	}

	stored, err := models.Subscriptions.Query(sm.OrderBy(models.Subscriptions.Columns.CreatedAt).Asc()).
		All(
			ctx, db)
	require.NoError(t, err)
	require.Len(t, stored, len(subs))

	for i, sub := range stored {
		s := subscriptionFromDB(sub)
		require.EqualValues(t, subs[i], s)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	r, db, ctx := testEnv(t)

	err := r.Delete(ctx, uuidext.NewV7())
	require.NoError(t, err)

	count := 10
	subs := factory.New().NewSubscription().MustCreateMany(ctx, db, count)

	err = r.Delete(ctx, subs[0].ID)
	require.NoError(t, err)

	_, err = models.FindSubscription(ctx, db, subs[0].ID)
	require.ErrorAs(t, err, new(sql.ErrNoRows))
}

func testEnv(t *testing.T) (Repository, bob.DB, context.Context) {
	db := testdb.New(t)
	r := NewRepository(db)
	ctx := t.Context()

	return r, db, ctx
}
