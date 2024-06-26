package zmiddlewares

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"github.com/zondax/golem/pkg/logger"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const (
	undefinedPath = "undefined"
)

func PathToRegexp(path string) *regexp.Regexp {
	escapedPath := regexp.QuoteMeta(path)
	escapedPath = strings.ReplaceAll(escapedPath, "\\{", "{")
	escapedPath = strings.ReplaceAll(escapedPath, "\\}", "}")

	pattern := regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(escapedPath, "[^/]+")
	return regexp.MustCompile("^" + pattern + "$")
}

func GetRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return undefinedPath
	}

	if pattern := rctx.RoutePattern(); pattern != "" && !strings.HasSuffix(pattern, "*") {
		return pattern
	}

	routePath := r.URL.Path
	tctx := chi.NewRouteContext()
	if !rctx.Routes.Match(tctx, r.Method, routePath) {
		return undefinedPath
	}

	// tctx has the updated pattern, since Match mutates it
	return tctx.RoutePattern()
}

func GetSubRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return undefinedPath
	}

	routePattern := rctx.RoutePattern()
	if strings.Contains(rctx.RoutePattern(), "*") {
		return routePattern
	}

	return getRoutePrefix(r.Context(), routePattern)
}

func getRoutePrefix(ctx context.Context, route string) string {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	segments := strings.Split(route, "/")

	// The first segment is empty due to the leading "/", so we check the second segment
	if len(segments) > 2 {
		// The first real segment is at index 1, return it with "/*"
		return "/" + segments[1] + "/*"
	}

	if len(segments) == 2 && segments[1] != "" {
		// There's only one segment in the route, return it with "/*"
		return "/" + segments[1] + "/*"
	}

	logger.GetLoggerFromContext(ctx).Errorf("Cannot detect the route prefix for %s", route)
	return "/*"
}

func getRequestBody(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

func generateBodyHash(body []byte) string {
	hasher := sha256.New()
	hasher.Write(body)
	fullHash := hex.EncodeToString(hasher.Sum(nil))

	return fullHash
}
