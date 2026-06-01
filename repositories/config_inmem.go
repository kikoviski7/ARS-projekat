package repositories

import (
	"context"
	"errors"
	"fmt"
	"projekat/model"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("config-repository")

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

func (r *ConfigInMem) Add(ctx context.Context, config model.Config) error {
	_, span := tracer.Start(ctx, "ConfigRepo.Add")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", config.Name),
		attribute.Int("config.version", config.Version),
	)

	key := config.Name + "_" + strconv.Itoa(config.Version)
	r.configs[key] = config
	return nil
}

func (r *ConfigInMem) Get(ctx context.Context, name string, version int) (model.Config, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
	)

	key := name + "_" + strconv.Itoa(version)
	config, exists := r.configs[key]
	if !exists {
		err := errors.New("config not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Config{}, err
	}

	return config, nil
}

func (r *ConfigInMem) GetAll(ctx context.Context) (map[string]model.Config, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.GetAll")
	defer span.End()

	span.SetAttributes(attribute.Int("config.count", len(r.configs)))

	return r.configs, nil
}

func (r *ConfigInMem) DeleteByVersion(ctx context.Context, name string, version int) (model.Config, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.DeleteByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
	)

	key := name + "_" + strconv.Itoa(version)
	config, exists := r.configs[key]
	if !exists {
		err := errors.New("config not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Config{}, err
	}

	delete(r.configs, key)
	return config, nil
}

// ======================================================
// CONFIG GROUP METHODS
// ======================================================

func (r *ConfigInMem) GetGroup(ctx context.Context, name string, version int) (model.ConfigGroup, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.GetGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
	)

	key := name + "_" + strconv.Itoa(version)
	group, exists := r.groups[key]
	if !exists {
		err := errors.New("group not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, err
	}

	return group, nil
}

func (r *ConfigInMem) GetAllGroups(ctx context.Context) (map[string]model.ConfigGroup, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.GetAllGroups")
	defer span.End()

	span.SetAttributes(attribute.Int("group.count", len(r.groups)))

	return r.groups, nil
}

func (r *ConfigInMem) DeleteGroupByVersion(ctx context.Context, name string, version int) (model.ConfigGroup, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.DeleteGroupByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
	)

	key := name + "_" + strconv.Itoa(version)
	group, exists := r.groups[key]
	if !exists {
		err := errors.New("group not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.ConfigGroup{}, err
	}

	delete(r.groups, key)
	return group, nil
}

func (r *ConfigInMem) AddGroup(ctx context.Context, group model.ConfigGroup) error {
	_, span := tracer.Start(ctx, "ConfigRepo.AddGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", group.Name),
		attribute.Int("group.version", group.Version),
	)

	key := group.Name + "_" + strconv.Itoa(group.Version)
	r.groups[key] = group
	return nil
}

func (r *ConfigInMem) UpdateGroup(ctx context.Context, group model.ConfigGroup) error {
	_, span := tracer.Start(ctx, "ConfigRepo.UpdateGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", group.Name),
		attribute.Int("group.version", group.Version),
	)

	key := group.Name + "_" + strconv.Itoa(group.Version)
	_, exists := r.groups[key]
	if !exists {
		err := errors.New("group not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	r.groups[key] = group
	return nil
}

func (r *ConfigInMem) PutGroup(ctx context.Context, group model.ConfigGroup, oldName string, oldVersion int) error {
	_, span := tracer.Start(ctx, "ConfigRepo.PutGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.old_name", oldName),
		attribute.Int("group.old_version", oldVersion),
		attribute.String("group.new_name", group.Name),
		attribute.Int("group.new_version", group.Version),
	)

	key := oldName + "_" + strconv.Itoa(oldVersion)
	_, exists := r.groups[key]
	if !exists {
		err := errors.New("group not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	newKey := group.Name + "_" + strconv.Itoa(group.Version)
	r.groups[newKey] = group
	return nil
}

func (r *ConfigInMem) GetConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) ([]model.Config, error) {
	_, span := tracer.Start(ctx, "ConfigRepo.GetConfigsByLabels")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("labels.count", len(labels)),
	)

	group, err := r.GetGroup(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var results []model.Config
	for _, config := range group.Configs {
		if matchesAllLabels(config.Labels, labels) {
			results = append(results, config)
		}
	}

	if len(results) == 0 {
		err := fmt.Errorf("no config found with labels: %v", labels)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("results.count", len(results)))
	return results, nil
}

func (r *ConfigInMem) DeleteConfigsByLabels(ctx context.Context, name string, version int, labels map[string]string) error {
	_, span := tracer.Start(ctx, "ConfigRepo.DeleteConfigsByLabels")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("labels.count", len(labels)),
	)

	group, err := r.GetGroup(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	var remaining []model.Config
	for _, config := range group.Configs {
		if !matchesAllLabels(config.Labels, labels) {
			remaining = append(remaining, config)
		}
	}

	if len(remaining) == len(group.Configs) {
		err := fmt.Errorf("no configs found with labels: %v", labels)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	group.Configs = remaining
	return r.UpdateGroup(ctx, group)
}

func matchesAllLabels(configLabels map[string]string, searchLabels map[string]string) bool {
	for key, value := range searchLabels {
		if configLabels[key] != value {
			return false
		}
	}
	return true
}
