package subscriptions

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/koitra/subserv/internal/cursor"
	"github.com/koitra/subserv/internal/uuidext"
)

type (
	Service interface {
		Add(ctx context.Context, sub NewSubscription) (Subscription, error)
		Find(ctx context.Context, criteria Criteria) (FindResult, error)
		Update(ctx context.Context, update UpdateSubscription) (Subscription, error)
		Remove(ctx context.Context, subID uuid.UUID) error
		Total(ctx context.Context, criteria TotalCriteria) (int64, error)
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

	UpdateSubscription struct {
		ID          uuid.UUID  `validate:"required,uuid"`
		ServiceName string     `validate:"required"`
		Price       int32      `validate:"gte=0"`
		UserID      uuid.UUID  `validate:"required,uuid"`
		StartDate   time.Time  `validate:"required"`
		EndDate     *time.Time `validate:"omitnil,gtfield=StartDate"`
	}

	Criteria struct {
		IDs          uuid.UUIDs
		ServiceNames []string
		UserIDs      uuid.UUIDs

		Cursor cursor.Cursor
	}
	FindResult struct {
		Subscriptions []Subscription

		Cursor *time.Time
	}

	TotalCriteria struct {
		UserID      *uuid.UUID
		ServiceName *string
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

	slog.Info("Created new subscription", slog.String("subscriptionID", sub.ID.String()))

	return sub, nil
}

func (s *Svc) Remove(ctx context.Context, subID uuid.UUID) error {
	err := s.r.Delete(ctx, subID)
	if err != nil {
		return err
	}

	slog.Info("Removed subscription", slog.String("subscriptionID", subID.String()))
	return nil
}

func (s *Svc) Find(ctx context.Context, criteria Criteria) (FindResult, error) {
	subs, err := s.r.Find(ctx, criteria)
	if err != nil {
		return FindResult{}, err
	}

	res := FindResult{
		Subscriptions: subs,
	}

	count := len(subs)
	if criteria.Cursor.Limit > 0 && count >= criteria.Cursor.Limit {
		res.Cursor = &res.Subscriptions[count-1].CreatedAt
	}

	return res, nil
}

func (s *Svc) Update(ctx context.Context, update UpdateSubscription) (Subscription, error) {
	err := s.validate.Struct(&update)
	if err != nil {
		return Subscription{}, fmt.Errorf("invalid subscription: %w", err)
	}

	subs, err := s.r.Find(ctx, Criteria{
		IDs: uuid.UUIDs{update.ID},
	})
	if err != nil {
		return Subscription{}, fmt.Errorf("find subscriptions: %w", err)
	}
	if len(subs) == 0 {
		return Subscription{}, UnknownSubscriptionError{ID: update.ID}
	}

	sub := subs[0]
	sub.ServiceName = update.ServiceName
	sub.Price = update.Price
	sub.UserID = update.UserID
	sub.StartDate = update.StartDate
	sub.EndDate = update.EndDate
	sub.UpdatedAt = time.Now()
	sub.normalizeTime()

	err = s.r.Update(ctx, sub)
	if err != nil {
		return Subscription{}, fmt.Errorf("update subscription: %w", err)
	}
	slog.Info("Updated subscription", slog.String("subscriptionID", sub.ID.String()))
	return sub, nil
}

func (s *Svc) Total(ctx context.Context, criteria TotalCriteria) (int64, error) {
	return s.r.Total(ctx, criteria)
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

type UnknownSubscriptionError struct {
	ID uuid.UUID
}

func (e UnknownSubscriptionError) Error() string {
	return fmt.Sprintf("subscription with id %v was not found", e.ID)
}
