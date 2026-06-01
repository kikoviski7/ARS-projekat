package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"projekat/model"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	capi "github.com/hashicorp/consul/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	configPrefix string = "config/"
	groupPrefix  string = "group/"
)

var (
	ErrConfigNotFound = fmt.Errorf("config not found")
	ErrGroupNotFound  = fmt.Errorf("group not found")
	consulTracer      = otel.Tracer("config-consul-repository")
)

type ConfigConsulRepository struct {
	cli     *capi.Client
	kv      *capi.KV
	session *capi.Session
}

func NewConfigConsulRepository() model.ConfigRepository {
	repoConfig := capi.DefaultConfig()
	repoConfig.Address = os.Getenv("DB") + ":" + os.Getenv("DBPORT")

	cli, err := capi.NewClient(repoConfig)
	if err != nil {
		fmt.Println("connection to client database failed: %w", err)
	}
	return ConfigConsulRepository{
		cli:     cli,
		kv:      cli.KV(),
		session: cli.Session(),
	}
}

func (c ConfigConsulRepository) SaveIdempotencyResult(ctx context.Context, key string, result interface{}, ttl time.Duration) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.SaveIdempotencyResult")
	defer span.End()

	sessionEntry := &capi.SessionEntry{
		TTL:      ttl.String(),
		Behavior: "delete",
	}
	sessionID, _, err := c.session.Create(sessionEntry, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to create session: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	acquired, _, err := c.kv.Acquire(&capi.KVPair{
		Key:     idempotencyKey(key),
		Value:   data,
		Session: sessionID,
	}, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to save idempotency result: %w", err)
	}
	if !acquired {
		return fmt.Errorf("failed to acquire lock for idempotency key")
	}

	/*_, err = c.kv.Put(&capi.KVPair{
		Key:     idempotencyKey(key),
		Value:   data,
		Session: sessionID,
	}, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to save idempotency result: %w", err)
	}*/

	return nil
}

func (c ConfigConsulRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.ConfigGroup, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.GetByIdempotencyKey")
	defer span.End()

	span.SetAttributes(attribute.String("idempotency.key", key))

	kvPair, _, err := c.kv.Get(idempotencyKey(key), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get idempotency result: %w", err)
	}

	if kvPair == nil {
		return nil, nil
	}

	var result model.ConfigGroup
	err = json.Unmarshal(kvPair.Value, &result)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to unmarshal idempotency result: %w", err)
	}

	return &result, nil
}

func idempotencyKey(key string) string {
	return "idempotency/" + key
}

func (c ConfigConsulRepository) Add(ctx context.Context, config model.Config) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.Add")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", config.Name),
		attribute.Int("config.version", config.Version),
	)

	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	data, err := json.Marshal(config)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = c.kv.Put(&capi.KVPair{Key: configKey(config.Name, config.Version), Value: data}, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to put config: %w", err)

	}

	key := configKey(config.Name, config.Version)

	pair := &capi.KVPair{
		Key:   key,
		Value: data,
	}

	_, err = c.kv.Put(pair, nil)
	if err != nil {
		fmt.Println("failed to put config:", err)
	}

	return nil
}

func (c ConfigConsulRepository) AddGroup(ctx context.Context, group model.ConfigGroup) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.AddGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", group.Name),
		attribute.Int("group.version", group.Version),
		attribute.Int("group.configs_count", len(group.Configs)),
	)

	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}

	groupForStorage := model.ConfigGroup{
		ID:      group.ID,
		Name:    group.Name,
		Version: group.Version,
		Configs: nil,
	}

	groupData, err := json.Marshal(groupForStorage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal group: %w", err)
	}

	_, err = c.kv.Put(&capi.KVPair{Key: groupKey(group.Name, group.Version), Value: groupData}, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to put group: %w", err)
	}

	for _, config := range group.Configs {
		if config.ID == uuid.Nil {
			config.ID = uuid.New()
		}

		configData, err := json.Marshal(config)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to marshal config inside group: %w", err)
		}

		_, err = c.kv.Put(&capi.KVPair{Key: groupConfigKey(group.Name, group.Version, config), Value: configData}, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to put config inside group: %w", err)
		}
	}

	return nil
}

func (c ConfigConsulRepository) DeleteByVersion(ctx context.Context, name string, version int) (model.Config, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.DeleteByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
	)

	config, err := c.Get(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, ErrConfigNotFound) {
			return model.Config{}, fmt.Errorf("nothing to delete: %w", err)
		}
		return model.Config{}, err
	}

	_, err = c.kv.Delete(configKey(name, version), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Config{}, fmt.Errorf("failed to delete config from consul: %w", err)
	}

	return config, nil
}

