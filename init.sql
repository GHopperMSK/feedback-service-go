CREATE DATABASE IF NOT EXISTS feedback_service;
USE feedback_service;

CREATE TABLE IF NOT EXISTS feedbacks(
    id INT NOT NULL AUTO_INCREMENT,
    parent_id INT DEFAULT NULL,
    sender_uuid BINARY(16) NOT NULL,
    sender_name VARCHAR(64) NOT NULL,
    sender_avater VARCHAR(128) NOT NULL,
    receiver_uuid BINARY(16) NOT NULL,
    receiver_name VARCHAR(16) NOT NULL,
    receiver_avater VARCHAR(128) NOT NULL,
    offer_hash CHAR(11) NOT NULL,
    offer_authorized BOOL NOT NULL,
    offer_owner_uuid BINARY(16) NOT NULL,
    offer_type ENUM('BUY', 'SELL'),
    offer_payment_method VARCHAR(64) NOT NULL,
    offer_payment_method_slug VARCHAR(64) NOT NULL,
    offer_fiat_code ENUM('USD', 'EUR', 'RUB', 'PLN', 'CNY', 'VES', 'NGN'),
    offer_crypto_code VARCHAR(12),
    offer_deleted_at TIMESTAMP NULL DEFAULT NULL,
    trade_hash CHAR(11) NOT NULL,
    trade_fiat_amount_requested_in_usd DECIMAL(10, 2),
    trade_status ENUM('RELEASED', 'CANCELLED', 'DISPUTED'),
    message TEXT NOT NULL,
    feedback_type ENUM('POSITIVE', 'NEGATIVE'),
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (parent_id) REFERENCES feedbacks (id) ON DELETE CASCADE,
    INDEX offer_hash_idx (offer_hash),
    INDEX trade_hash_idx (trade_hash),
    INDEX sender_receiver_payment_method_fiat_code_idx (sender_uuid, receiver_uuid, offer_payment_method_slug, offer_fiat_code)
);

CREATE TABLE IF NOT EXISTS feedback_stats(
    user_uuid BINARY(16) NOT NULL,
    positive INT DEFAULT 0,
    negative INT DEFAULT 0,
    initial INT DEFAULT 0
);
CREATE UNIQUE INDEX feedback_stats_user_id_uq ON feedback_stats (user_uuid);
