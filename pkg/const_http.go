package pkg

import (
	"encoding/json"
)

// --------------------------------------------------------- //

const (
	HTTP_CT_HINT             = "Content-Type"
	HTTP_CT_TEXT_PLAIN       = "text/plain"
	HTTP_CT_APPLICATION_JSON = "application/json"

	STATUS_RESP_MESSAGE_BAD_REQUEST             = "Bad Request"
	STATUS_RESP_MESSAGE_UNAUTHORIZED            = "Unauthorized"
	STATUS_RESP_MESSAGE_METHOD_NOT_ALLOWED      = "Method Not Allowed"
	STATUS_RESP_MESSAGE_INTERNAL_SERVER_ERROR   = "Internal Server Error"
	STATUS_RESP_MESSAGE_PRECONDITION_FAILED     = "Pre-Condition Failed"
	STATUS_RESP_MESSAGE_PRECONDITION_REQUIRED   = "Pre-Condition Required"
	STATUS_RESP_MESSAGE_JSON_BODY_NOT_VALID     = "Json Body Not Valid"
	STATUS_RESP_MESSAGE_JSON_BODY_FAIL_TO_PARSE = "Json Body Fail to Parse"

	STATUS_RESP_MESSAGE_INTERNAL_ERROR_DB_SQL = "Internal Error DB SQL"

	HTTP_HEADER_HOST          = "Host"
	HTTP_HEADER_ORIGIN        = "Origin"
	HTTP_HEADER_AUTHORIZATION = "Authorization"
)

// --------------------------------------------------------- //

type StatusBackend struct {
	Ok      bool   `json:"ok"`
	Version string `json:"version"`
}

// --------------------------------------------------------- //

type Response_t struct {
	Ok      bool
	Message string
	Data    json.RawMessage
}

type Response_tj struct {
	Ok      bool            `json:"ok"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}
