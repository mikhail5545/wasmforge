package plugin

import (
	"testing"

	"github.com/google/uuid"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRoutePlugin_ExecutionOrderUniquePerRoute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&routepluginmodel.RoutePlugin{}))

	routeIDOne, _ := uuid.NewV7()
	routeIDTwo, _ := uuid.NewV7()
	pluginIDOne, _ := uuid.NewV7()
	pluginIDTwo, _ := uuid.NewV7()
	pluginIDThree, _ := uuid.NewV7()

	require.NoError(t, db.Create(&routepluginmodel.RoutePlugin{
		RouteID:        routeIDOne,
		PluginID:       pluginIDOne,
		ExecutionOrder: 1,
	}).Error)

	err = db.Create(&routepluginmodel.RoutePlugin{
		RouteID:        routeIDOne,
		PluginID:       pluginIDTwo,
		ExecutionOrder: 1,
	}).Error
	require.Error(t, err)

	err = db.Create(&routepluginmodel.RoutePlugin{
		RouteID:        routeIDTwo,
		PluginID:       pluginIDThree,
		ExecutionOrder: 1,
	}).Error
	assert.NoError(t, err)
}
