package middleware

import (
	"net"
	"net/http"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/rs/zerolog"
)

// TrustedSubnet это middleware которое, если передана маска подсети, проверяет пришел ли запрос с доверенной подсети
func TrustedSubnet(cfg *config.ServerConfig, logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet != nil {
				realIp := net.ParseIP(r.RemoteAddr)
				if realIp == nil {
					logger.Error().Msg("trusted_subnet_middleware: failed to parse real ip")
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}

				if !cfg.TrustedSubnet.Contains(realIp) {
					logger.Error().Msg("trusted_subnet_middleware: request ip does not belongs to trusted subnet")
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
