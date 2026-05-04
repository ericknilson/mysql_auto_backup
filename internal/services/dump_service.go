package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jamf/go-mysqldump"
)

type DumpService struct{}

func NewDumpService() *DumpService { return &DumpService{} }

// Dump gera um arquivo .sql para as credenciais informadas e retorna o caminho.
// O caller é responsável por remover o arquivo após o uso.
func (s *DumpService) Dump(name string, creds Credentials) (string, error) {
	db, err := sql.Open("mysql", creds.DSN())
	if err != nil {
		return "", fmt.Errorf("abrir conexão alvo: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return "", fmt.Errorf("ping no banco alvo: %w", err)
	}

	ts := time.Now().UTC().Format("20060102_150405")
	fname := fmt.Sprintf("%s_%s.sql", sanitize(name), ts)
	fpath := filepath.Join(os.TempDir(), fname)

	f, err := os.Create(fpath)
	if err != nil {
		return "", err
	}

	dump := &mysqldump.Data{
		Out:        f,
		Connection: db,
	}
	if err := dump.Dump(); err != nil {
		f.Close()
		os.Remove(fpath)
		return "", fmt.Errorf("dump: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(fpath)
		return "", err
	}
	return fpath, nil
}

var safeChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "connection"
	}
	return safeChars.ReplaceAllString(s, "_")
}
