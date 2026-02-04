package pkg_middleware

import (
	"fmt"
	"net/http"
	"os"
	"slices"

	"showcase-backend-go/pkg"
	"showcase-backend-go/pkg/configs"
)

func CheckHttpOrigin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := pkg.ConfigServerLoad(config.BACKEND_API_CONFIG_JSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
			return
		}

		ok := slices.Contains(cfg.Security.WhitelistOrigin, r.Header.Get(pkg.HTTP_HEADER_ORIGIN))

		if !ok {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_UNAUTHORIZED, http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func CheckHttpHost(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := pkg.ConfigServerLoad(config.BACKEND_API_CONFIG_JSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
			return
		}

		ok := slices.Contains(cfg.Security.WhitelistHost, r.Host)

		if !ok {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_UNAUTHORIZED, http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func CheckContentTypeMustJson(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok := (r.Header.Get(pkg.HTTP_CT_HINT) == pkg.HTTP_CT_APPLICATION_JSON)

		if !ok {
			http.Error(w,
				string(pkg.STATUS_RESP_MESSAGE_PRECONDITION_FAILED+"; Content-Type must application/json"),
				http.StatusPreconditionFailed)
			return
		}

		next(w, r)
	}
}

func CheckHeaderAuthorization(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// IMPORTANT:
		// - in real world application, use at least jwt/jwe token or use specific block cipher
		// - or/then combined with base64 encode/decode
		// ---
		// jwt: https://www.jwt.io/libraries
		// block cipher: https://en.wikipedia.org/wiki/Block_cipherBlock_ciphers_(security_summary)4602
		// ---
		// this section is only user id with base64 enc/dec, mind your purpose

		// GET method skip, implementation directly from handler
		if r.Method == http.MethodGet {
			next(w, r)
			return
		}

		authorization := r.Header.Get(pkg.HTTP_HEADER_AUTHORIZATION)

		if len(authorization) <= 0 {
			http.Error(w,
				string(pkg.STATUS_RESP_MESSAGE_PRECONDITION_FAILED+"; Authorization header required"),
				http.StatusPreconditionFailed)
			return
		}

		next(w, r)
	}
}
