package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/agris/user-service/internal/dto"
	models "github.com/agris/user-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository implements repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, *dto.ServiceResponse) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*models.User), nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string, delete bool) (*models.User, *dto.ServiceResponse) {
	args := m.Called(ctx, email, delete)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*models.User), nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, *dto.ServiceResponse) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*dto.ServiceResponse)
	}
	return args.Get(0).(*models.User), nil
}

func (m *MockUserRepository) List(ctx context.Context, pageRequest *dto.PageRequest) (*dto.PageResponse, error) {
	args := m.Called(ctx, pageRequest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PageResponse), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// === TESTS ===

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name             string
		request          *dto.UserRequest
		existingUser     *models.User
		findByEmailErr   error
		createdUser      *models.User
		createErr        error
		shouldCallFindBy bool
		shouldCallCreate bool
		expectedErr      string
		expectSuccess    bool
	}{
		{
			name: "Success - Create New User",
			request: &dto.UserRequest{
				Name:         "John Doe",
				Email:        "john@example.com",
				PasswordHash: "password123",
			},
			existingUser:   nil,
			findByEmailErr: errors.New("not found"),
			createdUser: &models.User{
				ID:           uuid.New(),
				Name:         "John Doe",
				Email:        "john@example.com",
				PasswordHash: "hashedpassword",
				Role:         models.RoleUser,
				CreatedAt:    time.Now(),
			},
			createErr:        nil,
			shouldCallFindBy: true,
			shouldCallCreate: true,
			expectedErr:      "",
			expectSuccess:    true,
		},
		{
			name:             "Error - Nil Request",
			request:          nil,
			shouldCallFindBy: false,
			shouldCallCreate: false,
			expectedErr:      ErrInvalidData,
			expectSuccess:    false,
		},
		{
			name: "Error - Empty Email",
			request: &dto.UserRequest{
				Name:         "John Doe",
				Email:        "",
				PasswordHash: "password123",
			},
			shouldCallFindBy: false,
			shouldCallCreate: false,
			expectedErr:      ErrInvalidData,
			expectSuccess:    false,
		},
		{
			name: "Error - Empty Password",
			request: &dto.UserRequest{
				Name:         "John Doe",
				Email:        "john@example.com",
				PasswordHash: "",
			},
			shouldCallFindBy: false,
			shouldCallCreate: false,
			expectedErr:      ErrInvalidData,
			expectSuccess:    false,
		},
		{
			name: "Error - Empty Name",
			request: &dto.UserRequest{
				Name:         "",
				Email:        "john@example.com",
				PasswordHash: "password123",
			},
			shouldCallFindBy: false,
			shouldCallCreate: false,
			expectedErr:      ErrInvalidData,
			expectSuccess:    false,
		},
		{
			name: "Error - Email Already Exists",
			request: &dto.UserRequest{
				Name:         "John Doe",
				Email:        "existing@example.com",
				PasswordHash: "password123",
			},
			existingUser: &models.User{
				ID:    uuid.New(),
				Email: "existing@example.com",
			},
			findByEmailErr:   nil,
			shouldCallFindBy: true,
			shouldCallCreate: false,
			expectedErr:      ErrEmailExists,
			expectSuccess:    false,
		},
		{
			name: "Error - Database Create Fails",
			request: &dto.UserRequest{
				Name:         "John Doe",
				Email:        "john@example.com",
				PasswordHash: "password123",
			},
			existingUser:     nil,
			findByEmailErr:   errors.New("not found"),
			createdUser:      nil,
			createErr:        errors.New("database error"),
			shouldCallFindBy: true,
			shouldCallCreate: true,
			expectedErr:      "database error",
			expectSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			if tt.shouldCallFindBy {
				if tt.existingUser != nil {
					mockRepo.On("FindByEmail", mock.Anything, mock.Anything, true).
						Return(tt.existingUser, (*dto.ServiceResponse)(nil))
				} else {
					mockRepo.On("FindByEmail", mock.Anything, mock.Anything, true).
						Return(nil, &dto.ServiceResponse{Err: tt.findByEmailErr})
				}
			}

			if tt.shouldCallCreate {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
					return u.Email == tt.request.Email && u.Name == tt.request.Name
				})).Return(tt.createdUser, tt.createErr)
			}

			result, err := service.CreateUser(context.Background(), tt.request)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.createdUser.ID, result.Id)
				assert.Equal(t, tt.createdUser.Email, result.Email)
				assert.Equal(t, tt.createdUser.Name, result.Name)
				assert.Equal(t, tt.createdUser.Role, result.Role)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErr)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	validUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		userID         uuid.UUID
		mockUser       *models.User
		mockResponse   *dto.ServiceResponse
		shouldCallRepo bool
		expectedErr    string
		expectedStatus string
	}{
		{
			name:   "Success - Active User",
			userID: validUUID,
			mockUser: &models.User{
				ID:        validUUID,
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: now,
			},
			mockResponse:   nil,
			shouldCallRepo: true,
			expectedErr:    "",
			expectedStatus: "ACTIVE",
		},
		{
			name:           "Error - Nil UUID",
			userID:         uuid.Nil,
			mockUser:       nil,
			mockResponse:   nil,
			shouldCallRepo: false,
			expectedErr:    ErrInvalidData,
			expectedStatus: "",
		},
		{
			name:     "Error - User Not Found",
			userID:   validUUID,
			mockUser: nil,
			mockResponse: &dto.ServiceResponse{
				Status: http.StatusNotFound,
				Err:    errors.New(ErrUserNotFound),
			},
			shouldCallRepo: true,
			expectedErr:    ErrUserNotFound,
			expectedStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("FindByID", mock.Anything, tt.userID).
					Return(tt.mockUser, tt.mockResponse)
			}

			result, response := service.GetCurrentUser(context.Background(), tt.userID)

			if tt.expectedErr != "" {
				assert.Nil(t, result)
				assert.NotNil(t, response)
				assert.Contains(t, response.Err.Error(), tt.expectedErr)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, response)
				assert.Equal(t, tt.mockUser.ID, result.Id)
				assert.Equal(t, tt.mockUser.Email, result.Email)
				assert.Equal(t, tt.mockUser.Name, result.Name)
				if tt.expectedStatus != "" {
					assert.Equal(t, tt.expectedStatus, string(result.Status))
				}
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "FindByID")
			}
		})
	}
}

