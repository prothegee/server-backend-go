package pkg_middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"showcase-backend-go/pkg"
	db_rd "showcase-backend-go/pkg/databases/redis"
	db_rd_main_account_user "showcase-backend-go/pkg/databases/redis/main/key_value/account"

	"net/http"
	"strings"

	"github.com/google/uuid"
)

// --------------------------------------------------------- //

// https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml
const (
	AuthorizationHeadKey_basic         = "Basic"
	AuthorizationHeadKey_bearer        = "Bearer"
	AuthorizationHeadKey_concealed     = "Concealed"
	AuthorizationHeadKey_digest        = "Digest"
	AuthorizationHeadKey_dpop          = "DPoP"
	AuthorizationHeadKey_gnap          = "GNAP"
	AuthorizationHeadKey_hoba          = "HOBA"
	AuthorizationHeadKey_mutual        = "Mutual"
	AuthorizationHeadKey_negotiate     = "Negotiate"
	AuthorizationHeadKey_oauth         = "OAuth"
	AuthorizationHeadKey_privateToken  = "PrivateToken"
	AuthorizationHeadKey_scram_sha_1   = "SCRAM-SHA-1"
	AuthorizationHeadKey_scram_sha_256 = "SCRAM-SHA-256"
)

// --------------------------------------------------------- //

// check authorization header for Bearer
//
// has continuity with CheckAuthorizationHeaderBearerSession
//
// params:
// 	w http.ResponseWriter
// 	authorization string
//
// return: (uuid.UUID, error) - (actual id)
func CheckAuthorizationHeaderBearer(w http.ResponseWriter,
	authorization string) (uuid.UUID, error) {
	var (
		credentialByte []byte
		scheme         string
		credential     string
	)

	if len(authorization) <= 0 {
		return uuid.Nil, errors.New("requirement not satisfied, Authorization is required")
	}
	scheme, credential, err := pkg.ParseAuthorizationHeader(authorization)
	if err != nil {
		return uuid.Nil, err
	}
	if scheme != AuthorizationHeadKey_bearer {
		return uuid.Nil, errors.New("scheme doesn't match; tmp hint: use 'Bearer'")
	}
	credentialByte, err = base64.StdEncoding.DecodeString(credential)
	if err != nil {
		return uuid.Nil, errors.New("failed to decoded credential")
	}
	_, err = pkg.IsValidUuid(string(credentialByte))
	if err != nil {
		return uuid.Nil, errors.New("credential uuid is not supported")
	}

	_uid := strings.TrimSpace(string(credentialByte))
	uid, err := uuid.Parse(_uid)
	if err != nil {
		return uuid.Nil, errors.New("fail to parse uid")
	}

	return uid, nil
}

// check session from CheckAuthorizationHeaderBearer
//
// only use after CheckAuthorizationHeaderBearer
//
// params:
// 	w http.ResponseWriter
// 	resp *pkg.Response_tj
// 	ctx context.Context
// 	uid uuid.UUID
func CheckAuthorizationHeaderBearerSession(w http.ResponseWriter,
	resp *pkg.Response_tj,
	ctx context.Context,
	uid uuid.UUID) {
	/*
		example usage:
		```go
			authorization := r.Header.Get(pkg.HTTP_HEADER_AUTHORIZATION)
		uid, err := CheckAuthorizationHeaderBearer(w, authorization); if err != nil {
			resp.Message = err.Error()

			w.WriteHeader(http.StatusBadRequest)

			err = json.NewEncoder(w).Encode(resp); if err != nil {
				http.Error(w, STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
					http.StatusInternalServerError)
			}
			return
		}
		CheckAuthorizationHeaderBearerSession(w, &resp, ctx, uid)
		```
	*/
	userSession := db_rd_main_account_user.UserSession{}
	found, err := userSession.GetSessionExistence(db_rd.MainDb, ctx, uid)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusUnauthorized)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
	if !found {
		resp.Message = "session not found, create session first"

		w.WriteHeader(http.StatusUnauthorized)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
}
