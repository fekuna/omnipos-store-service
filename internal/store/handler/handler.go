package handler

import (
	"context"

	"github.com/fekuna/omnipos-pkg/logger"
	storev1 "github.com/fekuna/omnipos-proto/proto/store/v1"
	"github.com/fekuna/omnipos-store-service/internal/model"
	"github.com/fekuna/omnipos-store-service/internal/store/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StoreHandler struct {
	storev1.UnimplementedStoreServiceServer
	uc     usecase.Usecase
	logger logger.ZapLogger
}

func NewStoreHandler(uc usecase.Usecase, logger logger.ZapLogger) *StoreHandler {
	return &StoreHandler{
		uc:     uc,
		logger: logger,
	}
}

func (h *StoreHandler) mapToProto(m *model.Store) *storev1.Store {
	if m == nil {
		return nil
	}
	return &storev1.Store{
		Id:         m.ID,
		MerchantId: m.MerchantID,
		Name:       m.Name,
		Address:    m.Address,
		Phone:      m.Phone,
		CreatedAt:  timestamppb.New(m.CreatedAt),
		UpdatedAt:  timestamppb.New(m.UpdatedAt),
	}
}

func (h *StoreHandler) CreateStore(ctx context.Context, req *storev1.CreateStoreRequest) (*storev1.CreateStoreResponse, error) {
	// Map Request -> Model
	input := &model.Store{
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
	}

	if err := h.uc.CreateStore(ctx, input); err != nil {
		h.logger.Error("failed to create store", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create store")
	}

	return &storev1.CreateStoreResponse{Store: h.mapToProto(input)}, nil
}

func (h *StoreHandler) GetStore(ctx context.Context, req *storev1.GetStoreRequest) (*storev1.GetStoreResponse, error) {
	store, err := h.uc.GetStore(ctx, req.Id)
	if err != nil {
		// Usecase should return appropriate error status, but if it returns pure error, we might default to Internal
		// Ideally Usecase error is checked, but relying on Usecase to return status errors or repo to return known errors is fine if simple.
		// Detailed error mapping logic omitted for brevity, assuming Usecase handles logic errors.
		h.logger.Error("failed to get store", zap.Error(err))
		return nil, err
	}

	return &storev1.GetStoreResponse{Store: h.mapToProto(store)}, nil
}

func (h *StoreHandler) ListStores(ctx context.Context, req *storev1.ListStoresRequest) (*storev1.ListStoresResponse, error) {
	stores, total, err := h.uc.ListStores(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		h.logger.Error("failed to list stores", zap.Error(err))
		return nil, err
	}

	var protoStores []*storev1.Store
	for _, s := range stores {
		// Range var 's' is a copy, safe to take address if mapped immediately or use helper
		protoStores = append(protoStores, h.mapToProto(&s))
	}

	return &storev1.ListStoresResponse{
		Stores:   protoStores,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (h *StoreHandler) UpdateStore(ctx context.Context, req *storev1.UpdateStoreRequest) (*storev1.UpdateStoreResponse, error) {
	input := &model.Store{
		ID:      req.Id,
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
	}

	if err := h.uc.UpdateStore(ctx, input); err != nil {
		h.logger.Error("failed to update store", zap.Error(err))
		return nil, err
	}

	// Refetch or just return input (input won't have UpdatedAt unless usecase updates it in place pointer)
	// Usecase `UpdateStore` updates the pointer `store.UpdatedAt`.
	return &storev1.UpdateStoreResponse{Store: h.mapToProto(input)}, nil
}

func (h *StoreHandler) DeleteStore(ctx context.Context, req *storev1.DeleteStoreRequest) (*storev1.DeleteStoreResponse, error) {
	err := h.uc.DeleteStore(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to delete store", zap.Error(err))
		return nil, err
	}
	return &storev1.DeleteStoreResponse{Success: true}, nil
}
