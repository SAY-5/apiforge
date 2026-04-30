package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SAY-5/apiforge/spec"
)

func TestMockServerRespondsWithDeclared2xx(t *testing.T) {
	s, _ := spec.Parse([]byte(`{
		"paths": {
			"/x": {"get": {"operationId": "x", "responses": {"201": {"description": "created"}}}}
		}
	}`))
	srv := NewServer(s)
	req := httptest.NewRequest("GET", "/x", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestMockServerHandlesPathParam(t *testing.T) {
	s, _ := spec.Parse([]byte(`{
		"paths": {
			"/users/{id}": {
				"get": {"operationId": "u", "responses": {"200": {"description": "ok"}}}
			}
		}
	}`))
	srv := NewServer(s)
	req := httptest.NewRequest("GET", "/users/42", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("expected 200 on /users/42, got %d", rec.Code)
	}
}

func TestUnregisteredPathReturns404(t *testing.T) {
	s, _ := spec.Parse([]byte(`{"paths": {}}`))
	srv := NewServer(s)
	req := httptest.NewRequest("GET", "/nope", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
