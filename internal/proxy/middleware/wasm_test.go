package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/mikhail5545/wasmforge/internal/runtime/core"
	authsvc "github.com/mikhail5545/wasmforge/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockRuntime struct {
	invokeResult core.InvocationResult
	invokeErr    error
}

func (m *mockRuntime) Invoke(ctx context.Context, req core.InvocationRequest) (core.InvocationResult, error) {
	return m.invokeResult, m.invokeErr
}
func (m *mockRuntime) Preload(ctx context.Context, ref core.ModuleRef) error { return nil }
func (m *mockRuntime) Evict(ctx context.Context, ref core.ModuleRef) error   { return nil }
func (m *mockRuntime) Close(ctx context.Context) error                       { return nil }

func TestFactoryCreate_InvokeSuccess(t *testing.T) {
	ctx := context.Background()

	rt := &mockRuntime{
		invokeResult: core.InvocationResult{
			Action: core.ResponseActionContinue,
		},
	}
	f := NewFactory(rt, zap.NewNop())

	ref := core.ModuleRef{}
	mw, err := f.Create(ctx, ref, nil)
	require.NoError(t, err)

	nextCalls := 0
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/bench", nil)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, 1, nextCalls)
}

func TestWasmMiddleware_PreservesExistingAuthRequestState(t *testing.T) {
	ctx := context.Background()
	rt := &mockRuntime{
		invokeResult: core.InvocationResult{
			Action: core.ResponseActionContinue,
		},
	}

	f := NewFactory(rt, zap.NewNop())

	ref := core.ModuleRef{}
	mw, err := f.Create(ctx, ref, nil)
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
