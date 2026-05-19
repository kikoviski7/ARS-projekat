package services

import (
	"projekat/model"
)

type ConfigGroupService struct {
	repo model.ConfigRepository
}

func NewConfigGroupService(repo model.ConfigRepository) ConfigGroupService {
	return ConfigGroupService{
		repo: repo,
	}
}

// READ ONE
func (s ConfigGroupService) GetGroup(
	name string,
	version int,
) (model.ConfigGroup, error) {

	return s.repo.GetGroup(name, version)
}

func (s ConfigGroupService) GetAllGroups() (
	map[string]model.ConfigGroup,
	error,
) {

	return s.repo.GetAllGroups()
}

func (s ConfigGroupService) PostGroup(
	name string,
	version int,
	configs []model.Config,
) error {

	group := model.ConfigGroup{
		Name:    name,
		Version: version,
		Configs: configs,
	}

	for _, config := range configs {
		s.repo.Add(config)
	}
	s.repo.AddGroup(group)

	return nil
}

func (s ConfigGroupService) DeleteGroupByVersion(
	name string,
	version int,
) error {

	_, err := s.repo.DeleteGroupByVersion(name, version)

	return err
}

func (s ConfigGroupService) DeleteConfigByVersion(
	groupName string,
	groupVersion int,
	configName string,
	configVersion int,
) error {

	group, err := s.repo.GetGroup(groupName, groupVersion)
	if err != nil {
		return err
	}

	newConfigs := []model.Config{}

	for _, config := range group.Configs {

		if !(config.Name == configName &&
			config.Version == configVersion) {

			newConfigs = append(newConfigs, config)
		}
	}

	group.Configs = newConfigs

	return s.repo.UpdateGroup(group)
}

func (s ConfigGroupService) PutGroup(config model.Config, groupName string, groupVersion int) error {

	theConfig, err := s.repo.Get(config.Name, config.Version)
	if err != nil {
		return err
	}

	theConfig.Labels = config.Labels

	group, err := s.repo.GetGroup(groupName, groupVersion)
	if err != nil {
		return err
	}

	group.Configs = append(group.Configs, theConfig)

	return s.repo.UpdateGroup(group)
}

func (s *ConfigGroupService) GetConfigsByLabels(name string, version int, labels map[string]string) ([]model.Config, error) {
	return s.repo.GetConfigsByLabels(name, version, labels)
}

func (s *ConfigGroupService) DeleteConfigsByLabels(name string, version int, labels map[string]string) error {
	return s.repo.DeleteConfigsByLabels(name, version, labels)
}
