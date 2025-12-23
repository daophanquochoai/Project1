-- 1) Create databases
CREATE DATABASE user_service;
CREATE DATABASE product_service;

-- 2) Create users
CREATE USER user_service WITH PASSWORD 'user_service_pwd';
CREATE USER product_service WITH PASSWORD 'product_service_pwd';

-- 3) Grant privileges
GRANT ALL PRIVILEGES ON DATABASE user_service TO user_service;
GRANT ALL PRIVILEGES ON DATABASE product_service TO product_service;
