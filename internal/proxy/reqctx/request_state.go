/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
