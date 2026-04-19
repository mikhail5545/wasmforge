package database

import (
	"os"
	"testing"

	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewMigratesPluginVersions(t *testing.T) {
	dbPath := createLegacyPluginDB(t)

	db, err := New(dbPath)
	require.NoError(t, err)

	var existing pluginmodel.Plugin
	require.NoError(t, db.Where("id = ?", "0199f98c-4b09-7465-bf57-f807f2f9ed90").First(&existing).Error)
	require.Equal(t, pluginmodel.DefaultVersion, existing.Version)

	err = db.Create(&pluginmodel.Plugin{
		Name:     "auth_plugin",
		Version:  "1.0.0",
		Filename: "auth_plugin_v1.wasm",
		Checksum: "hash2",
	}).Error
	require.NoError(t, err)

	err = db.Create(&pluginmodel.Plugin{
		Name:     "auth_plugin",
		Version:  pluginmodel.DefaultVersion,
		Filename: "auth_plugin_copy.wasm",
		Checksum: "hash3",
	}).Error
	require.Error(t, err)
}

func createLegacyPluginDB(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(".", "legacy-plugin-*.db")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})

	db, err := gorm.Open(sqlite.Open(f.Name()), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE plugins (
			id uuid PRIMARY KEY,
			created_at datetime,
			name varchar(512) NOT NULL,
			filename varchar(512) NOT NULL,
			checksum text
		);
	`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX idx_plugins_name ON plugins(name);`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX idx_plugins_filename ON plugins(filename);`).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO plugins (id, name, filename, checksum)
		VALUES ('0199f98c-4b09-7465-bf57-f807f2f9ed90', 'auth_plugin', 'auth_plugin.wasm', 'hash1');
	`).Error)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return f.Name()
}
