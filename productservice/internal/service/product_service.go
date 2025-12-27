package service

import (
	"context"
	"productservice/internal/dto"
	"productservice/internal/model"
	"productservice/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProductService interface {
	GetList(ctx context.Context, request *dto.PageProdRequest) (*dto.PageResponse, *dto.ServiceResponse)
	GetProductById(ctx context.Context, productId uuid.UUID) (*model.Product, *dto.ServiceResponse)
	GetProductSimilar(ctx context.Context, productId uuid.UUID, pageRequest *dto.PageSimilarAndRelatedRequest, relationType *dto.RelationType) (*dto.PageResponse, *dto.ServiceResponse)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (p *productService) GetProductSimilar(ctx context.Context, productId uuid.UUID, pageRequest *dto.PageSimilarAndRelatedRequest, relationType *dto.RelationType) (*dto.PageResponse, *dto.ServiceResponse) {
	if productId == uuid.Nil {
		response := dto.ServiceResponse{
			Err: ErrNotFound,
		}
		return nil, &response
	}

	if pageRequest.Page <= 0 {
		pageRequest.Page = 1
	}

	if pageRequest.Limit < 5 || pageRequest.Limit > 20 {
		pageRequest.Limit = 5
	}

	ct, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return p.repo.GetProductSimilar(ct, productId, pageRequest, relationType)
}

func (p *productService) GetProductById(ctx context.Context, productId uuid.UUID) (*model.Product, *dto.ServiceResponse) {
	if productId == uuid.Nil {
		response := dto.ServiceResponse{
			Err: ErrNotFound,
		}
		return nil, &response
	}

	return p.repo.GetProductById(ctx, productId)
}

func (p *productService) GetList(ctx context.Context, request *dto.PageProdRequest) (*dto.PageResponse, *dto.ServiceResponse) {
	if request.Page == 0 {
		request.Page = 1
	}

	if request.Limit <= 0 || request.Limit > 100 {
		request.Limit = 20
	}

	validSortFields := map[string]bool{
		"average_rating": true,
		"name":           true,
		"price":          true,
		"created_at":     true,
	}

	if request.SortBy != "" {
		if !validSortFields[request.SortBy] {
			response := dto.ServiceResponse{
				Status: 400,
				Err:    ErrInvalid,
			}
			return nil, &response
		}
	} else {
		request.SortBy = "id"
	}

	if request.SortOrder != "" {
		request.SortOrder = strings.ToUpper(request.SortOrder)
		if request.SortOrder != "ASC" && request.SortOrder != "DESC" {
			response := dto.ServiceResponse{
				Status: 400,
				Err:    ErrInvalid,
			}
			return nil, &response
		}
	} else {
		request.SortOrder = "DESC"
	}

	if request.MinPrice < 0 {
		request.MinPrice = 0
	}
	if request.MaxPrice < 0 {
		request.MaxPrice = 0
	}
	if request.MinPrice > 0 && request.MaxPrice > 0 && request.MinPrice > request.MaxPrice {
		response := dto.ServiceResponse{
			Status: 400,
			Err:    ErrInvalid,
		}
		return nil, &response
	}

	if request.MinRate < 0 || request.MinRate > 5 {
		request.MinRate = 0
	}
	if request.MaxRate < 0 || request.MaxRate > 5 {
		request.MaxRate = 0
	}
	if request.MinRate > 0 && request.MaxRate > 0 && request.MinRate > request.MaxRate {
		response := dto.ServiceResponse{
			Status: 400,
			Err:    ErrInvalid,
		}
		return nil, &response
	}

	return p.repo.GetList(ctx, request)
}
