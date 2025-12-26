package service

import (
	"context"
	"net/http"
	"testing"

	"productservice/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRateRepository struct {
	mock.Mock
}

func (m *MockRateRepository) RatingProduct(ctx context.Context, rateProduct *dto.RateProduct) (*dto.RateResponse, *dto.ServiceResponse) {
	args := m.Called(ctx, rateProduct)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.RateResponse), args.Get(1).(*dto.ServiceResponse)
}

func (m *MockRateRepository) UpdateRatingProduct(ctx context.Context, ratingProduct *dto.UpdateRatingProduct) (*dto.RateResponse, *dto.ServiceResponse) {
	args := m.Called(ctx, ratingProduct)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.RateResponse), args.Get(1).(*dto.ServiceResponse)
}

func (m *MockRateRepository) DeleteRatingProduct(ctx context.Context, ratingId uuid.UUID) *dto.ServiceResponse {
	args := m.Called(ctx, ratingId)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*dto.ServiceResponse)
}

func (m *MockRateRepository) GetRateListOfProduct(ctx context.Context, filterRequest *dto.RateListOfProduct) (*dto.PageResponse, *dto.ServiceResponse) {
	args := m.Called(ctx, filterRequest)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.PageResponse), args.Get(1).(*dto.ServiceResponse)
}

func (m *MockRateRepository) GetRateStatisticOfProduct(ctx context.Context, productId uuid.UUID) (*dto.RatingSummaryResponse, *dto.ServiceResponse) {
	args := m.Called(ctx, productId)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.RatingSummaryResponse), args.Get(1).(*dto.ServiceResponse)
}

func (m *MockRateRepository) GetMyRatings(ctx context.Context, userId uuid.UUID, request *dto.MyRatingsRequest) (*dto.MyRatingsResponse, *dto.ServiceResponse) {
	args := m.Called(ctx, userId, request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.MyRatingsResponse), args.Get(1).(*dto.ServiceResponse)
}

// === TESTS ===

