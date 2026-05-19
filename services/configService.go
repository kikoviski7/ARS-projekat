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

func (s ConfigService) GetAll() (map[string]model.Config, error) {
	return s.repo.GetAll()
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
	s.repo.DeleteByVersion(name, version)
	return nil
}
