package services

import "fmt"

type Credentials struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func (c Credentials) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&timeout=5s&readTimeout=5s&writeTimeout=5s",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}
