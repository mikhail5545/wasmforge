package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	authsvc "github.com/mikhail5545/wasmforge/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

type countingRuntime struct {
	wazero.Runtime
	mu               sync.Mutex
	instantiateCalls int
}

func (r *countingRuntime) InstantiateModule(ctx context.Context, compiled wazero.CompiledModule, cfg wazero.ModuleConfig) (api.Module, error) {
	r.mu.Lock()
	r.instantiateCalls++
	r.mu.Unlock()
	return r.Runtime.InstantiateModule(ctx, compiled, cfg)
}

func (r *countingRuntime) InstantiateCalls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.instantiateCalls
}

func TestFactoryCreate_InstantiatesModuleOnceAndReusesItAcrossRequests(t *testing.T) {
	ctx := context.Background()
	baseRuntime := wazero.NewRuntime(ctx)
	t.Cleanup(func() {
		_ = baseRuntime.Close(ctx)
	})

	rt := &countingRuntime{Runtime: baseRuntime}
	f := NewFactory(rt, zap.NewNop())

	// Minimal module exporting on_request with empty body.
	wasmBytes := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
		0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
		0x03, 0x02, 0x01, 0x00,
		0x07, 0x0e, 0x01, 0x0a, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x00, 0x00,
		0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b,
	}

	mw, err := f.Create(ctx, wasmBytes, nil)
	require.NoError(t, err)
	require.Equal(t, 1, rt.InstantiateCalls(), "module should be instantiated at middleware creation")

	nextCalls := 0
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	}))

	for range 2 {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.com/bench", nil)
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}

	assert.Equal(t, 2, nextCalls)
	assert.Equal(t, 1, rt.InstantiateCalls(), "module should not be instantiated per request")
}

func TestWasmMiddleware_PreservesExistingAuthRequestState(t *testing.T) {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	t.Cleanup(func() {
		_ = rt.Close(ctx)
	})

	f := NewFactory(rt, zap.NewNop())
	wasmBytes := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
		0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
		0x03, 0x02, 0x01, 0x00,
		0x07, 0x0e, 0x01, 0x0a, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x00, 0x00,
		0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b,
	}

	mw, err := f.Create(ctx, wasmBytes, nil)
	require.NoError(t, err)

	state := &reqctx.RequestState{
		AuthContext: &reqctx.AuthContext{
			IsAuthenticated: true,
			Subject:         "plugin-user",
			ValidatedToken: &authsvc.ValidatedToken{
				Subject: "plugin-user",
				Claims: map[string]interface{}{
					"role": "admin",
				},
			},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/bench", nil)
	req = req.WithContext(reqctx.WithRequestState(req.Context(), state))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextState := reqctx.RequestStateFromContextSafe(r.Context())
		require.Same(t, state, nextState)
		require.NotNil(t, nextState.AuthContext)
		assert.True(t, nextState.AuthContext.IsAuthenticated)
		assert.Equal(t, "plugin-user", nextState.AuthContext.Subject)
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}
