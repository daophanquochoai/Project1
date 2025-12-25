\connect product_service;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- categories
CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- products
CREATE TABLE IF NOT EXISTS products (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL,
    description    TEXT,
    search_name    TEXT NOT NULL,
    price          NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    category_id    UUID NOT NULL REFERENCES categories(id),
    average_rating NUMERIC(3,2) NOT NULL DEFAULT 0,
    total_ratings  INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
    );

-- ratings
CREATE TABLE IF NOT EXISTS ratings (
                                       id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    user_id    UUID NOT NULL, -- from User Service (no FK across DB)
    rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

-- product_related (many-to-many)
CREATE TABLE IF NOT EXISTS product_related (
    product_id     UUID NOT NULL REFERENCES products(id),
    related_id     UUID NOT NULL REFERENCES products(id),
    relation_type  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, related_id, relation_type)
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_average_rating ON products(average_rating);
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);

CREATE INDEX IF NOT EXISTS idx_ratings_product_id ON ratings(product_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);

-- UNIQUE: 1 user rate 1 product (soft-delete aware)
CREATE UNIQUE INDEX IF NOT EXISTS idx_ratings_user_product
    ON ratings(user_id, product_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_related_product_id ON product_related(product_id);
CREATE INDEX IF NOT EXISTS idx_product_related_related_id ON product_related(related_id);

-- updated_at auto
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;
CREATE TRIGGER trg_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_ratings_updated_at ON ratings;
CREATE TRIGGER trg_ratings_updated_at
    BEFORE UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Recompute avg + count for a product (ignore soft-deleted ratings)
CREATE OR REPLACE FUNCTION recompute_product_rating(p_product_id UUID)
RETURNS VOID AS $$
BEGIN
UPDATE products p
SET
    average_rating = COALESCE((
                                  SELECT ROUND(AVG(r.rating)::numeric, 2)
                                  FROM ratings r
                                  WHERE r.product_id = p_product_id
                                    AND r.deleted_at IS NULL
                              ), 0),
    total_ratings = COALESCE((
                                 SELECT COUNT(*)
                                 FROM ratings r
                                 WHERE r.product_id = p_product_id
                                   AND r.deleted_at IS NULL
                             ), 0),
    updated_at = NOW()
WHERE p.id = p_product_id;
END;
$$ LANGUAGE plpgsql;

-- Trigger on ratings: insert/update/delete (soft delete)
CREATE OR REPLACE FUNCTION trg_ratings_recompute()
RETURNS TRIGGER AS $$
DECLARE
v_product_id UUID;
BEGIN
  v_product_id := COALESCE(NEW.product_id, OLD.product_id);
  PERFORM recompute_product_rating(v_product_id);
RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ratings_after_change ON ratings;
CREATE TRIGGER trg_ratings_after_change
    AFTER INSERT OR UPDATE OR DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION trg_ratings_recompute();

INSERT INTO public.categories (id,name,description,created_at,updated_at) VALUES
                                                                              ('550e8400-e29b-41d4-a716-446655440001'::uuid,'Điện thoại & Phụ kiện','Điện thoại, tai nghe, sạc dự phòng','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440002'::uuid,'Máy tính & Laptop','Laptop, PC, linh kiện máy tính','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440003'::uuid,'Thiết bị điện tử','Tivi, loa, camera, thiết bị thông minh','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440004'::uuid,'Thời trang nam','Quần áo, giày dép nam','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440005'::uuid,'Thời trang nữ','Quần áo, giày dép nữ','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440006'::uuid,'Mẹ & Bé','Đồ dùng cho mẹ và bé','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440007'::uuid,'Nhà cửa & Đời sống','Nội thất, đồ gia dụng','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440008'::uuid,'Sắc đẹp','Mỹ phẩm, chăm sóc da','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440009'::uuid,'Sức khỏe','Thực phẩm chức năng, dụng cụ y tế','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07'),
                                                                              ('550e8400-e29b-41d4-a716-446655440010'::uuid,'Thể thao & Du lịch','Dụng cụ thể thao, phụ kiện du lịch','2025-12-24 10:30:44.236844+07','2025-12-24 10:30:44.236844+07');



INSERT INTO public.products (id,name,description,price,category_id,average_rating,total_ratings,created_at,updated_at,deleted_at,search_name) VALUES
                                                                                                                                                  ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'iPhone 15 Pro Max 256GB','iPhone 15 Pro Max mới nhất, chip A17 Pro, camera 48MP',29990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.197514+07',NULL,'iphone 15 pro max 256gb'),
                                                                                                                                                  ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'Samsung Galaxy S24 Ultra 512GB','Galaxy S24 Ultra với bút S-Pen, màn hình 6.8 inch',31990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.213677+07',NULL,'samsung galaxy s24 ultra 512gb'),
                                                                                                                                                  ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'Xiaomi 14 Pro 12GB/512GB','Xiaomi 14 Pro Snapdragon 8 Gen 3, sạc nhanh 120W',18990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.218153+07',NULL,'xiaomi 14 pro 12gb/512gb'),
                                                                                                                                                  ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'OPPO Find X7 Ultra','OPPO Find X7 Ultra camera Hasselblad, zoom 6x',24990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.222576+07',NULL,'oppo find x7 ultra'),
                                                                                                                                                  ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'Vivo V30 Pro 5G','Vivo V30 Pro camera selfie 50MP, pin 5000mAh',12990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.226619+07',NULL,'vivo v30 pro 5g'),
                                                                                                                                                  ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'Tai nghe AirPods Pro 2','AirPods Pro thế hệ 2, chống ồn chủ động ANC',6490000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.234653+07',NULL,'tai nghe airpods pro 2'),
                                                                                                                                                  ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'Tai nghe Samsung Galaxy Buds 2 Pro','Galaxy Buds 2 Pro chống ồn, âm thanh 360',3990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.239569+07',NULL,'tai nghe samsung galaxy buds 2 pro'),
                                                                                                                                                  ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'Sạc dự phòng Anker 20000mAh','Pin sạc dự phòng Anker 20000mAh, sạc nhanh PD 20W',890000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.244099+07',NULL,'sac du phong anker 20000mah'),
                                                                                                                                                  ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'Ốp lưng iPhone 15 Pro Max Silicone','Ốp lưng chính hãng Apple Silicone mềm mại',1290000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.248616+07',NULL,'op lung iphone 15 pro max silicone'),
                                                                                                                                                  ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'Củ sạc nhanh 65W GaN','Củ sạc GaN 65W nhỏ gọn, sạc 3 cổng cùng lúc',590000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.258114+07',NULL,'cu sac nhanh 65w gan');
INSERT INTO public.products (id,name,description,price,category_id,average_rating,total_ratings,created_at,updated_at,deleted_at,search_name) VALUES
                                                                                                                                                  ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'Kính cường lực iPhone 15 Pro','Kính cường lực full màn hình, độ cứng 9H',150000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.262432+07',NULL,'kinh cuong luc iphone 15 pro'),
                                                                                                                                                  ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'Giá đỡ điện thoại ô tô Baseus','Giá đỡ điện thoại ô tô gắn cửa gió',190000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.266736+07',NULL,'gia do dien thoai o to baseus'),
                                                                                                                                                  ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'Realme 11 Pro 5G 8GB/256GB','Realme 11 Pro màn hình AMOLED 120Hz',8990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.270708+07',NULL,'realme 11 pro 5g 8gb/256gb'),
                                                                                                                                                  ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'Nokia G60 5G','Nokia G60 5G bền bỉ, pin 4500mAh',5990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.274342+07',NULL,'nokia g60 5g'),
                                                                                                                                                  ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'Poco X6 Pro 5G','Poco X6 Pro chip Dimensity 8300, màn hình 120Hz',7990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.27812+07',NULL,'poco x6 pro 5g'),
                                                                                                                                                  ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'OnePlus 12R','OnePlus 12R Snapdragon 8 Gen 2, sạc 100W',15990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.282164+07',NULL,'oneplus 12r'),
                                                                                                                                                  ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'Motorola Edge 40 Pro','Motorola Edge 40 Pro màn hình cong, camera 50MP',11990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.286083+07',NULL,'motorola edge 40 pro'),
                                                                                                                                                  ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'Google Pixel 8 Pro','Google Pixel 8 Pro AI camera, chip Tensor G3',25990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.28998+07',NULL,'google pixel 8 pro'),
                                                                                                                                                  ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'Asus ROG Phone 8','ROG Phone 8 gaming phone, tản nhiệt tốt',22990000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,0.00,0,'2025-12-24 10:36:04.895171+07','2025-12-24 10:38:11.294703+07',NULL,'asus rog phone 8'),
                                                                                                                                                  ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'Cáp sạc Type-C 100W Ugreen','Cáp sạc Type-C to Type-C 100W, dài 2m',290000.00,'550e8400-e29b-41d4-a716-446655440001'::uuid,2.00,1,'2025-12-24 10:36:04.895171+07','2025-12-25 11:47:00.533996+07',NULL,'cap sac type-c 100w ugreen');



INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:40:45.377523+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:40:45.377523+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('42547533-b854-47fa-936f-804e1aa42f4e'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'35d65544-0849-4ad7-9715-099a11099b42'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('35d65544-0849-4ad7-9715-099a11099b42'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('ad7bb8bd-cac0-4c2d-a2b3-483a133033a8'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'f4916aa8-0f67-45bb-b66f-bb69180671e8'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'abde1ace-1a1e-4b3e-8da7-d0119a04ee5d'::uuid,'similar','2025-12-24 10:41:13.893028+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('e0b3ed16-d6d7-41f5-951b-cbbd2feaf966'::uuid,'fe4d2de1-8841-4040-a1c7-7c1ee7bdc253'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'d9d8537f-3487-4891-aa9d-9556be8df973'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'a7306261-e49e-4146-a999-50a09c68e69b'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'1bce31a9-9c74-4ac5-9685-b53d5b192d39'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('afa08215-a6f2-48b3-a559-6edf8ce74e4a'::uuid,'76b510d1-579d-4d4e-a5e7-489e37b87358'::uuid,'similar','2025-12-24 10:41:13.893028+07'),
                                                                                        ('efcaeff2-060d-4a43-a317-2891ee7445a6'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'related','2025-12-24 10:40:45.404707+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'b16b776e-8291-4b21-ac56-222245bbed2b'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'06cf4e2b-f4ef-4da5-a014-fde741dd670f'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('ff3e718a-51d6-4f61-a272-330bc8868f3c'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'related','2025-12-24 10:40:45.404707+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'related','2025-12-24 10:40:45.413722+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'related','2025-12-24 10:40:45.413722+07'),
                                                                                        ('49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'related','2025-12-24 10:40:45.413722+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'49bb6c47-97ea-4d00-b228-59e94bd12bc7'::uuid,'related','2025-12-24 10:40:45.413722+07');
INSERT INTO public.product_related (product_id,related_id,relation_type,created_at) VALUES
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'266011b5-c93d-415e-aa8f-b949671aa0fb'::uuid,'related','2025-12-24 10:40:45.413722+07'),
                                                                                        ('f169add5-3586-4d89-85c9-4f545854ce28'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'related','2025-12-24 10:40:45.413722+07');


INSERT INTO public.ratings (id,product_id,user_id,rating,"comment",created_at,updated_at,deleted_at) VALUES
    ('bdd8c5de-26e3-4eb5-aea7-6cc8affe4e79'::uuid,'ed8ced0a-f1dd-405a-93a8-9d3b8b3266b4'::uuid,'2294c66d-79df-4d99-974e-6cac94bc8796'::uuid,2,'Cập nhật đánh giá sau khi dùng lâu hơn','2025-12-24 17:30:05.74572+07','2025-12-25 11:47:00.533996+07',NULL);
