package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	_ MetricsServiceServer = (*MetricsService)(nil)
)

type MetricsService struct {
	UnimplementedMetricsServiceServer
	Config *config.ServerConfig
}

func (ms *MetricsService) Get(ctx context.Context, in *model.GetMetricRequest) (*model.GetMetricResponse, error) {
	var response model.GetMetricResponse
	key, nmErr := metric.NewMetric(in.GetId(), in.GetType(), "0", "")
	if nmErr.Error != nil {
		ms.Config.Logger.Error().Err(nmErr.Error).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	m, err := ms.Config.Repo.FindByID(ctx, key)
	if err != nil {
		msg := fmt.Sprintf("metric_service: metric with ID: %s, MType: %s not found", key.ID, key.MType.String())
		ms.Config.Logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.NotFound, msg)
	}

	pbMetric := &model.Metric{}
	err = mapDtoWithSerialization(m, pbMetric)
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	response.Metric = pbMetric
	return &response, nil
}

func (ms *MetricsService) GetAll(ctx context.Context, empty *emptypb.Empty) (*model.GetAllMetricsResponse, error) {
	var response model.GetAllMetricsResponse

	m, err := ms.Config.Repo.FindAll(ctx)
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	pbMetrics := make([]*model.Metric, len(m))
	err = MapSliceWithSerialization(ToSliceOfPointers(m), pbMetrics)
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	response.Metrics = pbMetrics
	return &response, nil
}

func (ms *MetricsService) Save(ctx context.Context, in *model.Metric) (*model.Status, error) {
	var response model.Status

	m := metric.Metric{}
	err := mapDtoWithSerialization(in, &m)
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	err = validateMetric(m, ms.Config)
	if err != nil {
		return nil, err
	}

	err = ms.Config.Repo.Save(ctx, m)
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	response.Code = 0
	response.Message = "OK"
	return &response, nil
}

func (ms *MetricsService) SaveAll(ctx context.Context, in *model.Metrics) (*model.Status, error) {
	var response model.Status
	pbMetrics := in.GetMetrics()

	m := make([]*metric.Metric, len(pbMetrics))
	err := MapSliceWithSerialization(pbMetrics, m)
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	for _, mt := range m {
		err = validateMetric(*mt, ms.Config)
		if err != nil {
			return nil, err
		}
	}

	err = ms.Config.Repo.SaveAll(ctx, toSliceOfValues(m))
	if err != nil {
		ms.Config.Logger.Error().Err(err).Send()
		return nil, status.Errorf(codes.Internal, "Internal error")
	}

	response.Code = 0
	response.Message = "OK"
	return &response, nil
}

func MapSliceWithSerialization[T any, E any](in []*T, out []*E) error {
	if len(in) != len(out) {
		return errors.New("mapping error: slices must have equal length")
	}

	for i := range in {
		pb := new(E)
		err := mapDtoWithSerialization(in[i], pb)
		if err != nil {
			return err
		}
		out[i] = pb
	}

	return nil
}

func mapDtoWithSerialization[T any, E any](in T, out *E) error {
	b, err := json.Marshal(&in)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, out)
	if err != nil {
		return err
	}

	return nil
}

func ToSliceOfPointers[T any](in []T) []*T {
	out := make([]*T, len(in))
	for i := range in {
		out[i] = &in[i]
	}
	return out
}

func toSliceOfValues[T any](in []*T) []T {
	out := make([]T, len(in))
	for i := range in {
		out[i] = *in[i]
	}
	return out
}

func validateMetric(m metric.Metric, cfg *config.ServerConfig) error {
	if !metric.IsValid(m) {
		msg := fmt.Sprintf("metric_service: metric with ID: %s, MType: %s invalid", m.ID, m.MType.String())
		cfg.Logger.Error().Msg(msg)
		return status.Errorf(codes.InvalidArgument, msg)
	}

	ok, err := m.CheckHash(cfg.Key)
	if err != nil {
		cfg.Logger.Error().Err(err).Send()
		return status.Errorf(codes.Internal, "Internal error")
	}
	if !ok {
		msg := fmt.Sprint("metric_service: invalid metric hash")
		cfg.Logger.Error().Err(err).Msg(msg)
		return status.Errorf(codes.InvalidArgument, msg)
	}

	return nil
}
