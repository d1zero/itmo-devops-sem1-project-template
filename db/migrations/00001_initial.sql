-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS prices
(
    id          BIGSERIAL PRIMARY KEY,
    create_date TIMESTAMP,
    name        TEXT,
    category    TEXT,
    price       DECIMAL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
