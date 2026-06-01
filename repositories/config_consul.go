package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"projekat/model"
	"sort"
	"strconv"
	"strings"

	capi "github.com/hashicorp/consul/api"
)

//dizajn baze:
//groupPrefix/groupName/groupVersion/labelkey=labelvalue|labelkey=labelvalue/configPrefix/configName/configVersion

const (
	configPrefix string = "config/"
	groupPrefix  string = "group/"
)

var (
	ErrConfigNotFound = fmt.Errorf("config not found")
	ErrGroupNotFound  = fmt.Errorf("group not found")
)

type ConfigConsulRepository struct {
	cli *capi.Client
	kv  *capi.KV
}

func NewConfigConsulRepository() model.ConfigRepository {

	repoConfig := capi.DefaultConfig()
	repoConfig.Address = os.Getenv("DB") + ":" + os.Getenv("DBPORT")

	cli, err := capi.NewClient(repoConfig) //config for long live connection
	if err != nil {
		fmt.Println("connection to client database failed: %w", err)
	}
	return ConfigConsulRepository{
		cli: cli,
		kv:  cli.KV(),
	}

}

func (c ConfigConsulRepository) Add(config model.Config) {

	data, err := json.Marshal(config)
	if err != nil {
		fmt.Println("failed to marshal in put config: %w", err)
	}

	key := configKey(config.Name, config.Version)

	//key-value entry reprezentacija u consul-u
	pair := &capi.KVPair{
		Key:   key,
		Value: data,
	}

	//PUT poziv ka bazi
	_, err = c.kv.Put(pair, nil)
	if err != nil {
		fmt.Println("failed to put config: %w", err)
	}
}

func (c ConfigConsulRepository) AddGroup(group model.ConfigGroup) {

	groupForStorage := model.ConfigGroup{
		Name:    group.Name,
		Version: group.Version,
		Configs: nil,
	}

	groupData, err := json.Marshal(groupForStorage)
	if err != nil {
		fmt.Println("failed to marshal group:", err)
		return
	}

	groupPair := &capi.KVPair{
		Key:   groupKey(group.Name, group.Version),
		Value: groupData,
	}

	_, err = c.kv.Put(groupPair, nil)
	if err != nil {
		fmt.Println("failed to put group:", err)
		return
	}

	for _, config := range group.Configs {
		configData, err := json.Marshal(config)
		if err != nil {
			fmt.Println("failed to marshal config inside group:", err)
			continue
		}

		configPair := &capi.KVPair{
			Key:   groupConfigKey(group.Name, group.Version, config),
			Value: configData,
		}

		_, err = c.kv.Put(configPair, nil)
		if err != nil {
			fmt.Println("failed to put config inside group:", err)
		}
	}
}

func (c ConfigConsulRepository) DeleteByVersion(name string, version int) (model.Config, error) {

	config, err := c.Get(name, version)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			return model.Config{}, fmt.Errorf("nothing to delete: %w", err)
		}

		return model.Config{}, err
	}

	_, err = c.kv.Delete(configKey(name, version), nil)
	if err != nil {
		return model.Config{}, fmt.Errorf("failed to delete config from consul: %w", err)
	}

	return config, nil
}

func (c ConfigConsulRepository) DeleteConfigsByLabels(
	name string,
	version int,
	labels map[string]string,
) error {

	_, err := c.GetGroup(name, version)
	if err != nil {
		return err
	}

	prefix := groupConfigPrefix(name, version, labels)

	pairs, _, err := c.kv.List(prefix, nil)
	if err != nil {
		return fmt.Errorf("failed to get grouped configs from consul: %w", err)
	}

	if len(pairs) == 0 {
		return fmt.Errorf("nothing to delete: no configs found with labels %v", labels)
	}

	for _, pair := range pairs {
		_, err := c.kv.Delete(pair.Key, nil)
		if err != nil {
			return fmt.Errorf("failed to delete grouped config from consul: %w", err)
		}
	}

	return nil
}

func (c ConfigConsulRepository) DeleteGroupByVersion(
	name string,
	version int,
) (model.ConfigGroup, error) {

	group, err := c.GetGroup(name, version)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return model.ConfigGroup{}, fmt.Errorf("nothing to delete: %w", err)
		}

		return model.ConfigGroup{}, err
	}

	_, err = c.kv.DeleteTree(groupPrefixKey(name, version), nil)
	if err != nil {
		return model.ConfigGroup{}, fmt.Errorf("failed to delete grouped configs from consul: %w", err)
	}

	_, err = c.kv.Delete(groupKey(name, version), nil)
	if err != nil {
		return model.ConfigGroup{}, fmt.Errorf("failed to delete group metadata from consul: %w", err)
	}

	return group, nil
}

