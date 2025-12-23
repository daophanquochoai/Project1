# API Documentation - Hệ thống E-commerce Nông Nghiệp

## Lịch sử phiên bản

| Phiên bản | Ngày | Tác giả | Mô tả |
|-----------|------|---------|-------|
| 1.0.0 | 2025-12-22 | Trần Tiến Đạt | Tạo tài liệu |
| 1.0.1 | 2025-12-23 | Đào Phan Quốc Hoài | Viết tài liệu cụ thể cho task TOK-33 |

---

## 1. Tổng quan

### 1.1 Mục đích

Xây dựng hệ thống đánh giá (rating) và nhận xét (review) cho nền tảng e-commerce nông nghiệp:

- User đăng ký/đăng nhập bằng JWT
- Xem danh sách sản phẩm + tìm kiếm tiếng Việt
- User tạo/cập nhật/xóa rating cho sản phẩm
- Hệ thống tự động tính average rating, cung cấp thống kê + "popular products"

### 1.2 Phạm vi

- **User Service**: quản lý user, xác thực, phân quyền (RBAC)
- **Product Service**: quản lý product + rating/review
- **Gateway**: Traefik route request tới các service

### 1.3 Định nghĩa thuật ngữ

- **P90 Latency**: 90% request phải hoàn thành dưới ngưỡng (≤ 200ms)
- **gRPC**: Remote Procedure Call hiệu năng cao cho inter-service
- **JWT**: JSON Web Token
- **RBAC**: Role-based Access Control
- **Admin / User**: 2 role chính trong hệ thống

---

## 2. Thiết kế

### 2.1 Kiến trúc microservices

#### 2.1.1 UserService

**Cấu hình port:**
- HTTP API: 8005
- gRPC Server: 9005

**Chức năng:**
- Đăng ký và đăng nhập người dùng
- Tạo token và xác thực
- Quản lý thông tin của người dùng

#### 2.1.2 ProductService

**Cấu hình port:**
- HTTP API: 8010

**Chức năng:**
- Quản lý loại sản phẩm
- Tìm kiếm và lọc sản phẩm
- Quản lý đánh giá và bình luận của sản phẩm
- Khuyến nghị sản phẩm liên quan và tương tự

#### 2.1.3 Api Gateway (Traefik)

**Chức năng:**
- Điều hướng yêu cầu từ người dùng
- Cân bằng tải yêu cầu
- Giới hạn yêu cầu

---

### 2.2 Thiết kế cơ sở dữ liệu

#### 2.2.1 Tổng quan

##### a. Các thực thể chính

- **users**: Quản lý thông tin tài khoản và người dùng
- **products**: Quản lý thông tin sản phẩm
- **categories**: Quản lý thông tin loại sản phẩm
- **ratings**: Quản lý đánh giá và bình luận từ người dùng cho sản phẩm
- **product_related**: Quản lý thông tin sản phẩm liên quan, tương tự

##### b. Quan hệ tổng quát

- 1 category - N product
- 1 product - N ratings
- 1 user - N ratings
- 1 product - N product_relates

##### c. Định nghĩa các bảng

**Bảng người dùng (users)**

| Cột | Kiểu | Ràng buộc | Miêu tả |
|-----|------|-----------|---------|
| id | UUID | PK | Id người dùng |
| name | TEXT | NOT NULL | Tên hiển thị |
| email | TEXT | NOT NULL, UNIQUE | Email đăng nhập |
| password_hash | TEXT | NOT NULL | Mật khẩu đã hash |
| role | TEXT | NOT NULL | Chứa 2 giá trị "user" và "admin" |
| created_at | TIMESTAMPTZ | NOT NULL | Thời điểm tạo |
| updated_at | TIMESTAMPTZ | NOT NULL | Thời điểm cập nhật |
| deleted_at | TIMESTAMPTZ | NULL | Xóa mềm |

**Bảng loại sản phẩm (categories)**

| Cột | Kiểu | Ràng buộc | Miêu tả |
|-----|------|-----------|---------|
| id | UUID | PK | Id loại sản phẩm |
| name | TEXT | NOT NULL | Tên danh mục |
| description | TEXT | NULL | Mô tả |
| created_at | TIMESTAMPTZ | NOT NULL | Thời điểm tạo |
| updated_at | TIMESTAMPTZ | NOT NULL | Thời điểm cập nhật |

**Bảng sản phẩm (products)**

