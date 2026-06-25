package subscriptions

import (
	"context"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/stephenafamo/bob"

	"github.com/koitra/subserv/internal/dba/models"
)

type (
	Repository interface {
		Create(ctx context.Context, sub Subscription) error
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
	sub := Subscription{
		ID:          db.ID,
		ServiceName: db.ServiceName,
		Price:       db.Price,
		UserID:      db.UserID,
		StartDate:   db.StartDate,
		EndDate:     db.EndDate.Ptr(),
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
	sub.normalizeTime()
	return sub
}
