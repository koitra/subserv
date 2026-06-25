package subscriptions

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stretchr/testify/require"

	"github.com/koitra/subserv/internal/cursor"
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

func TestFind(t *testing.T) {
	t.Parallel()
	r, _, ctx := testEnv(t)

	now := time.Now()
	subs := []Subscription{
		{
			ID:          uuidext.NewV7(),
			ServiceName: "service 1",
			Price:       100,
			UserID:      uuidext.NewV7(),
			StartDate:   now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},

		{
			ID:          uuidext.NewV7(),
			ServiceName: "service 1",
			Price:       10,
			UserID:      uuidext.NewV7(),
			StartDate:   now,
			CreatedAt:   now.Add(time.Hour),
			UpdatedAt:   now.Add(time.Hour),
		},

		{
			ID:          uuidext.NewV7(),
			ServiceName: "service 2",
			Price:       30,
			UserID:      uuidext.NewV7(),
			StartDate:   now,
			CreatedAt:   now.Add(time.Hour * 2),
			UpdatedAt:   now.Add(time.Hour * 2),
		},
	}

	for i := range subs {
		sub := &subs[i]
		sub.normalizeTime()

		err := r.Create(ctx, *sub)
		require.NoError(t, err)
	}

	t.Run("empty criteria", func(t *testing.T) {
		t.Parallel()

		found, err := r.Find(ctx, Criteria{})
		require.NoError(t, err)
		require.Len(t, found, 3)
	})

	t.Run("empty result", func(t *testing.T) {
		t.Parallel()

		found, err := r.Find(ctx, Criteria{
			IDs: uuid.UUIDs{uuidext.NewV7()},
		})
		require.NoError(t, err)
		require.Len(t, found, 0)
	})

	t.Run("by IDs", func(t *testing.T) {
		t.Parallel()

		ids := uuid.UUIDs{subs[1].ID, subs[0].ID}
		found, err := r.Find(ctx, Criteria{
			IDs: ids,
		})
		require.NoError(t, err)
		require.Len(t, found, 2)
		for i := range ids {
			require.Equal(t, ids[i], found[i].ID)
		}
	})

	t.Run("by Names", func(t *testing.T) {
		t.Parallel()

		names := []string{"service 1"}
		found, err := r.Find(ctx, Criteria{
			ServiceNames: names,
		})
		require.NoError(t, err)
		require.Len(t, found, 2)
		for i := range names {
			require.Equal(t, names[i], found[i].ServiceName)
		}
	})

	t.Run("cursor", func(t *testing.T) {
		t.Parallel()

		found, err := r.Find(ctx, Criteria{Cursor: cursor.Cursor{Limit: 2}})
		require.NoError(t, err)
		require.Len(t, found, 2)

		found, err = r.Find(
			ctx,
			Criteria{Cursor: cursor.Cursor{Limit: 2, At: &found[len(found)-1].CreatedAt}},
		)
		require.NoError(t, err)
		require.Len(t, found, 1)
		time.Sleep(time.Second)
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("unknown subscription", func(t *testing.T) {
		t.Parallel()
		r, _, ctx := testEnv(t)

		s := Subscription{
			ID: uuidext.NewV7(),
		}
		err := r.Update(ctx, s)
		expect := UnknownSubscriptionError{ID: s.ID}
		require.Equal(t, expect, err)
	})

	t.Run("updates", func(t *testing.T) {
		t.Parallel()
		r, db, ctx := testEnv(t)

		s1T := time.Now().UTC()
		s2T := s1T.Add(time.Millisecond * 10)

		subs := []Subscription{
			{
				ID:          uuidext.NewV7(),
				ServiceName: "s1",
				Price:       10,
				UserID:      uuidext.NewV7(),
				StartDate:   time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				CreatedAt:   s1T,
				UpdatedAt:   s1T,
			},
			{
				ID:          uuidext.NewV7(),
				ServiceName: "s2",
				Price:       100,
				UserID:      uuidext.NewV7(),
				StartDate:   time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				CreatedAt:   s2T,
				UpdatedAt:   s2T,
			},
		}

		for i := range subs {
			sub := &subs[i]
			sub.normalizeTime()

			err := r.Create(ctx, *sub)
			require.NoError(t, err)
		}

		subs[1].ServiceName = "s2 updated"
		subs[1].EndDate = new(time.Now().Add(time.Hour * 10))
		subs[1].normalizeTime()

		err := r.Update(ctx, subs[1])
		require.NoError(t, err)

		stored, err := models.Subscriptions.Query(sm.OrderBy(models.Subscriptions.Columns.CreatedAt).Asc()).
			All(
				ctx, db)
		require.NoError(t, err)
		require.Len(t, stored, len(subs))

		for i, sub := range stored {
			s := subscriptionFromDB(sub)
			require.EqualValues(t, subs[i], s)
		}
	})
}

func testEnv(t *testing.T) (Repository, bob.DB, context.Context) {
	db := testdb.New(t)
	r := NewRepository(db)
	ctx := t.Context()

	return r, db, ctx
}
