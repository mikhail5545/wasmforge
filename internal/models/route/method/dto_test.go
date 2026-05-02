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

package method

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetRequestMethodSpec_Validate(t *testing.T) {
	t.Parallel()

	validTimeout := 1000
	validPayload := int64(1024)

	invalidTimeout := -1
	invalidPayload := int64(-1024)

	tests := []struct {
		name    string
		spec    SetRequestMethodSpec
		wantErr bool
	}{
		{
			name: "valid GET method",
			spec: SetRequestMethodSpec{
				Method: "GET",
			},
			wantErr: false,
		},
		{
			name: "valid TRACE method",
			spec: SetRequestMethodSpec{
				Method: "TRACE",
			},
			wantErr: false,
		},
		{
			name: "invalid method",
			spec: SetRequestMethodSpec{
				Method: "INVALID",
			},
			wantErr: true,
		},
		{
			name: "valid payload and timeouts",
			spec: SetRequestMethodSpec{
				Method:                 "POST",
				MaxRequestPayloadBytes: &validPayload,
				RequestTimeoutMs:       &validTimeout,
				ResponseTimeoutMs:      &validTimeout,
				RateLimitPerMinute:     &validTimeout,
			},
			wantErr: false,
		},
		{
			name: "invalid payload",
			spec: SetRequestMethodSpec{
				Method:                 "POST",
				MaxRequestPayloadBytes: &invalidPayload,
			},
			wantErr: true,
		},
		{
			name: "invalid request timeout",
			spec: SetRequestMethodSpec{
				Method:           "POST",
				RequestTimeoutMs: &invalidTimeout,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSetRequest_ValidateValidatesEachMethodSpec(t *testing.T) {
	t.Parallel()

	req := &SetRequest{
		RouteID: "00000000-0000-0000-0000-000000000001",
		Methods: []SetRequestMethodSpec{
			{Method: "GET"},
			{Method: "INVALID"},
		},
	}

	require.Error(t, req.Validate())
}
