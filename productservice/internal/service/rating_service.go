package service

import (
	"context"
	"net/http"
	"productservice/internal/dto"
	"productservice/internal/repository"
	"time"

	"github.com/google/uuid"
)

type RateService interface {
	RatingProduct(ctx context.Context, rateProduct *dto.RateProduct) (*dto.RateResponse, *dto.ServiceResponse)
	UpdateRatingProduct(ctx context.Context, ratingProduct *dto.UpdateRatingProduct) (*dto.RateResponse, *dto.ServiceResponse)
	DeleteRatingProduct(ctx context.Context, ratingId uuid.UUID) *dto.ServiceResponse
	GetRateListOfProduct(ctx context.Context, filterRequest *dto.RateListOfProduct) (*dto.PageResponse, *dto.ServiceResponse)
	GetRateStatisticOfProduct(ctx context.Context, productId uuid.UUID) (*dto.RatingSummaryResponse, *dto.ServiceResponse)
	GetMyRatings(ctx context.Context, userId uuid.UUID, request *dto.MyRatingsRequest) (*dto.MyRatingsResponse, *dto.ServiceResponse)
}

type rateService struct {
	rateRepo repository.RateRepository
}

func NewRateService(rateRepo repository.RateRepository) RateService {
	return &rateService{rateRepo: rateRepo}
}

func (s *rateService) GetMyRatings(ctx context.Context, userId uuid.UUID, request *dto.MyRatingsRequest) (*dto.MyRatingsResponse, *dto.ServiceResponse) {
	if request == nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    ErrInvalid,
		}
	}
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Limit < 5 || request.Limit > 20 {
		request.Limit = 5
	}

	return s.rateRepo.GetMyRatings(ctx, userId, request)
}

func (s *rateService) GetRateStatisticOfProduct(ctx context.Context, productId uuid.UUID) (*dto.RatingSummaryResponse, *dto.ServiceResponse) {
	if productId == uuid.Nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    ErrInvalid,
		}
	}
	return s.rateRepo.GetRateStatisticOfProduct(ctx, productId)
}

func (s *rateService) GetRateListOfProduct(ctx context.Context, filterRequest *dto.RateListOfProduct) (*dto.PageResponse, *dto.ServiceResponse) {
	return s.rateRepo.GetRateListOfProduct(ctx, filterRequest)
}

func (s *rateService) DeleteRatingProduct(ctx context.Context, ratingId uuid.UUID) *dto.ServiceResponse {
	if ratingId == uuid.Nil {
		return &dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    ErrInvalid,
		}
	}

	ct, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return s.rateRepo.DeleteRatingProduct(ct, ratingId)
}

func (s *rateService) UpdateRatingProduct(ctx context.Context, ratingProduct *dto.UpdateRatingProduct) (*dto.RateResponse, *dto.ServiceResponse) {
	if ratingProduct.RateId == uuid.Nil {
		response := dto.ServiceResponse{
			Status: http.StatusNotFound,
			Err:    ErrNotFound,
		}
		return nil, &response
	}

	if ratingProduct.Star < 1 || ratingProduct.Star > 5 {
		response := dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    ErrInvalid,
		}
		return nil, &response
	}

	return s.rateRepo.UpdateRatingProduct(ctx, ratingProduct)
}

func (s *rateService) RatingProduct(ctx context.Context, rateProduct *dto.RateProduct) (*dto.RateResponse, *dto.ServiceResponse) {
	if rateProduct.ProductId == uuid.Nil || rateProduct.UserId == uuid.Nil {
		response := dto.ServiceResponse{
			Status: http.StatusNotFound,
			Err:    ErrNotFound,
		}
		return nil, &response
	}

	if rateProduct.Star < 1 || rateProduct.Star > 5 {
		response := dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    ErrInvalid,
		}
		return nil, &response
	}

	return s.rateRepo.RatingProduct(ctx, rateProduct)
}
