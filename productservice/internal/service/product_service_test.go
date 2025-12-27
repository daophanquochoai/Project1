package service

import (
	"context"
	"productservice/internal/dto"
	"productservice/internal/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (p *MockRepository) GetProductSimilar(
	ctx context.Context,
	productId uuid.UUID,
	pageRequest *dto.PageSimilarAndRelatedRequest,
	relationType *dto.RelationType,
) (*dto.PageResponse, *dto.ServiceResponse) {
	args := p.Called(ctx, productId, pageRequest, relationType)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.PageResponse), args.Get(1).(*dto.ServiceResponse)
}

func (p *MockRepository) GetProductById(ctx context.Context, productId uuid.UUID) (*model.Product, *dto.ServiceResponse) {
	args := p.Called(ctx, productId)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*model.Product), args.Get(1).(*dto.ServiceResponse)
}

func (p *MockRepository) GetList(ctx context.Context, request *dto.PageProdRequest) (*dto.PageResponse, *dto.ServiceResponse) {
	args := p.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*dto.PageResponse), args.Get(1).(*dto.ServiceResponse)
}

// === TEST ===
func TestGetProductById(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name           string
		productId      uuid.UUID
		mockReturn     *model.Product
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    error
		expectedStatus int
	}{
		{
			name:      "Success - Valid UUID",
			productId: validUUID,
			mockReturn: &model.Product{
				ID:   validUUID,
				Name: "Test Product",
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
			expectedStatus: 0,
		},
		{
			name:           "Error - Nil UUID",
			productId:      uuid.Nil,
			mockReturn:     nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrNotFound,
			expectedStatus: 0,
		},
		{
			name:       "Error - Product Not Found in DB",
			productId:  validUUID,
			mockReturn: nil,
			mockResponse: &dto.ServiceResponse{
				Status: 404,
				Err:    ErrNotFound,
			},
			shouldCallRepo: true,
			expectedErr:    ErrNotFound,
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewProductService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("GetProductById", mock.Anything, tt.productId).
					Return(tt.mockReturn, tt.mockResponse)
			}

			product, response := service.GetProductById(context.Background(), tt.productId)

			if tt.expectedErr != nil {
				assert.Nil(t, product)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectedErr, response.Err)
				assert.Equal(t, tt.expectedStatus, response.Status)
			} else {
				assert.NotNil(t, product)
				assert.Equal(t, tt.productId, product.ID)
				assert.Nil(t, response)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "GetProductById")
			}
		})
	}
}

