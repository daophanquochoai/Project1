package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"productservice/internal/dto"
	"productservice/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	redis      redismock.ClientMock
	repository ProductRepository
	ctx        context.Context
}

func (suite *ProductRepositoryTestSuite) SetupSuite() {
	// SQLite for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	suite.Require().NoError(err, "Failed to open database")

	// Create tables manually for SQLite
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS categories (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		search_name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		category_id TEXT NOT NULL,
		average_rating REAL NOT NULL DEFAULT 0,
		total_ratings INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME,
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);

	CREATE TABLE IF NOT EXISTS ratings (
		id TEXT PRIMARY KEY,
		product_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
		comment TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted_at DATETIME,
		FOREIGN KEY (product_id) REFERENCES products(id)
	);

	CREATE TABLE IF NOT EXISTS product_relate (
		product_id TEXT NOT NULL,
		related_id TEXT NOT NULL,
		relation_type TEXT NOT NULL CHECK(relation_type IN ('related', 'similar')),
		created_at DATETIME NOT NULL,
		PRIMARY KEY (product_id, related_id),
		FOREIGN KEY (product_id) REFERENCES products(id),
		FOREIGN KEY (related_id) REFERENCES products(id)
	);

	CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);
	CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
	CREATE INDEX IF NOT EXISTS idx_products_search_name ON products(search_name);
	CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
	CREATE INDEX IF NOT EXISTS idx_products_rating ON products(average_rating);
	CREATE INDEX IF NOT EXISTS idx_ratings_product_id ON ratings(product_id);
	CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);
	CREATE INDEX IF NOT EXISTS idx_ratings_deleted_at ON ratings(deleted_at);
	`

	result := db.Exec(createTablesSQL)
	suite.Require().NoError(result.Error, "Failed to create tables")

	suite.db = db
	suite.ctx = context.Background()
}

func (suite *ProductRepositoryTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM product_relate")
	suite.db.Exec("DELETE FROM ratings")
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")

	rdb, mock := redismock.NewClientMock()
	suite.redis = mock
	suite.repository = NewProductRepository(suite.db, rdb)
}

func (suite *ProductRepositoryTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *ProductRepositoryTestSuite) cleanupData() {
	suite.db.Exec("DELETE FROM product_relate")
	suite.db.Exec("DELETE FROM ratings")
	suite.db.Exec("DELETE FROM products")
	suite.db.Exec("DELETE FROM categories")
}

// Test GetProductById
func (suite *ProductRepositoryTestSuite) TestGetProductById() {
	tests := []struct {
		name         string
		productID    uuid.UUID
		setup        func() *model.Product
		mockRedis    func(uuid.UUID, *model.Product)
		wantError    bool
		expectStatus int
		validate     func(*testing.T, *model.Product, *dto.ServiceResponse)
	}{
		{
			name:      "success - get from database",
			productID: uuid.New(),
			setup: func() *model.Product {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				product := &model.Product{
					ID:            uuid.New(),
					Name:          "Laptop",
					SearchName:    "laptop",
					Price:         1000,
					AverageRating: 4.5,
					TotalRatings:  10,
					CategoryID:    category.ID,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				suite.db.Create(product)
				return product
			},
			mockRedis: func(id uuid.UUID, product *model.Product) {
				key := baseProduct + id.String()
				suite.redis.ExpectGet(key).RedisNil()
				suite.redis.Regexp().ExpectSet(key, `.*`, time.Hour).SetVal("OK")
			},
			wantError: false,
			validate: func(t *testing.T, result *model.Product, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
				assert.Equal(t, "Laptop", result.Name)
				assert.Equal(t, float64(1000), result.Price)
				assert.Equal(t, 4.5, result.AverageRating)
			},
		},
		{
			name:      "success - get from cache",
			productID: uuid.New(),
			setup: func() *model.Product {
				suite.cleanupData()
				return &model.Product{
					ID:            uuid.New(),
					Name:          "Phone",
					SearchName:    "phone",
					Price:         500,
					AverageRating: 4.0,
					TotalRatings:  5,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
			},
			mockRedis: func(id uuid.UUID, product *model.Product) {
				key := baseProduct + id.String()
				productJSON, _ := json.Marshal(product)
				suite.redis.ExpectGet(key).SetVal(string(productJSON))
			},
			wantError: false,
			validate: func(t *testing.T, result *model.Product, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
				assert.Equal(t, "Phone", result.Name)
			},
		},
		{
			name:      "error - product not found",
			productID: uuid.New(),
			setup: func() *model.Product {
				suite.cleanupData()
				return nil
			},
			mockRedis: func(id uuid.UUID, product *model.Product) {
				key := baseProduct + id.String()
				suite.redis.ExpectGet(key).RedisNil()
			},
			wantError:    true,
			expectStatus: 404,
			validate: func(t *testing.T, result *model.Product, resp *dto.ServiceResponse) {
				assert.Nil(t, result)
				assert.NotNil(t, resp)
				assert.Equal(t, 404, resp.Status)
				assert.Equal(t, ErrNotFound, resp.Err)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			product := tt.setup()
			if product != nil {
				tt.mockRedis(product.ID, product)
				result, resp := suite.repository.GetProductById(suite.ctx, product.ID)
				tt.validate(suite.T(), result, resp)
			} else {
				tt.mockRedis(tt.productID, nil)
				result, resp := suite.repository.GetProductById(suite.ctx, tt.productID)
				tt.validate(suite.T(), result, resp)
			}
			assert.NoError(suite.T(), suite.redis.ExpectationsWereMet())
		})
	}
}

// Test GetList
func (suite *ProductRepositoryTestSuite) TestGetList() {
	tests := []struct {
		name        string
		setup       func()
		pageRequest func() *dto.PageProdRequest // ← Changed to function type
		validate    func(*testing.T, *dto.PageResponse, *dto.ServiceResponse)
	}{
		{
			name: "success - list all products",
			setup: func() {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				products := []*model.Product{
					{ID: uuid.New(), Name: "Laptop", SearchName: "laptop", Price: 1000, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Phone", SearchName: "phone", Price: 500, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Tablet", SearchName: "tablet", Price: 300, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				for _, p := range products {
					suite.db.Create(p)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:  1,
					Limit: 10,
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
				assert.Equal(t, int64(3), result.Total)
				assert.Len(t, result.Data, 3)
			},
		},
		{
			name: "success - search by name",
			setup: func() {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				products := []*model.Product{
					{ID: uuid.New(), Name: "Laptop Dell", SearchName: "laptop dell", Price: 1000, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Laptop HP", SearchName: "laptop hp", Price: 900, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Phone", SearchName: "phone", Price: 500, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				for _, p := range products {
					suite.db.Create(p)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:   1,
					Limit:  10,
					Search: "laptop",
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(2), result.Total)
				assert.Len(t, result.Data, 2)
			},
		},
		{
			name: "success - filter by category",
			setup: func() {
				suite.cleanupData()
				cat1 := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				cat2 := &model.Category{
					ID:        uuid.New(),
					Name:      "Books",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(cat1)
				suite.db.Create(cat2)

				products := []*model.Product{
					{ID: uuid.New(), Name: "Laptop", SearchName: "laptop", Price: 1000, CategoryID: cat1.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Book", SearchName: "book", Price: 20, CategoryID: cat2.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Phone", SearchName: "phone", Price: 500, CategoryID: cat1.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				for _, p := range products {
					suite.db.Create(p)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Fixed: No () at the end
				var cat model.Category
				suite.db.Where("name = ?", "Electronics").First(&cat)
				return &dto.PageProdRequest{
					Page:        1,
					Limit:       10,
					CategoryIds: []uuid.UUID{cat.ID},
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(2), result.Total)
				assert.Len(t, result.Data, 2)
			},
		},
		{
			name: "success - filter by price range",
			setup: func() {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				products := []*model.Product{
					{ID: uuid.New(), Name: "Laptop", SearchName: "laptop", Price: 1000, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Phone", SearchName: "phone", Price: 500, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Tablet", SearchName: "tablet", Price: 300, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				for _, p := range products {
					suite.db.Create(p)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:     1,
					Limit:    10,
					MinPrice: 400,
					MaxPrice: 600,
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(1), result.Total)
				assert.Len(t, result.Data, 1)
			},
		},
		{
			name: "success - filter by rating range",
			setup: func() {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				products := []*model.Product{
					{ID: uuid.New(), Name: "Product A", SearchName: "product a", Price: 100, AverageRating: 4.5, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Product B", SearchName: "product b", Price: 200, AverageRating: 3.5, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Product C", SearchName: "product c", Price: 300, AverageRating: 2.5, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				for _, p := range products {
					suite.db.Create(p)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:    1,
					Limit:   10,
					MinRate: 3.0,
					MaxRate: 5.0,
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(2), result.Total)
				assert.Len(t, result.Data, 2)
			},
		},
		{
			name: "success - pagination",
			setup: func() {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				for i := 0; i < 25; i++ {
					product := &model.Product{
						ID:         uuid.New(),
						Name:       fmt.Sprintf("Product %d", i),
						SearchName: fmt.Sprintf("product %d", i),
						Price:      float64(100 * i),
						CategoryID: category.ID,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}
					suite.db.Create(product)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:  2,
					Limit: 10,
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(25), result.Total)
				assert.Len(t, result.Data, 10)
			},
		},
		{
			name: "success - sort by price ascending",
			setup: func() {
				suite.cleanupData()
				category := &model.Category{
					ID:        uuid.New(),
					Name:      "Electronics",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				suite.db.Create(category)

				products := []*model.Product{
					{ID: uuid.New(), Name: "Product C", SearchName: "product c", Price: 300, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Product A", SearchName: "product a", Price: 100, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: uuid.New(), Name: "Product B", SearchName: "product b", Price: 200, CategoryID: category.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				for _, p := range products {
					suite.db.Create(p)
				}
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:      1,
					Limit:     10,
					SortBy:    "price",
					SortOrder: "ASC",
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(3), result.Total)
				products := result.Data.([]model.Product)
				assert.Equal(t, float64(100), products[0].Price)
				assert.Equal(t, float64(200), products[1].Price)
				assert.Equal(t, float64(300), products[2].Price)
			},
		},
		{
			name: "success - empty result",
			setup: func() {
				suite.cleanupData()
			},
			pageRequest: func() *dto.PageProdRequest { // ← Now a function
				return &dto.PageProdRequest{
					Page:  1,
					Limit: 10,
				}
			},
			validate: func(t *testing.T, result *dto.PageResponse, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, int64(0), result.Total)
				assert.Len(t, result.Data, 0)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setup()

			// Call pageRequest() function AFTER setup has run
			pageReq := tt.pageRequest()

			result, resp := suite.repository.GetList(suite.ctx, pageReq)
			tt.validate(suite.T(), result, resp)
		})
	}
}

func TestProductRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}
