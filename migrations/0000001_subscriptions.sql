-- +goose Up
-- +goose StatementBegin
CREATE TABLE subscriptions (
    id uuid PRIMARY KEY,
    service_name varchar NOT NULL,
    price integer NOT NULL,
    user_id uuid NOT NULL,
    start_date timestamptz NOT NULL,
    end_date timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX ON subscriptions (created_at);

CREATE INDEX ON subscriptions (user_id);

CREATE INDEX ON subscriptions (service_name);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE subscriptions;

-- +goose StatementEnd
