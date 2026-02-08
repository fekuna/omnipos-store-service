package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/fekuna/omnipos-store-service/internal/model"
	"github.com/jmoiron/sqlx"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateStore(ctx context.Context, store *model.Store) error {
	query := `
		INSERT INTO stores (merchant_id, name, address, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	var id string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query,
		store.MerchantID,
		store.Name,
		store.Address,
		store.Phone,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return err
	}

	store.ID = id
	store.CreatedAt = createdAt
	store.UpdatedAt = updatedAt
	return nil
}

func (r *postgresRepository) GetStore(ctx context.Context, id string) (*model.Store, error) {
	var m model.Store
	// Note: sqlx maps directly to model struct tags, but model tags might need sql.NullString handling
	// if using "address" string in domain model but DB has NULL.
	// For simplicity, let's assume domain model handles empty strings or DB columns are NOT NULL.
	// Assuming columns are NOT NULL or we accept empty string for NULL.
	// Using COALESCE is safer if columns are nullable.
	query := `
        SELECT id, merchant_id, name, COALESCE(address, '') as address, COALESCE(phone, '') as phone, created_at, updated_at 
        FROM stores WHERE id = $1
    `
	err := r.db.GetContext(ctx, &m, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) ListStores(ctx context.Context, merchantID string, page, pageSize int) ([]model.Store, int, error) {
	offset := (page - 1) * pageSize
	var stores []model.Store

	countQuery := `SELECT count(*) FROM stores WHERE merchant_id = $1`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, merchantID); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.Store{}, 0, nil
	}

	query := `
		SELECT id, merchant_id, name, COALESCE(address, '') as address, COALESCE(phone, '') as phone, created_at, updated_at 
		FROM stores 
        WHERE merchant_id = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &stores, query, merchantID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

func (r *postgresRepository) UpdateStore(ctx context.Context, store *model.Store) error {
	query := `
		UPDATE stores 
		SET name = $1, address = $2, phone = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		store.Name,
		store.Address,
		store.Phone,
		store.ID,
	).Scan(&updatedAt)

	if err != nil {
		return err
	}

	store.UpdatedAt = updatedAt
	return nil
}

func (r *postgresRepository) DeleteStore(ctx context.Context, id string) error {
	query := `DELETE FROM stores WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