func TestGetList(t *testing.T) {
	tests := []struct {
		name             string
		request          *dto.PageProdRequest
		expectedRequest  *dto.PageProdRequest
		mockPageResponse *dto.PageResponse
		mockServiceResp  *dto.ServiceResponse
		shouldCallRepo   bool
		expectedErr      error
		expectedStatus   int
	}{
		{
			name: "Success - Valid Request",
			request: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "price",
				SortOrder: "ASC",
			},
			expectedRequest: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "price",
				SortOrder: "ASC",
			},
			mockPageResponse: &dto.PageResponse{Total: 50, Filter: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "price",
				SortOrder: "ASC",
			}},
			mockServiceResp: nil,
			shouldCallRepo:  true,
			expectedErr:     nil,
			expectedStatus:  0,
		},
		{
			name: "Normalize Limit - Too Low",
			request: &dto.PageProdRequest{
				Page:      1,
				Limit:     -5,
				SortBy:    "average_rating",
				SortOrder: "DESC",
			},
			expectedRequest: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "average_rating",
				SortOrder: "DESC",
			},
			mockPageResponse: &dto.PageResponse{Total: 50, Filter: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "average_rating",
				SortOrder: "DESC",
			}},
			mockServiceResp: nil,
			shouldCallRepo:  true,
			expectedErr:     nil,
			expectedStatus:  0,
		},
		{
			name: "Normalize Limit - Too High",
			request: &dto.PageProdRequest{
				Page:  1,
				Limit: 150,
			},
			expectedRequest: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "id",
				SortOrder: "DESC",
			},
			mockPageResponse: &dto.PageResponse{Total: 50, Filter: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "id",
				SortOrder: "DESC",
			}},
			mockServiceResp: nil,
			shouldCallRepo:  true,
			expectedErr:     nil,
			expectedStatus:  0,
		},
		{
			name: "Error - Invalid SortBy",
			request: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "invalid_field",
				SortOrder: "DESC",
			},
			expectedRequest:  nil,
			mockPageResponse: nil,
			mockServiceResp:  nil,
			shouldCallRepo:   false,
			expectedErr:      ErrInvalid,
			expectedStatus:   400,
		},
		{
			name: "Error - Invalid SortOrder",
			request: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "price",
				SortOrder: "INVALID",
			},
			expectedRequest:  nil,
			mockPageResponse: nil,
			mockServiceResp:  nil,
			shouldCallRepo:   false,
			expectedErr:      ErrInvalid,
			expectedStatus:   400,
		},
		{
			name: "Normalize SortOrder - Lowercase to Uppercase",
			request: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "average_rating",
				SortOrder: "asc",
			},
			expectedRequest: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "average_rating",
				SortOrder: "ASC",
			},
			mockPageResponse: &dto.PageResponse{},
			mockServiceResp:  &dto.ServiceResponse{Status: 200},
			shouldCallRepo:   true,
			expectedErr:      nil,
			expectedStatus:   200,
		},
		{
			name: "Error - Invalid Price Range",
			request: &dto.PageProdRequest{
				Page:     1,
				Limit:    20,
				MinPrice: 100,
				MaxPrice: 50,
			},
			expectedRequest:  nil,
			mockPageResponse: nil,
			mockServiceResp:  nil,
			shouldCallRepo:   false,
			expectedErr:      ErrInvalid,
			expectedStatus:   400,
		},
		{
			name: "Error - Invalid Rating Range",
			request: &dto.PageProdRequest{
				Page:    1,
				Limit:   20,
				MinRate: 4,
				MaxRate: 2,
			},
			expectedRequest:  nil,
			mockPageResponse: nil,
			mockServiceResp:  nil,
			shouldCallRepo:   false,
			expectedErr:      ErrInvalid,
			expectedStatus:   400,
		},
		{
			name: "Normalize Negative Prices",
			request: &dto.PageProdRequest{
				Page:     1,
				Limit:    20,
				MinPrice: -10,
				MaxPrice: -5,
			},
			expectedRequest: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "id",
				SortOrder: "DESC",
				MinPrice:  0,
				MaxPrice:  0,
			},
			mockPageResponse: &dto.PageResponse{},
			mockServiceResp:  &dto.ServiceResponse{Status: 200},
			shouldCallRepo:   true,
			expectedErr:      nil,
			expectedStatus:   200,
		},
		{
			name: "Normalize Invalid Ratings",
			request: &dto.PageProdRequest{
				Page:    1,
				Limit:   20,
				MinRate: -1,
				MaxRate: 6,
			},
			expectedRequest: &dto.PageProdRequest{
				Page:      1,
				Limit:     20,
				SortBy:    "id",
				SortOrder: "DESC",
				MinRate:   0,
				MaxRate:   0,
			},
			mockPageResponse: &dto.PageResponse{},
			mockServiceResp:  &dto.ServiceResponse{Status: 200},
			shouldCallRepo:   true,
			expectedErr:      nil,
			expectedStatus:   200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewProductService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("GetList", mock.Anything, mock.MatchedBy(func(req *dto.PageProdRequest) bool {
					return req.Page == tt.expectedRequest.Page &&
						req.Limit == tt.expectedRequest.Limit &&
						req.SortBy == tt.expectedRequest.SortBy &&
						req.SortOrder == tt.expectedRequest.SortOrder &&
						req.MinPrice == tt.expectedRequest.MinPrice &&
						req.MaxPrice == tt.expectedRequest.MaxPrice &&
						req.MinRate == tt.expectedRequest.MinRate &&
						req.MaxRate == tt.expectedRequest.MaxRate
				})).Return(tt.mockPageResponse, tt.mockServiceResp)
			}

			pageResp, serviceResp := service.GetList(context.Background(), tt.request)

			if tt.expectedErr != nil {
				assert.Nil(t, pageResp)
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
				assert.Equal(t, tt.expectedStatus, serviceResp.Status)
			} else {
				assert.NotNil(t, pageResp)
				assert.Nil(t, nil)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "GetList")
			}
		})
	}
}