| Cột | Kiểu | Ràng buộc | Miêu tả |
|-----|------|-----------|---------|
| id | UUID | PK | Id sản phẩm |
| name | TEXT | NOT NULL | Tên sản phẩm |
| description | TEXT | NULL | Mô tả |
| price | NUMERIC(12,2) | NOT NULL | Giá |
| category_id | UUID | FK -> categories(id) | Danh mục |
| average_rating | NUMERIC(3,2) | NOT NULL | Điểm đánh giá trung bình |
| total_ratings | INTEGER | NOT NULL | Tổng lượt đánh giá |
| created_at | TIMESTAMPTZ | NOT NULL | Thời điểm tạo |
| updated_at | TIMESTAMPTZ | NOT NULL | Thời điểm cập nhật |
| deleted_at | TIMESTAMPTZ | NULL | Xóa mềm |

**Bảng đánh giá và bình luận (ratings)**

| Cột | Kiểu | Ràng buộc | Miêu tả |
|-----|------|-----------|---------|
| id | UUID | PK | Id đánh giá |
| product_id | UUID | NOT NULL, FK -> products(id) | Sản phẩm |
| user_id | UUID | NOT NULL | Id người dùng (userservice) |
| rating | INTEGER | NOT NULL | Giá trị từ 1-5 |
| comment | TEXT | NULL | Nhận xét |
| created_at | TIMESTAMPTZ | NOT NULL | Thời điểm tạo |
| updated_at | TIMESTAMPTZ | NOT NULL | Thời điểm cập nhật |
| deleted_at | TIMESTAMPTZ | NULL | Xóa mềm |

**Bảng sản phẩm liên quan (product_relates)**

| Cột | Kiểu | Ràng buộc | Miêu tả |
|-----|------|-----------|---------|
| product_id | UUID | NOT NULL, FK -> products(id) | Sản phẩm chính |
| related_id | UUID | NOT NULL, FK -> products(id) | Sản phẩm liên quan |
| relation_type | TEXT | NOT NULL | Chứa giá trị: "related", "similar" |
| created_at | TIMESTAMPTZ | NOT NULL | Thời điểm tạo |

---

## 3. API Reference

### 3.1 UserService

**Base URL**: `http://localhost:8005/users`

#### 3.1.1 Đăng ký người dùng

**Endpoint**: `POST /users/register`

**Request Body:**
```json
{
  "name": "Quốc Hoài",
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Quốc Hoài",
  "email": "user@example.com",
  "role": "user",
  "created_at": "2025-12-23T10:00:00Z"
}
```

**Error Responses:**
- `400`: Email đã tồn tại
- `400`: Email là bắt buộc
- `400`: Email không đúng mẫu
- `400`: Mật khẩu nhất định có ít nhất 6 ký tự
- `400`: Dữ liệu chưa đúng

#### 3.1.2 Đăng nhập tài khoản

**Endpoint**: `POST /users/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user"
  }
}
```

**Error Responses:**
- `401`: Tài khoản chưa được xác thực
- `403`: Tài khoản đã bị khóa
- `500`: Máy chủ bị lỗi

#### 3.1.3 Lấy tài khoản hiện tại

**Endpoint**: `GET /users/me`

**Headers:**
```
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Quốc Hoài",
  "email": "user@example.com",
  "role": "user",
  "status": "ACTIVE",
  "created_at": "2025-12-23T10:00:00Z"
}
```

**Error Responses:**
- `401`: Tài khoản chưa được xác thực

#### 3.1.4 Cập nhật vai trò tài khoản (Dành cho Admin)

**Endpoint**: `PATCH /users/{userId}/role`

**Headers:**
```
Authorization: Bearer {admin_token}
```

**Request Body:**
```json
{
  "role": "admin"
}
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "admin",
  "updated_at": "2025-12-23T11:00:00Z"
}
```

**Error Responses:**
- `401`: Tài khoản chưa được xác thực
- `400`: Không thể cập nhật tài khoản chính mình
- `404`: Tài khoản không thể tìm thấy
- `403`: Tài khoản không có quyền
- `500`: Máy chủ bị lỗi

#### 3.1.5 Xác thực token (gRPC)

```protobuf
syntax = "proto3";
package user;

service UserService {
  rpc Authenticate(AuthRequest) returns (AuthResponse);
}

message AuthRequest {
  string token = 1;
}

message AuthResponse {
  bool valid = 1;
  string user_id = 2;
  string role = 3;
  string error_message = 4;
}
```

---

### 3.2 ProductService

**Base URL**: `http://localhost:8010/products`

#### 3.2.1 Lấy danh sách sản phẩm

**Endpoint**: `GET /products`

**Query Parameters:**
- `page` (int, default: 1)
- `limit` (int, default: 20, max: 100)
- `category_id` (UUID, optional)
- `sort` (string, optional: "rating", "newest", "name")

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "Cà phê Robusta",
      "category": {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "name": "Cà phê"
      },
      "avg_rating": 4.5,
      "rating_count": 120,
      "created_at": "2025-12-20T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### 3.2.2 Tìm kiếm sản phẩm

