package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/erick_nilson/mysql_auto_backup/internal/crypto"
	"github.com/erick_nilson/mysql_auto_backup/internal/models"
	"github.com/erick_nilson/mysql_auto_backup/internal/repository"
)

var ErrConnectionUnreachable = errors.New("não foi possível conectar ao banco com as credenciais fornecidas")

type ConnectionService struct {
	repo   *repository.ConnectionRepository
	cipher *crypto.Cipher
}

func NewConnectionService(repo *repository.ConnectionRepository, cipher *crypto.Cipher) *ConnectionService {
	return &ConnectionService{repo: repo, cipher: cipher}
}

type DecryptedConnection struct {
	ID    uint64
	Name  string
	Creds Credentials
}

type CreateInput struct {
	Name     string
	Host     string
	Port     string
	User     string
	Password string
	Database string
	IsActive *bool
}

type UpdateInput struct {
	Name     *string
	Host     *string
	Port     *string
	User     *string
	Password *string
	Database *string
	IsActive *bool
}

func (s *ConnectionService) TestConnection(ctx context.Context, creds Credentials) error {
	db, err := sql.Open("mysql", creds.DSN())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionUnreachable, err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionUnreachable, err)
	}
	return nil
}

func (s *ConnectionService) Create(ctx context.Context, in CreateInput) (*models.Connection, error) {
	creds := Credentials{
		Host: in.Host, Port: in.Port, User: in.User,
		Password: in.Password, Database: in.Database,
	}
	if err := s.TestConnection(ctx, creds); err != nil {
		return nil, err
	}

	enc, err := s.encryptCreds(creds)
	if err != nil {
		return nil, err
	}

	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}

	c := &models.Connection{
		Name:     in.Name,
		Host:     enc.Host,
		Port:     enc.Port,
		User:     enc.User,
		Password: enc.Password,
		Database: enc.Database,
		IsActive: active,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ConnectionService) Update(ctx context.Context, id uint64, in UpdateInput) (*models.Connection, error) {
	current, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	creds, err := s.decryptCreds(current)
	if err != nil {
		return nil, err
	}

	if in.Host != nil {
		creds.Host = *in.Host
	}
	if in.Port != nil {
		creds.Port = *in.Port
	}
	if in.User != nil {
		creds.User = *in.User
	}
	if in.Password != nil {
		creds.Password = *in.Password
	}
	if in.Database != nil {
		creds.Database = *in.Database
	}

	if err := s.TestConnection(ctx, creds); err != nil {
		return nil, err
	}

	enc, err := s.encryptCreds(creds)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		current.Name = *in.Name
	}
	if in.IsActive != nil {
		current.IsActive = *in.IsActive
	}
	current.Host = enc.Host
	current.Port = enc.Port
	current.User = enc.User
	current.Password = enc.Password
	current.Database = enc.Database

	if err := s.repo.Update(current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *ConnectionService) Delete(id uint64) error {
	return s.repo.Delete(id)
}

func (s *ConnectionService) Get(id uint64) (*models.Connection, error) {
	return s.repo.FindByID(id)
}

func (s *ConnectionService) List() ([]models.Connection, error) {
	return s.repo.List()
}

func (s *ConnectionService) ListActiveDecrypted() ([]DecryptedConnection, error) {
	conns, err := s.repo.ListActive()
	if err != nil {
		return nil, err
	}
	return s.decryptAll(conns)
}

func (s *ConnectionService) FindActiveDecryptedByIDs(ids []uint64) ([]DecryptedConnection, error) {
	conns, err := s.repo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	active := make([]models.Connection, 0, len(conns))
	for _, c := range conns {
		if c.IsActive {
			active = append(active, c)
		}
	}
	return s.decryptAll(active)
}

func (s *ConnectionService) decryptAll(conns []models.Connection) ([]DecryptedConnection, error) {
	out := make([]DecryptedConnection, 0, len(conns))
	for _, c := range conns {
		creds, err := s.decryptCreds(&c)
		if err != nil {
			return nil, fmt.Errorf("decriptar conexão %d (%s): %w", c.ID, c.Name, err)
		}
		out = append(out, DecryptedConnection{ID: c.ID, Name: c.Name, Creds: creds})
	}
	return out, nil
}

func (s *ConnectionService) encryptCreds(creds Credentials) (Credentials, error) {
	enc := Credentials{}
	var err error
	if enc.Host, err = s.cipher.Encrypt(creds.Host); err != nil {
		return enc, err
	}
	if enc.Port, err = s.cipher.Encrypt(creds.Port); err != nil {
		return enc, err
	}
	if enc.User, err = s.cipher.Encrypt(creds.User); err != nil {
		return enc, err
	}
	if enc.Password, err = s.cipher.Encrypt(creds.Password); err != nil {
		return enc, err
	}
	if enc.Database, err = s.cipher.Encrypt(creds.Database); err != nil {
		return enc, err
	}
	return enc, nil
}

func (s *ConnectionService) decryptCreds(c *models.Connection) (Credentials, error) {
	out := Credentials{}
	var err error
	if out.Host, err = s.cipher.Decrypt(c.Host); err != nil {
		return out, err
	}
	if out.Port, err = s.cipher.Decrypt(c.Port); err != nil {
		return out, err
	}
	if out.User, err = s.cipher.Decrypt(c.User); err != nil {
		return out, err
	}
	if out.Password, err = s.cipher.Decrypt(c.Password); err != nil {
		return out, err
	}
	if out.Database, err = s.cipher.Decrypt(c.Database); err != nil {
		return out, err
	}
	return out, nil
}
