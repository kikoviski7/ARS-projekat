package repositories

import (
	"errors"
	"fmt"
	"projekat/model"
)

type ConfigInMemRepository struct {
	configs map[string]model.Config
}

// todo: dodaj implementaciju metoda iz interfejsa ConfigRepository

func NewConfigInMemRepository() model.ConfigRepository {
	return ConfigInMemRepository{
		configs: make(map[string]model.Config),
	}
}

// Add implements model.ConfigRepository.
func (c ConfigInMemRepository) Add(config model.Config) {
	key := fmt.Sprintf("%s/%d", config.Name, config.Version)
	c.configs[key] = config
}

// Get implements model.ConfigRepository.
func (c ConfigInMemRepository) Get(name string, version int) (model.Config, error) {
	key := fmt.Sprintf("%s/%d", name, version)
	config, ok := c.configs[key]
	if !ok {
		return model.Config{}, errors.New("config not found")
	}
	return config, nil
}

func (c ConfigInMemRepository) GetAll() (map[string]model.Config, error) {
	return c.configs, nil
}

func (c ConfigInMemRepository) DeleteByVersion(name string, version int) (model.Config, error) {
	key := fmt.Sprintf("%s/%d", name, version)
	config := c.configs[key]
	delete(c.configs, key)
	return config, nil
}

func (c ConfigInMemRepository) AddGroup(configGroup model.ConfigGroup) {

}

func (c ConfigInMemRepository) DeleteGroupByVersion(name string, version int) (model.ConfigGroup, error) {
	return model.ConfigGroup{}, nil
}

// func (c ConfigInMemRepository) DeleteByName(name string) (model.Config, error) {
// 	delete(c.configs, name)
// }
