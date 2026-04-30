// Package mock generates a stub HTTP server from a Spec. Each
// operation responds with a canned JSON body that satisfies the
// declared schema (or {} for unschemafied responses). Useful for
// frontend devs who want to start integrating before the backend
// implementation lands.
package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SAY-5/apiforge/spec"
)

func NewServer(s *spec.Spec) *http.ServeMux {
	mux := http.NewServeMux()
	for _, ref := range s.Operations() {
		path, method, op := ref.Path, ref.Method, ref.Op
		// Convert OpenAPI {param} to Go's net/http {param} pattern.
		// Go 1.22+ supports this natively.
		mux.HandleFunc(method+" "+path, func(w http.ResponseWriter, _ *http.Request) {
			// Pick the lowest 2xx response, fall back to 200.
			code := pick2xxCode(op)
			body := mockBodyFor(op, code)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(parseCode(code, 200))
			_ = json.NewEncoder(w).Encode(body)
		})
	}
	return mux
}

func pick2xxCode(op *spec.Operation) string {
	for code := range op.Responses {
		if strings.HasPrefix(code, "2") {
			return code
		}
	}
	return "200"
}

func parseCode(code string, fallback int) int {
	var n int
	if _, err := fmt.Sscanf(code, "%d", &n); err != nil {
		return fallback
	}
	return n
}

func mockBodyFor(op *spec.Operation, code string) any {
	resp, ok := op.Responses[code]
	if !ok {
		return map[string]any{}
	}
	return map[string]any{
		"_mock":       true,
		"description": resp.Description,
		"type":        resp.Type,
	}
}
