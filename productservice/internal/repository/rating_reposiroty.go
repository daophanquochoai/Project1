package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"productservice/internal/dto"
	"productservice/internal/grpc/client"
	"productservice/internal/model"
	"productservice/internal/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RateRepository interface {
	RatingProduct(ctx context.Context, rateProduct *dto.RateProduct) (*dto.RateResponse, *dto.ServiceResponse)
	UpdateRatingProduct(ctx context.Context, ratingProduct *dto.UpdateRatingProduct) (*dto.RateResponse, *dto.ServiceResponse)
	DeleteRatingProduct(ctx context.Context, ratingId uuid.UUID) *dto.ServiceResponse
	GetRateListOfProduct(ctx context.Context, filterRequest *dto.RateListOfProduct) (*dto.PageResponse, *dto.ServiceResponse)
	GetRateStatisticOfProduct(ctx context.Context, productId uuid.UUID) (*dto.RatingSummaryResponse, *dto.ServiceResponse)
	GetMyRatings(ctx context.Context, userId uuid.UUID, request *dto.MyRatingsRequest) (*dto.MyRatingsResponse, *dto.ServiceResponse)
}

type rateRepository struct {
	db         *gorm.DB
	rd         *redis.Client
	authClient *client.AuthClient
}

func NewRateRepository(db *gorm.DB, rd *redis.Client, authClient *client.AuthClient) RateRepository {
	return &rateRepository{db: db, rd: rd, authClient: authClient}
}