func (c ConfigConsulRepository) DeleteConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.DeleteConfigsByLabels")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("labels.count", len(labels)),
	)

	_, err := c.GetGroup(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	prefix := groupConfigPrefix(name, version, labels)

	pairs, _, err := c.kv.List(prefix, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to get grouped configs from consul: %w", err)
	}

	if len(pairs) == 0 {
		err := fmt.Errorf("nothing to delete: no configs found with labels %v", labels)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	for _, pair := range pairs {
		_, err := c.kv.Delete(pair.Key, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to delete grouped config from consul: %w", err)
		}
	}

	span.SetAttributes(attribute.Int("deleted.count", len(pairs)))
	return nil
}

func (c ConfigConsulRepository) DeleteGroupByVersion(ctx context.Context, name string, version int) (model.ConfigGroup, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.DeleteGroupByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
	)

	group, err := c.GetGroup(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, ErrGroupNotFound) {
			return model.ConfigGroup{}, fmt.Errorf("nothing to delete: %w", err)
		}
		return model.ConfigGroup{}, err
	}

	_, err = c.kv.DeleteTree(groupPrefixKey(name, version), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, fmt.Errorf("failed to delete grouped configs from consul: %w", err)
	}

	_, err = c.kv.Delete(groupKey(name, version), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, fmt.Errorf("failed to delete group metadata from consul: %w", err)
	}

	return group, nil
}

func (c ConfigConsulRepository) Get(ctx context.Context, name string, version int) (model.Config, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
	)

	pair, _, err := c.kv.Get(configKey(name, version), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Config{}, fmt.Errorf("failed to get config from consul: %w", err)
	}

	if pair == nil {
		err := fmt.Errorf("%w: name=%s version=%d", ErrConfigNotFound, name, version)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Config{}, err
	}

	var config model.Config
	err = json.Unmarshal(pair.Value, &config)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func (c ConfigConsulRepository) GetByName(ctx context.Context, name string) ([]model.Config, error) {

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

func (c ConfigConsulRepository) GetAll(ctx context.Context) (map[string]model.Config, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.GetAll")
	defer span.End()

	pairs, _, err := c.kv.List(configPrefix, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		configs[config.Name+"/"+strconv.Itoa(config.Version)] = config
	}

	span.SetAttributes(attribute.Int("configs.count", len(configs)))
	return configs, nil
}

func (c ConfigConsulRepository) Put(
	ctx context.Context,
	config model.Config,
	oldName string,
	oldVersion int,
) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.Put")
	defer span.End()

	oldConfig, err := c.Get(ctx, oldName, oldVersion)
	if err != nil {
		return err
	}

	if config.ID == uuid.Nil {
		config.ID = oldConfig.ID
	}

	oldKey := configKey(oldName, oldVersion)
	newKey := configKey(config.Name, config.Version)

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	pair := &capi.KVPair{
		Key:   newKey,
		Value: data,
	}

	_, err = c.kv.Put(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to put config in consul: %w", err)
	}

	if oldKey != newKey {
		_, err = c.kv.Delete(oldKey, nil)
		if err != nil {
			return fmt.Errorf("failed to delete old config from consul: %w", err)
		}
	}

	return nil
}

func (c ConfigConsulRepository) GetAllGroups(ctx context.Context) (map[string]model.ConfigGroup, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.GetAllGroups")
	defer span.End()

	pairs, _, err := c.kv.List(groupPrefix, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to unmarshal group: %w", err)
		}

		group.Configs = nil
		groups[group.Name+"/"+strconv.Itoa(group.Version)] = group
		configs, err := c.getConfigsForGroup(group.Name, group.Version)
		if err != nil {
			return nil, err
		}

		group.Configs = configs

		key := group.Name + "/" + strconv.Itoa(group.Version)

		groups[key] = group
	}

	span.SetAttributes(attribute.Int("groups.count", len(groups)))
	return groups, nil
}

func (c ConfigConsulRepository) GetConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) ([]model.Config, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.GetConfigsByLabels")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("labels.count", len(labels)),
	)

	_, err := c.GetGroup(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	prefix := groupConfigPrefix(name, version, labels)

	pairs, _, err := c.kv.List(prefix, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get grouped configs from consul: %w", err)
	}

	configs := make([]model.Config, 0)

	for _, pair := range pairs {
		var config model.Config
		err := json.Unmarshal(pair.Value, &config)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to unmarshal grouped config: %w", err)
		}
		configs = append(configs, config)
	}

	span.SetAttributes(attribute.Int("results.count", len(configs)))
	return configs, nil
}

