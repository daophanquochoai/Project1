# User Service Implementation Guide

**Service:** User Service  
**HTTP Port:** 8005  
**gRPC Port:** 9005  
**Database:** PostgreSQL (user_service)  
**Framework:** Fiber v2.x  
**ORM:** GORM  
**Authentication:** JWT

---

## 📋 Mục lục

1. [Tổng quan](#tổng-quan)
2. [Cấu trúc Project](#cấu-trúc-project)
3. [Setup & Dependencies](#setup--dependencies)
4. [Implementation Steps](#implementation-steps)
5. [API Endpoints](#api-endpoints)
6. [gRPC Service](#grpc-service)
7. [Testing](#testing)
8. [Deployment](#deployment)

---

## 🎯 Tổng quan

User Service là microservice đầu tiên trong hệ thống, chịu trách nhiệm:

-   Quản lý người dùng (User & Admin)
-   Authentication & Authorization (JWT)
-   gRPC service để validate tokens cho các services khác
-   Admin endpoints để quản lý users

### Requirements Summary

**Business Requirements:**

-   BR-001: User registration với email và password
-   BR-002: JWT-based authentication
-   BR-003: Phân biệt User và Admin roles
-   BR-004: Admin có thể xem danh sách users và thay đổi phân quyền
-   BR-005: Admin không thể tự thay đổi role của chính mình

**Technical Requirements:**

-   TC-001: Port 8005 HTTP, Port 9005 gRPC
-   TC-002: Expose gRPC endpoint Authenticate()
-   NFR-S01: JWT với expiry 1h, bcrypt min cost 10
-   NFR-S02: RBAC middleware validation

---

## 📁 Cấu trúc Project

```
user-service/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── database/
│   │   └── postgres.go             # Database connection
│   ├── models/
│   │   └── user.go                 # User model
│   ├── repository/
│   │   └── user_repository.go      # Data access layer
│   ├── service/
│   │   ├── auth_service.go         # Authentication logic
│   │   └── user_service.go         # User business logic
│   ├── handler/
│   │   ├── auth_handler.go         # Auth HTTP handlers
│   │   ├── user_handler.go         # User HTTP handlers
│   │   └── admin_handler.go        # Admin HTTP handlers
│   ├── middleware/
│   │   ├── auth.go                 # JWT authentication middleware
│   │   └── admin.go                # Admin authorization middleware
│   ├── grpc/
│   │   ├── auth.proto              # gRPC proto definition
│   │   ├── auth.pb.go              # Generated code
│   │   └── auth_server.go          # gRPC server implementation
│   └── utils/
│       ├── jwt.go                  # JWT utilities
│       ├── password.go             # Password hashing utilities
│       └── validator.go            # Input validation
├── pkg/
│   └── errors/
│       └── errors.go               # Error definitions
├── wire/
│   └── wire.go                     # Wire providers and injectors
├── migrations/
│   └── 001_create_users.up.sql     # Database migrations (optional)
├── tests/
│   ├── handler/
│   ├── service/
│   └── repository/
├── .env.example                    # Environment variables template
├── config.yaml.example             # YAML config template
├── .gitignore
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

---

## 🚀 Setup & Dependencies

### 1. Initialize Go Module

```bash
mkdir user-service
cd user-service
go mod init github.com/agris/user-service
```

### 2. Install Dependencies

```bash
# HTTP Framework
go get github.com/gofiber/fiber/v2

# Database
go get gorm.io/gorm
go get gorm.io/driver/postgres

# JWT
go get github.com/golang-jwt/jwt/v5

# Password Hashing
go get golang.org/x/crypto/bcrypt

# Configuration Management
go get github.com/spf13/viper

# Validation
go get github.com/go-playground/validator/v10

# gRPC
go get google.golang.org/grpc
go get google.golang.org/protobuf/proto
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# UUID
go get github.com/google/uuid

# Logger
go get github.com/gofiber/fiber/v2/middleware/logger
go get github.com/gofiber/fiber/v2/middleware/recover

# Dependency Injection
go get github.com/google/wire/cmd/wire
go get github.com/google/wire
```

### 3. Create Configuration Files

#### Option 1: YAML Config File (`config.yaml.example`)

```yaml
server:
    http_port: "8005"
    grpc_port: "9005"
    env: "development"

database:
    host: "localhost"
    port: "5432"
    user: "user_service"
    password: "user_service_pwd"
    name: "user_service"
    sslmode: "disable"
    timezone: "UTC"

jwt:
    secret: "your-super-secret-jwt-key-change-in-production-min-32-chars"
    expiry: "1h"

logging:
    level: "info"
```

#### Option 2: Environment Variables (`.env.example`)

```env
# Server Configuration
USER_SERVICE_SERVER_HTTP_PORT=8005
USER_SERVICE_SERVER_GRPC_PORT=9005
USER_SERVICE_SERVER_ENV=development

# Database Configuration
USER_SERVICE_DATABASE_HOST=localhost
USER_SERVICE_DATABASE_PORT=5432
USER_SERVICE_DATABASE_USER=user_service
USER_SERVICE_DATABASE_PASSWORD=user_service_pwd
USER_SERVICE_DATABASE_NAME=user_service
USER_SERVICE_DATABASE_SSLMODE=disable
USER_SERVICE_DATABASE_TIMEZONE=UTC

# JWT Configuration
USER_SERVICE_JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
USER_SERVICE_JWT_EXPIRY=1h

# Or use shorter names (need to bind explicitly)
HTTP_PORT=8005
GRPC_PORT=9005
DB_HOST=localhost
DB_PORT=5432
DB_USER=user_service
DB_PASSWORD=user_service_pwd
DB_NAME=user_service
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
JWT_EXPIRY=1h
```

**Note:** Code đã bind các env vars với tên ngắn hơn trong `Load()` function, nên bạn có thể dùng cả hai cách:

-   `HTTP_PORT` hoặc `USER_SERVICE_SERVER_HTTP_PORT`
-   `DB_HOST` hoặc `USER_SERVICE_DATABASE_HOST`
-   `JWT_SECRET` hoặc `USER_SERVICE_JWT_SECRET`

### 4. Create `.gitignore`

```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
user-service

# Test binary
*.test

# Output
*.out

# Go workspace file
go.work

# Environment & Config
.env
.env.local
config.yaml
config.local.yaml

# IDE
.idea/
.vscode/
*.swp
*.swo

# Generated files
*.pb.go
*.pb.gw.go
wire_gen.go
```

---

## 🔨 Implementation Steps

### Step 1: Configuration (`internal/config/config.go`)

```go
package config

import (
    "fmt"
    "time"

    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    HTTPPort string
    GRPCPort string
    Env      string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
    TimeZone string
}

type JWTConfig struct {
    Secret string
    Expiry time.Duration
}

func Load() (*Config, error) {
    viper.SetConfigName("config")        // name of config file (without extension)
    viper.SetConfigType("yaml")          // or "json", "toml", "env"
    viper.AddConfigPath(".")             // look for config in current directory
    viper.AddConfigPath("./config")       // optionally look for config in config directory
    viper.AddConfigPath("$HOME/.user-service") // call multiple times to add many search paths

    // Environment variables - support both prefixed and short names
    viper.SetEnvPrefix("USER_SERVICE")    // will be uppercased automatically
    viper.AutomaticEnv()                  // read in environment variables that match

    // Bind environment variables with shorter names (optional)
    // This allows using HTTP_PORT instead of USER_SERVICE_SERVER_HTTP_PORT
    viper.BindEnv("server.http_port", "HTTP_PORT")
    viper.BindEnv("server.grpc_port", "GRPC_PORT")
    viper.BindEnv("server.env", "ENV")
    viper.BindEnv("database.host", "DB_HOST")
    viper.BindEnv("database.port", "DB_PORT")
    viper.BindEnv("database.user", "DB_USER")
    viper.BindEnv("database.password", "DB_PASSWORD")
    viper.BindEnv("database.name", "DB_NAME")
    viper.BindEnv("database.sslmode", "DB_SSLMODE")
    viper.BindEnv("database.timezone", "DB_TIMEZONE")
    viper.BindEnv("jwt.secret", "JWT_SECRET")
    viper.BindEnv("jwt.expiry", "JWT_EXPIRY")

    // Set defaults
    setDefaults()

    // Try to read config file (optional - will use defaults if not found)
    if err := viper.ReadInConfig(); err != nil {
        // Config file not found; use defaults and environment variables
        fmt.Printf("Config file not found, using defaults and environment variables: %v\n", err)
    } else {
        fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
    }

    // Parse JWT expiry duration
    jwtExpiryStr := viper.GetString("jwt.expiry")
    if jwtExpiryStr == "" {
        jwtExpiryStr = "1h"
    }
    jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
    if err != nil {
        jwtExpiry = time.Hour // default to 1 hour
    }

    cfg := &Config{
        Server: ServerConfig{
            HTTPPort: viper.GetString("server.http_port"),
            GRPCPort: viper.GetString("server.grpc_port"),
            Env:      viper.GetString("server.env"),
        },
        Database: DatabaseConfig{
            Host:     viper.GetString("database.host"),
            Port:     viper.GetString("database.port"),
            User:     viper.GetString("database.user"),
            Password: viper.GetString("database.password"),
            Name:     viper.GetString("database.name"),
            SSLMode:  viper.GetString("database.sslmode"),
            TimeZone: viper.GetString("database.timezone"),
        },
        JWT: JWTConfig{
            Secret: viper.GetString("jwt.secret"),
            Expiry: jwtExpiry,
        },
    }

    // Validate required fields
    if err := validateConfig(cfg); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return cfg, nil
}

func setDefaults() {
    // Server defaults
    viper.SetDefault("server.http_port", "8005")
    viper.SetDefault("server.grpc_port", "9005")
    viper.SetDefault("server.env", "development")

    // Database defaults
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", "5432")
    viper.SetDefault("database.user", "user_service")
    viper.SetDefault("database.password", "user_service_pwd")
    viper.SetDefault("database.name", "user_service")
    viper.SetDefault("database.sslmode", "disable")
    viper.SetDefault("database.timezone", "UTC")

    // JWT defaults
    viper.SetDefault("jwt.secret", "default-secret-change-in-production")
    viper.SetDefault("jwt.expiry", "1h")
}

func validateConfig(cfg *Config) error {
    if cfg.JWT.Secret == "default-secret-change-in-production" && cfg.Server.Env == "production" {
        return fmt.Errorf("JWT_SECRET must be changed in production")
    }

    if cfg.JWT.Secret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }

    if len(cfg.JWT.Secret) < 32 {
        return fmt.Errorf("JWT_SECRET must be at least 32 characters")
    }

    return nil
}
```

**Viper Features:**

1. **Multiple Config Sources** (theo thứ tự ưu tiên):

    - Explicit calls to `viper.Set()`
    - Flags
    - Environment variables (có prefix `USER_SERVICE_` hoặc bind custom names)
    - Config file (yaml/json/toml/env)
    - Default values

2. **Watch Config Files** (optional - reload config khi file thay đổi):

```go
// In main.go, after Load()
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    fmt.Println("Config file changed:", e.Name)
    // Reload config or handle change
})
```

3. **Environment Variables Mapping:**

    - `USER_SERVICE_SERVER_HTTP_PORT` → `server.http_port`
    - `HTTP_PORT` → `server.http_port` (via BindEnv)
    - `DB_HOST` → `database.host` (via BindEnv)
    - `JWT_SECRET` → `jwt.secret` (via BindEnv)

4. **Unmarshal to Struct** (alternative approach):

```go
var cfg Config
if err := viper.Unmarshal(&cfg); err != nil {
    return nil, err
}
// Requires struct tags: `mapstructure:"server.http_port"`
```

### Step 2: Database Connection (`internal/database/postgres.go`)

```go
package database

import (
    "fmt"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/agris/user-service/internal/config"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
        cfg.Database.Host,
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Name,
        cfg.Database.Port,
        cfg.Database.SSLMode,
        cfg.Database.TimeZone,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return db, nil
}
```

### Step 3: User Model (`internal/models/user.go`)

```go
package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Role string

const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
)

type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
    Role         Role      `gorm:"type:varchar(20);default:'user';index"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
    return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    return nil
}
```

### Step 4: Repository Layer (`internal/repository/user_repository.go`)

```go
package repository

import (
    "github.com/agris/user-service/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uuid.UUID) (*models.User, error)
    FindByEmail(email string) (*models.User, error)
    Update(user *models.User) error
    Delete(id uuid.UUID) error
    List(offset, limit int, search string, role *models.Role) ([]*models.User, int64, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.Where("id = ?", id).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) List(offset, limit int, search string, role *models.Role) ([]*models.User, int64, error) {
    var users []*models.User
    var total int64

    query := r.db.Model(&models.User{})

    if search != "" {
        query = query.Where("email ILIKE ?", "%"+search+"%")
    }

    if role != nil {
        query = query.Where("role = ?", *role)
    }

    // Count total
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Get paginated results
    err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
    return users, total, err
}
```

### Step 5: Password Utilities (`internal/utils/password.go`)

```go
package utils

import (
    "golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Step 6: JWT Utilities (`internal/utils/jwt.go`)

```go
package utils

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    Role   string    `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, email, role, secret string, expiry time.Duration) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

### Step 7: Auth Service (`internal/service/auth_service.go`)

```go
package service

import (
    "errors"
    "time"

    "github.com/agris/user-service/internal/config"
    "github.com/agris/user-service/internal/models"
    "github.com/agris/user-service/internal/repository"
    "github.com/agris/user-service/internal/utils"
    "github.com/google/uuid"
)

type AuthService interface {
    Register(email, password string) (*models.User, string, error)
    Login(email, password string) (*models.User, string, error)
    ValidateToken(token string) (*models.User, error)
}

type authService struct {
    userRepo repository.UserRepository
    config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
    return &authService{
        userRepo: userRepo,
        config:   cfg,
    }
}

func (s *authService) Register(email, password string) (*models.User, string, error) {
    // Check if user exists
    existing, _ := s.userRepo.FindByEmail(email)
    if existing != nil {
        return nil, "", errors.New("email already registered")
    }

    // Hash password
    passwordHash, err := utils.HashPassword(password)
    if err != nil {
        return nil, "", err
    }

    // Create user
    user := &models.User{
        Email:        email,
        PasswordHash: passwordHash,
        Role:         models.RoleUser,
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, "", err
    }

    // Generate token
    token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.config.JWT.Secret, s.config.JWT.Expiry)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}

func (s *authService) Login(email, password string) (*models.User, string, error) {
    // Find user
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        return nil, "", errors.New("invalid email or password")
    }

    // Check password
    if !utils.CheckPasswordHash(password, user.PasswordHash) {
        return nil, "", errors.New("invalid email or password")
    }

    // Generate token
    token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.config.JWT.Secret, s.config.JWT.Expiry)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}

func (s *authService) ValidateToken(token string) (*models.User, error) {
    claims, err := utils.ValidateToken(token, s.config.JWT.Secret)
    if err != nil {
        return nil, err
    }

    user, err := s.userRepo.FindByID(claims.UserID)
    if err != nil {
        return nil, errors.New("user not found")
    }

    return user, nil
}
```

### Step 8: User Service (`internal/service/user_service.go`)

```go
package service

import (
    "errors"

    "github.com/agris/user-service/internal/models"
    "github.com/agris/user-service/internal/repository"
    "github.com/google/uuid"
)

type UserService interface {
    GetCurrentUser(userID uuid.UUID) (*models.User, error)
    ListUsers(offset, limit int, search string, role *models.Role) ([]*models.User, int64, error)
    UpdateUserRole(userID uuid.UUID, newRole models.Role, adminID uuid.UUID) error
}

type userService struct {
    userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
    return &userService{userRepo: userRepo}
}

func (s *userService) GetCurrentUser(userID uuid.UUID) (*models.User, error) {
    return s.userRepo.FindByID(userID)
}

func (s *userService) ListUsers(offset, limit int, search string, role *models.Role) ([]*models.User, int64, error) {
    return s.userRepo.List(offset, limit, search, role)
}

func (s *userService) UpdateUserRole(userID uuid.UUID, newRole models.Role, adminID uuid.UUID) error {
    // Prevent admin from changing own role
    if userID == adminID {
        return errors.New("cannot change your own role")
    }

    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return errors.New("user not found")
    }

    user.Role = newRole
    return s.userRepo.Update(user)
}
```

### Step 9: Middleware (`internal/middleware/auth.go`)

```go
package middleware

