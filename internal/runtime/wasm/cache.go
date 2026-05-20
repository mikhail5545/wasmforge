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

package wasm

import (
	"sync"

	"github.com/google/uuid"
	"github.com/tetratelabs/wazero/api"
)

type Cache struct {
	modules map[uuid.UUID]api.Module
	mu      sync.RWMutex
}

func (c *Cache) GetModule(id uuid.UUID) (api.Module, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m, ok := c.modules[id]
	return m, ok
}

func (c *Cache) SetModule(id uuid.UUID, m api.Module) {
	c.mu.Lock()
	c.modules[id] = m
	c.mu.Unlock()
}

func (c *Cache) RemoveModule(id uuid.UUID) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.modules[id]; found {
		delete(c.modules, id)
		return true
	}
	return false
}
