package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/agris/user-service/internal/dto"
	"github.com/agris/user-service/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	redis      redismock.ClientMock
	repository UserRepository
	ctx        context.Context
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	// SQLite for testing - PostgreSQL for production
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	suite.Require().NoError(err, "Failed to open database")

	// AutoMigrate sẽ FAIL với SQLite vì gen_random_uuid() là PostgreSQL function
	// Đây là EXPECTED behavior - không phải bug
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		// Expected: SQLite không hiểu PostgreSQL syntax
		// Giải pháp: Tạo table manually với SQL thuần SQLite
		suite.T().Log("AutoMigrate failed (expected for SQLite), creating table manually...")

		createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
		CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
		CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
		`
		result := db.Exec(createTableSQL)
		suite.Require().NoError(result.Error, "Failed to create table manually")

		suite.T().Log("✓ Table created successfully with SQLite syntax")
	}

	// Verify table exists
	var count int64
	db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	suite.Require().Equal(int64(1), count, "Users table was not created")

	suite.db = db
	suite.ctx = context.Background()
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM users")

	rdb, mock := redismock.NewClientMock()
	suite.redis = mock
	suite.repository = NewUserRepository(suite.db, rdb)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// Helper function to cleanup users
func (suite *UserRepositoryTestSuite) cleanupUsers() {
	suite.db.Exec("DELETE FROM users")
}

// Test Create with Table Driven
func (suite *UserRepositoryTestSuite) TestCreate() {
	tests := []struct {
		name      string
		user      *model.User
		setup     func()
		wantError bool
		validate  func(*testing.T, *model.User, error)
	}{
		{
			name: "success - create new user",
			user: &model.User{
				ID:           uuid.New(),
				Email:        "test@example.com",
				PasswordHash: "hashedpassword",
				Role:         "user",
			},
			setup: func() {
				suite.cleanupUsers()
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test@example.com", result.Email)
				assert.NotEqual(t, uuid.Nil, result.ID)
			},
		},
		{
			name: "error - duplicate email",
			user: &model.User{
				ID:           uuid.New(),
				Email:        "duplicate@example.com",
				PasswordHash: "password",
			},
			setup: func() {
				suite.cleanupUsers()
				existingUser := &model.User{
					ID:           uuid.New(),
					Email:        "duplicate@example.com",
					PasswordHash: "password123",
				}
				suite.db.Create(existingUser)
			},
			wantError: true,
			validate: func(t *testing.T, result *model.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name: "success - create admin user",
			user: &model.User{
				ID:           uuid.New(),
				Email:        "admin@example.com",
				PasswordHash: "adminpass",
				Role:         model.RoleAdmin,
			},
			setup: func() {
				suite.cleanupUsers()
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, model.RoleAdmin, result.Role)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setup()
			result, err := suite.repository.Create(suite.ctx, tt.user)
			tt.validate(suite.T(), result, err)
		})
	}
}

// Test FindByID with Table Driven
func (suite *UserRepositoryTestSuite) TestFindByID() {
	tests := []struct {
		name         string
		userID       uuid.UUID
		setup        func() *model.User
		mockRedis    func(uuid.UUID, *model.User)
		wantError    bool
		expectStatus int
		validate     func(*testing.T, *model.User, *dto.ServiceResponse)
	}{
		{
			name:   "success - find from database",
			userID: uuid.New(),
			setup: func() *model.User {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "findbyid@example.com",
					PasswordHash: "password",
					Role:         model.RoleUser,
				}
				suite.db.Create(user)
				return user
			},
			mockRedis: func(id uuid.UUID, user *model.User) {
				key := baseUser + id.String()
				suite.redis.ExpectGet(key).RedisNil()
				suite.redis.Regexp().ExpectSet(key, `.*`, 30*time.Minute).SetVal("OK")
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
				assert.Equal(t, "findbyid@example.com", result.Email)
			},
		},
		{
			name:   "success - find from cache",
			userID: uuid.New(),
			setup: func() *model.User {
				suite.cleanupUsers()
				return &model.User{
					ID:           uuid.New(),
					Email:        "cached@example.com",
					PasswordHash: "password",
					Role:         model.RoleUser,
				}
			},
			mockRedis: func(id uuid.UUID, user *model.User) {
				key := baseUser + id.String()
				userJSON, _ := json.Marshal(user)
				suite.redis.ExpectGet(key).SetVal(string(userJSON))
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
			},
		},
		{
			name:   "error - user not found",
			userID: uuid.New(),
			setup: func() *model.User {
				suite.cleanupUsers()
				return nil
			},
			mockRedis: func(id uuid.UUID, user *model.User) {
				key := baseUser + id.String()
				suite.redis.ExpectGet(key).RedisNil()
			},
			wantError:    true,
			expectStatus: 404,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, result)
				assert.NotNil(t, resp)
				assert.Equal(t, 404, resp.Status)
				assert.Equal(t, ErrNotFound, resp.Err)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			user := tt.setup()
			if user != nil {
				tt.mockRedis(user.ID, user)
				result, resp := suite.repository.FindByID(suite.ctx, user.ID)
				tt.validate(suite.T(), result, resp)
			} else {
				tt.mockRedis(tt.userID, nil)
				result, resp := suite.repository.FindByID(suite.ctx, tt.userID)
				tt.validate(suite.T(), result, resp)
			}
			assert.NoError(suite.T(), suite.redis.ExpectationsWereMet())
		})
	}
}

// Test FindByEmail with Table Driven
func (suite *UserRepositoryTestSuite) TestFindByEmail() {
	tests := []struct {
		name           string
		email          string
		includeDeleted bool
		setup          func() *model.User
		mockRedis      func(string, *model.User)
		wantError      bool
		expectStatus   int
		validate       func(*testing.T, *model.User, *dto.ServiceResponse)
	}{
		{
			name:           "success - find by email",
			email:          "findbyemail@example.com",
			includeDeleted: false,
			setup: func() *model.User {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "findbyemail@example.com",
					PasswordHash: "password",
					Role:         model.RoleUser,
				}
				suite.db.Create(user)
				return user
			},
			mockRedis: func(email string, user *model.User) {
				key := baseUserEmail + email
				suite.redis.ExpectGet(key).RedisNil()
				suite.redis.Regexp().ExpectSet(key, `.*`, 30*time.Minute).SetVal("OK")
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
				assert.Equal(t, "findbyemail@example.com", result.Email)
			},
		},
		{
			name:           "success - find deleted user with flag",
			email:          "deleted@example.com",
			includeDeleted: true,
			setup: func() *model.User {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "deleted@example.com",
					PasswordHash: "password",
				}
				suite.db.Create(user)
				suite.db.Delete(user)
				return user
			},
			mockRedis: func(email string, user *model.User) {
				key := baseUserEmail + email
				suite.redis.ExpectGet(key).RedisNil()
				suite.redis.Regexp().ExpectSet(key, `.*`, 30*time.Minute).SetVal("OK")
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
			},
		},
		{
			name:           "success - find from cache",
			email:          "cached@example.com",
			includeDeleted: false,
			setup: func() *model.User {
				suite.cleanupUsers()
				return &model.User{
					ID:           uuid.New(),
					Email:        "cached@example.com",
					PasswordHash: "password",
				}
			},
			mockRedis: func(email string, user *model.User) {
				key := baseUserEmail + email
				userJSON, _ := json.Marshal(user)
				suite.redis.ExpectGet(key).SetVal(string(userJSON))
			},
			wantError: false,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
			},
		},
		{
			name:           "error - email not found",
			email:          "notfound@example.com",
			includeDeleted: false,
			setup: func() *model.User {
				suite.cleanupUsers()
				return nil
			},
			mockRedis: func(email string, user *model.User) {
				key := baseUserEmail + email
				suite.redis.ExpectGet(key).RedisNil()
			},
			wantError:    true,
			expectStatus: 404,
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, result)
				assert.NotNil(t, resp)
				assert.Equal(t, 404, resp.Status)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			user := tt.setup()
			tt.mockRedis(tt.email, user)
			result, resp := suite.repository.FindByEmail(suite.ctx, tt.email, tt.includeDeleted)
			tt.validate(suite.T(), result, resp)
			assert.NoError(suite.T(), suite.redis.ExpectationsWereMet())
		})
	}
}

// Test Update with Table Driven
func (suite *UserRepositoryTestSuite) TestUpdate() {
	tests := []struct {
		name      string
		setup     func() *model.User
		update    func(*model.User) *model.User
		mockRedis func(*model.User)
		validate  func(*testing.T, *model.User, *dto.ServiceResponse)
	}{
		{
			name: "success - update role",
			setup: func() *model.User {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "update@example.com",
					PasswordHash: "oldpassword",
					Role:         model.RoleUser,
				}
				suite.db.Create(user)
				return user
			},
			update: func(user *model.User) *model.User {
				user.Role = model.RoleAdmin
				return user
			},
			mockRedis: func(user *model.User) {
				suite.redis.ExpectDel(baseUser + user.ID.String()).SetVal(1)
				suite.redis.ExpectDel(baseUserEmail + user.ID.String()).SetVal(1)
			},
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.NotNil(t, result)
				assert.Equal(t, model.RoleAdmin, result.Role)
			},
		},
		{
			name: "success - update password",
			setup: func() *model.User {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "updatepass@example.com",
					PasswordHash: "oldpass",
					Role:         model.RoleUser,
				}
				suite.db.Create(user)
				return user
			},
			update: func(user *model.User) *model.User {
				user.PasswordHash = "newhashedpass"
				return user
			},
			mockRedis: func(user *model.User) {
				suite.redis.ExpectDel(baseUser + user.ID.String()).SetVal(1)
				suite.redis.ExpectDel(baseUserEmail + user.ID.String()).SetVal(1)
			},
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, "newhashedpass", result.PasswordHash)
			},
		},
		{
			name: "success - update multiple fields",
			setup: func() *model.User {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "multi@example.com",
					PasswordHash: "pass",
					Role:         model.RoleUser,
				}
				suite.db.Create(user)
				return user
			},
			update: func(user *model.User) *model.User {
				user.Role = model.RoleAdmin
				user.PasswordHash = "newpass"
				return user
			},
			mockRedis: func(user *model.User) {
				suite.redis.ExpectDel(baseUser + user.ID.String()).SetVal(1)
				suite.redis.ExpectDel(baseUserEmail + user.ID.String()).SetVal(1)
			},
			validate: func(t *testing.T, result *model.User, resp *dto.ServiceResponse) {
				assert.Nil(t, resp)
				assert.Equal(t, model.RoleAdmin, result.Role)
				assert.Equal(t, "newpass", result.PasswordHash)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			user := tt.setup()
			updatedUser := tt.update(user)
			tt.mockRedis(updatedUser)
			result, resp := suite.repository.Update(suite.ctx, updatedUser)
			tt.validate(suite.T(), result, resp)
		})
	}
}

// Test Delete with Table Driven
func (suite *UserRepositoryTestSuite) TestDelete() {
	tests := []struct {
		name      string
		setup     func() uuid.UUID
		mockRedis func(uuid.UUID)
		wantError bool
		validate  func(*testing.T, error, uuid.UUID)
	}{
		{
			name: "success - delete existing user",
			setup: func() uuid.UUID {
				suite.cleanupUsers()
				user := &model.User{
					ID:           uuid.New(),
					Email:        "delete@example.com",
					PasswordHash: "password",
				}
				suite.db.Create(user)
				return user.ID
			},
			mockRedis: func(id uuid.UUID) {
				suite.redis.ExpectDel(baseUser + id.String()).SetVal(1)
				suite.redis.ExpectDel(baseUserEmail + id.String()).SetVal(1)
			},
			wantError: false,
			validate: func(t *testing.T, err error, id uuid.UUID) {
				assert.NoError(t, err)
				var deletedUser model.User
				dbErr := suite.db.First(&deletedUser, id).Error
				assert.Error(t, dbErr)
				assert.Equal(t, gorm.ErrRecordNotFound, dbErr)
			},
		},
		{
			name: "success - delete non-existing user (soft delete)",
			setup: func() uuid.UUID {
				suite.cleanupUsers()
				return uuid.New()
			},
			mockRedis: func(id uuid.UUID) {
				suite.redis.ExpectDel(baseUser + id.String()).SetVal(0)
				suite.redis.ExpectDel(baseUserEmail + id.String()).SetVal(0)
			},
			wantError: false,
			validate: func(t *testing.T, err error, id uuid.UUID) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			id := tt.setup()
			tt.mockRedis(id)
			err := suite.repository.Delete(suite.ctx, id)
			tt.validate(suite.T(), err, id)
		})
	}
}

// Test List with Table Driven
func (suite *UserRepositoryTestSuite) TestList() {
	tests := []struct {
		name        string
		setup       func()
		pageRequest *dto.PageRequest
		validate    func(*testing.T, *dto.PageResponse, error)
	}{
		{
			name: "success - list all users",
			setup: func() {
				suite.cleanupUsers()
				users := []*model.User{
					{ID: uuid.New(), Email: "user1@example.com", PasswordHash: "pass", Role: "user"},
					{ID: uuid.New(), Email: "user2@example.com", PasswordHash: "pass", Role: "admin"},
					{ID: uuid.New(), Email: "user3@example.com", PasswordHash: "pass", Role: "user"},
				}
				for _, u := range users {
					suite.db.Create(u)
				}
			},
			pageRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, int64(3), result.Total)
				assert.Len(t, result.Data, 3)
			},
		},
		{
			name: "success - search by email",
			setup: func() {
				suite.cleanupUsers()
				users := []*model.User{
					{ID: uuid.New(), Email: "john@example.com", PasswordHash: "pass"},
					{ID: uuid.New(), Email: "jane@example.com", PasswordHash: "pass"},
					{ID: uuid.New(), Email: "alice@test.com", PasswordHash: "pass"},
				}
				for _, u := range users {
					suite.db.Create(u)
				}
			},
			pageRequest: &dto.PageRequest{
				Page:   1,
				Limit:  10,
				Search: "example",
			},
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(2), result.Total)
				assert.Len(t, result.Data, 2)
			},
		},
		{
			name: "success - filter by role",
			setup: func() {
				suite.cleanupUsers()
				users := []*model.User{
					{ID: uuid.New(), Email: "user1@example.com", PasswordHash: "pass", Role: "user"},
					{ID: uuid.New(), Email: "admin1@example.com", PasswordHash: "pass", Role: "admin"},
					{ID: uuid.New(), Email: "user2@example.com", PasswordHash: "pass", Role: "user"},
				}
				for _, u := range users {
					suite.db.Create(u)
				}
			},
			pageRequest: func() *dto.PageRequest {
				role := model.RoleAdmin
				return &dto.PageRequest{
					Page:  1,
					Limit: 10,
					Role:  &role,
				}
			}(),
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), result.Total)
				assert.Len(t, result.Data, 1)
			},
		},
		{
			name: "success - pagination page 1",
			setup: func() {
				suite.cleanupUsers()
				for i := 0; i < 25; i++ {
					user := &model.User{
						ID:           uuid.New(),
						Email:        fmt.Sprintf("user%d@example.com", i),
						PasswordHash: "pass",
						Role:         "user",
					}
					suite.db.Create(user)
				}
			},
			pageRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(25), result.Total)
				assert.Len(t, result.Data, 10)
			},
		},
		{
			name: "success - pagination page 3",
			setup: func() {
				suite.cleanupUsers()
				for i := 0; i < 25; i++ {
					user := &model.User{
						ID:           uuid.New(),
						Email:        fmt.Sprintf("user%d@test.com", i),
						PasswordHash: "pass",
						Role:         "user",
					}
					suite.db.Create(user)
				}
			},
			pageRequest: &dto.PageRequest{
				Page:  3,
				Limit: 10,
			},
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(25), result.Total)
				assert.Len(t, result.Data, 5)
			},
		},
		{
			name: "success - empty result",
			setup: func() {
				suite.cleanupUsers()
				// No users created
			},
			pageRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(0), result.Total)
				assert.Len(t, result.Data, 0)
			},
		},
		{
			name: "success - search with role filter",
			setup: func() {
				suite.cleanupUsers()
				users := []*model.User{
					{ID: uuid.New(), Email: "admin1@example.com", PasswordHash: "pass", Role: "admin"},
					{ID: uuid.New(), Email: "admin2@test.com", PasswordHash: "pass", Role: "admin"},
					{ID: uuid.New(), Email: "user1@example.com", PasswordHash: "pass", Role: "user"},
				}
				for _, u := range users {
					suite.db.Create(u)
				}
			},
			pageRequest: func() *dto.PageRequest {
				role := model.RoleAdmin
				return &dto.PageRequest{
					Page:   1,
					Limit:  10,
					Search: "example",
					Role:   &role,
				}
			}(),
			validate: func(t *testing.T, result *dto.PageResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), result.Total)
				assert.Len(t, result.Data, 1)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setup()
			result, err := suite.repository.List(suite.ctx, tt.pageRequest)
			tt.validate(suite.T(), result, err)
		})
	}
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
