CREATE TABLE backup_logs (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  connection_id  BIGINT UNSIGNED NOT NULL,
  file_name      VARCHAR(512) NOT NULL,
  status         VARCHAR(20) NOT NULL,
  error_message  TEXT NULL,
  sent_at        DATETIME NOT NULL,
  created_at     DATETIME NULL,
  updated_at     DATETIME NULL,
  KEY idx_backup_logs_connection_id (connection_id),
  KEY idx_backup_logs_sent_at (sent_at),
  CONSTRAINT fk_backup_logs_connection FOREIGN KEY (connection_id)
    REFERENCES connections(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
