package ports

import "yogo/internal/domain"

type ConfigService interface {
	Load() (domain.Config, error)
}
