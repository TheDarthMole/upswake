package repository

import "github.com/TheDarthMole/UPSWake/internal/domain/entity"

//go:generate mockgen -package mocks -source config.go -destination mocks/config_mock.go ConfigRepository

type ConfigRepository interface {
	Load() (*entity.Config, error)
}
