package backend_api_game1

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"showcase-backend-go/pkg"
	db_pg "showcase-backend-go/pkg/databases/postgres"
	db_pg_main_game1_stash "showcase-backend-go/pkg/databases/postgres/main/schema_table/game1"
	mw "showcase-backend-go/pkg/middleware"

	"github.com/google/uuid"
)

// --------------------------------------------------------- //

type postGame1StashRequestData struct {
	Name string `json:"name"`
}

type patchGame1StashRequestData struct {
	Name     string                                         `json:"name"`
	Operand  db_pg_main_game1_stash.Game1StashItemOperand_e `json:"operand"`
	Item     string                                         `json:"item"`
	Quantity uint64                                         `json:"quantity"`
}

type deleteGame1StashRequestData struct {
	StashId string `json:"stash_id"`
}

// --------------------------------------------------------- //

func getGame1Stash(w http.ResponseWriter, r *http.Request) {
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

	stashIdStr := r.URL.Query().Get("id")

	if len(stashIdStr) <= 0 {
		resp.Message = "required param/s: id (as the stash id or all)"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}

	game1Stash := db_pg_main_game1_stash.Stash{}
	uuidType, err := pkg.IsValidUuid(stashIdStr) // skip err since `all` is possible

	// reserving error when stashId is not valid
	resp.Message = "default error from stash id"

	if stashIdStr == "all" {
		// query all stash by authorization id
		data, err := game1Stash.SelectAllStashByUid(db_pg.MainDb, ctx, uid)
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

			w.WriteHeader(http.StatusBadRequest)

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
	}

	if stashIdStr != "all" && uuidType != pkg.UUID_UNDEFINED {
		stashId, err := uuid.FromBytes([]byte(strings.TrimSpace(stashIdStr)))
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

		data, err := game1Stash.SelectStashByIdAndUid(db_pg.MainDb,
			ctx, stashId, uid)
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

			w.WriteHeader(http.StatusBadRequest)

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
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func postGame1Stash(w http.ResponseWriter, r *http.Request) {
	req := postGame1StashRequestData{}
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

	if len(req.Name) <= 0 {
		resp.Message = "\"name\" can't be empty"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
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

	game1Stash := db_pg_main_game1_stash.Stash{}

	err = game1Stash.InsertNewStash(db_pg.MainDb, ctx, uid, req.Name)
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
	resp.Message = "created"

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func patchGame1Stash(w http.ResponseWriter, r *http.Request) {
	req := patchGame1StashRequestData{}
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

	if len(req.Name) <= 0 {
		resp.Message = "req \"name\" can't be empty"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
	if req.Operand == db_pg_main_game1_stash.GAME1_STASH_ITEM_OPERAND_UNDEFINED {
		resp.Message = "req \"operand\" must 1 (addition) or 2 (substraction)"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
	if len(req.Item) <= 0 {
		resp.Message = "req \"item\" can't be empty"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
	}
	if req.Quantity <= 0 {
		resp.Message = "req \"quantity\" can't be less or equal than 0"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
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

	stashItem := db_pg_main_game1_stash.Stash{}

	err, extErrMsg := stashItem.UpdateStashByUidAndName(db_pg.MainDb, ctx, uid, req.Name,
		db_pg_main_game1_stash.StashItem_t{Item: req.Item, Quantity: req.Quantity},
		req.Operand)
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
	resp.Message = "updated"

	// expecting conditional where it's substraction but item name doesn't exists
	if len(extErrMsg) > 0 {
		resp.Message = extErrMsg
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}

func deleteGame1Stash(w http.ResponseWriter, r *http.Request) {
	req := deleteGame1StashRequestData{}
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

	if len(req.StashId) <= 0 {
		resp.Message = "req \"stash_id\" can't be empty"

		w.WriteHeader(http.StatusPreconditionRequired)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
				http.StatusInternalServerError)
		}
		return
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

	stash := db_pg_main_game1_stash.Stash{}
	stashId, err := uuid.FromBytes([]byte(strings.TrimSpace(req.StashId)))
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

	err = stash.DeleteStashById(db_pg.MainDb, ctx, stashId)
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

const BackendApiGame1StashHint = "/api/game1/stash"

func BackendApiGame1Stash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(pkg.HTTP_CT_HINT, pkg.HTTP_CT_APPLICATION_JSON)

	switch method := r.Method; method {
	case http.MethodGet:
		{
			getGame1Stash(w, r)
		}
	case http.MethodPost:
		{
			postGame1Stash(w, r)
		}
	case http.MethodPatch:
		{
			patchGame1Stash(w, r)
		}
	case http.MethodDelete:
		{
			deleteGame1Stash(w, r)
		}
	default:
		{
			http.Error(w, pkg.STATUS_RESP_MESSAGE_METHOD_NOT_ALLOWED,
				http.StatusMethodNotAllowed)
		}
	}
}
