package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/agris/user-service/internal/dto"
	"github.com/agris/user-service/internal/model"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, *dto.ServiceResponse)
	FindByEmail(ctx context.Context, email string, delete bool) (*model.User, *dto.ServiceResponse)
	Update(ctx context.Context, user *model.User) (*model.User, *dto.ServiceResponse)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, pageRequest *dto.PageRequest) (*dto.PageResponse, error)
}

type userRepository struct {
	db *gorm.DB
	rd *redis.Client
}

func NewUserRepository(db *gorm.DB, rd *redis.Client) UserRepository {
	return &userRepository{db: db, rd: rd}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, *dto.ServiceResponse) {
	// redis
	key := baseUser + id.String()
	cached, err := r.rd.Get(ctx, key).Bytes()
	if err == nil {
		var user model.User
		er := json.Unmarshal(cached, &user)
		if er == nil {
			log.Info("Using data of redis")
			return &user, nil
		}
	}

	// query
	var user model.User
	errSelect := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if errSelect != nil {
		log.Error(errSelect.Error())
		response := dto.ServiceResponse{
			Status: 404,
			Err:    ErrNotFound,
		}
		return nil, &response
	}

	// save to redis
	_ = r.rd.Set(ctx, key, user, 30*time.Minute)

	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string, delete bool) (*model.User, *dto.ServiceResponse) {

	// redis
	key := baseUserEmail + email
	cached, err := r.rd.Get(ctx, key).Bytes()
	if err == nil {
		var user model.User
		if er := json.Unmarshal(cached, &user); er == nil {
			log.Info("Using data of redis")
			return &user, nil
		}
	}

	// query
	var user model.User
	query := r.db.WithContext(ctx)
	if delete {
		query = query.Unscoped()
	}
	errSelect := query.Where("email = ? ", email).First(&user).Error
	if errSelect != nil {
		log.Error(errSelect.Error())
		response := dto.ServiceResponse{
			Status: 404,
			Err:    ErrNotFound,
		}
		return nil, &response
	}

	// save to redis
	_ = r.rd.Set(ctx, key, user, 30*time.Minute)

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) (*model.User, *dto.ServiceResponse) {

	// Update và tự động trả về dữ liệu mới nhất (PostgreSQL)
	if err := r.db.WithContext(ctx).Clauses(clause.Returning{}).Save(user).Error; err != nil {
		response := dto.ServiceResponse{
			Status: http.StatusInternalServerError,
			Err:    ErrInternalServer,
		}
		return nil, &response
	}

	// delete redis
	_ = r.rd.Del(ctx, baseUser+user.ID.String())
	_ = r.rd.Del(ctx, baseUserEmail+user.ID.String())

	return user, nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&model.User{}, id).Error
	if err != nil {
		log.Error("[ERROR] : [USERREPOSITORY} : 106 : " + err.Error())
		return err
	}

	// delete redis
	_ = r.rd.Del(ctx, baseUser+id.String())
	_ = r.rd.Del(ctx, baseUserEmail+id.String())
	return nil
}

func (r *userRepository) List(ctx context.Context, pageRequest *dto.PageRequest) (*dto.PageResponse, error) {
	var users []*model.User
	var total int64

	query := r.db.Model(&model.User{})

	if pageRequest.Search != "" {
		if r.db.Dialector.Name() == "postgres" {
			query = query.Where("email ILIKE ?", "%"+pageRequest.Search+"%")
		} else {
			// SQLite, MySQL
			query = query.Where("email LIKE ?", "%"+pageRequest.Search+"%")
		}
	}

	if pageRequest.Role != nil {
		query = query.Where("role = ?", *pageRequest.Role)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get paginated results
	err := query.WithContext(ctx).Offset((pageRequest.Page - 1) * pageRequest.Limit).Limit(pageRequest.Limit).Order("created_at DESC").Find(&users).Error

	response := dto.PageResponse{
		Total:  total,
		Data:   users,
		Filter: pageRequest,
	}
	return &response, err
}