func (r *rateRepository) GetMyRatings(ctx context.Context, userId uuid.UUID, request *dto.MyRatingsRequest) (*dto.MyRatingsResponse, *dto.ServiceResponse) {

	// cache
	key := baseRateOfUser + userId.String()
	if cached, err := r.rd.Get(ctx, key).Bytes(); err == nil && cached != nil {
		var response dto.MyRatingsResponse
		if err := json.Unmarshal(cached, &response); err == nil {
			log.Info("Using redis...")
			return &response, nil
		}
	}

	query := r.db.WithContext(ctx)

	// Base query với user filter
	baseQuery := query.Model(&model.Rating{}).
		Preload("Product").
		Preload("Product.Category").
		Where("user_id = ?", userId)

	// Get total count
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// Nếu không có rating nào
	if total == 0 {

		response := dto.MyRatingsResponse{
			Data:       nil,
			Total:      total,
			Pagination: request,
			Summary: &dto.MyRatingSummary{
				TotalRatings:  0,
				AvgStarsGiven: 0,
			},
		}

		// =====  Cache response =====
		if b, err := json.Marshal(response); err == nil {
			log.Info("Update/Save redis...")
			_ = r.rd.Set(context.Background(), key, b, time.Hour).Err()
		}

		return &response, nil
	}

	// Apply sorting
	sortOrder := "created_at DESC" // default
	switch request.SortBy {
	case "newest":
		sortOrder = "created_at DESC"
	case "oldest":
		sortOrder = "created_at ASC"
	case "highest":
		sortOrder = "rating DESC, created_at DESC"
	case "lowest":
		sortOrder = "rating ASC, created_at DESC"
	}
	baseQuery = baseQuery.Order(sortOrder)

	// Apply pagination
	page := request.Page
	if page < 1 {
		page = 1
	}
	limit := request.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Fetch ratings
	var ratings []model.Rating
	if err := baseQuery.Offset(offset).Limit(limit).Find(&ratings).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// Tính tổng stars để có avg
	var totalStars int64
	if err := query.Model(&model.Rating{}).
		Where("user_id = ? AND deleted_at IS NULL", userId).
		Select("COALESCE(SUM(rating), 0)").
		Scan(&totalStars).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	avgStarsGiven := float64(totalStars) / float64(total)
	avgStarsGiven = float64(int(avgStarsGiven*100)) / 100

	// Map to response
	data := make([]*dto.RateResponse, 0, len(ratings))
	for _, rating := range ratings {
		comment := ""
		if rating.Comment != nil {
			comment = *rating.Comment
		}

		item := &dto.RateResponse{
			Id: rating.ID,
			Product: dto.ProductOfRate{
				ID:   rating.Product.ID,
				Name: rating.Product.Name,
			},
			Star:     rating.Rating,
			Comment:  comment,
			CreateAt: rating.CreatedAt,
			UpdateAt: rating.UpdatedAt,
		}
		data = append(data, item)
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	response := dto.MyRatingsResponse{
		Data:       data,
		Pagination: request,
		Total:      total,
		Summary: &dto.MyRatingSummary{
			TotalRatings:  total,
			AvgStarsGiven: avgStarsGiven,
		},
	}

	// =====  Cache response =====
	if b, err := json.Marshal(response); err == nil {
		log.Info("Update/Save redis...")
		_ = r.rd.Set(context.Background(), key, b, time.Hour).Err()
	}

	return &response, nil
}

func (r *rateRepository) GetRateStatisticOfProduct(ctx context.Context, productId uuid.UUID) (*dto.RatingSummaryResponse, *dto.ServiceResponse) {

	// cache
	key := baseRateStatistic + productId.String()
	if cached, err := r.rd.Get(ctx, key).Bytes(); err == nil && cached != nil {
		var response dto.RatingSummaryResponse
		if err := json.Unmarshal(cached, &response); err == nil {
			log.Info("Using redis...")
			return &response, nil
		}
	}

	query := r.db.WithContext(ctx)

	// Kiểm tra product có tồn tại không
	var product model.Product
	if err := query.Where("id = ? AND deleted_at IS NULL", productId).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dto.ServiceResponse{
				Status: http.StatusNotFound,
				Err:    ErrNotFound,
			}
		}
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// Lấy tất cả ratings của product
	var ratings []model.Rating
	if err := query.Where("product_id = ? AND deleted_at IS NULL", productId).Find(&ratings).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	total := int64(len(ratings))

	// Nếu không có rating nào
	if total == 0 {
		response := dto.RatingSummaryResponse{
			ProductID: productId.String(),
			Summary: &dto.RatingSummary{
				AvgRating:    0,
				TotalRatings: 0,
			},
			Distribution:      make(map[string]*dto.StarDetail),
			DistributionChart: []*dto.DistributionChart{},
		}

		// =====  Cache response =====
		if b, err := json.Marshal(response); err == nil {
			log.Info("Update/Save redis...")
			_ = r.rd.Set(context.Background(), key, b, time.Hour).Err()
		}

		return &response, nil
	}

	// Đếm số lượng mỗi loại sao
	starCounts := map[int]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	var sumRating float64

	for _, rating := range ratings {
		starCounts[rating.Rating]++
		sumRating += float64(rating.Rating)
	}

	// Tính average rating
	avgRating := sumRating / float64(total)
	// Làm tròn đến 2 chữ số thập phân
	avgRating = float64(int(avgRating*100)) / 100

	// Tạo distribution map và chart
	distribution := make(map[string]*dto.StarDetail)
	var distributionChart []*dto.DistributionChart

	for star := 5; star >= 1; star-- {
		count := starCounts[star]
		percentage := (float64(count) / float64(total)) * 100
		// Làm tròn percentage đến 2 chữ số thập phân
		percentage = float64(int(percentage*100)) / 100

		starKey := fmt.Sprintf("%d", star)
		distribution[starKey] = &dto.StarDetail{
			Count:      count,
			Percentage: percentage,
		}

		distributionChart = append(distributionChart, &dto.DistributionChart{
			Stars:      star,
			Count:      count,
			Percentage: percentage,
		})
	}

	response := dto.RatingSummaryResponse{
		ProductID: productId.String(),
		Summary: &dto.RatingSummary{
			AvgRating:    avgRating,
			TotalRatings: total,
		},
		Distribution:      distribution,
		DistributionChart: distributionChart,
	}

	// =====  Cache response =====
	if b, err := json.Marshal(response); err == nil {
		log.Info("Update/Save redis...")
		_ = r.rd.Set(context.Background(), key, b, time.Hour).Err()
	}

	return &response, nil
}

