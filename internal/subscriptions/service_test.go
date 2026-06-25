package subscriptions

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	var svc Service = NewService(nil, validate)
	ctx := t.Context()

	t.Run("validation", func(t *testing.T) {
		_, err := svc.Add(ctx, NewSubscription{})
		require.ErrorAs(t, err, new(validator.ValidationErrors))
		errs, _ := errors.AsType[validator.ValidationErrors](err)
		require.Len(t, errs, 3)

		fields := []string{}
		for _, e := range errs {
			fields = append(fields, e.Field())
		}

		require.Contains(t, fields, "ServiceName")
		require.Contains(t, fields, "UserID")
		require.Contains(t, fields, "StartDate")
	})
}
