// Copyright The E2BGateway Authors
// SPDX-License-Identifier: Apache-2.0

package opensandbox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/e2bgateway/e2bgateway/internal/config"
)

func TestFactoryAPIKeyFile(t *testing.T) {
	for _, key := range []string{"apiKeyFile", "apikeyfile"} {
		t.Run(key, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "key")
			require.NoError(t, os.WriteFile(path, []byte(" file-test-key\n"), 0o600))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("OPEN-SANDBOX-API-KEY") != "file-test-key" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"items":[],"pagination":{"page":1,"pageSize":1,"totalItems":0,"totalPages":0}}`))
			}))
			defer server.Close()
			a, err := NewAdapterFromConfig(config.BackendConfig{Name: "test", Config: map[string]any{
				"baseURL": server.URL, "apiKey": "inline-must-not-win", key: path,
			}}, nil)
			require.NoError(t, err)
			require.NoError(t, a.HealthCheck(context.Background()))
		})
	}
}

func TestFactoryAPIKeyFileFailsClosed(t *testing.T) {
	for _, missing := range []bool{true, false} {
		path := filepath.Join(t.TempDir(), "key")
		if !missing {
			require.NoError(t, os.WriteFile(path, []byte(" \n"), 0o600))
		}
		_, err := NewAdapterFromConfig(config.BackendConfig{Name: "test", Config: map[string]any{
			"baseURL": "http://127.0.0.1:1", "apiKey": "inline-secret", "apiKeyFile": path,
		}}, nil)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "inline-secret")
	}
}
