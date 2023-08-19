package service

import (
	"database/sql"
)

type RepositoryService interface {
	LoadCount() (any, error)
}

type repositoryService struct {
	db *sql.DB
}

func NewRepositoryService(db *sql.DB) RepositoryService {
	return &repositoryService{db: db}
}

func (b *repositoryService) LoadCount() (any, error) {
	var count int
	err := b.db.QueryRow("SELECT count FROM counter").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
