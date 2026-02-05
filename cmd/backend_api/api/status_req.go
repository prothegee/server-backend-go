package backend_api

import (
	"encoding/json"
	"net/http"

	"showcase-backend-go/pkg"
	"showcase-backend-go/pkg/configs"
)

const BackendApiStatusHint = "/api/status"

func BackendApiStatus(w http.ResponseWriter, r *http.Request) {
	cfg, err := pkg.ConfigServerLoad(config.BACKEND_API_CONFIG_JSON)
	if err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_METHOD_NOT_ALLOWED,
			http.StatusMethodNotAllowed)
		return
	}

	resp := pkg.StatusBackend{
		Ok:      true,
		Version: cfg.Version,
	}

	w.Header().Set(pkg.HTTP_CT_HINT, pkg.HTTP_CT_APPLICATION_JSON)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError)
	}
}
