package migration

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Service struct {
	migrate *migrate.Migrate
}

func NewService(databaseURL string, migrationsPath string) (*Service, error) {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		databaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return &Service{migrate: m}, nil
}

func (s *Service) Up() error {
	if err := s.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

func (s *Service) Down() error {
	if err := s.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations down: %w", err)
	}
	return nil
}

func (s *Service) Status() (uint, bool, error) {
	version, dirty, err := s.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

func (s *Service) Close() error {
	sourceErr, dbErr := s.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}
