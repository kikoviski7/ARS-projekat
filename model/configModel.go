package model

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Params         map[string]string `json:"params"`
	Version        int               `json:"version"`
	Labels         map[string]string `json:"labels"`
	IdempotencyKey string            `json:"-"`
}

type ConfigGroup struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Configs        []Config  `json:"configs"`
	Version        int       `json:"version"`
	IdempotencyKey string    `json:"-"`
}

type ConfigRepository interface {
	GetByIdempotencyKey(ctx context.Context, key string) (*ConfigGroup, error)
	SaveIdempotencyResult(ctx context.Context, key string, result interface{}, ttl time.Duration) error

	Add(ctx context.Context, config Config) error
	Get(ctx context.Context, name string, version int) (Config, error)
	GetAll(ctx context.Context) (map[string]Config, error)
	DeleteByVersion(ctx context.Context, name string, version int) (Config, error)

	Put(ctx context.Context, config Config, oldName string, oldVersion int) error
	GetByName(ctx context.Context, name string) ([]Config, error)

	AddGroup(ctx context.Context, configGroup ConfigGroup) error
	GetGroup(ctx context.Context, name string, version int) (ConfigGroup, error)
	GetAllGroups(ctx context.Context) (map[string]ConfigGroup, error)
	DeleteGroupByVersion(ctx context.Context, name string, version int) (ConfigGroup, error)
	UpdateGroup(ctx context.Context, group ConfigGroup) error
	PutGroup(ctx context.Context, group ConfigGroup, oldName string, oldVersion int) error

	GetConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) ([]Config, error)
	DeleteConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) error
}

var ErrConfigAlreadyExists = errors.New("config already exists")
var ErrGroupAlreadyExists = errors.New("group already exists")