import (
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/agris/user-service/internal/config"
    "github.com/agris/user-service/internal/utils"
)

func AuthMiddleware(cfg *config.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "missing authorization header",
            })
        }

        // Extract token from "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid authorization header format",
            })
        }

        token := parts[1]
        claims, err := utils.ValidateToken(token, cfg.JWT.Secret)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid or expired token",
            })
        }

        // Store user info in context
        c.Locals("user_id", claims.UserID)
        c.Locals("email", claims.Email)
        c.Locals("role", claims.Role)

        return c.Next()
    }
}

func AdminMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        role := c.Locals("role")
        if role != "admin" {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "error": "admin access required",
            })
        }
        return c.Next()
    }
}
```

### Step 10: HTTP Handlers

#### Auth Handler (`internal/handler/auth_handler.go`)

```go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/agris/user-service/internal/service"
)

type AuthHandler struct {
    authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
    User  interface{} `json:"user"`
    Token string      `json:"token"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
    var req RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    user, token, err := h.authService.Register(req.Email, req.Password)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(AuthResponse{
        User: fiber.Map{
            "id":    user.ID,
            "email": user.Email,
            "role":  user.Role,
        },
        Token: token,
    })
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    user, token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(AuthResponse{
        User: fiber.Map{
            "id":    user.ID,
            "email": user.Email,
            "role":  user.Role,
        },
        Token: token,
    })
}
```

#### User Handler (`internal/handler/user_handler.go`)

```go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/agris/user-service/internal/service"
)

type UserHandler struct {
    userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(uuid.UUID)

    user, err := h.userService.GetCurrentUser(userID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "user not found",
        })
    }

    return c.JSON(fiber.Map{
        "id":    user.ID,
        "email": user.Email,
        "role":  user.Role,
    })
}
```

#### Admin Handler (`internal/handler/admin_handler.go`)

```go
package handler

import (
    "strconv"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/agris/user-service/internal/models"
    "github.com/agris/user-service/internal/service"
)

type AdminHandler struct {
    userService service.UserService
}

func NewAdminHandler(userService service.UserService) *AdminHandler {
    return &AdminHandler{userService: userService}
}

type UpdateRoleRequest struct {
    Role string `json:"role" validate:"required,oneof=user admin"`
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
    // Pagination
    page, _ := strconv.Atoi(c.Query("page", "1"))
    limit, _ := strconv.Atoi(c.Query("limit", "20"))
    offset := (page - 1) * limit

    // Search & Filter
    search := c.Query("search", "")
    roleStr := c.Query("role", "")

    var role *models.Role
    if roleStr != "" {
        r := models.Role(roleStr)
        role = &r
    }

    users, total, err := h.userService.ListUsers(offset, limit, search, role)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to fetch users",
        })
    }

    return c.JSON(fiber.Map{
        "data": users,
        "pagination": fiber.Map{
            "page":  page,
            "limit": limit,
            "total": total,
        },
    })
}

func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
    userID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid user id",
        })
    }

    adminID := c.Locals("user_id").(uuid.UUID)

    var req UpdateRoleRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invalid request body",
        })
    }

    err = h.userService.UpdateUserRole(userID, models.Role(req.Role), adminID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "user role updated successfully",
    })
}
```

### Step 11: gRPC Service

#### Proto Definition (`internal/grpc/auth.proto`)

```protobuf
syntax = "proto3";

package auth;

option go_package = "github.com/agris/user-service/internal/grpc";

service AuthService {
  rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse);
}

message AuthenticateRequest {
  string token = 1;
}

message AuthenticateResponse {
  bool valid = 1;
  string user_id = 2;
  string email = 3;
  string role = 4;
  string error = 5;
}
```

#### Generate gRPC Code

```bash
cd internal/grpc
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       auth.proto
```

#### gRPC Server (`internal/grpc/auth_server.go`)

```go
package grpc

import (
    "context"

    "google.golang.org/grpc"
    "github.com/agris/user-service/internal/config"
    "github.com/agris/user-service/internal/service"
)

type AuthServer struct {
    authService service.AuthService
}

func NewAuthServer(authService service.AuthService) *AuthServer {
    return &AuthServer{authService: authService}
}

func (s *AuthServer) Authenticate(ctx context.Context, req *AuthenticateRequest) (*AuthenticateResponse, error) {
    user, err := s.authService.ValidateToken(req.Token)
    if err != nil {
        return &AuthenticateResponse{
            Valid: false,
            Error: err.Error(),
        }, nil
    }

    return &AuthenticateResponse{
        Valid:  true,
        UserId: user.ID.String(),
        Email:  user.Email,
        Role:   string(user.Role),
    }, nil
}

func RegisterAuthServer(s *grpc.Server, authService service.AuthService) {
    RegisterAuthServiceServer(s, NewAuthServer(authService))
}
```

### Step 12: Wire Dependency Injection (`wire/wire.go`)

Tạo Wire providers và injectors để tự động quản lý dependencies:

```go
//go:build wireinject
// +build wireinject

package wire

import (
    "github.com/google/wire"
    "gorm.io/gorm"
    "github.com/agris/user-service/internal/config"
    "github.com/agris/user-service/internal/database"
    "github.com/agris/user-service/internal/handler"
    "github.com/agris/user-service/internal/repository"
    "github.com/agris/user-service/internal/service"
)

// InitializeConfig loads configuration
func InitializeConfig() (*config.Config, error) {
    return config.Load()
}

// InitializeDatabase creates database connection
func InitializeDatabase(cfg *config.Config) (*gorm.DB, error) {
    return database.NewPostgresDB(cfg)
}

// RepositorySet is a Wire provider set for repositories
var RepositorySet = wire.NewSet(
    repository.NewUserRepository,
)

// ServiceSet is a Wire provider set for services
var ServiceSet = wire.NewSet(
    service.NewAuthService,
    service.NewUserService,
)

// HandlerSet is a Wire provider set for handlers
var HandlerSet = wire.NewSet(
    handler.NewAuthHandler,
    handler.NewUserHandler,
    handler.NewAdminHandler,
)

// InitializeApp creates the complete application with all dependencies
func InitializeApp() (*App, error) {
    wire.Build(
        InitializeConfig,
        InitializeDatabase,
        RepositorySet,
        ServiceSet,
        HandlerSet,
        NewApp,
    )
    return nil, nil
}

// App holds all application dependencies
type App struct {
    Config        *config.Config
    DB            *gorm.DB
    AuthHandler   *handler.AuthHandler
    UserHandler   *handler.UserHandler
    AdminHandler  *handler.AdminHandler
    AuthService   service.AuthService
    UserService   service.UserService
}

func NewApp(
    cfg *config.Config,
    db *gorm.DB,
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    adminHandler *handler.AdminHandler,
    authService service.AuthService,
    userService service.UserService,
) *App {
    return &App{
        Config:       cfg,
        DB:           db,
        AuthHandler:  authHandler,
        UserHandler:  userHandler,
        AdminHandler: adminHandler,
        AuthService:  authService,
        UserService:  userService,
    }
}
```

**Generate Wire code:**

```bash
cd wire
wire
```

Điều này sẽ tạo file `wire_gen.go` với code được generate tự động.

**Lợi ích của Wire:**

1. **Compile-time Safety**: Wire kiểm tra dependencies tại compile-time, không phải runtime
2. **Clean Code**: Giảm boilerplate code, tự động generate dependency injection code
3. **Easy Testing**: Dễ dàng inject mocks cho testing
4. **Type Safety**: Đảm bảo type safety với Go's type system
5. **No Runtime Overhead**: Code được generate tại compile-time, không có reflection overhead

**Wire Workflow:**

1. Define providers (constructor functions)
2. Create provider sets (group related providers)
3. Define injector function với `wire.Build()`
4. Run `wire` command để generate `wire_gen.go`
5. Use generated code trong `main.go`

**Note:** File `wire.go` phải có build tag `//go:build wireinject` để Wire biết đây là file cần generate.

**Makefile để tự động generate Wire code:**

```makefile
.PHONY: wire
wire:
	cd wire && wire

.PHONY: build
build: wire
	go build ./cmd/server

.PHONY: run
run: wire
	go run ./cmd/server

.PHONY: test
test:
	go test ./... -v

.PHONY: test-coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
```

Hoặc sử dụng script:

```bash
#!/bin/bash
# scripts/build.sh

echo "Generating Wire code..."
cd wire && wire

echo "Building application..."
cd ..
go build ./cmd/server
```

### Step 13: Main Application (`cmd/server/main.go`)

```go
package main

import (
    "log"
    "net"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "google.golang.org/grpc"

    "github.com/agris/user-service/internal/middleware"
    grpcHandler "github.com/agris/user-service/internal/grpc"
    "github.com/agris/user-service/wire"
)

func main() {
    // Initialize application with Wire dependency injection
    app, err := wire.InitializeApp()
    if err != nil {
        log.Fatal("Failed to initialize app:", err)
    }

    // Setup HTTP server
    httpApp := fiber.New(fiber.Config{
        AppName: "User Service",
    })

    // Middleware
    httpApp.Use(recover.New())
    httpApp.Use(logger.New())

    // Health check
    httpApp.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // Public routes
    api := httpApp.Group("/api")
    api.Post("/users/register", app.AuthHandler.Register)
    api.Post("/users/login", app.AuthHandler.Login)

    // Protected routes
    protected := api.Group("", middleware.AuthMiddleware(app.Config))
    protected.Get("/users/me", app.UserHandler.GetMe)

    // Admin routes
    admin := protected.Group("", middleware.AdminMiddleware())
    admin.Get("/users", app.AdminHandler.ListUsers)
    admin.Put("/users/:id/role", app.AdminHandler.UpdateUserRole)

    // Start HTTP server
    go func() {
        log.Printf("HTTP server starting on port %s", app.Config.Server.HTTPPort)
        if err := httpApp.Listen(":" + app.Config.Server.HTTPPort); err != nil {
            log.Fatal("Failed to start HTTP server:", err)
        }
    }()

    // Setup gRPC server
    lis, err := net.Listen("tcp", ":"+app.Config.Server.GRPCPort)
    if err != nil {
        log.Fatal("Failed to listen:", err)
    }

    grpcServer := grpc.NewServer()
    grpcHandler.RegisterAuthServer(grpcServer, app.AuthService)

    log.Printf("gRPC server starting on port %s", app.Config.Server.GRPCPort)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatal("Failed to start gRPC server:", err)
    }
}
```

**Note:** File `wire/wire.go` cần import `gorm.io/gorm` và các packages cần thiết. Đảm bảo generate wire code trước khi build:

```bash
# Generate Wire code
cd wire
wire

# Build application
cd ..
go build ./cmd/server
```

---

## 📡 API Endpoints

### Public Endpoints

#### POST `/api/users/register`

Register a new user.

**Request:**

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Response:** `201 Created`

```json
{
    "user": {
        "id": "uuid",
        "email": "user@example.com",
        "role": "user"
    },
    "token": "jwt-token"
}
```

#### POST `/api/users/login`

Login user.

**Request:**

```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

**Response:** `200 OK`

```json
{
    "user": {
        "id": "uuid",
        "email": "user@example.com",
        "role": "user"
    },
    "token": "jwt-token"
}
```

### Protected Endpoints (Require JWT)

#### GET `/api/users/me`

Get current user info.

**Headers:**

```
Authorization: Bearer <token>
```

**Response:** `200 OK`

```json
{
    "id": "uuid",
    "email": "user@example.com",
    "role": "user"
}
```

### Admin Endpoints (Require Admin Role)

#### GET `/api/users`

List all users with pagination.

**Query Parameters:**

-   `page` (default: 1)
-   `limit` (default: 20)
-   `search` (optional, search by email)
-   `role` (optional, filter by role: user|admin)

**Response:** `200 OK`

```json
{
    "data": [
        {
            "id": "uuid",
            "email": "user@example.com",
            "role": "user",
            "created_at": "2025-12-22T00:00:00Z"
        }
    ],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 100
    }
}
```

#### PUT `/api/users/:id/role`

Update user role.

**Request:**

```json
{
    "role": "admin"
}
```

**Response:** `200 OK`

```json
{
    "message": "user role updated successfully"
}
```

---

## 🔧 gRPC Service

### Authenticate Method

**Request:**

```protobuf
message AuthenticateRequest {
  string token = 1;
}
```

**Response:**

```protobuf
message AuthenticateResponse {
  bool valid = 1;
  string user_id = 2;
  string email = 3;
  string role = 4;
  string error = 5;
}
```

**Usage from Product Service:**

```go
conn, _ := grpc.Dial("localhost:9005", grpc.WithInsecure())
client := grpc.NewAuthServiceClient(conn)
resp, _ := client.Authenticate(ctx, &grpc.AuthenticateRequest{Token: token})
```

---

## 🧪 Testing

### Unit Tests Example

```go
// internal/service/auth_service_test.go
package service_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    // ... imports
)

