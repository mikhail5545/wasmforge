package plugin

import (
	"context"
	"testing"

	"github.com/google/uuid"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	mockpluginrepo "github.com/mikhail5545/wasmforge/internal/mocks/database/plugin"
	mockrouterepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route"
	mockroutepluginrepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route/plugin"
	mockproxy "github.com/mikhail5545/wasmforge/internal/mocks/proxy"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceCreateResolvesHighestMatchingPluginVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	baseRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	basePluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	factory := mockproxy.NewMockFactory(ctrl)
	svc := New(baseRepo, ServiceParams{RouteRepo: baseRouteRepo, PluginRepo: basePluginRepo, RouteFactory: factory}, zap.NewNop())

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	txRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	txRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	txPluginRepo := mockpluginrepo.NewMockRepository(ctrl)

	routeID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14520")
	basePluginID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14521")
	highestPluginID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14522")
	route := &routemodel.Route{ID: routeID, Enabled: false}
	basePlugin := &pluginmodel.Plugin{ID: basePluginID, Name: "auth_filter", Version: "1.1.0"}
	candidates := []*pluginmodel.Plugin{
		basePlugin,
		{ID: highestPluginID, Name: "auth_filter", Version: "1.8.0"},
		{ID: uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14523"), Name: "auth_filter", Version: "2.0.0"},
	}

	baseRepo.EXPECT().DB().Return(db).Times(1)
	baseRepo.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepo).Times(1)
	baseRouteRepo.EXPECT().WithTx(gomock.Any()).Return(txRouteRepo).Times(1)
	basePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txPluginRepo).Times(1)

	txRouteRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(route, nil).Times(1)
	txPluginRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(basePlugin, nil).Times(2)
	txPluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return(candidates, nil).Times(1)
	txRoutePluginRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, rp *routepluginmodel.RoutePlugin) error {
		require.Equal(t, highestPluginID, rp.PluginID)
		require.Equal(t, "^1.0", rp.VersionConstraint)
		return nil
	}).Times(1)

	res, err := svc.Create(context.Background(), &routepluginmodel.CreateRequest{
		RouteID:           routeID.String(),
		PluginID:          basePluginID.String(),
		VersionConstraint: "^1.0",
		ExecutionOrder:    1,
	})
	require.NoError(t, err)
	require.Equal(t, highestPluginID, res.PluginID)
	require.Equal(t, "^1.0", res.VersionConstraint)
	require.Equal(t, "1.8.0", res.ResolvedPluginVersion)
}

func TestServiceUpdateResolvesAndPersistsPluginVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	baseRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	basePluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	factory := mockproxy.NewMockFactory(ctrl)
	svc := New(baseRepo, ServiceParams{RouteRepo: baseRouteRepo, PluginRepo: basePluginRepo, RouteFactory: factory}, zap.NewNop())

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	txRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	txRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	txPluginRepo := mockpluginrepo.NewMockRepository(ctrl)

	routeID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14530")
	routePluginID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14531")
	currentPluginID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14532")
	resolvedPluginID := uuid.MustParse("01983a0a-74d6-76af-a377-56f8a6f14533")

	baseRepo.EXPECT().DB().Return(db).Times(1)
	baseRepo.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepo).Times(1)
	baseRouteRepo.EXPECT().WithTx(gomock.Any()).Return(txRouteRepo).Times(1)
	basePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txPluginRepo).Times(1)

	txRoutePluginRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&routepluginmodel.RoutePlugin{
		ID:                routePluginID,
		RouteID:           routeID,
		PluginID:          currentPluginID,
		VersionConstraint: "^1.0",
		ExecutionOrder:    1,
	}, nil).Times(1)
	txPluginRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&pluginmodel.Plugin{
		ID:      currentPluginID,
		Name:    "auth_filter",
		Version: "1.1.0",
	}, nil).Times(1)
	txPluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return([]*pluginmodel.Plugin{
		{ID: currentPluginID, Name: "auth_filter", Version: "1.1.0"},
		{ID: resolvedPluginID, Name: "auth_filter", Version: "1.9.0"},
	}, nil).Times(1)
	txRoutePluginRepo.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updates map[string]any, _ ...any) (int64, error) {
		require.Equal(t, resolvedPluginID, updates["plugin_id"])
		return 1, nil
	}).Times(1)
	txRouteRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&routemodel.Route{ID: routeID, Enabled: false}, nil).Times(1)

	updates, err := svc.Update(context.Background(), &routepluginmodel.UpdateRequest{
		ID: routePluginID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, resolvedPluginID, updates["plugin_id"])
	require.Equal(t, "^1.0", updates["version_constraint"])
	require.Equal(t, "1.9.0", updates["resolved_plugin_version"])
}

func TestServiceCreateRejectsInvalidVersionConstraint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := New(
		mockroutepluginrepo.NewMockRepository(ctrl),
		ServiceParams{
			RouteRepo:    mockrouterepo.NewMockRepository(ctrl),
			PluginRepo:   mockpluginrepo.NewMockRepository(ctrl),
			RouteFactory: mockproxy.NewMockFactory(ctrl),
		},
		zap.NewNop(),
	)

	_, err := svc.Create(context.Background(), &routepluginmodel.CreateRequest{
		RouteID:           "01983a0a-74d6-76af-a377-56f8a6f14540",
		PluginID:          "01983a0a-74d6-76af-a377-56f8a6f14541",
		VersionConstraint: "not-a-constraint",
		ExecutionOrder:    1,
	})
	require.ErrorIs(t, err, inerrors.ErrValidationFailed)
}
