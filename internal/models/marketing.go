package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Price struct {
	ID           int64           `db:"id"`
	CreationDate time.Time       `db:"create_date"`
	Name         string          `db:"name"`
	Category     string          `db:"category"`
	Price        decimal.Decimal `db:"price"`
}

type AggPriceData struct {
	TotalItems      int             `json:"total_items" db:"total_items"`
	TotalCategories int             `json:"total_categories" db:"total_categories"`
	TotalPrice      decimal.Decimal `json:"total_price" db:"total_price"`
}
