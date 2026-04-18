package plugin

import (
	"context"
	"os"
	"testing"

	"github.com/mikhail5545/wasmforge/internal/database"
	plugindb "github.com/mikhail5545/wasmforge/internal/database/plugin"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceGetWithVersion(t *testing.T) {
	dbPath := createTestDBPath(t)
	db, err := database.New(dbPath)
	require.NoError(t, err)

	repo := plugindb.New(db)
	ctx := context.Background()
	require.NoError(t, repo.Create(ctx, &pluginmodel.Plugin{Name: "auth_filter", Version: "1.0.0", Filename: "auth_filter_1.wasm"}))
	require.NoError(t, repo.Create(ctx, &pluginmodel.Plugin{Name: "auth_filter", Version: "2.0.0", Filename: "auth_filter_2.wasm"}))

	svc := New(Dependencies{PluginRepo: repo}, zap.NewNop())
	name := "auth_filter"
	version := "2.0.0"
	p, err := svc.Get(ctx, &pluginmodel.GetRequest{Name: &name, Version: &version})
	require.NoError(t, err)
	require.Equal(t, "2.0.0", p.Version)
}

func TestServiceListWithVersionsFilter(t *testing.T) {
	dbPath := createTestDBPath(t)
	db, err := database.New(dbPath)
	require.NoError(t, err)

	repo := plugindb.New(db)
	ctx := context.Background()
	require.NoError(t, repo.Create(ctx, &pluginmodel.Plugin{Name: "auth_filter", Version: "1.0.0", Filename: "auth_filter_1.wasm"}))
	require.NoError(t, repo.Create(ctx, &pluginmodel.Plugin{Name: "auth_filter", Version: "2.0.0", Filename: "auth_filter_2.wasm"}))
	require.NoError(t, repo.Create(ctx, &pluginmodel.Plugin{Name: "rate_limit", Version: "1.0.0", Filename: "rate_limit_1.wasm"}))

	svc := New(Dependencies{PluginRepo: repo}, zap.NewNop())
	plugins, _, err := svc.List(ctx, &pluginmodel.ListRequest{
		Versions:       []string{"2.0.0"},
		OrderField:     pluginmodel.OrderFieldVersion,
		OrderDirection: "asc",
		PageSize:       10,
	})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	require.Equal(t, "2.0.0", plugins[0].Version)
	require.Equal(t, "auth_filter", plugins[0].Name)
}

func createTestDBPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(".", "plugin-version-test-*.db")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})
	return f.Name()
}
