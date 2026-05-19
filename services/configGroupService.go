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

func (s ConfigGroupService) PutGroup(
	name string,
	version int,
	updatedGroup model.ConfigGroup,
) error {

	group, err := s.repo.GetGroup(name, version)
	if err != nil {
		return err
	}

	group.Name = updatedGroup.Name
	group.Configs = updatedGroup.Configs
	group.Version = updatedGroup.Version

	return s.repo.UpdateGroup(group)
}

func (s *ConfigGroupService) GetConfigsByLabels(name string, version int, labels map[string]string) ([]model.Config, error) {
    return s.repo.GetConfigsByLabels(name, version, labels)
}
