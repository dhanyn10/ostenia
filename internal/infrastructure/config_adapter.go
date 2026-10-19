package infrastructure

import (
	"ostenia/internal/config"
	"ostenia/internal/domain"
)

type ConfigAdapter struct{}

func NewConfigAdapter() *ConfigAdapter {
	return &ConfigAdapter{}
}

func (a *ConfigAdapter) LoadConfig() (*domain.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return (*domain.Config)(cfg), nil
}

func (a *ConfigAdapter) SaveConfig(cfg *domain.Config) error {
	c := (*config.Config)(cfg)
	return config.SaveConfig(c)
}

func (a *ConfigAdapter) GetBaseDir() string {
	return config.GetBaseDir()
}
