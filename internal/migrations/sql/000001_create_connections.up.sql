CREATE TABLE connections (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(255) NOT NULL,
  host        TEXT NOT NULL,
  port        TEXT NOT NULL,
  `user`      TEXT NOT NULL,
  password    TEXT NOT NULL,
  `database`  TEXT NOT NULL,
  is_active   TINYINT(1) NOT NULL DEFAULT 1,
  created_at  DATETIME NULL,
  updated_at  DATETIME NULL,
  deleted_at  DATETIME NULL,
  KEY idx_connections_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
