package model

import "github.com/google/uuid"

type Config struct {
	ID      uuid.UUID         `json:"id"`
	Name    string            `json:"name"`
	Params  map[string]string `json:"params"`
	Version int               `json:"version"`
	Labels  map[string]string `json:"labels"`
}

type ConfigGroup struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Configs []Config  `json:"configs"`
	Version int       `json:"version"`
}

type ConfigRepository interface {
	Add(config Config)
	Get(name string, version int) (Config, error)
	GetByName(name string) ([]Config, error)
	GetAll() (map[string]Config, error)
	DeleteByVersion(name string, version int) (Config, error)
	Put(config Config, oldName string, oldVersion int) error

	AddGroup(configGroup ConfigGroup)
	GetGroup(name string, version int) (ConfigGroup, error)
	GetAllGroups() (map[string]ConfigGroup, error)
	DeleteGroupByVersion(name string, version int) (ConfigGroup, error)
	UpdateGroup(group ConfigGroup) error
	PutGroup(group ConfigGroup, oldName string, oldVersion int) error

	GetConfigsByLabels(name string, version int, labels map[string]string) ([]Config, error)
	DeleteConfigsByLabels(name string, version int, labels map[string]string) error
}
