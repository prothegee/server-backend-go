package backend_path

import (
	"fmt"
	"io"
	"net/http"
	"showcase-backend-go/pkg"
	"strings"
)

const (
	BackendPathDynamicFirstPathHint      = "/path/"
	BackendPathDynamicFirstPathHintSplit = "path"
)

func BackendPathDynamic(w http.ResponseWriter, r *http.Request) {
	pathChunks := strings.Split(r.URL.Path, "/")
	pathChunksLen := len(pathChunks)
	pathChunksLenMax := 5

	// NOTE:
	// it's applied as: /path/{p1}/{p2}
	// you may modified if len 5 need to be included as: /path/{p1}/{p2}/

	// only allowed first path as `path`
	if pathChunks[1] != BackendPathDynamicFirstPathHintSplit {
		// 404
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, pkg.STATUS_RESP_MESSAGE_METHOD_NOT_ALLOWED,
			http.StatusMethodNotAllowed)
		return
	}

	if pathChunksLen >= pathChunksLenMax {
		http.Error(w, "path chunks exceed the length",
			http.StatusInternalServerError)
		return
	}

	// just a text plain
	w.Header().Set(pkg.HTTP_CT_HINT, pkg.HTTP_CT_TEXT_PLAIN)

	var resp strings.Builder

	resp.WriteString("we're at: ")

	// remove first index which ""
	pathChunks = append(pathChunks[:0], pathChunks[1:]...)

	for i := range pathChunks {
		ext := fmt.Sprintf("/%s", pathChunks[i])
		resp.WriteString(ext)
	}
	resp.WriteString("\n")

	io.WriteString(w, resp.String())
}
