package interceptors

import (
	"context"
	"net"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TrustedSubnetUnaryServerInterceptor(cfg *config.ServerConfig) func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if p, ok := peer.FromContext(ctx); ok {
			ip := getIPFromPeer(p)
			if cfg.TrustedSubnet.Contains(ip) {
				return handler(ctx, req)
			}

			msg := "trusted_subnet: request ip does not belongs to trusted subnet"
			cfg.Logger.Error().Msg(msg)
			return nil, status.Error(codes.PermissionDenied, msg)
		}

		msg := "trusted_subnet: failed to resolve ip address from context"
		cfg.Logger.Error().Msg(msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
}

func getIPFromPeer(p *peer.Peer) net.IP {
	addr := p.Addr
	switch addr := addr.(type) {
	case *net.UDPAddr:
		return addr.IP
	case *net.TCPAddr:
		return addr.IP
	case *net.IPAddr:
		return addr.IP
	}

	return nil
}
