# Sequence Diagrams - Product Review & Rating System

## 1. User Registration Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway (Traefik)
    participant US as User Service
    participant DB as Database

    U->>G: POST /api/users/register<br/>{email, password}
    G->>US: POST /api/users/register<br/>{email, password}
    US->>US: Validate input
    US->>DB: Check if email exists
    DB-->>US: Email not found
    US->>US: Hash password (bcrypt)
    US->>DB: Create user
    DB-->>US: User created
    US-->>G: 201 Created<br/>{id, email, role}
    G-->>U: 201 Created<br/>{id, email, role}
```

## 2. User Login Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant US as User Service
    participant DB as Database

    U->>G: POST /api/users/login<br/>{email, password}
    G->>US: POST /api/users/login<br/>{email, password}
    US->>DB: Find user by email
    DB-->>US: User found
    US->>US: Verify password (bcrypt)
    US->>US: Generate JWT token
    US-->>G: 200 OK<br/>{token, user}
    G-->>U: 200 OK<br/>{token, user}
```

## 3. Token Validation Flow (gRPC)

```mermaid
sequenceDiagram
    participant PS as Product Service
    participant US as User Service (gRPC)
    participant DB as Database

    PS->>US: Authenticate(token)
    US->>US: Validate JWT signature
    US->>US: Check token expiry
    US->>DB: Get user by ID
    DB-->>US: User info
    US->>US: Check user status
    US-->>PS: UserInfo{id, email, role}<br/>+ Valid
    Note over PS: Use user info in request context
```

## 4. Get Products (Authenticated)

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant US as User Service (gRPC)
    participant DB as Database

    U->>G: GET /api/products<br/>Authorization: Bearer {token}
    G->>PS: GET /api/products<br/>Authorization: Bearer {token}
    PS->>PS: Extract token
    PS->>US: Authenticate(token)
    US-->>PS: UserInfo{id, email, role}
    PS->>PS: Check permissions
    PS->>DB: Get products (paginated)
    DB-->>PS: Products list
    PS-->>G: 200 OK<br/>{products, pagination}
    G-->>U: 200 OK<br/>{products, pagination}
```

## 5. Search Products Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant R as Redis Cache
    participant DB as Database

    U->>G: GET /api/products/search?q=máy cày
    G->>PS: GET /api/products/search?q=máy cày
    PS->>PS: Normalize Vietnamese text<br/>(remove diacritics)
    PS->>R: Get cache(key: "search:máy cày")
    alt Cache Hit
        R-->>PS: Cached results
        PS-->>G: 200 OK<br/>{products}
        G-->>U: 200 OK<br/>{products}
    else Cache Miss
        R-->>PS: Cache miss
        PS->>DB: Search products<br/>(case-insensitive, diacritic-insensitive)
        DB-->>PS: Products list
        PS->>R: Set cache(key: "search:máy cày", TTL: 1h)
        PS-->>G: 200 OK<br/>{products}
        G-->>U: 200 OK<br/>{products}
    end
```

## 6. Create Rating Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant US as User Service (gRPC)
    participant R as Redis Cache
    participant DB as Database

    U->>G: POST /api/products/:id/ratings<br/>{rating: 5, comment: "..."}
    G->>PS: POST /api/products/:id/ratings<br/>{rating: 5, comment: "..."}
    PS->>US: Authenticate(token)
    US-->>PS: UserInfo{id, email, role}
    PS->>DB: Check existing rating<br/>(user_id, product_id)
    alt Rating Exists
        DB-->>PS: Existing rating found
        PS-->>G: 409 Conflict<br/>"Rating already exists"
        G-->>U: 409 Conflict
    else No Rating Exists
        DB-->>PS: No existing rating
        PS->>DB: Create rating
        DB-->>PS: Rating created
        Note over DB: Trigger updates<br/>average_rating
        PS->>R: Invalidate cache<br/>("product:{id}", "popular", "similar:{id}")
        PS-->>G: 201 Created<br/>{rating}
        G-->>U: 201 Created<br/>{rating}
    end
```

## 7. Update Rating Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant US as User Service (gRPC)
    participant R as Redis Cache
    participant DB as Database

    U->>G: PUT /api/ratings/:id<br/>{rating: 4, comment: "..."}
    G->>PS: PUT /api/ratings/:id<br/>{rating: 4, comment: "..."}
    PS->>US: Authenticate(token)
    US-->>PS: UserInfo{id, email, role}
    PS->>DB: Get rating by ID
    DB-->>PS: Rating found
    PS->>PS: Check ownership<br/>(user_id == rating.user_id)
    alt Not Owner
        PS-->>G: 403 Forbidden
        G-->>U: 403 Forbidden
    else Is Owner
        PS->>DB: Update rating
        DB-->>PS: Rating updated
        Note over DB: Trigger updates<br/>average_rating
        PS->>R: Invalidate cache<br/>("product:{id}", "popular")
        PS-->>G: 200 OK<br/>{rating}
        G-->>U: 200 OK<br/>{rating}
    end
```

