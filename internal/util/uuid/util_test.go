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

package uuid

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMustParseSlice(t *testing.T) {
	t.Run("successful parsing", func(t *testing.T) {
		strIDs := []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440001",
		}
		expected := uuid.UUIDs{}
		for _, strID := range strIDs {
			uid, _ := uuid.Parse(strID)
			expected = append(expected, uid)
		}

		uids := MustParseSlice(strIDs)
		assert.Equal(t, expected, uids)
	})

	t.Run("some invalid UUIDs", func(t *testing.T) {
		strIDs := []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"invalid-uuid-string",
			"550e8400-e29b-41d4-a716-446655440001",
		}
		expected := uuid.UUIDs{}
		for _, strID := range strIDs {
			if uid, err := uuid.Parse(strID); err == nil {
				expected = append(expected, uid)
			}
		}

		uids := MustParseSlice(strIDs)
		assert.Equal(t, expected, uids)
	})

	t.Run("empty input", func(t *testing.T) {
		var strIDs []string

		uids := MustParseSlice(strIDs)
		assert.Nil(t, uids)
	})
}
