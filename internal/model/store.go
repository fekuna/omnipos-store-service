package model

import "time"

type Store struct {
	ID         string    `db:"id"`
	MerchantID string    `db:"merchant_id"`
	Name       string    `db:"name"`
	Address    string    `db:"address"`
	Phone      string    `db:"phone"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
