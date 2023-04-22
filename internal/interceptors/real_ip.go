package interceptors

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var trueClientIP = http.CanonicalHeaderKey("True-Client-IP")
var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

func RealIPUnaryServerInterceptor() func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ip := getRealIP(md)
			if p, ok := peer.FromContext(ctx); ok && ip != "" {
				p.Addr, err = net.ResolveIPAddr("", ip)
				if err != nil {
					return nil, status.Error(codes.InvalidArgument, "real_ip: failed to resolve ip address from metadata")
				}

				outCtx := peer.NewContext(ctx, p)
				return handler(outCtx, req)
			}
		}

		return handler(ctx, req)
	}
}

func getRealIP(md metadata.MD) string {
	var ip string

	if tcIP := md.Get(trueClientIP); len(tcIP) > 0 {
		ip = tcIP[0]
	} else if xrIP := md.Get(xRealIP); len(xrIP) > 0 {
		ip = xrIP[0]
	} else if xff := md.Get(xForwardedFor); len(xff) > 0 {
		ip = xff[0]
	}
	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
