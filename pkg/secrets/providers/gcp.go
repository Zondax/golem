// Package providers contains implementations of the SecretProvider interface for different secret backends.
// GcpProvider allows resolving secrets from Google Cloud Secret Manager.
//
// Usage:
//
//	import (
//	    "github.com/zondax/golem/pkg/secrets"
//	    "github.com/zondax/golem/pkg/secrets/providers"
//	)
//	secrets.RegisterProvider(providers.GcpProvider{})
package providers

import (
	"context"
	"strings"
	"sync"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var (
	defaultGcpClient *secretmanager.Client
	gcpClientOnce    sync.Once
)

const (
	GcpSecretPrefix = "gcp_"
)

// GcpProvider implements the SecretProvider interface for GCP Secret Manager.
// Register this provider using secrets.RegisterProvider(providers.GcpProvider{}).
type GcpProvider struct{}

// IsSecretKey returns true if the key is a GCP secret reference (starts with GcpSecretPrefix).
// It supports both top-level and nested keys (e.g., "gcp_secret" or "database.gcp_secret").
func (GcpProvider) IsSecretKey(_ context.Context, key string) bool {
	lastPart := key
	if idx := strings.LastIndex(key, "."); idx != -1 {
		lastPart = key[idx+1:]
	}
	return strings.HasPrefix(lastPart, GcpSecretPrefix)
}

// GetSecret fetches the actual secret value from GCP Secret Manager
func (GcpProvider) GetSecret(ctx context.Context, secretPath string) (string, error) {
	client, err := getDefaultGcpClient(ctx)
	if err != nil {
		return "", err
	}
	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	})
	if err != nil {
		return "", err
	}
	return string(resp.Payload.Data), nil
}

func getDefaultGcpClient(ctx context.Context) (*secretmanager.Client, error) {
	var err error
	gcpClientOnce.Do(func() {
		defaultGcpClient, err = secretmanager.NewClient(ctx)
	})
	return defaultGcpClient, err
}
