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

package common

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func DecodeMasterKey(raw string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("master key is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil {
		if len(decoded) != 32 {
			return nil, fmt.Errorf("master key must decode to 32 bytes")
		}
		return decoded, nil
	}
	if len(trimmed) != 32 {
		return nil, fmt.Errorf("master key must be a base64 string or raw 32-byte value")
	}
	return []byte(trimmed), nil
}