func (c ConfigConsulRepository) GetGroup(ctx context.Context, name string, version int) (model.ConfigGroup, error) {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.GetGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
	)

	pair, _, err := c.kv.Get(groupKey(name, version), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, fmt.Errorf("failed to get group from consul: %w", err)
	}

	if pair == nil {
		err := fmt.Errorf("%w: name=%s version=%d", ErrGroupNotFound, name, version)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, err
	}

	var group model.ConfigGroup
	err = json.Unmarshal(pair.Value, &group)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, fmt.Errorf("failed to unmarshal group: %w", err)
	}

	group.Configs = nil
	configs, err := c.getConfigsForGroup(name, version)
	if err != nil {
		return model.ConfigGroup{}, err
	}

	group.Configs = configs

	return group, nil
}

func (c ConfigConsulRepository) PutGroup(ctx context.Context, group model.ConfigGroup, oldName string, oldVersion int) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.PutGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.old_name", oldName),
		attribute.Int("group.old_version", oldVersion),
		attribute.String("group.new_name", group.Name),
		attribute.Int("group.new_version", group.Version),
	)

	oldGroup, err := c.GetGroup(ctx, oldName, oldVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if group.ID == uuid.Nil {
		group.ID = oldGroup.ID
	}

	oldKey := groupKey(oldName, oldVersion)
	newKey := groupKey(group.Name, group.Version)

	if oldKey != newKey {
		_, err = c.kv.DeleteTree(groupPrefixKey(oldName, oldVersion), nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to delete old grouped configs from consul: %w", err)
		}

		_, err = c.kv.Delete(oldKey, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to delete old group metadata from consul: %w", err)
		}
	} else {
		_, err = c.kv.DeleteTree(groupPrefixKey(group.Name, group.Version), nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to delete old grouped configs from consul: %w", err)
		}
	}

	groupForStorage := model.ConfigGroup{
		ID:      group.ID,
		Name:    group.Name,
		Version: group.Version,
		Configs: nil,
	}

	groupData, err := json.Marshal(groupForStorage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal group: %w", err)
	}

	_, err = c.kv.Put(&capi.KVPair{Key: newKey, Value: groupData}, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to put group metadata in consul: %w", err)
	}

	for _, config := range group.Configs {
		if config.ID == uuid.Nil {
			config.ID = uuid.New()
		}

		configData, err := json.Marshal(config)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to marshal grouped config: %w", err)
		}

		_, err = c.kv.Put(&capi.KVPair{Key: groupConfigKey(group.Name, group.Version, config), Value: configData}, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to put grouped config in consul: %w", err)
		}
	}

	return nil
}

func (c ConfigConsulRepository) UpdateGroup(ctx context.Context, group model.ConfigGroup) error {
	_, span := consulTracer.Start(ctx, "ConfigConsulRepo.UpdateGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", group.Name),
		attribute.Int("group.version", group.Version),
		attribute.Int("group.configs_count", len(group.Configs)),
	)

	oldGroup, err := c.GetGroup(ctx, group.Name, group.Version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if group.ID == uuid.Nil {
		group.ID = oldGroup.ID
	}

	_, err = c.kv.DeleteTree(groupPrefixKey(group.Name, group.Version), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to delete old grouped configs from consul: %w", err)
	}

	groupForStorage := model.ConfigGroup{
		ID:      group.ID,
		Name:    group.Name,
		Version: group.Version,
		Configs: nil,
	}

	groupData, err := json.Marshal(groupForStorage)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal group: %w", err)
	}

	_, err = c.kv.Put(&capi.KVPair{Key: groupKey(group.Name, group.Version), Value: groupData}, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to update group metadata in consul: %w", err)
	}

	for _, config := range group.Configs {
		if config.ID == uuid.Nil {
			config.ID = uuid.New()
		}

		configData, err := json.Marshal(config)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to marshal grouped config: %w", err)
		}

		_, err = c.kv.Put(&capi.KVPair{Key: groupConfigKey(group.Name, group.Version, config), Value: configData}, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to update grouped config in consul: %w", err)
		}
	}

	return nil
}

// ***************************************************************
// HELPER FUNCTIONS
// ***************************************************************

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

func groupConfigKey(groupName string, groupVersion int, config model.Config) string {
	return groupKey(groupName, groupVersion) +
		"/" +
		buildLabelsKey(config.Labels) +
		"/" +
		config.Name +
		"/" +
		strconv.Itoa(config.Version)
}

func groupConfigPrefix(groupName string, groupVersion int, labels map[string]string) string {
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
