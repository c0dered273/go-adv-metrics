package middleware

import (
	"net"
	"net/http"

	"github.com/c0dered273/go-adv-metrics/internal/config"
)

// TrustedSubnet это middleware которое, если передана маска подсети, проверяет пришел ли запрос с доверенной подсети
func TrustedSubnet(cfg *config.ServerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet != nil {
				realIP := net.ParseIP(r.RemoteAddr)
				if realIP == nil {
					cfg.Logger.Error().Msg("trusted_subnet_middleware: failed to parse real ip")
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}

				if !cfg.TrustedSubnet.Contains(realIP) {
					cfg.Logger.Error().Msg("trusted_subnet_middleware: request ip does not belongs to trusted subnet")
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
