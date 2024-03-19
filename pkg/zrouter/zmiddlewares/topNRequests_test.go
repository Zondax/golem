package zmiddlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zondax/golem/pkg/zcache"
	"github.com/zondax/golem/pkg/zrouter/auth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testToken = "Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImtleTAwMSIsInR5cCI6IkpXVCJ9.eyJyb2xlcyI6W10sImlzcyI6IlpvbmRheCIsImF1ZCI6WyJiZXJ5eCJdLCJleHAiOjE3MTA2MDU5NjksImp0aSI6IkVtbWFudWVsLGVtbWFudWVsbTQxQGdtYWlsLmNvbSJ9.LoM5lrl9wscuCphqTHoKus5jrBd-YdcgsckLY_-PUBKdsPNxv-G2uR8YmR5WPRn94MqKdbbpOve0h5ttj4H1Hw"
	jtiTest   = "Emmanuel,emmanuelm41@gmail.com"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name           string
		headerValue    string
		expectedToken  string
		expectedErrMsg string
	}{
		{"ValidToken", "Bearer token123", "token123", ""},
		{"MissingToken", "", "", "authorization header is missing"},
		{"InvalidFormat", "invalid123", "", "invalid authorization header format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			if tt.headerValue != "" {
				req.Header.Add("Authorization", tt.headerValue)
			}

			token, err := extractBearerToken(req)
			if tt.expectedErrMsg != "" {
				require.EqualError(t, err, tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestGetTokenDetails(t *testing.T) {
	mockCache := &zcache.MockZCache{}
	ctx := context.TODO()

	hash := sha256.Sum256([]byte(testToken))
	shaToken := hex.EncodeToString(hash[:])

	expectedDetails := auth.TokenDetails{JTI: jtiTest}

	mockCache.On("Get", ctx, shaToken, &auth.TokenDetails{}).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*auth.TokenDetails)
		*arg = expectedDetails
	}).Return(nil).Once()

	details, err := getTokenDetails(ctx, mockCache, testToken, tokenDetailsTTLDefault)
	require.NoError(t, err)
	require.Equal(t, expectedDetails, details)

	mockCache.On("Get", ctx, shaToken, &auth.TokenDetails{}).Return(errors.New("not found")).Once()
	mockCache.On("Set", ctx, shaToken, expectedDetails, tokenDetailsTTLDefault).Return(nil).Once()

	details, err = getTokenDetails(ctx, mockCache, testToken, tokenDetailsTTLDefault)
	require.NoError(t, err)
	require.Equal(t, expectedDetails, details)

	mockCache.AssertExpectations(t)
}

func TestIncrementUsageCount(t *testing.T) {
	mockCache := &zcache.MockZCache{}
	ctx := context.TODO()
	jti := "jti123"
	path := "/test/path"
	ttl := time.Hour

	metricKey := jti + ":" + path

	mockCache.On("ZIncrBy", ctx, PathUsageByJWTKey, metricKey, 1.0).Return(1.0, nil).Once()
	mockCache.On("Expire", ctx, metricKey, ttl).Return(true, nil).Once()

	incrementUsageCount(ctx, mockCache, jti, path, ttl)

	mockCache.AssertExpectations(t)
}

func TestTopRequestTokensMiddleware(t *testing.T) {
	mockCache := &zcache.MockZCache{}
	tokenDetailsTTL := 45 * time.Minute
	usageMetricTTL := time.Hour
	expectedPath := "Emmanuel,emmanuelm41@gmail.com:/"
	expectedTTL := time.Hour

	hash := sha256.Sum256([]byte(strings.TrimPrefix(testToken, "Bearer ")))
	shaToken := hex.EncodeToString(hash[:])

	expectedDetails := auth.TokenDetails{JTI: jtiTest}
	mockCache.On("Get", mock.Anything, shaToken, mock.AnythingOfType("*auth.TokenDetails")).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*auth.TokenDetails)
		*arg = expectedDetails
	}).Return(nil).Once()

	mockCache.On("ZIncrBy", mock.Anything, "jwt_path_usage", mock.AnythingOfType("string"), mock.AnythingOfType("float64")).Return(float64(1), nil).Once()
	mockCache.On("Expire", mock.Anything, expectedPath, expectedTTL).Return(true, nil).Once()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := TopRequestTokensMiddleware(mockCache, tokenDetailsTTL, usageMetricTTL)(handler)

	testServer := httptest.NewServer(middleware)
	defer testServer.Close()

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", testToken)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	mockCache.AssertExpectations(t)
}
