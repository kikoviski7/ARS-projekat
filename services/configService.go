package services

import (
	"projekat/model"
	"time"

	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ConfigService struct {
	repo   model.ConfigRepository
	tracer trace.Tracer
}

func NewConfigService(repo model.ConfigRepository) ConfigService {
	return ConfigService{
		repo:   repo,
		tracer: otel.Tracer("config-service"),
	}
}

func (s *ConfigService) Add(ctx context.Context, config model.Config) error {
	ctx, span := s.tracer.Start(ctx, "Add")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", config.Name),
		attribute.Int("config.version", config.Version),
		attribute.Int("config.params.count", len(config.Params)),
	)

	err := s.repo.Add(ctx, config)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to add config")
		return err
	}

	span.SetStatus(codes.Ok, "config added successfully")
	return nil
}

func (s *ConfigService) Get(ctx context.Context, name string, version int) (model.Config, error) {
	ctx, span := s.tracer.Start(ctx, "ConfigService.Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
	)

	config, err := s.repo.Get(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get config")
		return model.Config{}, err
	}

	span.SetStatus(codes.Ok, "config retrieved successfully")
	span.SetAttributes(
		attribute.Int("config.params.count", len(config.Params)),
	)

	return config, nil
}

func (s *ConfigService) GetAll(ctx context.Context) (map[string]model.Config, error) {
	ctx, span := s.tracer.Start(ctx, "ConfigService.GetAll")
	defer span.End()

	configs, err := s.repo.GetAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get all configs")
		return nil, err
	}

	span.SetStatus(codes.Ok, "all configs retrieved successfully")
	span.SetAttributes(
		attribute.Int("configs.count", len(configs)),
	)

	return configs, nil
}

func (s *ConfigService) Post(ctx context.Context, name string, version int, params map[string]string, idempotencyKey string) error {
	ctx, span := s.tracer.Start(ctx, "ConfigService.Post")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
		attribute.Int("config.params.count", len(params)),
		attribute.String("idempotency.key", idempotencyKey),
	)

	existingConfig, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err == nil && existingConfig != nil {
		span.AddEvent("Group already exists for this idempotency key - returning cached result")
		span.SetStatus(codes.Ok, "group already exists")
		return nil
	}

	config := model.Config{
		Name:           name,
		Version:        version,
		Params:         params,
		IdempotencyKey: idempotencyKey,
	}

	err = s.repo.Add(ctx, config)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to add config")
		return err
	}

	const idempotencyTTL = 30 * time.Second
	if err := s.repo.SaveIdempotencyResult(ctx, idempotencyKey, config, idempotencyTTL); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save idempotency result")
		return err
	}

	span.SetStatus(codes.Ok, "config created successfully")
	return nil
}

func (s *ConfigService) DeleteByVersion(ctx context.Context, name string, version int) error {
	ctx, span := s.tracer.Start(ctx, "ConfigService.DeleteByVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("config.name", name),
		attribute.Int("config.version", version),
	)

	_, err := s.repo.DeleteByVersion(ctx, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete config")
		return err
	}

	span.SetStatus(codes.Ok, "config deleted successfully")
	return nil
}
