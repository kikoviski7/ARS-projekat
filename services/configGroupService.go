package services

import (
	"fmt"
	"projekat/model"
)

type ConfigGroupService struct {
	repo          model.ConfigRepository
	configService ConfigService
}

func NewConfigGroupService(repo model.ConfigRepository, service ConfigService) ConfigGroupService {
	return ConfigGroupService{
		repo:          repo,
		configService: service,
	}
}

func (s ConfigGroupService) AddGroup(config model.Config) {
	s.repo.Add(config)
}

func (s ConfigGroupService) GetGroup(name string, version int) (model.Config, error) {
	return s.repo.Get(name, version)
}

func (s ConfigGroupService) GetAllGroups() (map[string]model.Config, error) {
	return s.repo.GetAll()
}

func (s ConfigGroupService) PostGroup(name string, version int, params []map[string]any) error {
	configs, _ := s.configService.GetAll()

	data := make([]model.Config, 0)

	for _, param := range params {
		data = append(data, configs[fmt.Sprintf("%s/%d", param["name"], int(param["version"].(float64)))])
	}

	configGroup := model.ConfigGroup{
		Name:    name,
		Version: version,
		Configs: data,
	}
	s.repo.AddGroup(configGroup)
	return nil
}

func (s ConfigGroupService) DeleteGroupByVersion(name string, version int) error {
	s.repo.DeleteGroupByVersion(name, version)
	return nil
}