**Endpoint**: `GET /products/search`

**Query Parameters:**
- `q` (string, required): Search query
- `page` (int, default: 1)
- `limit` (int, default: 20)
- `category_id` (UUID, optional)

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "Cà phê Robusta",
      "category": {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "name": "Cà phê"
      },
      "avg_rating": 4.5,
      "rating_count": 120,
      "created_at": "2025-12-20T10:00:00Z"
    }
  ],
  "pagination": {
    "q": "name",
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### 3.2.3 Lấy chi tiết sản phẩm

**Endpoint**: `GET /products/{productId}`

**Response (200 OK):**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "name": "Cà phê Robusta",
  "description": "Cà phê Robusta chất lượng cao từ Tây Nguyên",
  "category": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "name": "Cà phê"
  },
  "avg_rating": 4.5,
  "rating_count": 120,
  "created_at": "2025-12-20T10:00:00Z",
  "updated_at": "2025-12-23T10:00:00Z"
}
```

**Error Response:**
- `400`: Sản phẩm không tìm thấy

#### 3.2.4 Lấy sản phẩm tương tự

**Endpoint**: `GET /products/{productId}/similar`

**Query Parameters:**
- `limit` (int, default: 5, max: 20)

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "680e8400-e29b-41d4-a716-446655440000",
      "name": "Cà phê Arabica",
      "category": {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "name": "Cà phê"
      },
      "avg_rating": 4.7,
      "rating_count": 89
    }
  ]
}
```

#### 3.2.5 Lấy sản phẩm liên quan

**Endpoint**: `GET /products/{productId}/related`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "690e8400-e29b-41d4-a716-446655440000",
      "name": "Máy xay cà phê",
      "type": "related",
      "avg_rating": 4.3,
      "rating_count": 45
    },
    {
      "id": "700e8400-e29b-41d4-a716-446655440000",
      "name": "Cà phê hòa tan",
      "type": "related",
      "avg_rating": 4.1,
      "rating_count": 67
    }
  ]
}
```

#### 3.2.6 Lấy sản phẩm phổ biến

**Endpoint**: `GET /products/popular`

**Query Parameters:**
- `category_id` (UUID, optional)
- `limit` (int, default: 10, max: 50)

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "690e8400-e29b-41d4-a716-446655440000",
      "name": "Máy xay cà phê",
      "type": "related",
      "avg_rating": 4.3,
      "rating_count": 45
    },
    {
      "id": "700e8400-e29b-41d4-a716-446655440000",
      "name": "Cà phê hòa tan",
      "type": "related",
      "avg_rating": 4.1,
      "rating_count": 67
    }
  ]
}
```

#### 3.2.7 Đánh giá sản phẩm

**Endpoint**: `POST /products/{productId}/ratings`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "stars": 5,
  "comment": "Sản phẩm rất tốt, đóng gói cẩn thận"
}
```

**Response (201 Created):**
```json
{
  "id": "710e8400-e29b-41d4-a716-446655440000",
  "product_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "stars": 5,
  "comment": "Sản phẩm rất tốt, đóng gói cẩn thận",
  "created_at": "2025-12-23T11:00:00Z"
}
```

**Error Responses:**
- `401`: Token không được xác thực
- `400`: Đánh giá không hợp lệ
- `409`: Tài khoản đã đánh giá sản phẩm này

#### 3.2.8 Cập nhật đánh giá sản phẩm

**Endpoint**: `PUT /products/{productId}/ratings/{ratingId}`

**Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "stars": 4,
  "comment": "Cập nhật đánh giá sau khi dùng lâu hơn"
}
```

**Response (200 OK):**
```json
{
  "id": "710e8400-e29b-41d4-a716-446655440000",
  "product_id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "stars": 4,
  "comment": "Cập nhật đánh giá sau khi dùng lâu hơn",
  "created_at": "2025-12-23T11:00:00Z",
  "updated_at": "2025-12-23T15:30:00Z"
}
```

**Error Responses:**
- `401`: Token không thể xác thực
- `403`: Tài khoản không có quyền
- `404`: Đánh giá không tìm thấy
- `404`: Sản phẩm không tìm thấy
- `404`: Tài khoản không tìm thấy
- `400`: Đánh giá không hợp lệ

#### 3.2.9 Xóa đánh giá sản phẩm

**Endpoint**: `DELETE /products/{productId}/ratings/{ratingId}`

**Headers:**
```
Authorization: Bearer {token}
```

**Response**: `204 No Content`

