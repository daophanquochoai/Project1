# Product Review & Rating System - Task Breakdown & Implementation Plan

**Author:** Trần Tiến Đạt  
**Version:** 1.0.0  
**Date:** 22 Dec 2025  
**Timeline:** 1 sprint (1 week)

---

## 📋 Mục lục

1. [Xác nhận Requirements](#xác-nhận-requirements)
2. [Thiết kế ERD](#thiết-kế-erd)
3. [Sequence Diagrams](#sequence-diagrams)
4. [Task Breakdown](#task-breakdown)
5. [Technical Checklist](#technical-checklist)
6. [Non-Functional Requirements Checklist](#non-functional-requirements-checklist)
7. [Workflow & Git Strategy](#workflow--git-strategy)

---

## ✅ Xác nhận Requirements

### 1.1 Business Requirements

#### User Management

-   ✅ **BR-001**: User registration với email và password
-   ✅ **BR-002**: JWT-based authentication
-   ✅ **BR-003**: Phân biệt User và Admin roles
-   ✅ **BR-004**: Admin có thể xem danh sách users và thay đổi phân quyền
-   ✅ **BR-005**: Admin không thể tự thay đổi role của chính mình

#### Product Management

-   ✅ **BR-006**: Hiển thị danh sách sản phẩm với phân trang
-   ✅ **BR-007**: Tìm kiếm sản phẩm theo tên (không phân biệt dấu và hoa thường tiếng Việt)
-   ✅ **BR-008**: Gợi ý sản phẩm tương tự (cùng category)
-   ✅ **BR-009**: Hiển thị sản phẩm đi kèm/phụ kiện liên quan
-   ✅ **BR-010**: Hiển thị sản phẩm được yêu thích dựa vào rating

#### Rating & Review

-   ✅ **BR-011**: Chỉ user đã đăng nhập mới được đánh giá
-   ✅ **BR-012**: Một user chỉ được tạo 1 rating cho mỗi sản phẩm
-   ✅ **BR-013**: User có thể cập nhật rating của chính mình
-   ✅ **BR-014**: Rating phải là số nguyên từ 1-5 sao
-   ✅ **BR-015**: Tự động tính toán và cập nhật average rating
-   ✅ **BR-016**: User chỉ xóa được rating của chính mình, Admin xóa được mọi rating

### 1.2 Technical Requirements

-   ✅ **TC-001**: 2 services độc lập (User Service: 8005 HTTP, 9005 gRPC | Product Service: 8010 HTTP)
-   ✅ **TC-002**: User Service expose gRPC endpoint Authenticate()
-   ✅ **TC-003**: Product Service/Gateway sử dụng gRPC để validate token
-   ✅ **TC-004**: Traefik làm API Gateway
-   ✅ **TC-005**: Mỗi service có database riêng
-   ✅ **TC-008**: Search hỗ trợ tiếng Việt (không phân biệt dấu, hoa thường, partial match)
-   ✅ **TC-009**: Seed database (5,000-10,000 products, 20 categories, ~5 ratings/product)

### 1.3 Non-Functional Requirements

#### Performance

-   ✅ **NFR-P01**: P90 Latency ≤ 200ms
-   ✅ **NFR-P02**: Caching Strategy (Redis, TTL: similar 1h, popular 30m, related 1h, hit rate ≥80%)
-   ✅ **NFR-P03**: Database Optimization (indexes, query <50ms, connection pooling)

#### Security

-   ✅ **NFR-S01**: JWT authentication (expiry 1h), bcrypt (min cost 10)
-   ✅ **NFR-S02**: RBAC (User vs Admin), middleware validation
-   ✅ **NFR-S03**: Input validation, SQL injection prevention, data sanitization

#### Code Quality

-   ✅ **NFR-Q01**: Clean architecture (handler → service → repository)
-   ✅ **NFR-Q02**: Unit tests (≥50% coverage)
-   ✅ **NFR-Q03**: API documentation (Swagger/Postman), README với setup instructions

---

## 🗄️ Thiết kế ERD

### 2.1 User Service Database Schema

```
┌─────────────────┐
│     users       │
├─────────────────┤
│ id (PK)         │
│ email (UNIQUE)  │
│ password_hash   │
│ role            │  (user, admin)
│ created_at      │
│ updated_at      │
│ deleted_at      │
└─────────────────┘
```

**Indexes:**

-   `idx_users_email` on `email`
-   `idx_users_role` on `role`

### 2.2 Product Service Database Schema

```
┌─────────────────┐
│   categories    │
├─────────────────┤
│ id (PK)         │
│ name            │
│ description     │
│ created_at      │
│ updated_at      │
└─────────────────┘
         │
         │ 1:N
         │
┌─────────────────┐
│    products     │
├─────────────────┤
│ id (PK)         │
│ name            │
│ description     │
│ price           │
│ category_id(FK) │
│ average_rating  │  (computed, default 0)
│ total_ratings   │  (count, default 0)
│ created_at      │
│ updated_at      │
│ deleted_at      │
└─────────────────┘
         │
         │ 1:N
         │
┌─────────────────┐
│     ratings     │
├─────────────────┤
│ id (PK)         │
│ product_id (FK) │
│ user_id         │  (from User Service)
│ rating          │  (1-5)
│ comment         │  (optional)
│ created_at      │
│ updated_at      │
│ deleted_at      │
└─────────────────┘

┌─────────────────┐
│ product_related │  (Many-to-Many)
├─────────────────┤
│ product_id (FK) │
│ related_id (FK) │
│ relation_type   │  (accessory, bundle, etc.)
│ created_at      │
└─────────────────┘
```

**Indexes:**

-   `idx_products_category_id` on `category_id`
-   `idx_products_average_rating` on `average_rating`
-   `idx_products_name` on `name` (for search)
-   `idx_ratings_product_id` on `product_id`
-   `idx_ratings_user_id` on `user_id`
-   `UNIQUE idx_ratings_user_product` on `(user_id, product_id)` where `deleted_at IS NULL`
-   `idx_product_related_product_id` on `product_id`
-   `idx_product_related_related_id` on `related_id`

**Triggers/Functions:**

-   Function to update `average_rating` and `total_ratings` on `products` when ratings change

---

## 🔄 Sequence Diagrams

### 3.1 User Registration Flow

```
User          Gateway        User Service      Database
  |              |                |               |
  |--POST /register-->|            |               |
  |              |--POST /register-->|             |
  |              |                |--Validate---->|
  |              |                |--Hash Password|
  |              |                |--Create User->|
  |              |                |<--User Created|
  |              |<--201 Created--|               |
  |<--201 Created--|                |               |
```

### 3.2 User Login Flow

```
User          Gateway        User Service      Database
  |              |                |               |
  |--POST /login-->|              |               |
  |              |--POST /login-->|               |
  |              |                |--Find User--->|
  |              |                |<--User Found--|
  |              |                |--Verify Pass->|
  |              |                |--Generate JWT-|
  |              |<--200 + JWT----|               |
  |<--200 + JWT----|                |               |
```

### 3.3 Token Validation Flow (gRPC)

```
Product Service    User Service (gRPC)
      |                    |
      |--Authenticate(token)-->|
      |                    |--Validate JWT--|
      |                    |--Check User---|
      |                    |<--User Info---|
      |<--User Info + Valid--|
      |                    |
```

### 3.4 Get Products with Authentication

```
User          Gateway        Product Service   User Service (gRPC)   Database
  |              |                |                    |               |
  |--GET /products-->|            |                    |               |
  |              |--GET /products-->|                  |               |
  |              |                |--Validate Token--->|
  |              |                |                    |--Check JWT--->|
  |              |                |<--User Info--------|               |
  |              |                |--Get Products------>|
  |              |                |<--Products----------|
  |              |<--200 + Products|                    |               |
  |<--200 + Products--|                |                    |               |
```

### 3.5 Create Rating Flow

```
User          Gateway        Product Service   User Service (gRPC)   Database
  |              |                |                    |               |
  |--POST /ratings-->|            |                    |               |
  |              |--POST /ratings-->|                  |               |
  |              |                |--Validate Token--->|
  |              |                |<--User Info--------|
  |              |                |--Check Existing--->|
  |              |                |<--No Existing------|
  |              |                |--Create Rating---->|
  |              |                |--Update Avg Rating>|
  |              |                |--Invalidate Cache--|
  |              |                |<--Rating Created---|
  |              |<--201 Created--|                    |               |
  |<--201 Created--|                |                    |               |
```

### 3.6 Search Products Flow (with Cache)

```
User          Gateway        Product Service   Redis      Database
  |              |                |              |           |
  |--GET /search?q=...-->|        |              |           |
  |              |--GET /search-->|              |           |
  |              |                |--Check Cache->|
  |              |                |<--Cache Miss--|
  |              |                |--Search DB---->|
  |              |                |<--Results------|
  |              |                |--Set Cache---->|
  |              |<--200 + Results|              |           |
  |<--200 + Results--|                |              |           |
```

---

## 📦 Task Breakdown

### Epic: Product Review & Rating System Implementation

#### Phase 1: Project Setup & Infrastructure

**EPIC-001: Project Setup**

-   **ID-001**: Setup project structure (Go modules, folder structure)
    -   Create `user-service/` và `product-service/` directories
    -   Initialize Go modules
    -   Setup `.gitignore`
    -   Create `docker-compose.yml` for local development
-   **ID-002**: Setup Traefik Gateway configuration
    -   Configure Traefik routing rules
    -   Setup SSL/TLS (optional for dev)
    -   Health check endpoints
-   **ID-003**: Setup PostgreSQL databases
    -   Create `user_db` và `product_db` databases
    -   Setup connection pooling configuration
    -   Migration tool setup (golang-migrate or GORM migrations)
-   **ID-004**: Setup Redis configuration
    -   Redis connection setup
    -   Cache utility functions
    -   TTL configuration

#### Phase 2: User Service Implementation

**EPIC-002: User Service Core**

-   **ID-101**: User Service - Database Models & Migrations
    -   Create `User` model
    -   Database migration scripts
    -   Indexes creation
-   **ID-102**: User Service - Repository Layer
    -   User repository interface
    -   CRUD operations implementation
    -   Query optimization
-   **ID-103**: User Service - Authentication Service
    -   Password hashing (bcrypt)
    -   JWT token generation
    -   JWT token validation
    -   Token expiry management
-   **ID-104**: User Service - HTTP Handlers
    -   POST `/api/users/register` - User registration
    -   POST `/api/users/login` - User login
    -   GET `/api/users/me` - Get current user info
    -   Middleware for authentication
-   **ID-105**: User Service - gRPC Service
    -   Define `.proto` file for Authenticate service
    -   Implement gRPC Authenticate endpoint
    -   gRPC server setup (port 9005)
-   **ID-106**: User Service - Admin Endpoints
    -   GET `/api/users` - List all users (pagination, search, filter)
    -   PUT `/api/users/:id/role` - Update user role
    -   Admin middleware validation
    -   Prevent self-role change logic
-   **ID-107**: User Service - Input Validation & Error Handling
    -   Request validation
    -   Error response formatting
    -   Input sanitization

#### Phase 3: Product Service Implementation

**EPIC-003: Product Service Core**

-   **ID-201**: Product Service - Database Models & Migrations
    -   Create `Category`, `Product`, `Rating`, `ProductRelated` models
    -   Database migration scripts
    -   Indexes creation (including Vietnamese search indexes)
    -   Trigger/Function for average rating calculation
-   **ID-202**: Product Service - Repository Layer
    -   Product repository interface
    -   Category repository
    -   Rating repository
    -   CRUD operations implementation
    -   Query optimization
-   **ID-203**: Product Service - Vietnamese Search Implementation
    -   Vietnamese text normalization function
    -   Case-insensitive search
    -   Diacritic removal for search
    -   Partial match support
    -   Search index optimization
-   **ID-204**: Product Service - gRPC Client Setup
    -   gRPC client for User Service
    -   Connection pooling
    -   Error handling & retry logic
-   **ID-205**: Product Service - Authentication Middleware
    -   gRPC token validation middleware
    -   User context injection
    -   Protected route handling
-   **ID-206**: Product Service - HTTP Handlers - Product Management
    -   GET `/api/products` - List products (pagination)
    -   GET `/api/products/search` - Search products by name
    -   GET `/api/products/:id` - Get product detail
    -   GET `/api/products/:id/similar` - Get similar products
    -   GET `/api/products/:id/related` - Get related products
    -   GET `/api/products/popular` - Get popular products
-   **ID-207**: Product Service - HTTP Handlers - Rating Management
    -   POST `/api/products/:id/ratings` - Create rating
    -   PUT `/api/ratings/:id` - Update own rating
    -   DELETE `/api/ratings/:id` - Delete rating
    -   GET `/api/products/:id/ratings` - Get product ratings (pagination, filter)
    -   GET `/api/products/:id/ratings/stats` - Get rating statistics
    -   GET `/api/users/me/ratings` - Get my ratings
-   **ID-208**: Product Service - Rating Business Logic
    -   Unique constraint enforcement (1 user = 1 rating per product)
    -   Average rating calculation
    -   Rating update logic
    -   Rating deletion logic
-   **ID-209**: Product Service - Cache Implementation
    -   Redis cache for similar products (TTL: 1h)
    -   Redis cache for popular products (TTL: 30m)
    -   Redis cache for related products (TTL: 1h)
    -   Cache invalidation on rating changes
    -   Cache key strategy

#### Phase 4: Database Seeding

**EPIC-004: Data Seeding**

-   **ID-301**: Seed Script - Categories
    -   Create 20 categories
    -   Category data structure
-   **ID-302**: Seed Script - Products
    -   Generate 5,000-10,000 products
    -   Distribute across categories
    -   Realistic product data
-   **ID-303**: Seed Script - Ratings
    -   Generate ~5 ratings per product average
    -   Random rating distribution (1-5 stars)
    -   User assignment logic

#### Phase 5: Performance Optimization

**EPIC-005: Performance & Optimization**

-   **ID-401**: Database Query Optimization
    -   Analyze slow queries
    -   Add missing indexes
    -   Query optimization
    -   Connection pooling tuning
-   **ID-402**: Cache Strategy Refinement
    -   Cache hit rate monitoring
    -   TTL adjustment
    -   Cache warming strategies
-   **ID-403**: Load Testing & Performance Tuning
    -   Load test setup (10,000 products, 100 concurrent users)
    -   P90 latency measurement
    -   Performance bottleneck identification
    -   Optimization implementation

#### Phase 6: Testing

**EPIC-006: Testing**

-   **ID-501**: Unit Tests - User Service
    -   Repository layer tests
    -   Service layer tests
    -   Handler tests
    -   Mock implementations
-   **ID-502**: Unit Tests - Product Service
    -   Repository layer tests
    -   Service layer tests
    -   Handler tests
    -   Vietnamese search tests
    -   Rating logic tests
-   **ID-503**: Integration Tests
    -   End-to-end API tests
    -   gRPC communication tests
    -   Database integration tests
-   **ID-504**: Test Coverage Report
    -   Generate coverage report
    -   Ensure ≥50% coverage
    -   Coverage analysis

#### Phase 7: Documentation

**EPIC-007: Documentation**

-   **ID-601**: API Documentation
    -   Swagger/OpenAPI specification
    -   Postman collection
    -   API endpoint documentation
-   **ID-602**: Setup Documentation
    -   README.md với setup instructions
    -   Environment variables documentation
    -   Docker setup guide
    -   Database migration guide
-   **ID-603**: Architecture Documentation
    -   System architecture diagram
    -   Service communication flow
    -   Database schema documentation

#### Phase 8: Security & Validation

**EPIC-008: Security Hardening**

-   **ID-701**: Security Audit
    -   SQL injection prevention review
    -   XSS prevention
    -   Input validation audit
    -   Authentication/Authorization review
-   **ID-702**: Account Security Features
    -   Account lockout after N failed login attempts (optional)
    -   Password strength validation
    -   Rate limiting (optional)

---

## ✅ Technical Checklist

### Infrastructure

-   [ ] Go 1.25+ installed
-   [ ] PostgreSQL 17+ installed/configured
-   [ ] Redis 7+ installed/configured
-   [ ] Traefik 2.10+ configured
-   [ ] Docker & Docker Compose setup
-   [ ] Environment variables configuration

### User Service (Port 8005 HTTP, 9005 gRPC)

-   [ ] Fiber framework setup
-   [ ] GORM configured
-   [ ] Database connection established
-   [ ] User model created
-   [ ] Migrations run
-   [ ] JWT library integrated
-   [ ] bcrypt password hashing
-   [ ] gRPC server setup
-   [ ] `.proto` file defined
-   [ ] HTTP endpoints implemented
-   [ ] gRPC Authenticate endpoint implemented
-   [ ] Admin endpoints implemented
-   [ ] Middleware implemented
-   [ ] Error handling implemented

### Product Service (Port 8010 HTTP)

-   [ ] Fiber framework setup
-   [ ] GORM configured
-   [ ] Database connection established
-   [ ] Models created (Product, Category, Rating, ProductRelated)
-   [ ] Migrations run
-   [ ] Indexes created
-   [ ] Vietnamese search implemented
-   [ ] gRPC client configured
-   [ ] Authentication middleware implemented
-   [ ] HTTP endpoints implemented
-   [ ] Rating logic implemented
-   [ ] Cache implementation (Redis)
-   [ ] Cache invalidation logic

### Gateway

-   [ ] Traefik configured
-   [ ] Routing rules defined
-   [ ] Health checks configured
-   [ ] Load balancing configured

### Database

-   [ ] User database created
-   [ ] Product database created
-   [ ] All migrations applied
-   [ ] Indexes created
-   [ ] Seed data loaded (5,000-10,000 products, 20 categories, ~5 ratings/product)

### Testing

-   [ ] Unit tests written (≥50% coverage)
-   [ ] Integration tests written
-   [ ] Load tests performed
-   [ ] P90 latency verified (≤200ms)

### Documentation

-   [ ] API documentation (Swagger/Postman)
-   [ ] README.md with setup instructions
-   [ ] Architecture documentation

---

## ✅ Non-Functional Requirements Checklist

### Performance (NFR-P01, P02, P03)

-   [ ] P90 Latency ≤ 200ms verified
-   [ ] Redis caching implemented
    -   [ ] Similar products cache (TTL: 1h)
    -   [ ] Popular products cache (TTL: 30m)
    -   [ ] Related products cache (TTL: 1h)
-   [ ] Cache hit rate ≥80% achieved
-   [ ] Database indexes created
-   [ ] Query performance <50ms verified
-   [ ] Connection pooling configured

### Security (NFR-S01, S02, S03)

-   [ ] JWT tokens with 1h expiry
-   [ ] bcrypt password hashing (min cost 10)
-   [ ] Token validation on protected routes
-   [ ] RBAC implemented (User vs Admin)
-   [ ] Admin middleware validation
-   [ ] Input validation implemented
-   [ ] SQL injection prevention (GORM parameterized queries)
-   [ ] Data sanitization implemented

### Code Quality (NFR-Q01, Q02, Q03)

-   [ ] Clean architecture (handler → service → repository)
-   [ ] Unit tests ≥50% coverage
-   [ ] API documentation (Swagger/Postman)
-   [ ] README with setup instructions

---

## 🔀 Workflow & Git Strategy

### Branch Naming Convention

**Format:** `<type>/<TASK-ID>: <description>`

**Types:**

-   `feat/ID-XXX`: Implement new feature
-   `improve/ID-XXX`: Improve existing functionality
-   `fix/ID-XXX`: Fix bug
-   `docs/ID-XXX`: Documentation changes
-   `test/ID-XXX`: Test-related changes
-   `refactor/ID-XXX`: Code refactoring

**Examples:**

-   `feat/ID-101`: User Service - Database Models & Migrations
-   `feat/ID-102`: User Service - Repository Layer
-   `improve/ID-401`: Database Query Optimization
-   `fix/ID-789`: Fix rating calculation bug

### Commit Message Convention

**Format:** `<prefix>: <description>`

**Prefixes:**

-   `#add`: Thêm mới
-   `#update`: Sửa đổi
-   `#remove`: Xóa
-   `#fix`: Sửa lỗi

**Examples:**

-   `#add: User registration endpoint`
-   `#update: Improve Vietnamese search algorithm`
-   `#remove: Remove unused cache keys`
-   `#fix: Fix average rating calculation bug`

### Workflow Process

1. **Create Epic/Story in Jira**

    - Mentor creates epic/story
    - Assign to developer
    - Developer breaks down into tasks

2. **Start New Feature**

    - Checkout from `main`: `git checkout main`
    - Pull latest: `git pull origin main`
    - Create new branch: `git checkout -b feat/ID-101`
    - Work on task

3. **Commit Changes**

    - Commit with proper prefix: `git commit -m "#add: User model and migration"`
    - Push branch: `git push origin feat/ID-101`

4. **Create Pull Request**

    - Create PR from feature branch to `main`
    - Request review from mentor
    - Address review comments

5. **Merge After Approval**

    - Mentor approves PR
    - Merge to `main`
    - Delete feature branch

6. **Never commit directly to `main`**
    - All changes must go through PR and review

### Task Tracking

-   Each task should have corresponding Jira ticket
-   Branch name must match Jira task ID
-   Commit messages should reference task ID when relevant
-   PR title should include task ID and description

---

## 📊 Progress Tracking

### Phase Completion Status

-   [ ] Phase 1: Project Setup & Infrastructure
-   [ ] Phase 2: User Service Implementation
-   [ ] Phase 3: Product Service Implementation
-   [ ] Phase 4: Database Seeding
-   [ ] Phase 5: Performance Optimization
-   [ ] Phase 6: Testing
-   [ ] Phase 7: Documentation
-   [ ] Phase 8: Security & Validation

### Key Milestones

1. **Milestone 1**: Infrastructure setup complete
2. **Milestone 2**: User Service fully functional
3. **Milestone 3**: Product Service fully functional
4. **Milestone 4**: All features implemented
5. **Milestone 5**: Performance requirements met
6. **Milestone 6**: Testing complete
7. **Milestone 7**: Documentation complete
8. **Milestone 8**: Ready for production

---

## 🎯 Success Criteria

### Functional Requirements

-   ✅ All business requirements (BR-001 to BR-016) implemented
-   ✅ All functional features (F-U01 to F-R06) working
-   ✅ All technical requirements (TC-001 to TC-009) met

### Non-Functional Requirements

-   ✅ P90 Latency ≤ 200ms
-   ✅ Cache hit rate ≥80%
-   ✅ Database queries <50ms
-   ✅ Unit test coverage ≥50%
-   ✅ Security requirements met
-   ✅ Code quality standards met

### Deliverables

-   ✅ Working system with all features
-   ✅ API documentation
-   ✅ Setup documentation
-   ✅ Test coverage report
-   ✅ Performance test results

---

## 📝 Notes

-   **Timeline**: 1 sprint (1 week)
-   **Priority**: Focus on core features first (User Service → Product Service → Rating)
-   **Performance**: Critical requirement - must meet P90 latency target
-   **Code Review**: All PRs must be reviewed and approved by mentor
-   **Testing**: Write tests alongside implementation, not after

---

**Last Updated**: 22 Dec 2025  
**Status**: Planning Phase
