package repositories

import (
	"encoding/json"
	"fmt"
	"os"
	"projekat/model"
	"strconv"

	capi "github.com/hashicorp/consul/api"
)

const configPrefix string = "config/"
const groupPrefix string = "group/"

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

	key := configPrefix +
		config.Name +
		"/" +
		strconv.Itoa(config.Version)

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

	data, err := json.Marshal(group)
	if err != nil {
		fmt.Println(err)
		return
	}

	pair := &capi.KVPair{
		Key:   groupKey(group.Name, group.Version),
		Value: data,
	}

	_, err = c.kv.Put(pair, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (c ConfigConsulRepository) DeleteByVersion(name string, version int) (model.Config, error) {

	config, err := c.Get(name, version)
	if err != nil {
		fmt.Println("config not found")
		return model.Config{}, err
	}

	key := configPrefix + name + "/" + strconv.Itoa(version)

	_, err = c.kv.Delete(key, nil)
	if err != nil {
		return model.Config{}, err
	}

	return config, nil
}

func (c ConfigConsulRepository) DeleteConfigsByLabels(
	name string,
	version int,
	labels map[string]string,
) error {

	group, err := c.GetGroup(name, version)
	if err != nil {
		return err
	}

	newConfigs := []model.Config{}

	for _, config := range group.Configs {
		if !labelsMatch(config.Labels, labels) {
			newConfigs = append(newConfigs, config)
		}
	}

	group.Configs = newConfigs

	return c.UpdateGroup(group)
}

func (c ConfigConsulRepository) DeleteGroupByVersion(name string, version int) (model.ConfigGroup, error) {

	group, err := c.GetGroup(name, version)
	if err != nil {
		fmt.Println("group not found")
		return model.ConfigGroup{}, err
	}

	_, err = c.kv.Delete(groupKey(name, version), nil)
	if err != nil {
		return model.ConfigGroup{}, err
	}

	return group, nil
}

func (c ConfigConsulRepository) Get(name string, version int) (model.Config, error) {

	//pravljenje kljuca
	key := configPrefix + name + "/" + strconv.Itoa(version)

	//GET poziv ka bazi
	pair, _, err := c.kv.Get(key, nil)

	//error handling
	if err != nil {
		return model.Config{}, fmt.Errorf("failed to get config: %w", err)
	}

	if pair == nil {
		return model.Config{}, fmt.Errorf("config not found")
	}

	var config model.Config

	//unmarshalovanje (ceo konfig je u JSON-u)
	err = json.Unmarshal(pair.Value, &config)
	if err != nil {
		return model.Config{}, err
	}

	return config, nil
}

func (c ConfigConsulRepository) GetAll() (map[string]model.Config, error) {

	pairs, _, err := c.kv.List(configPrefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs: %w", err)
	}

	configs := make(map[string]model.Config)

	for _, pair := range pairs {

		var config model.Config

		err := json.Unmarshal(pair.Value, &config)
		if err != nil {
			return nil, err
		}

		key := config.Name + "/" + strconv.Itoa(config.Version)

		configs[key] = config
	}

	return configs, nil
}

func (c ConfigConsulRepository) GetAllGroups() (map[string]model.ConfigGroup, error) {

	pairs, _, err := c.kv.List(groupPrefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	groups := make(map[string]model.ConfigGroup)

	for _, pair := range pairs {

		var group model.ConfigGroup

		err := json.Unmarshal(pair.Value, &group)
		if err != nil {
			return nil, err
		}

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

	group, err := c.GetGroup(name, version)
	if err != nil {
		return nil, err
	}

	configs := []model.Config{}

	for _, config := range group.Configs {
		if labelsMatch(config.Labels, labels) {
			configs = append(configs, config)
		}
	}

	return configs, nil
}

func (c ConfigConsulRepository) GetGroup(name string, version int) (model.ConfigGroup, error) {

	pair, _, err := c.kv.Get(groupKey(name, version), nil)
	if err != nil {
		return model.ConfigGroup{}, fmt.Errorf("failed to get group: %w", err)
	}

	if pair == nil {
		return model.ConfigGroup{}, fmt.Errorf("group not found")
	}

	var group model.ConfigGroup

	err = json.Unmarshal(pair.Value, &group)
	if err != nil {
		return model.ConfigGroup{}, err
	}

	return group, nil
}

func (c ConfigConsulRepository) PutGroup(group model.ConfigGroup, oldName string, oldVersion int) error {

	oldKey := groupKey(oldName, oldVersion)
	newKey := groupKey(group.Name, group.Version)

	data, err := json.Marshal(group)
	if err != nil {
		return err
	}

	pair := &capi.KVPair{
		Key:   newKey,
		Value: data,
	}

	_, err = c.kv.Put(pair, nil)
	if err != nil {
		return err
	}

	if oldKey != newKey {
		_, err = c.kv.Delete(oldKey, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
func (c ConfigConsulRepository) UpdateGroup(group model.ConfigGroup) error {

	data, err := json.Marshal(group)
	if err != nil {
		return err
	}

	pair := &capi.KVPair{
		Key:   groupKey(group.Name, group.Version),
		Value: data,
	}

	_, err = c.kv.Put(pair, nil)

	return err
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

func labelsMatch(configLabels map[string]string, searchLabels map[string]string) bool {
	for key, value := range searchLabels {
		if configLabels[key] != value {
			return false
		}
	}

	return true
}