func TestRatingProduct(t *testing.T) {
	productId := uuid.New()
	userId := uuid.New()

	tests := []struct {
		name           string
		rateProduct    *dto.RateProduct
		mockReturn     *dto.RateResponse
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    error
		expectedStatus int
	}{
		{
			name: "Success - Valid Rating",
			rateProduct: &dto.RateProduct{
				ProductId: productId,
				UserId:    userId,
				Star:      4,
				Comment:   "Great product!",
			},
			mockReturn: &dto.RateResponse{
				Id:   uuid.New(),
				Star: 4,
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
		{
			name: "Error - Nil Product ID",
			rateProduct: &dto.RateProduct{
				ProductId: uuid.Nil,
				UserId:    userId,
				Star:      4,
				Comment:   "Great product!",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrNotFound,
			expectedStatus: 0,
		},
		{
			name: "Error - Nil User ID",
			rateProduct: &dto.RateProduct{
				ProductId: productId,
				UserId:    uuid.Nil,
				Star:      4,
				Comment:   "Great product!",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrNotFound,
			expectedStatus: 0,
		},
		{
			name: "Error - Star Below Range",
			rateProduct: &dto.RateProduct{
				ProductId: productId,
				UserId:    userId,
				Star:      0,
				Comment:   "Bad product!",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: 0,
		},
		{
			name: "Error - Star Above Range",
			rateProduct: &dto.RateProduct{
				ProductId: productId,
				UserId:    userId,
				Star:      6,
				Comment:   "Great product!",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: 0,
		},
		{
			name: "Success - All Valid Stars",
			rateProduct: &dto.RateProduct{
				ProductId: productId,
				UserId:    userId,
				Star:      5,
				Comment:   "Perfect!",
			},
			mockReturn:     &dto.RateResponse{Id: uuid.New(), Star: 5},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			service := NewRateService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("RatingProduct", mock.Anything, tt.rateProduct).
					Return(tt.mockReturn, tt.mockResponse)
			}

			rateResp, serviceResp := service.RatingProduct(context.Background(), tt.rateProduct)

			if tt.expectedErr != nil {
				assert.Nil(t, rateResp)
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
			} else {
				assert.NotNil(t, rateResp)
				assert.Nil(t, serviceResp)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "RatingProduct")
			}
		})
	}
}

func TestUpdateRatingProduct(t *testing.T) {
	ratingId := uuid.New()

	tests := []struct {
		name           string
		updateRating   *dto.UpdateRatingProduct
		mockReturn     *dto.RateResponse
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    error
		expectedStatus int
	}{
		{
			name: "Success - Valid Update",
			updateRating: &dto.UpdateRatingProduct{
				RateId:  ratingId,
				Star:    3,
				Comment: "Updated comment",
			},
			mockReturn:     &dto.RateResponse{Id: ratingId, Star: 3},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
		{
			name: "Error - Nil Rating ID",
			updateRating: &dto.UpdateRatingProduct{
				RateId:  uuid.Nil,
				Star:    3,
				Comment: "Updated comment",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Error - Star Below Range",
			updateRating: &dto.UpdateRatingProduct{
				RateId:  ratingId,
				Star:    0,
				Comment: "Updated comment",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Star Above Range",
			updateRating: &dto.UpdateRatingProduct{
				RateId:  ratingId,
				Star:    6,
				Comment: "Updated comment",
			},
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Success - Minimum Valid Star",
			updateRating: &dto.UpdateRatingProduct{
				RateId:  ratingId,
				Star:    1,
				Comment: "Bad",
			},
			mockReturn:     &dto.RateResponse{Id: ratingId, Star: 1},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
		{
			name: "Success - Maximum Valid Star",
			updateRating: &dto.UpdateRatingProduct{
				RateId:  ratingId,
				Star:    5,
				Comment: "Perfect",
			},
			mockReturn:     &dto.RateResponse{Id: ratingId, Star: 5},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			service := NewRateService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("UpdateRatingProduct", mock.Anything, tt.updateRating).
					Return(tt.mockReturn, tt.mockResponse)
			}

			rateResp, serviceResp := service.UpdateRatingProduct(context.Background(), tt.updateRating)

			if tt.expectedErr != nil {
				assert.Nil(t, rateResp)
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
				assert.Equal(t, http.StatusBadRequest, serviceResp.Status)
			} else {
				assert.NotNil(t, rateResp)
				assert.Nil(t, serviceResp)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "UpdateRatingProduct")
			}
		})
	}
}

func TestDeleteRatingProduct(t *testing.T) {
	ratingId := uuid.New()

	tests := []struct {
		name           string
		ratingId       uuid.UUID
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    error
		expectedStatus int
	}{
		{
			name:           "Success - Valid Delete",
			ratingId:       ratingId,
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
		{
			name:           "Error - Nil Rating ID",
			ratingId:       uuid.Nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			service := NewRateService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("DeleteRatingProduct", mock.Anything, tt.ratingId).
					Return(tt.mockResponse)
			}

			serviceResp := service.DeleteRatingProduct(context.Background(), tt.ratingId)

			if tt.expectedErr != nil {
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
				assert.Equal(t, tt.expectedStatus, serviceResp.Status)
			} else {
				assert.Nil(t, serviceResp)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "DeleteRatingProduct")
			}
		})
	}
}

func TestGetRateStatisticOfProduct(t *testing.T) {
	productId := uuid.New()

	tests := []struct {
		name           string
		productId      uuid.UUID
		mockReturn     *dto.RatingSummaryResponse
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    error
		expectedStatus int
	}{
		{
			name:      "Success - Valid Product ID",
			productId: productId,
			mockReturn: &dto.RatingSummaryResponse{
				ProductID: productId.String(),
				Summary: &dto.RatingSummary{
					AvgRating:    4.5,
					TotalRatings: 10,
				},
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
		{
			name:           "Error - Nil Product ID",
			productId:      uuid.Nil,
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			service := NewRateService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("GetRateStatisticOfProduct", mock.Anything, tt.productId).
					Return(tt.mockReturn, tt.mockResponse)
			}

			summaryResp, serviceResp := service.GetRateStatisticOfProduct(context.Background(), tt.productId)

			if tt.expectedErr != nil {
				assert.Nil(t, summaryResp)
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
				assert.Equal(t, tt.expectedStatus, serviceResp.Status)
			} else {
				assert.NotNil(t, summaryResp)
				assert.Nil(t, serviceResp)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "GetRateStatisticOfProduct")
			}
		})
	}
}

func TestGetMyRatings(t *testing.T) {
	userId := uuid.New()

	tests := []struct {
		name           string
		userId         uuid.UUID
		request        *dto.MyRatingsRequest
		mockReturn     *dto.MyRatingsResponse
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    error
		expectedStatus int
		expectedPage   int
		expectedLimit  int
	}{
		{
			name:    "Success - Valid Request",
			userId:  userId,
			request: &dto.MyRatingsRequest{Page: 1, Limit: 10, SortBy: "newest"},
			mockReturn: &dto.MyRatingsResponse{
				Total: 5,
				Data:  []*dto.RateResponse{},
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
			expectedPage:   1,
			expectedLimit:  10,
		},
		{
			name:           "Error - Nil Request",
			userId:         userId,
			request:        nil,
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Normalize Page - Zero",
			userId:  userId,
			request: &dto.MyRatingsRequest{Page: 0, Limit: 10, SortBy: "newest"},
			mockReturn: &dto.MyRatingsResponse{
				Total: 0,
				Data:  []*dto.RateResponse{},
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
			expectedPage:   1,
			expectedLimit:  10,
		},
		{
			name:    "Normalize Limit - Too Low",
			userId:  userId,
			request: &dto.MyRatingsRequest{Page: 1, Limit: 3, SortBy: "newest"},
			mockReturn: &dto.MyRatingsResponse{
				Total: 0,
				Data:  []*dto.RateResponse{},
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
			expectedPage:   1,
			expectedLimit:  5,
		},
		{
			name:    "Normalize Limit - Too High",
			userId:  userId,
			request: &dto.MyRatingsRequest{Page: 1, Limit: 25, SortBy: "newest"},
			mockReturn: &dto.MyRatingsResponse{
				Total: 0,
				Data:  []*dto.RateResponse{},
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
			expectedPage:   1,
			expectedLimit:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRateRepository)
			service := NewRateService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("GetMyRatings", mock.Anything, tt.userId, mock.MatchedBy(func(req *dto.MyRatingsRequest) bool {
					return req.Page == tt.expectedPage && req.Limit == tt.expectedLimit
				})).Return(tt.mockReturn, tt.mockResponse)
			}

			myRatingsResp, serviceResp := service.GetMyRatings(context.Background(), tt.userId, tt.request)

			if tt.expectedErr != nil {
				assert.Nil(t, myRatingsResp)
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
				assert.Equal(t, tt.expectedStatus, serviceResp.Status)
			} else {
				assert.NotNil(t, myRatingsResp)
				assert.Nil(t, serviceResp)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "GetMyRatings")
			}
		})
	}
}
