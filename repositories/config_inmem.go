package repositories

import (
	"errors"
	"projekat/model"
	"strconv"
	"fmt"
)

type ConfigInMem struct {
	configs map[string]model.Config
	groups  map[string]model.ConfigGroup
}

func NewConfigInMemRepository() *ConfigInMem {
	return &ConfigInMem{
		configs: make(map[string]model.Config),
		groups:  make(map[string]model.ConfigGroup),
	}
}

// ======================================================
// CONFIG METHODS
// ======================================================

func (r *ConfigInMem) Add(config model.Config) {

	key := config.Name + "_" + strconv.Itoa(config.Version)

	r.configs[key] = config
}

func (r *ConfigInMem) Get(
	name string,
	version int,
) (model.Config, error) {

	key := name + "_" + strconv.Itoa(version)

	config, exists := r.configs[key]

	if !exists {
		return model.Config{}, errors.New("config not found")
	}

	return config, nil
}

func (r *ConfigInMem) GetAll() (
	map[string]model.Config,
	error,
) {

	return r.configs, nil
}

func (r *ConfigInMem) DeleteByVersion(
	name string,
	version int,
) (model.Config, error) {

	key := name + "_" + strconv.Itoa(version)

	config, exists := r.configs[key]

	if !exists {
		return model.Config{}, errors.New("config not found")
	}

	delete(r.configs, key)

	return config, nil
}

// ======================================================
// CONFIG GROUP METHODS
// ======================================================

func (r *ConfigInMem) GetGroup(
	name string,
	version int,
) (model.ConfigGroup, error) {

	key := name + "_" + strconv.Itoa(version)

	group, exists := r.groups[key]

	if !exists {
		return model.ConfigGroup{}, errors.New("group not found")
	}

	return group, nil
}

func (r *ConfigInMem) GetAllGroups() (
	map[string]model.ConfigGroup,
	error,
) {

	return r.groups, nil
}

func (r *ConfigInMem) DeleteGroupByVersion(
	name string,
	version int,
) (model.ConfigGroup, error) {

	key := name + "_" + strconv.Itoa(version)

	group, exists := r.groups[key]

	if !exists {
		return model.ConfigGroup{}, errors.New("group not found")
	}

	delete(r.groups, key)

	return group, nil
}

func (r *ConfigInMem) AddGroup(group model.ConfigGroup) {

	key := group.Name + "_" + strconv.Itoa(group.Version)

	r.groups[key] = group
}

// dodati ili u update group ili u put group sistem za tabele
// labele se dodaju kada se konfiguracija dodaje u grupu
func (r *ConfigInMem) UpdateGroup(group model.ConfigGroup) error {

	key := group.Name + "_" + strconv.Itoa(group.Version)

	_, exists := r.groups[key]

	if !exists {
		return errors.New("group not found")
	}

	r.groups[key] = group

	return nil
}

func (r *ConfigInMem) PutGroup(group model.ConfigGroup, oldName string, oldVersion int) error {
	key := oldName + "_" + strconv.Itoa(oldVersion)

	_, exists := r.groups[key]

	if !exists {
		return errors.New("group not found")
	}
	newKey := group.Name + "_" + strconv.Itoa(group.Version)

	r.groups[newKey] = group

	return nil
}

func (r *ConfigInMem) GetConfigsByLabels(name string, version int, labels map[string]string) ([]model.Config, error) {
    group, err := r.GetGroup(name, version)
    if err != nil {
        return nil, err
    }

    var results []model.Config
    for _, config := range group.Configs {
        if matchesAllLabels(config.Labels, labels) {
            results = append(results, config)
        }
    }
    if len(results) == 0 {
        return nil, fmt.Errorf("no config found with labels: %v", labels)
    }
    return results, nil
}

func (r *ConfigInMem) DeleteConfigsByLabels(name string, version int, labels map[string]string) (error) {
    group, err := r.GetGroup(name, version)
    if err != nil {
        return err
    }

    var remaining []model.Config
    for _, config := range group.Configs {
        if !matchesAllLabels(config.Labels, labels) {
            remaining = append(remaining, config)
        }
    }

    if len(remaining) == len(group.Configs) {
        return fmt.Errorf("no configs found with labels: %v", labels)
    }

    group.Configs = remaining
    return r.UpdateGroup(group)
}

func matchesAllLabels(configLabels map[string]string, searchLabels map[string]string) bool {
	for key, value := range searchLabels {
		if configLabels[key] != value {
			return false
		}
	}
	return true
}