func (r *rateRepository) GetRateListOfProduct(ctx context.Context, filterRequest *dto.RateListOfProduct) (*dto.PageResponse, *dto.ServiceResponse) {

	// cache
	key := baseRateOfProduct + filterRequest.ProductId.String() +
		"page:" + strconv.Itoa(filterRequest.Page) +
		"limit:" + strconv.Itoa(filterRequest.Limit) +
		"star:" + strconv.Itoa(filterRequest.Stars) +
		"sort: " + filterRequest.SortBy
	if cached, err := r.rd.Get(ctx, key).Bytes(); err == nil && cached != nil {
		var response dto.PageResponse
		if err := json.Unmarshal(cached, &response); err == nil {
			log.Info("Using redis...")
			return &response, nil
		}
	}

	query := r.db.WithContext(ctx)

	// Base query with product filter
	baseQuery := query.Model(&model.Rating{}).Where("product_id = ? AND deleted_at IS NULL", filterRequest.ProductId)

	// Apply star filter if specified
	if filterRequest.Stars > 0 && filterRequest.Stars <= 5 {
		baseQuery = baseQuery.Where("rating = ?", filterRequest.Stars)
	}

	// Get total count
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// Apply sorting
	sortOrder := "created_at DESC" // default sort
	switch filterRequest.SortBy {
	case "newest":
		sortOrder = "created_at DESC"
	case "oldest":
		sortOrder = "created_at ASC"
	case "highest":
		sortOrder = "stars DESC, created_at DESC"
	case "lowest":
		sortOrder = "stars ASC, created_at DESC"
	}
	baseQuery = baseQuery.Order(sortOrder)

	// Apply pagination
	page := filterRequest.Page
	if page < 1 {
		page = 1
	}
	limit := filterRequest.Limit
	if limit < 1 {
		limit = 10 // default limit
	}
	offset := (page - 1) * limit

	// Fetch ratings with pagination
	var ratings []*model.Rating
	if err := baseQuery.Offset(offset).Limit(limit).Find(&ratings).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// Return paginated response
	response := dto.PageResponse{
		Total:  total,
		Data:   ratings,
		Filter: filterRequest,
	}
	// =====  Cache response =====
	if b, err := json.Marshal(response); err == nil {
		log.Info("Update/Save redis...")
		_ = r.rd.Set(context.Background(), key, b, time.Hour).Err()
	}

	return &response, nil
}

func (r *rateRepository) DeleteRatingProduct(ctx context.Context, ratingId uuid.UUID) *dto.ServiceResponse {
	query := r.db.WithContext(ctx)

	// Tìm rating hiện tại
	var rating model.Rating
	if err := query.Where("id = ? AND deleted_at IS NULL", ratingId).First(&rating).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.ServiceResponse{
				Status: http.StatusNotFound,
				Err:    ErrNotFound,
			}
		}
		return &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	userId := ctx.Value("userId")

	if userId != rating.UserID {
		return &dto.ServiceResponse{
			Status: http.StatusForbidden,
			Err:    ErrForbidden,
		}
	}

	if err := query.Delete(&rating).Error; err != nil {
		return &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// =====  Cache response =====
	log.Info("Delete redis...")
	_ = r.rd.Del(ctx, baseRateStatistic+rating.ProductID.String()).Err()
	_ = r.rd.Del(ctx, baseRateOfUser+rating.UserID.String())
	_ = utils.DeleteCacheByPattern(ctx, r.rd, baseRateOfProduct+rating.ProductID.String())

	return nil

}

