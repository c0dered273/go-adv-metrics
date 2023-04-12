package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"
)

// RSADecrypt это middleware которое, если передан приватный ключ, пытается расшифровать тело запроса алгоритмом RSA
func RSADecrypt(key *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if key != nil {
				encryptedBody, err := io.ReadAll(r.Body)
				if err != nil {
					return
				}
				defer func() {
					_ = r.Body.Close()
				}()

				if len(encryptedBody) != 0 {
					plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, encryptedBody, nil)
					if err != nil {
						return
					}

					r.Body = io.NopCloser(bytes.NewReader(plainText))
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