**Error Responses:**
- `401`: Token không thể xác thực
- `403`: Tài khoản không có quyền
- `404`: Đánh giá không tìm thấy
- `404`: Sản phẩm không tìm thấy
- `404`: Tài khoản không tìm thấy

#### 3.2.10 Lấy đánh giá của sản phẩm

**Endpoint**: `GET /products/{productId}/ratings`

**Query Parameters:**
- `page` (int, default: 1)
- `limit` (int, default: 20, max: 100)
- `stars` (int, optional: 1-5) - Filter by star rating
- `sort` (string, optional: "newest", "oldest", "highest", "lowest")

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "710e8400-e29b-41d4-a716-446655440000",
      "product_id": "660e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "stars": 5,
      "comment": "Sản phẩm rất tốt, đóng gói cẩn thận",
      "created_at": "2025-12-23T11:00:00Z",
      "updated_at": "2025-12-23T11:00:00Z"
    },
    {
      "id": "720e8400-e29b-41d4-a716-446655440000",
      "product_id": "660e8400-e29b-41d4-a716-446655440000",
      "user_id": "560e8400-e29b-41d4-a716-446655440000",
      "stars": 4,
      "comment": "Tốt nhưng giao hàng hơi lâu",
      "created_at": "2025-12-22T14:30:00Z",
      "updated_at": "2025-12-22T14:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 120,
    "total_pages": 6
  },
  "filter": {
    "stars": null
  }
}
```

#### 3.2.11 Lấy thống kê xếp hạng

**Endpoint**: `GET /products/{productId}/ratings/statistics`

**Response (200 OK):**
```json
{
  "product_id": "660e8400-e29b-41d4-a716-446655440000",
  "summary": {
    "avg_rating": 4.5,
    "total_ratings": 120
  },
  "distribution": {
    "5": {
      "count": 65,
      "percentage": 54.17
    },
    "4": {
      "count": 35,
      "percentage": 29.17
    },
    "3": {
      "count": 12,
      "percentage": 10.0
    },
    "2": {
      "count": 5,
      "percentage": 4.17
    },
    "1": {
      "count": 3,
      "percentage": 2.5
    }
  },
  "distribution_chart": [
    {
      "stars": 5,
      "count": 65,
      "percentage": 54.17
    },
    {
      "stars": 4,
      "count": 35,
      "percentage": 29.17
    },
    {
      "stars": 3,
      "count": 12,
      "percentage": 10.0
    },
    {
      "stars": 2,
      "count": 5,
      "percentage": 4.17
    },
    {
      "stars": 1,
      "count": 3,
      "percentage": 2.5
    }
  ]
}
```

**Error Response:**
- `400`: Sản phẩm không tìm thấy

#### 3.2.12 Lấy đánh giá của tài khoản

**Endpoint**: `GET /ratings/me`

**Headers:**
```
Authorization: Bearer {token}
```

**Query Parameters:**
- `page` (int, default: 1)
- `limit` (int, default: 20, max: 100)
- `sort` (string, optional: "newest", "oldest", "highest", "lowest")

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "710e8400-e29b-41d4-a716-446655440000",
      "product": {
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "name": "Cà phê Robusta",
        "category": {
          "id": "770e8400-e29b-41d4-a716-446655440000",
          "name": "Cà phê"
        }
      },
      "stars": 5,
      "comment": "Sản phẩm rất tốt, đóng gói cẩn thận",
      "created_at": "2025-12-23T11:00:00Z",
      "updated_at": "2025-12-23T11:00:00Z"
    },
    {
      "id": "730e8400-e29b-41d4-a716-446655440000",
      "product": {
        "id": "680e8400-e29b-41d4-a716-446655440000",
        "name": "Cà phê Arabica",
        "category": {
          "id": "770e8400-e29b-41d4-a716-446655440000",
          "name": "Cà phê"
        }
      },
      "stars": 4,
      "comment": "Hương vị thơm nhẹ, phù hợp uống buổi sáng",
      "created_at": "2025-12-20T09:15:00Z",
      "updated_at": "2025-12-20T09:15:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 15,
    "total_pages": 1
  },
  "summary": {
    "total_ratings": 15,
    "avg_stars_given": 4.3
  }
}
```

**Error Responses:**
- `401`: Token không thể xác thực

---

## Ghi chú

- Tất cả timestamp sử dụng định dạng ISO 8601 với timezone (TIMESTAMPTZ)
- UUID được sử dụng làm primary key cho tất cả các bảng
- Hệ thống hỗ trợ soft delete thông qua trường `deleted_at`
- Rating có giá trị từ 1-5 sao
- Average rating được tính tự động khi có rating mới/cập nhật/xóa
