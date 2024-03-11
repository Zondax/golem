package zmiddlewares

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRoutePattern(t *testing.T) {
	r := chi.NewRouter()

	routePattern := "/test/{param}"
	r.Get(routePattern, func(w http.ResponseWriter, r *http.Request) {
		routePattern := GetRoutePattern(r)

		assert.Equal(t, routePattern, "/test/{param}", "The returned route pattern should match the one configured in the router.")
	})

	req := httptest.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "The expected status code should be 200 OK.")
}
