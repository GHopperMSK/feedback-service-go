DROP DATABASE IF EXISTS feedback_service_test;
CREATE DATABASE feedback_service_test;

USE feedback_service_test;
CREATE TABLE feedbacks(
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