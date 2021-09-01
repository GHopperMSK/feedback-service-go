CREATE DATABASE IF NOT EXISTS feedback_service;
USE feedback_service;
CREATE TABLE IF NOT EXISTS feedbacks(
    id INT NOT NULL AUTO_INCREMENT, 
    parent_id INT DEFAULT NULL,
    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,
    trade_id INT NOT NULL,
    message TEXT NOT NULL,
    type ENUM('positive', 'negative'),
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (id)
);

GRANT ALL PRIVILEGES ON *.* TO 'db_user'@'%' IDENTIFIED BY 'secret';