func TestAuthService_Register(t *testing.T) {
    // Setup mock repository
    mockRepo := new(MockUserRepository)
    mockRepo.On("FindByEmail", mock.Anything).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything).Return(nil)

    // Test registration
    cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret", Expiry: time.Hour}}
    authService := service.NewAuthService(mockRepo, cfg)
    user, token, err := authService.Register("test@example.com", "password123")

    // Assert results
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.NotEmpty(t, token)
}
```

### Integration Tests

```bash
go test ./... -v
go test -cover ./...
```

### Testing with Wire

Wire giúp dễ dàng inject mocks cho testing:

```go
// wire/wire_test.go
//go:build !wireinject
// +build !wireinject

package wire

import (
    "github.com/google/wire"
    // ... imports
)

// TestApp for testing with mocks
func InitializeTestApp(mockRepo repository.UserRepository) (*App, error) {
    wire.Build(
        InitializeConfig,
        ServiceSet,
        HandlerSet,
        NewApp,
        // Provide mock repository
        func() repository.UserRepository { return mockRepo },
    )
    return nil, nil
}
```

---

## 🚀 Deployment

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

# Install Wire
RUN go install github.com/google/wire/cmd/wire@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Wire code
RUN cd wire && wire

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /productservice .
EXPOSE 8005 9005

CMD ["./user-service"]
```

