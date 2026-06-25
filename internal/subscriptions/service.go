package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/koitra/subserv/internal/uuidext"
)

type (
	Service interface {
		Add(ctx context.Context, sub NewSubscription) (Subscription, error)
	}

	Subscription struct {
		ID          uuid.UUID
		ServiceName string
		Price       int32
		UserID      uuid.UUID
		StartDate   time.Time
		EndDate     *time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	NewSubscription struct {
		ServiceName string     `validate:"required"`
		Price       int32      `validate:"gte=0"`
		UserID      uuid.UUID  `validate:"required,uuid"`
		StartDate   time.Time  `validate:"required"`
		EndDate     *time.Time `validate:"omitnil,required,gtefield=StartDate"`
	}

	Svc struct {
		r        Repository
		validate *validator.Validate
	}
)

func NewService(r Repository, validate *validator.Validate) *Svc {
	return &Svc{r, validate}
}

func (s *Svc) Add(ctx context.Context, add NewSubscription) (Subscription, error) {
	err := s.validate.Struct(&add)
	if err != nil {
		return Subscription{}, fmt.Errorf("invalid subscription: %w", err)
	}

	now := time.Now()
	sub := newSubscription(
		uuidext.NewV7(),
		add.ServiceName,
		add.Price,
		add.UserID,
		add.StartDate,
		add.EndDate,
		now,
		now,
	)

	err = s.r.Create(ctx, sub)
	if err != nil {
		return Subscription{}, fmt.Errorf("create new subscription: %w", err)
	}

	return sub, nil
}

func (sub *Subscription) normalizeTime() {
	sub.StartDate = sub.StartDate.Truncate(time.Millisecond).UTC()
	if sub.EndDate != nil {
		*sub.EndDate = sub.EndDate.Truncate(time.Millisecond).UTC()
	}

	sub.UpdatedAt = sub.UpdatedAt.Truncate(time.Millisecond).UTC()
	sub.CreatedAt = sub.CreatedAt.Truncate(time.Millisecond).UTC()
}

func newSubscription(
	id uuid.UUID,
	serviceName string,
	price int32,
	userID uuid.UUID,
	startDate time.Time,
	endDate *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Subscription {
	s := Subscription{
		ID:          id,
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	s.normalizeTime()

	return s
}
