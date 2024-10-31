package repository

import "github.com/TheDarthMole/UPSWake/internal/domain/entity"

type ConfigRepository interface {
	Load() (*entity.Config, error)
	GetNutServers() ([]entity.NutServer, error)
}