func TestGetProductSimilar(t *testing.T) {
	validUUID := uuid.New()
	relationType := dto.Relation_Similar

	tests := []struct {
		name            string
		productId       uuid.UUID
		pageRequest     *dto.PageSimilarAndRelatedRequest
		expectedPageReq *dto.PageSimilarAndRelatedRequest
		relationType    *dto.RelationType
		mockReturn      *dto.PageResponse
		mockResponse    *dto.ServiceResponse
		shouldCallRepo  bool
		expectedErr     error
	}{
		{
			name:      "Success - Valid Request",
			productId: validUUID,
			pageRequest: &dto.PageSimilarAndRelatedRequest{
				Page:  1,
				Limit: 10,
			},
			expectedPageReq: &dto.PageSimilarAndRelatedRequest{
				Page:  1,
				Limit: 10,
			},
			relationType:   &relationType,
			mockReturn:     &dto.PageResponse{Total: 5, Data: []model.Product{}},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
		},
		{
			name:            "Error - Nil Product ID",
			productId:       uuid.Nil,
			pageRequest:     &dto.PageSimilarAndRelatedRequest{Page: 1, Limit: 10},
			expectedPageReq: nil,
			relationType:    &relationType,
			mockReturn:      nil,
			mockResponse:    nil,
			shouldCallRepo:  false,
			expectedErr:     ErrNotFound,
		},
		{
			name:      "Normalize Page and Limit",
			productId: validUUID,
			pageRequest: &dto.PageSimilarAndRelatedRequest{
				Page:  0,
				Limit: 3,
			},
			expectedPageReq: &dto.PageSimilarAndRelatedRequest{
				Page:  1,
				Limit: 5,
			},
			relationType:   &relationType,
			mockReturn:     &dto.PageResponse{Total: 0, Data: []model.Product{}},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
		},
		{
			name:      "Normalize High Limit",
			productId: validUUID,
			pageRequest: &dto.PageSimilarAndRelatedRequest{
				Page:  2,
				Limit: 25,
			},
			expectedPageReq: &dto.PageSimilarAndRelatedRequest{
				Page:  2,
				Limit: 5,
			},
			relationType:   &relationType,
			mockReturn:     &dto.PageResponse{Total: 0, Data: []model.Product{}},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewProductService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("GetProductSimilar", mock.Anything, tt.productId, mock.MatchedBy(func(req *dto.PageSimilarAndRelatedRequest) bool {
					return req.Page == tt.expectedPageReq.Page && req.Limit == tt.expectedPageReq.Limit
				}), tt.relationType).Return(tt.mockReturn, tt.mockResponse)
			}

			pageResp, serviceResp := service.GetProductSimilar(context.Background(), tt.productId, tt.pageRequest, tt.relationType)

			if tt.expectedErr != nil {
				assert.Nil(t, pageResp)
				assert.NotNil(t, serviceResp)
				assert.Equal(t, tt.expectedErr, serviceResp.Err)
			} else {
				assert.NotNil(t, pageResp)
				assert.Nil(t, serviceResp)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "GetProductSimilar")
			}
		})
	}
}
