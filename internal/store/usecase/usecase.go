package usecase

import (
	"context"

	"github.com/fekuna/omnipos-store-service/internal/auth"
	"github.com/fekuna/omnipos-store-service/internal/model"
	"github.com/fekuna/omnipos-store-service/internal/store/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Usecase interface {
	CreateStore(ctx context.Context, store *model.Store) error
	GetStore(ctx context.Context, id string) (*model.Store, error)
	ListStores(ctx context.Context, page, pageSize int) ([]model.Store, int, error)
	UpdateStore(ctx context.Context, store *model.Store) error
	DeleteStore(ctx context.Context, id string) error
}

type storeUsecase struct {
	repo repository.Repository
}

func NewStoreUsecase(repo repository.Repository) Usecase {
	return &storeUsecase{repo: repo}
}

func (u *storeUsecase) CreateStore(ctx context.Context, store *model.Store) error {
	merchantID := auth.GetMerchantID(ctx)
	if merchantID == "" {
		return status.Error(codes.Unauthenticated, "merchant ID not found in context")
	}
	store.MerchantID = merchantID

	return u.repo.CreateStore(ctx, store)
}

func (u *storeUsecase) GetStore(ctx context.Context, id string) (*model.Store, error) {
	// Optional: Check if store belongs to merchant
	store, err := u.repo.GetStore(ctx, id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, status.Error(codes.NotFound, "store not found")
	}

	// Tenant check
	// Ideally we filter by merchant_id in Repo, but here is also a safeguard
	merchantID := auth.GetMerchantID(ctx)
	if merchantID != "" && store.MerchantID != merchantID {
		return nil, status.Error(codes.PermissionDenied, "store does not belong to merchant")
	}

	return store, nil
}

func (u *storeUsecase) ListStores(ctx context.Context, page, pageSize int) ([]model.Store, int, error) {
	merchantID := auth.GetMerchantID(ctx)
	if merchantID == "" {
		return nil, 0, status.Error(codes.Unauthenticated, "merchant ID not found in context")
	}
	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return u.repo.ListStores(ctx, merchantID, page, pageSize)
}

func (u *storeUsecase) UpdateStore(ctx context.Context, update *model.Store) error {
	// Verify ownership or existence via GetStore first
	current, err := u.GetStore(ctx, update.ID)
	if err != nil {
		return err
	}

	// Update allowed fields
	current.Name = update.Name
	current.Address = update.Address
	current.Phone = update.Phone

	return u.repo.UpdateStore(ctx, current)
}

func (u *storeUsecase) DeleteStore(ctx context.Context, id string) error {
	// Verify existence/ownership
	_, err := u.GetStore(ctx, id)
	if err != nil {
		return err // NotFound or PermissionDenied
	}
	return u.repo.DeleteStore(ctx, id)
}
