package subscriptions

import (
	"context"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"

	"github.com/koitra/subserv/internal/cursor"
	"github.com/koitra/subserv/internal/dba"
	"github.com/koitra/subserv/internal/dba/models"
)

type (
	Repository interface {
		Create(ctx context.Context, sub Subscription) error
		Find(ctx context.Context, criteria Criteria) ([]Subscription, error)
		Update(ctx context.Context, sub Subscription) error
		Delete(ctx context.Context, subID uuid.UUID) error
	}

	PgRepository struct {
		db bob.DB
	}
)

func NewRepository(db bob.DB) *PgRepository {
	return &PgRepository{db}
}

func (r *PgRepository) Create(ctx context.Context, sub Subscription) error {
	_, err := models.Subscriptions.Insert(sub.setter()).Exec(ctx, r.db)
	if err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}
	return nil
}

func (r *PgRepository) Delete(ctx context.Context, subID uuid.UUID) error {
	_, err := models.Subscriptions.Delete(models.DeleteWhere.Subscriptions.ID.EQ(subID)).
		Exec(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	return nil
}

func (r *PgRepository) Find(ctx context.Context, criteria Criteria) ([]Subscription, error) {
	q := models.Subscriptions.Query(
		dba.InIfNotEmpty(models.SelectWhere.Subscriptions.ID, criteria.IDs...),
		dba.InIfNotEmpty(models.SelectWhere.Subscriptions.UserID, criteria.UserIDs...),
		dba.InIfNotEmpty(models.SelectWhere.Subscriptions.ServiceName, criteria.ServiceNames...),
		cursor.SelectMod(criteria.Cursor, models.Subscriptions.Columns.CreatedAt),
	)

	subs, err := q.All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("select subscriptions: %w", err)
	}

	count := len(subs)
	res := make([]Subscription, count)

	for i, s := range subs {
		sub := subscriptionFromDB(s)
		res[i] = sub
	}

	return res, nil
}

func (r *PgRepository) Update(ctx context.Context, sub Subscription) error {
	setter := sub.setter()
	setter.ID = omit.Val[uuid.UUID]{}

	rowsCount, err := models.Subscriptions.Update(
		models.UpdateWhere.Subscriptions.ID.EQ(sub.ID),
		setter.UpdateMod(),
	).Exec(ctx, r.db)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}

	if rowsCount == 0 {
		return UnknownSubscriptionError{ID: sub.ID}
	}

	return nil
}

func (sub *Subscription) setter() *models.SubscriptionSetter {
	return &models.SubscriptionSetter{
		ID:          omit.From(sub.ID),
		ServiceName: omit.From(sub.ServiceName),
		Price:       omit.From(sub.Price),
		UserID:      omit.From(sub.UserID),
		StartDate:   omit.From(sub.StartDate),
		EndDate:     omitnull.FromPtr(sub.EndDate),
		CreatedAt:   omit.From(sub.CreatedAt),
		UpdatedAt:   omit.From(sub.UpdatedAt),
	}
}

func subscriptionFromDB(db *models.Subscription) Subscription {
	return newSubscription(
		db.ID,
		db.ServiceName,
		db.Price,
		db.UserID,
		db.StartDate,
		db.EndDate.Ptr(),
		db.CreatedAt,
		db.UpdatedAt,
	)
}
