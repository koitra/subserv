package subscriptions

import (
	"context"
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

	huma.Post(g, "", h.create, statuses)
	huma.Delete(g, "/{subscriptionId}", h.delete, statuses)
	huma.Get(g, "", h.index, statuses)
}

type (
	APISubscription struct {
		ID          uuid.UUID `json:"subscriptionId"`
		ServiceName string    `json:"serviceName"`
		Price       int32     `json:"price"`
		UserID      uuid.UUID `json:"userId"`
		StartDate   Date      `json:"startDate"`
		EndDate     *Date     `json:"endDate,omitempty"`
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
			ServiceName string    `json:"serviceName" minLength:"1"`
			Price       int32     `json:"price" minimum:"0"`
			UserID      uuid.UUID `json:"userId" format:"uuid"`
			StartDate   Date      `json:"startDate"`
			EndDate     *Date     `json:"endDate,omitempty"`
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
		return nil, err
	}

	var out SubscriptionsDeleteOut
	return &out, nil
}

type (
	SubscriptionsDeleteIn struct {
		SubscriptionID uuid.UUID `path:"subscriptionId" format:"uuid"`
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
		UserIDs humaext.UUIDs `query:"userIds"`

		ServerNames []string `query:"serverNames"`
		cursor.WebCursorIn
	}

	SubscriptionsIndexOut struct {
		Body struct {
			Subscriptions []APISubscription `json:"subscriptions"`
			Cursor        *time.Time        `json:"cursor,omitempty"`
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