func TestUpdateUserRole(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	now := time.Now()

	tests := []struct {
		name             string
		request          *dto.UpdateRoleRequest
		mockUser         *models.User
		mockFindErr      error
		mockUpdatedUser  *models.User
		mockUpdateResp   *dto.ServiceResponse
		shouldCallFind   bool
		shouldCallUpdate bool
		expectedErr      string
		expectedStatus   int
	}{
		{
			name: "Success - Update Role",
			request: &dto.UpdateRoleRequest{
				UserId:    userID,
				AccountId: accountID,
				Role:      models.RoleAdmin,
			},
			mockUser: &models.User{
				ID:    userID,
				Email: "user@example.com",
				Role:  models.RoleUser,
			},
			mockFindErr: nil,
			mockUpdatedUser: &models.User{
				ID:        userID,
				Email:     "user@example.com",
				Role:      models.RoleAdmin,
				UpdatedAt: now,
			},
			mockUpdateResp:   nil,
			shouldCallFind:   true,
			shouldCallUpdate: true,
			expectedErr:      "",
			expectedStatus:   0,
		},
		{
			name: "Error - Cannot Update Own Account",
			request: &dto.UpdateRoleRequest{
				UserId:    userID,
				AccountId: userID,
				Role:      models.RoleAdmin,
			},
			shouldCallFind:   false,
			shouldCallUpdate: false,
			expectedErr:      ErrCantUpdateOwnAccount,
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name: "Error - User Not Found",
			request: &dto.UpdateRoleRequest{
				UserId:    userID,
				AccountId: accountID,
				Role:      models.RoleAdmin,
			},
			mockUser:         nil,
			mockFindErr:      errors.New("not found"),
			shouldCallFind:   true,
			shouldCallUpdate: false,
			expectedErr:      ErrUserNotFound,
			expectedStatus:   http.StatusNotFound,
		},
		{
			name: "Error - Update Fails",
			request: &dto.UpdateRoleRequest{
				UserId:    userID,
				AccountId: userID,
				Role:      models.RoleAdmin,
			},
			mockUser: &models.User{
				ID:    userID,
				Email: "user@example.com",
				Role:  models.RoleUser,
			},
			mockFindErr:     nil,
			mockUpdatedUser: nil,
			mockUpdateResp: &dto.ServiceResponse{
				Status: http.StatusBadRequest,
				Err:    errors.New(ErrCantUpdateOwnAccount),
			},
			shouldCallFind:   false,
			shouldCallUpdate: false,
			expectedErr:      ErrCantUpdateOwnAccount,
			expectedStatus:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			if tt.shouldCallFind {
				if tt.mockFindErr != nil {
					mockRepo.On("FindByID", mock.Anything, tt.request.UserId).
						Return(tt.mockUser, &dto.ServiceResponse{Err: tt.mockFindErr})
				} else {
					mockRepo.On("FindByID", mock.Anything, tt.request.UserId).
						Return(tt.mockUser, (*dto.ServiceResponse)(nil))
				}
			}

			if tt.shouldCallUpdate {
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
					return u.ID == tt.request.UserId && u.Role == tt.request.Role
				})).Return(tt.mockUpdatedUser, tt.mockUpdateResp)
			}

			result, response := service.UpdateUserRole(context.Background(), tt.request)

			if tt.expectedErr != "" {
				assert.Nil(t, result)
				assert.NotNil(t, response)
				assert.Contains(t, response.Err.Error(), tt.expectedErr)
				assert.Equal(t, tt.expectedStatus, response.Status)
			} else {
				assert.NotNil(t, result)
				assert.Nil(t, response)
				assert.Equal(t, tt.mockUpdatedUser.ID, result.Id)
				assert.Equal(t, tt.mockUpdatedUser.Email, result.Email)
				assert.Equal(t, tt.mockUpdatedUser.Role, result.Role)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name             string
		request          *dto.PageRequest
		expectedRequest  *dto.PageRequest
		mockPageResponse *dto.PageResponse
		mockErr          error
		shouldCallRepo   bool
		expectedErr      string
	}{
		{
			name: "Success - Valid Request",
			request: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			expectedRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			mockPageResponse: &dto.PageResponse{
				Total: 50,
				Data:  []interface{}{},
			},
			mockErr:        nil,
			shouldCallRepo: true,
			expectedErr:    "",
		},
		{
			name:             "Error - Nil Request",
			request:          nil,
			expectedRequest:  nil,
			mockPageResponse: nil,
			mockErr:          nil,
			shouldCallRepo:   false,
			expectedErr:      ErrInvalidData,
		},
		{
			name: "Normalize Page - Less Than 1",
			request: &dto.PageRequest{
				Page:  0,
				Limit: 10,
			},
			expectedRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			mockPageResponse: &dto.PageResponse{Total: 50},
			mockErr:          nil,
			shouldCallRepo:   true,
			expectedErr:      "",
		},
		{
			name: "Normalize Limit - Less Than 1",
			request: &dto.PageRequest{
				Page:  1,
				Limit: 0,
			},
			expectedRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			mockPageResponse: &dto.PageResponse{Total: 50},
			mockErr:          nil,
			shouldCallRepo:   true,
			expectedErr:      "",
		},
		{
			name: "Normalize Limit - Greater Than 100",
			request: &dto.PageRequest{
				Page:  1,
				Limit: 150,
			},
			expectedRequest: &dto.PageRequest{
				Page:  1,
				Limit: 100,
			},
			mockPageResponse: &dto.PageResponse{Total: 50},
			mockErr:          nil,
			shouldCallRepo:   true,
			expectedErr:      "",
		},
		{
			name: "Error - Repository Error",
			request: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			expectedRequest: &dto.PageRequest{
				Page:  1,
				Limit: 10,
			},
			mockPageResponse: nil,
			mockErr:          errors.New("database error"),
			shouldCallRepo:   true,
			expectedErr:      "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo)

			if tt.shouldCallRepo {
				mockRepo.On("List", mock.Anything, mock.MatchedBy(func(req *dto.PageRequest) bool {
					return req.Page == tt.expectedRequest.Page && req.Limit == tt.expectedRequest.Limit
				})).Return(tt.mockPageResponse, tt.mockErr)
			}

			result, err := service.ListUsers(context.Background(), tt.request)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockPageResponse.Total, result.Total)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "List")
			}
		})
	}
}

// Helper test to verify password hashing
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	assert.NoError(t, err)
	assert.NotEqual(t, password, string(hashedPassword))

	// Verify hash
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	assert.NoError(t, err)
}