func (c ConfigConsulRepository) Get(name string, version int) (model.Config, error) {

	key := configKey(name, version)

	pair, _, err := c.kv.Get(key, nil)
	if err != nil {
		return model.Config{}, fmt.Errorf("failed to get config from consul: %w", err)
	}

	if pair == nil {
		return model.Config{}, fmt.Errorf("%w: name=%s version=%d", ErrConfigNotFound, name, version)
	}

	var config model.Config

	err = json.Unmarshal(pair.Value, &config)
	if err != nil {
		return model.Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func (c ConfigConsulRepository) GetByName(name string) ([]model.Config, error) {

	prefix := configPrefix + name + "/"

	pairs, _, err := c.kv.List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs by name from consul: %w", err)
	}

	configs := make([]model.Config, 0)

	for _, pair := range pairs {

		parts := strings.Split(pair.Key, "/")

		// Only standalone config keys:
		// config/{configName}/{configVersion}
		if len(parts) != 3 {
			continue
		}

		var config model.Config

		err := json.Unmarshal(pair.Value, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		configs = append(configs, config)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("%w: name=%s", ErrConfigNotFound, name)
	}

	return configs, nil
}

func (c ConfigConsulRepository) GetAll() (map[string]model.Config, error) {

	pairs, _, err := c.kv.List(configPrefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs: %w", err)
	}

	configs := make(map[string]model.Config)

	for _, pair := range pairs {

		parts := strings.Split(pair.Key, "/")

		// We only want standalone config keys:
		// config/{configName}/{configVersion}
		if len(parts) != 3 {
			continue
		}

		if parts[0] != "config" {
			continue
		}

		var config model.Config

		err := json.Unmarshal(pair.Value, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		key := config.Name + "/" + strconv.Itoa(config.Version)

		configs[key] = config
	}

	return configs, nil
}

func (c ConfigConsulRepository) GetAllGroups() (map[string]model.ConfigGroup, error) {

	pairs, _, err := c.kv.List(groupPrefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups from consul: %w", err)
	}

	groups := make(map[string]model.ConfigGroup)

	for _, pair := range pairs {

		parts := strings.Split(pair.Key, "/")

		//groupPart/labelPart/configPart
		if len(parts) != 3 {
			continue
		}

		var group model.ConfigGroup

		err := json.Unmarshal(pair.Value, &group)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal group: %w", err)
		}

		configs, err := c.getConfigsForGroup(group.Name, group.Version)
		if err != nil {
			return nil, err
		}

		group.Configs = configs

		key := group.Name + "/" + strconv.Itoa(group.Version)

		groups[key] = group
	}

	return groups, nil
}

func (c ConfigConsulRepository) GetConfigsByLabels(
	name string,
	version int,
	labels map[string]string,
) ([]model.Config, error) {

	_, err := c.GetGroup(name, version)
	if err != nil {
		return nil, err
	}

	prefix := groupConfigPrefix(name, version, labels)

	pairs, _, err := c.kv.List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get grouped configs from consul: %w", err)
	}

	configs := make([]model.Config, 0)

	for _, pair := range pairs {
		var config model.Config

		err := json.Unmarshal(pair.Value, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal grouped config: %w", err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}

func (c ConfigConsulRepository) GetGroup(name string, version int) (model.ConfigGroup, error) {

	pair, _, err := c.kv.Get(groupKey(name, version), nil)
	if err != nil {
		return model.ConfigGroup{}, fmt.Errorf("failed to get group from consul: %w", err)
	}

	if pair == nil {
		return model.ConfigGroup{}, fmt.Errorf("%w: name=%s version=%d", ErrGroupNotFound, name, version)
	}

	var group model.ConfigGroup

	err = json.Unmarshal(pair.Value, &group)
	if err != nil {
		return model.ConfigGroup{}, fmt.Errorf("failed to unmarshal group: %w", err)
	}

	configs, err := c.getConfigsForGroup(name, version)
	if err != nil {
		return model.ConfigGroup{}, err
	}

	group.Configs = configs

	return group, nil
}

func (c ConfigConsulRepository) PutGroup(
	group model.ConfigGroup,
	oldName string,
	oldVersion int,
) error {

	_, err := c.GetGroup(oldName, oldVersion)
	if err != nil {
		return err
	}

	oldKey := groupKey(oldName, oldVersion)
	newKey := groupKey(group.Name, group.Version)

	if oldKey != newKey {
		_, err = c.kv.DeleteTree(groupPrefixKey(oldName, oldVersion), nil)
		if err != nil {
			return fmt.Errorf("failed to delete old grouped configs from consul: %w", err)
		}

		_, err = c.kv.Delete(oldKey, nil)
		if err != nil {
			return fmt.Errorf("failed to delete old group metadata from consul: %w", err)
		}
	} else {
		_, err = c.kv.DeleteTree(groupPrefixKey(group.Name, group.Version), nil)
		if err != nil {
			return fmt.Errorf("failed to delete old grouped configs from consul: %w", err)
		}
	}

	groupForStorage := model.ConfigGroup{
		Name:    group.Name,
		Version: group.Version,
		Configs: nil,
	}

	groupData, err := json.Marshal(groupForStorage)
	if err != nil {
		return fmt.Errorf("failed to marshal group: %w", err)
	}

	groupPair := &capi.KVPair{
		Key:   newKey,
		Value: groupData,
	}

	_, err = c.kv.Put(groupPair, nil)
	if err != nil {
		return fmt.Errorf("failed to put group metadata in consul: %w", err)
	}

	for _, config := range group.Configs {
		configData, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal grouped config: %w", err)
		}

		configPair := &capi.KVPair{
			Key:   groupConfigKey(group.Name, group.Version, config),
			Value: configData,
		}

		_, err = c.kv.Put(configPair, nil)
		if err != nil {
			return fmt.Errorf("failed to put grouped config in consul: %w", err)
		}
	}

	return nil
}
func (c ConfigConsulRepository) UpdateGroup(group model.ConfigGroup) error {

	_, err := c.GetGroup(group.Name, group.Version)
	if err != nil {
		return err
	}

	_, err = c.kv.DeleteTree(groupPrefixKey(group.Name, group.Version), nil)
	if err != nil {
		return fmt.Errorf("failed to delete old grouped configs from consul: %w", err)
	}

	groupForStorage := model.ConfigGroup{
		Name:    group.Name,
		Version: group.Version,
		Configs: nil,
	}

	groupData, err := json.Marshal(groupForStorage)
	if err != nil {
		return fmt.Errorf("failed to marshal group: %w", err)
	}

	groupPair := &capi.KVPair{
		Key:   groupKey(group.Name, group.Version),
		Value: groupData,
	}

	_, err = c.kv.Put(groupPair, nil)
	if err != nil {
		return fmt.Errorf("failed to update group metadata in consul: %w", err)
	}

	for _, config := range group.Configs {
		configData, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal grouped config: %w", err)
		}

		configPair := &capi.KVPair{
			Key:   groupConfigKey(group.Name, group.Version, config),
			Value: configData,
		}

		_, err = c.kv.Put(configPair, nil)
		if err != nil {
			return fmt.Errorf("failed to update grouped config in consul: %w", err)
		}
	}

	return nil
}

//***************************************************************
//HELPER FUNCTION
//***************************************************************

func configKey(name string, version int) string {
	return configPrefix + name + "/" + strconv.Itoa(version)
}

func groupKey(name string, version int) string {
	return groupPrefix + name + "/" + strconv.Itoa(version)
}

func groupPrefixKey(name string, version int) string {
	return groupKey(name, version) + "/"
}

func buildLabelsKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))

	for key := range labels {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	parts := make([]string, 0, len(keys))

	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}

	return strings.Join(parts, "|")
}

func groupConfigKey(
	groupName string,
	groupVersion int,
	config model.Config,
) string {
	return groupKey(groupName, groupVersion) +
		"/" +
		buildLabelsKey(config.Labels) +
		"/" +
		config.Name +
		"/" +
		strconv.Itoa(config.Version)
}

func groupConfigPrefix(
	groupName string,
	groupVersion int,
	labels map[string]string,
) string {
	return groupKey(groupName, groupVersion) +
		"/" +
		buildLabelsKey(labels) +
		"/"
}

func (c ConfigConsulRepository) getConfigsForGroup(
	name string,
	version int,
) ([]model.Config, error) {

	pairs, _, err := c.kv.List(groupPrefixKey(name, version), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs for group from consul: %w", err)
	}

	configs := make([]model.Config, 0)

	for _, pair := range pairs {
		var config model.Config

		err := json.Unmarshal(pair.Value, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal grouped config: %w", err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}
