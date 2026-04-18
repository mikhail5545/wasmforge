package plugin

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	mockpluginrepo "github.com/mikhail5545/wasmforge/internal/mocks/database/plugin"
	mockrouterepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route"
	mockroutepluginrepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route/plugin"
	mockproxy "github.com/mikhail5545/wasmforge/internal/mocks/proxy"
	mockuploads "github.com/mikhail5545/wasmforge/internal/mocks/uploads"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceCreateAutoSwitchesMatchingRoutePluginsAndReassemblesEnabledRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	baseRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	baseRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	uploadManager := mockuploads.NewMockManager(ctrl)
	factory := mockproxy.NewMockFactory(ctrl)

	svc := New(Dependencies{
		PluginRepo:      basePluginRepo,
		RouteRepo:       baseRouteRepo,
		RoutePluginRepo: baseRoutePluginRepo,
		RouteFactory:    factory,
		UploadManager:   uploadManager,
	}, zap.NewNop())

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	txPluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	txRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	txRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)

	publishedPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f11")
	previousPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f12")
	otherPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f13")
	routeID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f14")
	routePluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f15")

	basePluginRepo.EXPECT().DB().Return(db).Times(1)
	basePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txPluginRepo).Times(1)
	baseRouteRepo.EXPECT().WithTx(gomock.Any()).Return(txRouteRepo).Times(1)
	baseRoutePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepo).Times(1)
	uploadManager.EXPECT().FromMultipartFile(gomock.Any(), "auth_filter_1_2_0\\.wasm", uploads.PluginUpload).Return("checksum", nil).Times(1)

	gomock.InOrder(
		txPluginRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, plugin *pluginmodel.Plugin) error {
			plugin.ID = publishedPluginID
			return nil
		}),
		txPluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return([]*pluginmodel.Plugin{
			{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			{ID: publishedPluginID, Name: "auth_filter", Version: "1.2.0"},
			{ID: otherPluginID, Name: "auth_filter", Version: "2.0.0"},
		}, nil),
		txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routepluginmodel.RoutePlugin{
			{
				ID:                routePluginID,
				RouteID:           routeID,
				PluginID:          previousPluginID,
				VersionConstraint: "^1.0",
				Plugin:            pluginmodel.Plugin{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			},
		}, nil),
		txRoutePluginRepo.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, updates map[string]any, _ ...any) (int64, error) {
				require.Equal(t, publishedPluginID, updates["plugin_id"])
				return 1, nil
			},
		),
		txRouteRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routemodel.Route{
			{ID: routeID, Path: "/secured", Enabled: true},
		}, nil),
		txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routepluginmodel.RoutePlugin{
			{
				ID:                routePluginID,
				RouteID:           routeID,
				PluginID:          publishedPluginID,
				VersionConstraint: "^1.0",
				ExecutionOrder:    1,
				Plugin:            pluginmodel.Plugin{ID: publishedPluginID, Name: "auth_filter", Version: "1.2.0"},
			},
		}, nil),
		factory.EXPECT().Reassemble(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error {
				require.Equal(t, routeID, route.ID)
				require.Len(t, plugins, 1)
				require.Equal(t, publishedPluginID, plugins[0].PluginID)
				require.Equal(t, "1.2.0", plugins[0].Plugin.Version)
				return nil
			},
		),
	)

	created, err := svc.Create(context.Background(), &multipart.FileHeader{}, &pluginmodel.CreateRequest{
		Name:     "auth_filter",
		Version:  "1.2.0",
		Filename: "auth_filter_1_2_0\\.wasm",
	})
	require.NoError(t, err)
	require.Equal(t, publishedPluginID, created.ID)
	require.Equal(t, "checksum", created.Checksum)
}

func TestServiceCreateSkipsAutoSwitchWhenNoMatchingConstraints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	baseRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	baseRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	uploadManager := mockuploads.NewMockManager(ctrl)
	factory := mockproxy.NewMockFactory(ctrl)

	svc := New(Dependencies{
		PluginRepo:      basePluginRepo,
		RouteRepo:       baseRouteRepo,
		RoutePluginRepo: baseRoutePluginRepo,
		RouteFactory:    factory,
		UploadManager:   uploadManager,
	}, zap.NewNop())

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	txPluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	txRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	txRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)

	publishedPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f21")
	previousPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f22")

	basePluginRepo.EXPECT().DB().Return(db).Times(1)
	basePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txPluginRepo).Times(1)
	baseRouteRepo.EXPECT().WithTx(gomock.Any()).Return(txRouteRepo).Times(1)
	baseRoutePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepo).Times(1)
	uploadManager.EXPECT().FromMultipartFile(gomock.Any(), "auth_filter_1_2_0\\.wasm", uploads.PluginUpload).Return("checksum", nil).Times(1)

	gomock.InOrder(
		txPluginRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, plugin *pluginmodel.Plugin) error {
			plugin.ID = publishedPluginID
			return nil
		}),
		txPluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return([]*pluginmodel.Plugin{
			{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			{ID: publishedPluginID, Name: "auth_filter", Version: "1.2.0"},
		}, nil),
		txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routepluginmodel.RoutePlugin{
			{
				ID:                uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f23"),
				RouteID:           uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f24"),
				PluginID:          previousPluginID,
				VersionConstraint: "^2.0",
				Plugin:            pluginmodel.Plugin{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			},
		}, nil),
	)

	txRoutePluginRepo.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	txRouteRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	factory.EXPECT().Reassemble(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	_, err = svc.Create(context.Background(), &multipart.FileHeader{}, &pluginmodel.CreateRequest{
		Name:     "auth_filter",
		Version:  "1.2.0",
		Filename: "auth_filter_1_2_0\\.wasm",
	})
	require.NoError(t, err)
}

