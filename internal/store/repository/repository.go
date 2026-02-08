package repository

import (
	"context"

	"github.com/fekuna/omnipos-store-service/internal/model"
)

type Repository interface {
	CreateStore(ctx context.Context, store *model.Store) error
	GetStore(ctx context.Context, id string) (*model.Store, error)
	ListStores(ctx context.Context, merchantID string, page, pageSize int) ([]model.Store, int, error)
	UpdateStore(ctx context.Context, store *model.Store) error
	DeleteStore(ctx context.Context, id string) error
}
