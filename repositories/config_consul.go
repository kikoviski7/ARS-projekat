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

// AddGroup implements [model.ConfigRepository].
func (c ConfigConsulRepository) AddGroup(configGroup model.ConfigGroup) {
	panic("unimplemented")
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

// DeleteConfigsByLabels implements [model.ConfigRepository].
func (c ConfigConsulRepository) DeleteConfigsByLabels(name string, version int, labels map[string]string) error {
	panic("unimplemented")
}

// DeleteGroupByVersion implements [model.ConfigRepository].
func (c ConfigConsulRepository) DeleteGroupByVersion(name string, version int) (model.ConfigGroup, error) {
	panic("unimplemented")
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

// GetAllGroups implements [model.ConfigRepository].
func (c ConfigConsulRepository) GetAllGroups() (map[string]model.ConfigGroup, error) {
	panic("unimplemented")
}

// GetConfigsByLabels implements [model.ConfigRepository].
func (c ConfigConsulRepository) GetConfigsByLabels(name string, version int, labels map[string]string) ([]model.Config, error) {
	panic("unimplemented")
}

// GetGroup implements [model.ConfigRepository].
func (c ConfigConsulRepository) GetGroup(name string, version int) (model.ConfigGroup, error) {
	panic("unimplemented")
}

// PutGroup implements [model.ConfigRepository].
func (c ConfigConsulRepository) PutGroup(group model.ConfigGroup, oldName string, oldVersion int) error {
	panic("unimplemented")
}

// UpdateGroup implements [model.ConfigRepository].
func (c ConfigConsulRepository) UpdateGroup(group model.ConfigGroup) error {
	panic("unimplemented")
}
