package subscriptions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/koitra/subserv/internal/cursor"
	"github.com/koitra/subserv/internal/humaext"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc}
}

func (h *Handler) Register(hapi huma.API) {
	g := huma.NewGroup(hapi, "/subscriptions")

	statuses := func(o *huma.Operation) {
		o.Errors = append(
			o.Errors,
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		)
	}

	huma.Get(g, "", h.index, statuses)
	huma.Post(g, "", h.create, statuses)
	huma.Put(g, "/{subscription_id}", h.update, statuses)
	huma.Delete(g, "/{subscription_id}", h.delete, statuses)
	huma.Get(g, "/total", h.total, statuses)
}

type (
	APISubscription struct {
		ID          uuid.UUID `json:"subscription_id"`
		ServiceName string    `json:"service_name"`
		Price       int32     `json:"price"`
		UserID      uuid.UUID `json:"user_id"`
		StartDate   Date      `json:"start_date"`
		EndDate     *Date     `json:"end_date,omitempty"`
	}
)

func (h *Handler) create(
	ctx context.Context,
	in *SubscriptionsCreateIn,
) (*SubscriptionsCreateOut, error) {
	sub, err := h.svc.Add(ctx, NewSubscription{
		ServiceName: in.Body.ServiceName,
		Price:       in.Body.Price,
		UserID:      in.Body.UserID,
		StartDate:   time.Time(in.Body.StartDate),
		EndDate:     (*time.Time)(in.Body.EndDate),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Create failed", slog.String("error", err.Error()))
		err, ok := humaext.ValidateToHuma(err)
		if ok {
			return nil, err
		}
		return nil, err
	}

	var out SubscriptionsCreateOut
	out.Body.Subscription = apiSubscriptionFromSubscription(&sub)

	return &out, nil
}

type (
	SubscriptionsCreateIn struct {
		Body struct {
			ServiceName string    `json:"service_name" minLength:"1"`
			Price       int32     `json:"price" minimum:"0"`
			UserID      uuid.UUID `json:"user_id" format:"uuid"`
			StartDate   Date      `json:"start_date"`
			EndDate     *Date     `json:"end_date,omitempty"`
		}
	}
	SubscriptionsCreateOut struct {
		Body struct {
			Subscription APISubscription `json:"subscription"`
		}
	}
)

func (h *Handler) delete(
	ctx context.Context,
	in *SubscriptionsDeleteIn,
) (*SubscriptionsDeleteOut, error) {
	err := h.svc.Remove(ctx, in.SubscriptionID)
	if err != nil {
		slog.ErrorContext(ctx, "Delete failed", slog.String("error", err.Error()))
		return nil, err
	}

	var out SubscriptionsDeleteOut
	return &out, nil
}

type (
	SubscriptionsDeleteIn struct {
		SubscriptionID uuid.UUID `path:"subscription_id" format:"uuid"`
	}

	SubscriptionsDeleteOut struct{}
)

func (h *Handler) index(
	ctx context.Context,
	in *SubscriptionsIndexIn,
) (*SubscriptionsIndexOut, error) {
	subs, err := h.svc.Find(ctx, Criteria{
		IDs:          in.IDs.IDs,
		ServiceNames: in.ServerNames,
		UserIDs:      in.UserIDs.IDs,
		Cursor:       in.WebCursorIn.ToCursor().WithMaxLimit(100),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Index failed", slog.String("error", err.Error()))
		return nil, err
	}

	var out SubscriptionsIndexOut
	out.Body.Subscriptions = make([]APISubscription, len(subs.Subscriptions))
	for i, sub := range subs.Subscriptions {
		out.Body.Subscriptions[i] = apiSubscriptionFromSubscription(&sub)
	}
	out.Body.Cursor = subs.Cursor

	return &out, nil
}

type (
	SubscriptionsIndexIn struct {
		IDs     humaext.UUIDs `query:"ids"`
		UserIDs humaext.UUIDs `query:"user_ids"`

		ServerNames []string `query:"server_names"`
		cursor.WebCursorIn
	}

	SubscriptionsIndexOut struct {
		Body struct {
			Subscriptions []APISubscription `json:"subscriptions"`
			Cursor        *time.Time        `json:"cursor,omitempty"`
		}
	}
)

func (h *Handler) update(
	ctx context.Context,
	in *SubscriptionsUpdateIn,
) (*SubscriptionsUpdateOut, error) {
	sub, err := h.svc.Update(ctx, UpdateSubscription{
		ID:          in.SubscriptionID,
		ServiceName: in.Body.ServiceName,
		Price:       in.Body.Price,
		UserID:      in.Body.UserID,
		StartDate:   time.Time(in.Body.StartDate),
		EndDate:     (*time.Time)(in.Body.EndDate),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Update failed", slog.String("error", err.Error()))
		err, ok := humaext.ValidateToHuma(err)
		if ok {
			return nil, err
		}
		if e, ok := errors.AsType[UnknownSubscriptionError](err); ok {
			return nil, huma.Error404NotFound("Subscription not found", e)
		}
		return nil, err
	}

	var out SubscriptionsUpdateOut
	out.Body.Subscription = apiSubscriptionFromSubscription(&sub)

	return &out, nil
}

type (
	SubscriptionsUpdateIn struct {
		SubscriptionID uuid.UUID `path:"subscription_id"`
		Body           struct {
			ServiceName string    `json:"service_name"`
			Price       int32     `json:"price"`
			UserID      uuid.UUID `json:"user_id"`
			StartDate   Date      `json:"start_date"`
			EndDate     *Date     `json:"end_date,omitempty"`
		}
	}
	SubscriptionsUpdateOut struct {
		Body struct {
			Subscription APISubscription `json:"subscription"`
		}
	}
)

func (h *Handler) total(
	ctx context.Context,
	in *SubscriptionsTotalIn,
) (*SubscriptionsTotalOut, error) {
	total, err := h.svc.Total(ctx, TotalCriteria{
		UserID:      in.UserID.Ptr(),
		ServiceName: in.ServiceName.Ptr(),
		PeriodStart: time.Time(in.PeriodStart),
		PeriodEnd:   time.Time(in.PerodEnd),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Total failed", slog.String("error", err.Error()))
		if err, ok := humaext.ValidateToHuma(err); ok {
			return nil, err
		}
		return nil, fmt.Errorf("get total: %w", err)
	}

	var out SubscriptionsTotalOut
	out.Body.Total = total
	return &out, nil
}

type (
	SubscriptionsTotalIn struct {
		UserID      humaext.Opt[uuid.UUID] `query:"user_id"      format:"uuid"`
		ServiceName humaext.Opt[string]    `query:"service_name"               minLength:"1"`
		PeriodStart Date                   `query:"period_start"                             required:"true"`
		PerodEnd    Date                   `query:"period_end"                               required:"true"`
	}
	SubscriptionsTotalOut struct {
		Body struct {
			Total int64 `json:"total"`
		}
	}
)

func apiSubscriptionFromSubscription(sub *Subscription) APISubscription {
	return APISubscription{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   Date(sub.StartDate),
		EndDate:     (*Date)(sub.EndDate),
	}
}