## 8. Get Similar Products Flow (Cached)

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant R as Redis Cache
    participant DB as Database

    U->>G: GET /api/products/:id/similar
    G->>PS: GET /api/products/:id/similar
    PS->>R: Get cache(key: "similar:{product_id}")
    alt Cache Hit
        R-->>PS: Cached similar products
        PS-->>G: 200 OK<br/>{products}
        G-->>U: 200 OK<br/>{products}
    else Cache Miss
        R-->>PS: Cache miss
        PS->>DB: Get product category
        DB-->>PS: Category ID
        PS->>DB: Get products with same category<br/>(exclude current product)
        DB-->>PS: Similar products
        PS->>R: Set cache(key: "similar:{id}", TTL: 1h)
        PS-->>G: 200 OK<br/>{products}
        G-->>U: 200 OK<br/>{products}
    end
```

## 9. Get Popular Products Flow (Cached)

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant R as Redis Cache
    participant DB as Database

    U->>G: GET /api/products/popular?limit=10
    G->>PS: GET /api/products/popular?limit=10
    PS->>R: Get cache(key: "popular:10")
    alt Cache Hit
        R-->>PS: Cached popular products
        PS-->>G: 200 OK<br/>{products}
        G-->>U: 200 OK<br/>{products}
    else Cache Miss
        R-->>PS: Cache miss
        PS->>DB: Get products<br/>ORDER BY average_rating DESC<br/>LIMIT 10
        DB-->>PS: Popular products
        PS->>R: Set cache(key: "popular:10", TTL: 30m)
        PS-->>G: 200 OK<br/>{products}
        G-->>U: 200 OK<br/>{products}
    end
```

## 10. Admin - List All Users Flow

```mermaid
sequenceDiagram
    participant A as Admin
    participant G as Gateway
    participant US as User Service
    participant DB as Database

    A->>G: GET /api/users?page=1&limit=20&search=john
    G->>US: GET /api/users?page=1&limit=20&search=john
    US->>US: Validate admin role
    alt Not Admin
        US-->>G: 403 Forbidden
        G-->>A: 403 Forbidden
    else Is Admin
        US->>DB: Query users<br/>(pagination, search, filter)
        DB-->>US: Users list + total count
        US-->>G: 200 OK<br/>{users, pagination}
        G-->>A: 200 OK<br/>{users, pagination}
    end
```

## 11. Admin - Update User Role Flow

```mermaid
sequenceDiagram
    participant A as Admin
    participant G as Gateway
    participant US as User Service
    participant DB as Database

    A->>G: PUT /api/users/:id/role<br/>{role: "admin"}
    G->>US: PUT /api/users/:id/role<br/>{role: "admin"}
    US->>US: Validate admin role
    US->>DB: Get current user (admin)
    DB-->>US: Admin user info
    US->>US: Check if admin trying to<br/>change own role
    alt Self Role Change
        US-->>G: 400 Bad Request<br/>"Cannot change own role"
        G-->>A: 400 Bad Request
    else Valid Request
        US->>DB: Update user role
        DB-->>US: User updated
        US-->>G: 200 OK<br/>{user}
        G-->>A: 200 OK<br/>{user}
    end
```

## 12. Delete Rating Flow

```mermaid
sequenceDiagram
    participant U as User
    participant G as Gateway
    participant PS as Product Service
    participant US as User Service (gRPC)
    participant R as Redis Cache
    participant DB as Database

    U->>G: DELETE /api/ratings/:id
    G->>PS: DELETE /api/ratings/:id
    PS->>US: Authenticate(token)
    US-->>PS: UserInfo{id, email, role}
    PS->>DB: Get rating by ID
    DB-->>PS: Rating found
    PS->>PS: Check permissions<br/>(owner or admin)
    alt Not Authorized
        PS-->>G: 403 Forbidden
        G-->>U: 403 Forbidden
    else Authorized
        PS->>DB: Soft delete rating<br/>(set deleted_at)
        DB-->>PS: Rating deleted
        Note over DB: Trigger updates<br/>average_rating
        PS->>R: Invalidate cache<br/>("product:{id}", "popular")
        PS-->>G: 200 OK
        G-->>U: 200 OK
    end
```
