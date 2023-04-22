package interceptors

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func errorHandler() func(any) error {
	return func(p any) (err error) {
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
}

func GetRecoveryOpts() []recovery.Option {
	return []recovery.Option{
		recovery.WithRecoveryHandler(errorHandler()),
	}
}
