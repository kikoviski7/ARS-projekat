package services

import (
	"projekat/model"
)

type ConfigService struct {
	repo model.ConfigRepository
}

func NewConfigService(repo model.ConfigRepository) ConfigService {
	return ConfigService{
		repo: repo,
	}
}

func (s ConfigService) Add(config model.Config) {
	s.repo.Add(config)
}

func (s ConfigService) Get(name string, version int) (model.Config, error) {
	return s.repo.Get(name, version)
}

func (s *ConfigService) GetByName(name string) ([]model.Config, error) {
	return s.repo.GetByName(name)
}

func (s ConfigService) GetAll() (map[string]model.Config, error) {
	return s.repo.GetAll()
}

func (s ConfigService) Put(
	config model.Config,
	oldName string,
	oldVersion int,
) error {
	return s.repo.Put(config, oldName, oldVersion)
}

func (s ConfigService) Post(name string, version int, params map[string]string) error {
	config := model.Config{
		Name:    name,
		Version: version,
		Params:  params,
	}
	s.repo.Add(config)
	return nil
}

func (s ConfigService) DeleteByVerison(name string, version int) error {
	_, err := s.repo.DeleteByVersion(name, version)
	return err
}
