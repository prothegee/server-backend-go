package backend_api_auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"showcase-backend-go/pkg"
	"showcase-backend-go/pkg/databases/postgres"
	"showcase-backend-go/pkg/databases/postgres/main/schema_table/account"
	"showcase-backend-go/pkg/databases/redis"
	"showcase-backend-go/pkg/databases/redis/main/key_value/account"

	mw "showcase-backend-go/pkg/middleware"
)

// --------------------------------------------------------- //

func getAuthSession(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	resp := pkg.Response_tj{
		Ok:      false,
		Message: "n/a",
		Data:    json.RawMessage("null"),
	}

	authorization := r.Header.Get(pkg.HTTP_HEADER_AUTHORIZATION)
	uid, err := mw.CheckAuthorizationHeaderBearer(w, authorization)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	userSession := db_rd_main_account_user.UserSession{}
	found, err := userSession.GetSessionExistence(db_rd.MainDb, ctx, uid)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	if !found {
		resp.Message = "session not found"
		// resp.Message = "session not found for " + uid.String()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	data, err := userSession.GetSessionData(db_rd.MainDb, ctx, uid)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	payload, err := json.Marshal(data)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusInternalServerError)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	resp.Ok = true
	resp.Message = "found"
	resp.Data = json.RawMessage(payload)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func postAuthSession(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	resp := pkg.Response_tj{
		Ok:      false,
		Message: "n/a",
		Data:    json.RawMessage("null"),
	}

	// expecting no body data
	bodyReq, err := io.ReadAll(r.Body)
	if len(string(bodyReq)) > 0 {
		resp.Message = "body data should be empty"

		w.WriteHeader(http.StatusPreconditionFailed)

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
	defer r.Body.Close()

	authorization := r.Header.Get(pkg.HTTP_HEADER_AUTHORIZATION)
	uid, err := mw.CheckAuthorizationHeaderBearer(w, authorization)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	account := db_pg_main_account_user.User{}

	ok, err := account.SelectIdIfExists(db_pg.MainDb, ctx, uid)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
	}
	if !ok {
		resp.Message = "user id not found/doesn't exists"

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	// create new session
	userSession := db_rd_main_account_user.UserSession{}
	err = userSession.SetNewSession(db_rd.MainDb, ctx, uid)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	resp.Ok = true
	resp.Message = "session created"

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func deleteAuthSession(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	resp := pkg.Response_tj{
		Ok:      false,
		Message: "n/a",
		Data:    json.RawMessage("null"),
	}

	authorization := r.Header.Get(pkg.HTTP_HEADER_AUTHORIZATION)
	uid, err := mw.CheckAuthorizationHeaderBearer(w, authorization)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	userSession := db_rd_main_account_user.UserSession{}
	total, err := userSession.DeleteSession(db_rd.MainDb, ctx, uid)
	if err != nil {
		resp.Message = err.Error()

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	if total <= 0 {
		resp.Message = "data not found/doesn't exists"

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	resp.Ok = true
	resp.Message = "deleted"

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

// --------------------------------------------------------- //

const BackendApiAuthSessionHint = "/api/auth/session"

func BackendApiAuthSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(pkg.HTTP_CT_HINT, pkg.HTTP_CT_APPLICATION_JSON)

	switch method := r.Method; method {
	case http.MethodGet:
		{
			getAuthSession(w, r)
		}
	case http.MethodPost:
		{
			postAuthSession(w, r)
		}
	case http.MethodDelete:
		{
			deleteAuthSession(w, r)
		}
	default:
		{
			http.Error(w, pkg.STATUS_RESP_MESSAGE_METHOD_NOT_ALLOWED,
				http.StatusMethodNotAllowed)
		}
	}
}
