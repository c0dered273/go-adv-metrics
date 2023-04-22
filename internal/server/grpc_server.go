package server

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/interceptors"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewServerCredentials(cfg *config.ServerConfig) (credentials.TransportCredentials, error) {
	caPem, err := os.ReadFile(cfg.CACertFile)
	if err != nil {
		cfg.Logger.Error().Err(err).Msg("server: error reading CA certificate")
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		cfg.Logger.Error().Msg("server: error loading CA to cert pool")
		return nil, err
	}

	serverCert, err := tls.LoadX509KeyPair(cfg.ServerCertFile, cfg.ServerKeyFile)
	if err != nil {
		cfg.Logger.Error().Err(err).Msg("server: error reading server key pair")
		return nil, err
	}

	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{serverCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		ClientCAs:          certPool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsConf), nil
}

func NewGRPCServerOptions(cfg *config.ServerConfig) ([]grpc.ServerOption, error) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.RealIPUnaryServerInterceptor(),
			interceptors.TrustedSubnetUnaryServerInterceptor(cfg),
			logging.UnaryServerInterceptor(interceptors.InterceptorLogger(cfg.Logger), interceptors.GetLoggerOpts()...),
			recovery.UnaryServerInterceptor(interceptors.GetRecoveryOpts()...),
		),
	}

	if cfg.IsTLSEnabled {
		tlsCredentials, err := NewServerCredentials(cfg)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(tlsCredentials))
		cfg.Logger.Info().Msg("grpc_server: TLS enabled")
	}

	return opts, nil
}

func NewGRPCServer(cfg *config.ServerConfig) (*grpc.Server, error) {
	grpcOpts, err := NewGRPCServerOptions(cfg)
	if err != nil {
		cfg.Logger.Fatal().Err(err).Msg("server: configuration error")
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	service.RegisterMetricsServiceServer(grpcServer, &service.MetricsService{
		Config: cfg,
	})

	return grpcServer, err
}
