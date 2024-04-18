package zmiddlewares

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const (
	notDefinedPath = "not_defined"
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
		return notDefinedPath
	}

	if pattern := rctx.RoutePattern(); pattern != "" && !strings.HasSuffix(pattern, "*") {
		return pattern
	}

	routePath := r.URL.Path
	tctx := chi.NewRouteContext()
	if !rctx.Routes.Match(tctx, r.Method, routePath) {
		return notDefinedPath
	}

	// tctx has the updated pattern, since Match mutates it
	return tctx.RoutePattern()
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
