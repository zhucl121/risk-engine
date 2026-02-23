-- Migration 002: Create scene_extra_params table
-- Stores the parameter specification for each Extra field per scene.
-- Compatible with MySQL 8.0+ and PostgreSQL 14+

CREATE TABLE IF NOT EXISTS scene_extra_params (
    id           BIGINT       NOT NULL AUTO_INCREMENT,
    scene_code   VARCHAR(128) NOT NULL,            -- scene identifier, e.g. "PAYMENT_CHECKOUT"
    param_key    VARCHAR(128) NOT NULL,            -- Extra field name, e.g. "merchant_id"
    param_type   VARCHAR(16)  NOT NULL DEFAULT 'string', -- string | int | float | bool
    required     TINYINT      NOT NULL DEFAULT 0,  -- 1=required, 0=optional
    default_val  VARCHAR(512)          DEFAULT NULL, -- default value for optional fields (stored as string)
    description  VARCHAR(512)          DEFAULT NULL,
    status       TINYINT      NOT NULL DEFAULT 1,  -- 1=active, 0=disabled
    version      INT          NOT NULL DEFAULT 1,  -- optimistic lock counter
    created_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
                                      ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uk_scene_param (scene_code, param_key),  -- one spec per (scene, key)
    KEY idx_scene_status (scene_code, status),           -- hot-path query filter
    KEY idx_updated_at   (updated_at)                    -- hot-reload watcher range scan
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Example seed data for the payment scene
INSERT IGNORE INTO scene_extra_params
    (scene_code, param_key, param_type, required, default_val, description)
VALUES
    ('payment', 'merchant_id',  'string', 1, NULL,  '商户 ID，必填'),
    ('payment', 'product_type', 'string', 0, 'GOODS', '商品类型，默认 GOODS'),
    ('payment', 'amount_usd',   'float',  0, '0.0',  '美元金额，可选'),
    ('payment', 'is_recurring', 'bool',   0, 'false','是否周期扣款，默认否'),
    ('payment', 'order_count',  'int',    0, '0',    '历史订单数，默认 0');
