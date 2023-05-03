package agent

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/interceptors"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/go-resty/resty/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Agent interface {
	SendAllMetricsContinuously([]metric.UpdatableMetric)
}

type Client interface {
	PostMetric([]metric.UpdatableMetric) error
}

func NewGRPCClient(ctx context.Context, cfg *config.AgentConfig) (Client, error) {
	caPem, err := os.ReadFile(cfg.CACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)

	targetURL, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	connectParams := grpc.ConnectParams{
		MinConnectTimeout: connTimeout,
	}
	conn, err := grpc.Dial(
		targetURL.Host,
		grpc.WithConnectParams(connectParams),
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(interceptors.InterceptorLogger(cfg.Logger), interceptors.GetLoggerOpts()...),
		),
	)
	if err != nil {
		return nil, err
	}

	grpcClient := service.NewMetricsServiceClient(conn)

	return &GRPCClient{
		ctx:          ctx,
		cfg:          cfg,
		metricClient: grpcClient,
	}, nil
}

func NewHTTPClient(ctx context.Context, cfg *config.AgentConfig) (Client, error) {
	restyClient := resty.New()
	restyClient.
		SetTimeout(connTimeout).
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWaitTime).
		SetRetryMaxWaitTime(retryMaxWaitTime)

	return &HTTPClient{
		ctx:    ctx,
		config: cfg,
		client: restyClient,
	}, nil
}
