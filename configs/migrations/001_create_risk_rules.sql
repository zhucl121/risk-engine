-- Migration 001: Create risk_rules table
-- Compatible with MySQL 8.0+ and PostgreSQL 14+

CREATE TABLE IF NOT EXISTS risk_rules (
    id               BIGINT       NOT NULL AUTO_INCREMENT,
    rule_key         VARCHAR(128) NOT NULL,           -- e.g. "DEVICE_MULTI_ACCOUNT_001"
    name             VARCHAR(256) NOT NULL,
    group_name       VARCHAR(128) NOT NULL,           -- e.g. "payment_rules"
    scene_code       VARCHAR(128) NOT NULL,           -- e.g. "PAYMENT_CHECKOUT"
    priority         INT          NOT NULL DEFAULT 100,
    condition_dsl    TEXT         NOT NULL,           -- DSL string compiled by pkg/dsl
    condition_ast    JSON                  DEFAULT NULL, -- visual builder JSON (nullable)
    action_decision  VARCHAR(32)  NOT NULL,           -- REJECT | MANUAL_REVIEW | PASS
    action_risk_code VARCHAR(128)          DEFAULT NULL,
    action_score     INT          NOT NULL DEFAULT 0,
    status           TINYINT      NOT NULL DEFAULT 1, -- 1=enabled, 0=disabled
    version          INT          NOT NULL DEFAULT 1, -- optimistic lock counter
    created_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
                                          ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uk_rule_key (rule_key),
    KEY idx_scene_status (scene_code, status),        -- hot-path query filter
    KEY idx_updated_at   (updated_at)                 -- hot-reload watcher range scan
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed: migrate existing payment_rules.yaml conditions
-- Run manually after schema creation if you want to port existing YAML rules.
INSERT IGNORE INTO risk_rules
    (rule_key, name, group_name, scene_code, priority, condition_dsl,
     action_decision, action_risk_code, action_score)
VALUES
    ('HISTORY_FRAUD_001',           '历史欺诈用户',           'payment_rules', 'PAYMENT_CHECKOUT', 300,
     'features[''user.history_fraud_count''] > 0',
     'REJECT',        'USER_HISTORY_FRAUD',          950),
    ('DEVICE_MULTI_ACCOUNT_001',    '设备关联多账号检测',     'payment_rules', 'PAYMENT_CHECKOUT', 200,
     'features[''device.linked_account_count_7d''] > 5 && features[''user.register_days''] < 30',
     'REJECT',        'DEVICE_MULTI_ACCOUNT',        850),
    ('HIGH_VELOCITY_PAY_001',       '短时高频支付检测',       'payment_rules', 'PAYMENT_CHECKOUT', 190,
     'features[''velocity.pay_count_1min''] > 5',
     'REJECT',        'HIGH_VELOCITY_PAYMENT',       900),
    ('DATACENTER_IP_LARGE_AMT_001', '数据中心IP大额交易',     'payment_rules', 'PAYMENT_CHECKOUT', 180,
     'features[''ip.is_datacenter''] == true && amount > 100000',
     'MANUAL_REVIEW', 'DATACENTER_IP_HIGH_AMOUNT',   700),
    ('NEW_DEVICE_HIGH_AMT_001',     '新设备大额交易',         'payment_rules', 'PAYMENT_CHECKOUT', 170,
     'features[''user.register_days''] < 7 && amount > 50000',
     'MANUAL_REVIEW', 'NEW_USER_HIGH_AMOUNT',         650);