### Docker Compose (for User Service)

```yaml
user-service:
    build: ./user-service
    ports:
        - "8005:8005"
        - "9005:9005"
    environment:
        - DB_HOST=postgres
        - DB_PORT=5432
        - DB_USER=user_service
        - DB_PASSWORD=user_service_pwd
        - DB_NAME=user_service
        - JWT_SECRET=${JWT_SECRET}
    depends_on:
        - postgres
    networks:
        - agris-network
```

---

## ✅ Checklist

-   [ ] Go module initialized
-   [ ] Dependencies installed (including Wire)
-   [ ] Configuration setup (Viper)
-   [ ] Database connection
-   [ ] Models created
-   [ ] Repository layer implemented
-   [ ] Service layer implemented
-   [ ] HTTP handlers implemented
-   [ ] Middleware implemented
-   [ ] gRPC proto defined
-   [ ] gRPC server implemented
-   [ ] Wire providers and injectors created
-   [ ] Wire code generated (`wire_gen.go`)
-   [ ] Main application setup (using Wire)
-   [ ] Unit tests written
-   [ ] Integration tests written
-   [ ] API documentation
-   [ ] Dockerfile created (with Wire generation)
-   [ ] Deployment configured

---

**Next Steps:**

1. Implement Product Service
2. Setup Traefik Gateway
3. Integration testing
4. Performance testing
5. Production deployment
