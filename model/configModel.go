package model

import "context"

type Config struct {
	Name    string            `json:"name"`
	Params  map[string]string `json:"params"`
	Version int               `json:"version"`
	Labels  map[string]string `json:"labels"`
}

type ConfigGroup struct {
	Name    string   `json:"name"`
	Configs []Config `json:"configs"`
	Version int      `json:"version"`
}

type ConfigRepository interface {
	Add(ctx context.Context, config Config) error
	Get(ctx context.Context, name string, version int) (Config, error)
	GetAll(ctx context.Context) (map[string]Config, error)
	DeleteByVersion(ctx context.Context, name string, version int) (Config, error)

	AddGroup(ctx context.Context, configGroup ConfigGroup) error
	GetGroup(ctx context.Context, name string, version int) (ConfigGroup, error)
	GetAllGroups(ctx context.Context) (map[string]ConfigGroup, error)
	DeleteGroupByVersion(ctx context.Context, name string, version int) (ConfigGroup, error)
	UpdateGroup(ctx context.Context, group ConfigGroup) error
	PutGroup(ctx context.Context, group ConfigGroup, oldName string, oldVersion int) error

	GetConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) ([]Config, error)
	DeleteConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) error
}
