package service

import "github.com/nika/soccer-manager-api/repository"

type Service struct {
	Repo *repository.DB
}

func NewService(repo *repository.DB) *Service {
	return &Service{Repo: repo}
}
