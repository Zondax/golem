package zmiddlewares

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"regexp"
	"strings"
)

func PathToRegexp(path string) *regexp.Regexp {
	escapedPath := regexp.QuoteMeta(path)
	escapedPath = strings.ReplaceAll(escapedPath, "\\{", "{")
	escapedPath = strings.ReplaceAll(escapedPath, "\\}", "}")

	pattern := regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(escapedPath, "[^/]+")
	return regexp.MustCompile("^" + pattern + "$")
}

func GetRoutePattern(r *http.Request) string {
	return chi.RouteContext(r.Context()).RoutePattern()
}