func TestServiceCreateReturnsErrorWhenRouteReassemblyFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	baseRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	baseRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	uploadManager := mockuploads.NewMockManager(ctrl)
	factory := mockproxy.NewMockFactory(ctrl)

	svc := New(Dependencies{
		PluginRepo:      basePluginRepo,
		RouteRepo:       baseRouteRepo,
		RoutePluginRepo: baseRoutePluginRepo,
		RouteFactory:    factory,
		UploadManager:   uploadManager,
	}, zap.NewNop())

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	txPluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	txRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	txRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)

	publishedPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f31")
	previousPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f32")
	routeID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f33")

	basePluginRepo.EXPECT().DB().Return(db).Times(1)
	basePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txPluginRepo).Times(1)
	baseRouteRepo.EXPECT().WithTx(gomock.Any()).Return(txRouteRepo).Times(1)
	baseRoutePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepo).Times(1)
	uploadManager.EXPECT().FromMultipartFile(gomock.Any(), "auth_filter_1_2_0\\.wasm", uploads.PluginUpload).Return("checksum", nil).Times(1)

	gomock.InOrder(
		txPluginRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, plugin *pluginmodel.Plugin) error {
			plugin.ID = publishedPluginID
			return nil
		}),
		txPluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return([]*pluginmodel.Plugin{
			{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			{ID: publishedPluginID, Name: "auth_filter", Version: "1.2.0"},
		}, nil),
		txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routepluginmodel.RoutePlugin{
			{
				ID:                uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f34"),
				RouteID:           routeID,
				PluginID:          previousPluginID,
				VersionConstraint: "^1.0",
				Plugin:            pluginmodel.Plugin{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			},
		}, nil),
		txRoutePluginRepo.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil),
		txRouteRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routemodel.Route{
			{ID: routeID, Path: "/secured", Enabled: true},
		}, nil),
		txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routepluginmodel.RoutePlugin{
			{
				ID:                uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f34"),
				RouteID:           routeID,
				PluginID:          publishedPluginID,
				VersionConstraint: "^1.0",
				ExecutionOrder:    1,
				Plugin:            pluginmodel.Plugin{ID: publishedPluginID, Name: "auth_filter", Version: "1.2.0"},
			},
		}, nil),
		factory.EXPECT().Reassemble(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("reassemble failed")),
	)

	_, err = svc.Create(context.Background(), &multipart.FileHeader{}, &pluginmodel.CreateRequest{
		Name:     "auth_filter",
		Version:  "1.2.0",
		Filename: "auth_filter_1_2_0\\.wasm",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to reassemble enabled route after auto-switching")
}

func TestServiceCreateSkipsReassemblyWhenAffectedRoutesAreDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	baseRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	baseRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	uploadManager := mockuploads.NewMockManager(ctrl)
	factory := mockproxy.NewMockFactory(ctrl)

	svc := New(Dependencies{
		PluginRepo:      basePluginRepo,
		RouteRepo:       baseRouteRepo,
		RoutePluginRepo: baseRoutePluginRepo,
		RouteFactory:    factory,
		UploadManager:   uploadManager,
	}, zap.NewNop())

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	txPluginRepo := mockpluginrepo.NewMockRepository(ctrl)
	txRouteRepo := mockrouterepo.NewMockRepository(ctrl)
	txRoutePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)

	publishedPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f41")
	previousPluginID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f42")
	routeID := uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f43")

	basePluginRepo.EXPECT().DB().Return(db).Times(1)
	basePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txPluginRepo).Times(1)
	baseRouteRepo.EXPECT().WithTx(gomock.Any()).Return(txRouteRepo).Times(1)
	baseRoutePluginRepo.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepo).Times(1)
	uploadManager.EXPECT().FromMultipartFile(gomock.Any(), "auth_filter_1_2_0\\.wasm", uploads.PluginUpload).Return("checksum", nil).Times(1)

	gomock.InOrder(
		txPluginRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, plugin *pluginmodel.Plugin) error {
			plugin.ID = publishedPluginID
			return nil
		}),
		txPluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return([]*pluginmodel.Plugin{
			{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			{ID: publishedPluginID, Name: "auth_filter", Version: "1.2.0"},
		}, nil),
		txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routepluginmodel.RoutePlugin{
			{
				ID:                uuid.MustParse("0198d1bf-f8b8-7b9c-8d60-e4cf6f873f44"),
				RouteID:           routeID,
				PluginID:          previousPluginID,
				VersionConstraint: "^1.0",
				Plugin:            pluginmodel.Plugin{ID: previousPluginID, Name: "auth_filter", Version: "1.1.0"},
			},
		}, nil),
		txRoutePluginRepo.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil),
		txRouteRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*routemodel.Route{}, nil),
	)

	txRoutePluginRepo.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	factory.EXPECT().Reassemble(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	_, err = svc.Create(context.Background(), &multipart.FileHeader{}, &pluginmodel.CreateRequest{
		Name:     "auth_filter",
		Version:  "1.2.0",
		Filename: "auth_filter_1_2_0\\.wasm",
	})
	require.NoError(t, err)
}
