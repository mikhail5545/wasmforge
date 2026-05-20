/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
)

type (
	ResponseAction string
	ExecutionMode  string
	ExecutionPhase string

	InvocationRequest struct {
		Context Context

		Ref ModuleRef

		ExecutionMode  ExecutionMode
		ExecutionPhase ExecutionPhase
	}

	InvocationResult struct {
		Action      ResponseAction
		Interrupted bool
		Mutated     bool

		RequestMutations  RequestMutations
		ResponseMutations ResponseMutations
	}

	RequestMutations struct {
		Headers map[string][]string
		Method  string
		Body    []byte
	}

	ResponseMutations struct {
		Headers    map[string][]string
		Method     string
		Body       []byte
		StatusCode int
	}

	// ModuleRef references individual module by one of the unique identifiers.
	// Must specify on of:
	//
	// 	- ID
	// 	- Name, Version and ProjectID
	// 	- Ref
	ModuleRef struct {
		ID        *uuid.UUID
		ProjectID *uuid.UUID
		Name      *string
		Version   *string
		Ref       *core.ObjectRef
	}
)

const (
	// ResponseActionContinue continue middleware chain (for [ExecutionModePlugin])
	ResponseActionContinue ResponseAction = "continue"
	// ResponseActionRespond interrupt middleware chain and respond immediately with the provided
	// response (for [ExecutionModePlugin]), for [ExecutionModeFunction] this is expected behaviour.
	ResponseActionRespond ResponseAction = "respond"
	// ResponseActionReject return policy rejection
	ResponseActionReject ResponseAction = "reject"
	// ResponseActionError internal module error / unexpected behaviour.
	ResponseActionError ResponseAction = "error"

	ExecutionModePlugin   ExecutionMode = "plugin"
	ExecutionModeFunction ExecutionMode = "function"

	ExecutionPhaseBeforeAuth ExecutionPhase = "before-auth"
	ExecutionPhaseAfterAuth  ExecutionPhase = "after-auth"
	ExecutionPhaseOnRequest  ExecutionPhase = "on-request"
	// ExecutionPhaseBeforeUpstream execute module before call
	// to the upstream service - exclusive for [ExecutionModePlugin].
	ExecutionPhaseBeforeUpstream ExecutionPhase = "before-upstream"
	// ExecutionPhaseAfterUpstream execute module after receiving response
	// from the upstream service - exclusive for [ExecutionModePlugin].
	ExecutionPhaseAfterUpstream ExecutionPhase = "after-upstream"
	ExecutionPhaseOnResponse    ExecutionPhase = "on-response"
	ExecutionPhaseOnError       ExecutionPhase = "on-error"
)

type Runtime interface {
	Invoke(ctx context.Context, req InvocationRequest) (InvocationResult, error)
	Preload(ctx context.Context, ref ModuleRef) error
	Evict(ctx context.Context, ref ModuleRef) error
	Close(ctx context.Context) error
}

func (a ResponseAction) String() string {
	return string(a)
}
