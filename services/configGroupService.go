package services

import (
	"projekat/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ConfigGroupService struct {
	repo model.ConfigRepository
	tracer trace.Tracer
}

func NewConfigGroupService(repo model.ConfigRepository) ConfigGroupService {
	return ConfigGroupService{
		repo: repo,
		tracer: otel.Tracer("config-group-service"),
	}
}

func (s ConfigGroupService) GetGroup(
	ctx context.Context,
	name string,
	version int,
) (model.ConfigGroup, error) {

	ctx, span := s.tracer.Start(ctx, "GetGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
	)

	group, err := s.repo.GetGroup(ctx, name, version)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get group")
		return model.ConfigGroup{}, err
	}

	span.SetStatus(codes.Ok, "group retrieved successfully")
	span.SetAttributes(
		attribute.Int("group.config_count", len(group.Configs)),
	)

	return group, nil
}

func (s ConfigGroupService) GetAllGroups(
	ctx context.Context,
) (
	map[string]model.ConfigGroup,
	error,
) {
	ctx, span := s.tracer.Start(ctx, "GetAllGroups")
	defer span.End()

	groups, err := s.repo.GetAllGroups(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get all groups")
		return nil, err
	}

	span.SetStatus(codes.Ok, "all groups retrieved successfully")
	span.SetAttributes(
		attribute.Int("groups.count", len(groups)),
	)

	return groups, nil
}

func (s ConfigGroupService) PostGroup(
	ctx context.Context,
	name string,
	version int,
	configs []model.Config,
) error {
	ctx, span := s.tracer.Start(ctx, "PostGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("configs.count", len(configs)),
	)

	group := model.ConfigGroup{
		Name:    name,
		Version: version,
		Configs: configs,
	}

	for i, config := range configs {
		span.AddEvent("adding_config", trace.WithAttributes(
			attribute.String("config.name", config.Name),
			attribute.Int("config.version", config.Version),
			attribute.Int("config.index", i),
		))

		err := s.repo.Add(ctx, config)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to add config")
			return err
		}
	}

	err := s.repo.AddGroup(ctx, group)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to add group")
		return err
	}

	span.SetStatus(codes.Ok, "group and configs created successfully")
	return nil
}

func (s ConfigGroupService) DeleteGroupByVersion(
	ctx context.Context,
	name string,
	version int,
) error {
	ctx, span := s.tracer.Start(ctx, "DeleteGroupByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
	)

	_, err := s.repo.DeleteGroupByVersion(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete group")
		return err
	}

	span.SetStatus(codes.Ok, "group deleted successfully")
	return nil
}

func (s ConfigGroupService) DeleteConfigByVersion(
	ctx context.Context,
	groupName string,
	groupVersion int,
	configName string,
	configVersion int,
) error {
	ctx, span := s.tracer.Start(ctx, "DeleteConfigByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", groupName),
		attribute.Int("group.version", groupVersion),
		attribute.String("config.name", configName),
		attribute.Int("config.version", configVersion),
	)

	group, err := s.repo.GetGroup(ctx, groupName, groupVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get group")
		return err
	}

	newConfigs := []model.Config{}

	for _, config := range group.Configs {
		if !(config.Name == configName &&
			config.Version == configVersion) {
			newConfigs = append(newConfigs, config)
		}
	}

	span.SetAttributes(
		attribute.Int("configs.removed", len(group.Configs)-len(newConfigs)),
	)

	group.Configs = newConfigs

	err = s.repo.UpdateGroup(ctx, group)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update group")
		return err
	}

	span.SetStatus(codes.Ok, "config deleted from group successfully")
	return nil
}

func (s *ConfigGroupService) PutGroup(
	ctx context.Context,
	config model.Config,
	groupName string,
	groupVersion int,
) error {
	ctx, span := s.tracer.Start(ctx, "PutGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", config.Name),
		attribute.Int("config.version", config.Version),
		attribute.String("group.name", groupName),
		attribute.Int("group.version", groupVersion),
	)

	theConfig, err := s.repo.Get(ctx, config.Name, config.Version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get config")
		return err
	}

	theConfig.Labels = config.Labels

	group, err := s.repo.GetGroup(ctx, groupName, groupVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get group")
		return err
	}

	group.Configs = append(group.Configs, theConfig)

	err = s.repo.UpdateGroup(ctx, group)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update group")
		return err
	}

	span.SetStatus(codes.Ok, "group updated with config successfully")
	return nil
}

func (s *ConfigGroupService) GetConfigsByLabels(
	ctx context.Context,
	name string,
	version int,
	labels map[string]string,
) ([]model.Config, error) {
	ctx, span := s.tracer.Start(ctx, "GetConfigsByLabels")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("labels.count", len(labels)),
	)

	configs, err := s.repo.GetConfigsByLabels(ctx, name, version, labels)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get configs by labels")
		return nil, err
	}

	span.SetStatus(codes.Ok, "configs retrieved by labels successfully")
	span.SetAttributes(
		attribute.Int("configs.found", len(configs)),
	)

	return configs, nil
}

func (s *ConfigGroupService) DeleteConfigsByLabels(
	ctx context.Context,
	name string,
	version int,
	labels map[string]string,
) error {
	ctx, span := s.tracer.Start(ctx, "DeleteConfigsByLabels")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.name", name),
		attribute.Int("group.version", version),
		attribute.Int("labels.count", len(labels)),
	)

	err := s.repo.DeleteConfigsByLabels(ctx, name, version, labels)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete configs by labels")
		return err
	}

	span.SetStatus(codes.Ok, "configs deleted by labels successfully")
	return nil
}
