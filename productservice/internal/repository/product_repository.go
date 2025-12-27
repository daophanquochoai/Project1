package repository

import (
	"context"
	"encoding/json"
	"errors"
	"productservice/internal/dto"
	"productservice/internal/model"
	"productservice/internal/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProductRepository interface {
	GetList(ctx context.Context, request *dto.PageProdRequest) (*dto.PageResponse, *dto.ServiceResponse)
	GetProductById(ctx context.Context, productId uuid.UUID) (*model.Product, *dto.ServiceResponse)
	GetProductSimilar(ctx context.Context, productId uuid.UUID, pageRequest *dto.PageSimilarAndRelatedRequest, relationType *dto.RelationType) (*dto.PageResponse, *dto.ServiceResponse)
}

type productRepository struct {
	db *gorm.DB
	rd *redis.Client
}

func NewProductRepository(db *gorm.DB, rd *redis.Client) ProductRepository {
	return &productRepository{db: db, rd: rd}
}

func (p *productRepository) GetProductSimilar(
	ctx context.Context,
	productId uuid.UUID,
	pageRequest *dto.PageSimilarAndRelatedRequest,
	relationType *dto.RelationType,
) (*dto.PageResponse, *dto.ServiceResponse) {

	// =====  Validate input =====
	if pageRequest.Page <= 0 {
		pageRequest.Page = 1
	}
	if pageRequest.Limit <= 0 {
		pageRequest.Limit = 10
	}

	offset := (pageRequest.Page - 1) * pageRequest.Limit

	// ===== Cache key =====
	var baseKey string
	if *relationType == dto.Relation_Related {
		baseKey = baseProductRelated
	} else {
		baseKey = baseProductSimilar
	}
	key := baseKey +
		":product:" + productId.String() +
		":page:" + strconv.Itoa(pageRequest.Page) +
		":limit:" + strconv.Itoa(pageRequest.Limit)

	// ===== Try cache =====
	if cached, err := p.rd.Get(ctx, key).Bytes(); err == nil && cached != nil {
		var response dto.PageResponse
		if err := json.Unmarshal(cached, &response); err == nil {
			log.Info("Using redis...")
			return &response, nil
		}
	}

	// ===== Count total =====
	var total int64
	err := p.db.WithContext(ctx).
		Model(&model.Product{}).
		Joins("INNER JOIN product_related pr ON pr.related_id = products.id").
		Where(`
			pr.product_id = ?
			AND pr.relation_type = ?
			AND products.deleted_at IS NULL
		`, productId, relationType).
		Count(&total).Error

	if err != nil {
		return nil, &dto.ServiceResponse{
			Status: 500,
			Err:    ErrInternalServerError,
		}
	}

	if total == 0 {
		return &dto.PageResponse{
			Total: total,
			Data:  []model.Product{},
		}, nil
	}

	// ===== Query data =====
	var products []model.Product
	err = p.db.WithContext(ctx).
		Model(&model.Product{}).
		Preload("Category").
		Joins("INNER JOIN product_related pr ON pr.related_id = products.id").
		Where(`
			pr.product_id = ?
			AND pr.relation_type = ?
			AND products.deleted_at IS NULL
		`, productId, relationType).
		Order("products.average_rating DESC, products.created_at DESC").
		Limit(pageRequest.Limit).
		Offset(offset).
		Find(&products).Error

	if err != nil {
		return nil, &dto.ServiceResponse{
			Status: 500,
			Err:    ErrInternalServerError,
		}
	}

	// =====  Build response =====
	response := dto.PageResponse{
		Total:  total,
		Data:   products,
		Filter: pageRequest,
	}

	// =====  Cache response =====
	if b, err := json.Marshal(response); err == nil {
		log.Info("Update/Save redis...")
		_ = p.rd.Set(context.Background(), key, b, time.Hour).Err()
	}

	return &response, nil
}

func (p *productRepository) GetProductById(ctx context.Context, productId uuid.UUID) (*model.Product, *dto.ServiceResponse) {
	var product model.Product
	key := baseProduct + productId.String()

	// 1. Kiểm tra cache trước
	cachedData, err := p.rd.Get(ctx, key).Result()
	if err == nil && cachedData != "" {
		if err := json.Unmarshal([]byte(cachedData), &product); err == nil {
			return &product, nil
		}
	}

	// 2. Cache miss - query từ database
	err = p.db.WithContext(ctx).
		Preload("Category").
		Where("id = ? AND deleted_at IS NULL", productId).
		First(&product).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dto.ServiceResponse{
				Err:    ErrNotFound,
				Status: 404,
			}
		}

		return nil, &dto.ServiceResponse{
			Err:    ErrInternalServerError,
			Status: 500,
		}
	}

	// 3. Lưu vào cache (TTL 1 giờ)
	productJSON, err := json.Marshal(product)
	if err == nil {
		log.Info("Update redis...")
		p.rd.Set(ctx, key, productJSON, time.Hour)
	}

	return &product, nil
}

func (p *productRepository) GetList(ctx context.Context, request *dto.PageProdRequest) (*dto.PageResponse, *dto.ServiceResponse) {
	var products []model.Product
	var total int64

	// Validate và set default values
	if request.Page < 1 {
		request.Page = 1
	}
	if request.Limit < 1 {
		request.Limit = 10
	}

	// Tính offset cho phân trang
	offset := (request.Page - 1) * request.Limit

	// Query builder
	query := p.db.WithContext(ctx).Model(&model.Product{})

	// Filter theo Search (hỗ trợ tiếng Việt)
	if request.Search != "" {
		normalizedSearch := utils.NormalizeSearchText(request.Search)
		query = query.Where("search_name LIKE ?", "%"+normalizedSearch+"%")
	}

	// Filter theo danh sách CategoryIds
	if len(request.CategoryIds) > 0 {
		query = query.Where("category_id IN ?", request.CategoryIds)
	}

	// Filter theo khoảng giá
	if request.MinPrice > 0 {
		query = query.Where("price >= ?", request.MinPrice)
	}

	if request.MaxPrice > 0 {
		query = query.Where("price <= ?", request.MaxPrice)
	}

	// Filter theo khoảng đánh giá (rating)
	if request.MinRate > 0 {
		query = query.Where("average_rating >= ?", request.MinRate)
	}

	if request.MaxRate > 0 {
		query = query.Where("average_rating <= ?", request.MaxRate)
	}

	// Đếm tổng số bản ghi
	if err := query.Count(&total).Error; err != nil {
		response := dto.ServiceResponse{
			Status: 500,
			Err:    ErrInternalServerError,
		}
		return nil, &response
	}

	// Áp dụng sắp xếp
	if request.SortBy != "" {
		order := request.SortBy
		if request.SortOrder != "" {
			order += " " + strings.ToUpper(request.SortOrder)
		} else {
			order += " ASC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// Lấy dữ liệu với phân trang
	if err := query.Offset(offset).Limit(request.Limit).Find(&products).Error; err != nil {
		response := dto.ServiceResponse{
			Status: 500,
			Err:    ErrInternalServerError,
		}
		return nil, &response
	}

	return &dto.PageResponse{
		Data:   products,
		Total:  total,
		Filter: request,
	}, nil
}
