-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS prices
(
    id          INTEGER,
    create_date DATE,
    name        TEXT,
    category    TEXT,
    price       DECIMAL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