func (r *rateRepository) UpdateRatingProduct(ctx context.Context, ratingProduct *dto.UpdateRatingProduct) (*dto.RateResponse, *dto.ServiceResponse) {
	query := r.db.WithContext(ctx)

	// Tìm rating hiện tại
	var rating model.Rating
	if err := query.Preload("Product").Where("id = ? AND deleted_at IS NULL", ratingProduct.RateId).First(&rating).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dto.ServiceResponse{
				Status: http.StatusNotFound,
				Err:    ErrNotFound,
			}
		}
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	token := ctx.Value("token")
	user, err := r.authClient.GetCurrentUserInfo(ctx, token.(string))
	if err != nil {
		response := dto.ServiceResponse{
			Status: http.StatusUnauthorized,
			Err:    ErrNotFound,
		}
		return nil, &response
	}
	userUuid, err := uuid.Parse(user.Id)
	if err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	if rating.UserID != userUuid {
		return nil, &dto.ServiceResponse{
			Status: http.StatusForbidden,
			Err:    ErrForbidden,
		}
	}

	// Cập nhật thông tin
	rating.Rating = ratingProduct.Star
	rating.Comment = &ratingProduct.Comment
	rating.UpdatedAt = time.Now()

	// Lưu vào database
	if err := query.Save(&rating).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	product := rating.Product

	// Trả về response
	var description string
	if rating.Comment != nil {
		description = *rating.Comment
	}

	var productDesc string
	if product.Description != nil {
		productDesc = *product.Description
	}

	response := &dto.RateResponse{
		Id:      rating.ID,
		Star:    rating.Rating,
		Comment: description,
		Product: dto.ProductOfRate{
			ID:            product.ID,
			Name:          product.Name,
			Description:   productDesc,
			Price:         product.Price,
			AverageRating: product.AverageRating,
			TotalRatings:  product.TotalRatings,
		},
		User: dto.UserOfRate{
			Id:   userUuid,
			Name: user.Name,
		},
		UpdateAt: rating.UpdatedAt,
	}

	// =====  Cache response =====
	log.Info("Delete redis...")
	_ = r.rd.Del(ctx, baseRateStatistic+rating.ProductID.String()).Err()
	_ = r.rd.Del(ctx, baseRateOfUser+rating.UserID.String())
	_ = utils.DeleteCacheByPattern(ctx, r.rd, baseRateOfProduct+rating.ProductID.String())

	return response, nil
}

func (r *rateRepository) RatingProduct(ctx context.Context, rateProduct *dto.RateProduct) (*dto.RateResponse, *dto.ServiceResponse) {
	query := r.db.WithContext(ctx)
	// check product
	var product model.Product
	if err := query.Where("id = ? AND deleted_at IS NULL", rateProduct.ProductId).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response := dto.ServiceResponse{
				Status: http.StatusNotFound,
				Err:    ErrNotFound,
			}
			return nil, &response
		}
		response := dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
		return nil, &response
	}

	// check was rating
	var existingRating model.Rating
	err := query.Where("user_id = ? AND product_id = ? AND deleted_at IS NULL",
		rateProduct.UserId, rateProduct.ProductId).
		First(&existingRating).Error

	if err == nil {
		response := dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    ErrWasRateProduct,
		}
		return nil, &response
	}

	if err != gorm.ErrRecordNotFound {
		return nil, &dto.ServiceResponse{
			Status: fiber.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	token := ctx.Value("token")
	resp, err := r.authClient.GetCurrentUserInfo(ctx, token.(string))
	if err != nil {
		response := dto.ServiceResponse{
			Status: http.StatusUnauthorized,
			Err:    ErrNotFound,
		}
		return nil, &response
	}

	newRating := model.Rating{
		ID:        uuid.New(),
		ProductID: rateProduct.ProductId,
		UserID:    rateProduct.UserId,
		Rating:    rateProduct.Star,
		Comment:   &rateProduct.Comment,
	}

	if err := query.Create(&newRating).Error; err != nil {
		return nil, &dto.ServiceResponse{
			Status: fiber.StatusInternalServerError,
			Err:    ErrInternalServerError,
		}
	}

	// Invalidate cache của product (nếu có cache)
	go func() {
		cacheCtx := context.Background()
		cacheKey := fmt.Sprintf("product:%s", rateProduct.ProductId.String())
		r.rd.Del(cacheCtx, cacheKey)

	}()

	return &dto.RateResponse{
		Id: newRating.ID,
		Product: dto.ProductOfRate{
			ID:            product.ID,
			Name:          product.Name,
			Price:         product.Price,
			Description:   *product.Description,
			AverageRating: product.AverageRating,
			TotalRatings:  product.TotalRatings,
		},
		User: dto.UserOfRate{
			Id:   rateProduct.UserId,
			Name: resp.Name,
		},
		Star:     newRating.Rating,
		Comment:  *newRating.Comment,
		CreateAt: newRating.CreatedAt,
	}, nil
}
