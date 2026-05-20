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
	"github.com/mikhail5545/wasmforge/internal/storage/artifacts"
	"github.com/tetratelabs/wazero"
	"go.uber.org/zap"
)

type HostServices struct {
	// Here will be services runtime depends on
}

type Runtime struct {
	rt       wazero.Runtime
	provider artifacts.Provider
	services HostServices
	cache    *Cache
	logger   *zap.Logger
}
