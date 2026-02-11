package reqctx

import "context"

// RequestState holds the state of a request as it is processed through the middleware chain.
type RequestState struct {
	Interrupted bool   // Did the plugin interrupt the request?
	StatusCode  int    // If interrupted, the status code to return
	Body        []byte // If interrupted, the body to return
}

type stateKey struct{}

// WithRequestState adds the RequestState to the reqctx.
func WithRequestState(ctx context.Context, state *RequestState) context.Context {
	return context.WithValue(ctx, stateKey{}, state)
}

// RequestStateFromContext retrieves the RequestState from the reqctx.
// It panics if the RequestState is not present, so it should only be called after ensuring it has been set.
func RequestStateFromContext(ctx context.Context) *RequestState {
	return ctx.Value(stateKey{}).(*RequestState)
}
