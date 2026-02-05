package backend_api_account

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"showcase-backend-go/pkg"
	"showcase-backend-go/pkg/databases/postgres"
	mw "showcase-backend-go/pkg/middleware"

	"showcase-backend-go/pkg/databases/postgres/main/schema_table/account"
)

// --------------------------------------------------------- //

type postAccountUserRequestData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type patchAccountUserRequestData struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type deleteAccountUserRequestData = patchAccountUserRequestData

// --------------------------------------------------------- //

func getAccountUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	resp := pkg.Response_tj{
		Ok:      false,
		Message: "n/a",
		Data:    json.RawMessage("null"),
	}

	email := r.URL.Query().Get("email")

	if len(email) <= 0 || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		resp.Message = "required param/s: email"

		w.WriteHeader(http.StatusPreconditionRequired)

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	user := db_pg_main_account_user.User{}
	id, err := user.SelectIdByEmail(db_pg.MainDb, ctx, email)
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

	// given info to data field/key
	data, err := json.Marshal(map[string]string{"id": id.String()})
	if err != nil {
		http.Error(w, "Marshal Failed", http.StatusInternalServerError)
		return
	}

	resp.Ok = true
	resp.Message = "found"
	resp.Data = json.RawMessage(data)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func postAccountUser(w http.ResponseWriter, r *http.Request) {
	req := postAccountUserRequestData{}
	ctx := context.Background()
	resp := pkg.Response_tj{
		Ok:      false,
		Message: "n/a",
		Data:    json.RawMessage("null"),
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_JSON_BODY_NOT_VALID,
			http.StatusBadRequest)
		return
	}
	_, err = json.Marshal(req)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_JSON_BODY_FAIL_TO_PARSE,
			http.StatusBadRequest)
		return
	}

	// minimal: a@b.c
	if len(req.Email) < 5 || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		resp.Message = "email format is wrong"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
	// minimal password requirements: 6 length
	if len(req.Password) < 6 {
		resp.Message = "password required 6 characters length"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	accountUser := db_pg_main_account_user.User{}

	err = accountUser.InsertNewUserByEmail(db_pg.MainDb,
		ctx, req.Email, req.Password)
	if err != nil {
		resp.Message = "email already in used"

		w.WriteHeader(http.StatusBadRequest)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	resp.Ok = true
	resp.Message = "created"

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func pathAccountUser(w http.ResponseWriter, r *http.Request) {
	req := patchAccountUserRequestData{}
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
	mw.CheckAuthorizationHeaderBearerSession(w, &resp, ctx, uid)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_JSON_BODY_NOT_VALID,
			http.StatusBadRequest)
		return
	}
	_, err = json.Marshal(req)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_JSON_BODY_FAIL_TO_PARSE,
			http.StatusBadRequest)
		return
	}

	accountUser := db_pg_main_account_user.User{}

	_, err = accountUser.SelectEmailIfExists(db_pg.MainDb, ctx, req.Email)
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

	err = accountUser.UpdateEmailById(db_pg.MainDb, ctx, req.Id, req.Email)
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
	resp.Message = "patched"

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func deleteAccountUser(w http.ResponseWriter, r *http.Request) {
	req := deleteAccountUserRequestData{}
	ctx := context.Background()
	resp := pkg.Response_tj{
		Ok:      false,
		Message: "n/a",
		Data:    json.RawMessage("null"),
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_JSON_BODY_NOT_VALID,
			http.StatusBadRequest)
		return
	}
	_, err = json.Marshal(req)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_JSON_BODY_FAIL_TO_PARSE,
			http.StatusBadRequest)
		return
	}

	accountUser := db_pg_main_account_user.User{}

	_, err = accountUser.SelectIdByEmail(db_pg.MainDb, ctx, req.Email)
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
	_, err = accountUser.SelectEmailIfExists(db_pg.MainDb, ctx, req.Email)
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

	err = accountUser.DeleteDataByIdAndEmail(db_pg.MainDb, ctx, req.Id, req.Email)
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
	resp.Message = "deleted"

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

// --------------------------------------------------------- //

const BackendApiAccountUserHint = "/api/account/user"

func BackendApiAccountUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(pkg.HTTP_CT_HINT, pkg.HTTP_CT_APPLICATION_JSON)

	switch method := r.Method; method {
	case http.MethodGet:
		{
			getAccountUser(w, r)
		}
	case http.MethodPost:
		{
			postAccountUser(w, r)
		}
	case http.MethodPatch:
		{
			pathAccountUser(w, r)
		}
	case http.MethodDelete:
		{
			deleteAccountUser(w, r)
		}
	default:
		{
			http.Error(w, pkg.STATUS_RESP_MESSAGE_METHOD_NOT_ALLOWED,
				http.StatusMethodNotAllowed)
		}
	}
}
