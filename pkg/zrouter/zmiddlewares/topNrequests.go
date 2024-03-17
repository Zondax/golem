package zmiddlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/zondax/golem/pkg/logger"
	"github.com/zondax/golem/pkg/metrics"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/auth"
	"net/http"
	"strings"
	"time"
)

const (
	tokenDetailsTTLDefault = 45 * time.Minute
	PathUsageByJWTKey      = "jwt_path_usage"
	defaultTTL             = time.Hour
)

func TopRequestTokensMiddleware(zCache zcache.RemoteCache, metricsServer metrics.TaskMetrics, usageMetricName string, tokenDetailsTTL, usageMetricTTL time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				logger.GetLoggerFromContext(r.Context()).Errorf("Error extracting bearer token %v", err.Error())
			}

			if token != "" {
				details, err := getTokenDetails(r.Context(), zCache, token, tokenDetailsTTL)
				if err != nil {
					logger.GetLoggerFromContext(r.Context()).Errorf("Error getting token details %v", err.Error())
				}

				if details.JTI != "" {
					if usageMetricTTL == 0 {
						usageMetricTTL = defaultTTL
					}
					incrementUsageCount(r.Context(), zCache, details.JTI, r.URL.Path, usageMetricTTL)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(auth.Header)
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return tokenParts[1], nil
}

func getTokenDetails(ctx context.Context, zCache zcache.ZCache, token string, tokenDetailsTTL time.Duration) (auth.TokenDetails, error) {
	var details auth.TokenDetails

	hash := sha256.Sum256([]byte(token))
	shaToken := hex.EncodeToString(hash[:])

	err := zCache.Get(ctx, shaToken, &details)
	if err != nil || (details.JTI == "") {
		payload, err := auth.DecodeJWT(token)
		if err != nil {
			return auth.TokenDetails{}, err
		}

		details.JTI, _ = payload["jti"].(string)

		if tokenDetailsTTL == 0 {
			tokenDetailsTTL = tokenDetailsTTLDefault
		}
		if err = zCache.Set(ctx, shaToken, details, tokenDetailsTTL); err != nil {
			logger.GetLoggerFromContext(ctx).Errorf("Cache error setting JWT details %v", err.Error())
			return auth.TokenDetails{}, err
		}
	}

	return details, nil
}

func incrementUsageCount(ctx context.Context, zCache zcache.RemoteCache, jti, path string, ttl time.Duration) {
	metricKey := fmt.Sprintf("%s:%s", jti, path)

	if _, err := zCache.ZIncrBy(ctx, PathUsageByJWTKey, metricKey, 1); err != nil {
		logger.GetLoggerFromContext(ctx).Errorf("Error incrementing usage count in cache %v", err.Error())
		return
	}

	if _, err := zCache.Expire(ctx, metricKey, ttl); err != nil {
		logger.GetLoggerFromContext(ctx).Errorf("Error setting expire in cache %v", err.Error())
		return
	}
}
