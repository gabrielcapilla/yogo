package config

import (
	"errors"
	"os"
	"path/filepath"
	"yogo/internal/domain"
	"yogo/internal/logger"
	"yogo/internal/ports"

	"github.com/spf13/viper"
)

type ViperConfigService struct{}

func NewViperConfigService() ports.ConfigService {
	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Log.Warn().Err(err).Msg("Could not find user config directory, using current directory")
	}

	if configDir != "" {
		yogoConfigDir := filepath.Join(configDir, "yogo")
		if err := os.MkdirAll(yogoConfigDir, 0755); err != nil {
			logger.Log.Error().Err(err).Msg("Could not create yogo config directory")
		} else {
			viper.AddConfigPath(yogoConfigDir)
		}
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.SetDefault("cookiesPath", "")
	viper.SetDefault("historyLimit", 50)

	return &ViperConfigService{}
}

func (s *ViperConfigService) Load() (domain.Config, error) {
	var cfg domain.Config

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			logger.Log.Info().Msg("Config file not found, creating with default values.")
			if err := viper.SafeWriteConfig(); err != nil {
				return cfg, err
			}
		} else {
			return cfg, err
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
