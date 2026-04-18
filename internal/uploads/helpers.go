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

package uploads

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func (m *manager) buildFullPath(uploadType UploadType, filename string) (string, error) {
	var baseDir string
	switch uploadType {
	case PluginUpload:
		baseDir = m.pluginUploadDir
	case CertUpload:
		baseDir = m.certUploadDir
	default:
		return "", fmt.Errorf("invalid upload type: %d", uploadType)
	}

	if filename == "" {
		return "", fmt.Errorf("invalid filename: cannot be empty")
	}
	if filepath.IsAbs(filename) || strings.Contains(filename, "/") || strings.Contains(filename, `\`) {
		return "", fmt.Errorf("invalid filename: path separators are not allowed")
	}
	cleanFilename := filepath.Clean(filename)
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename != filename {
		return "", fmt.Errorf("invalid filename: path traversal is not allowed")
	}

	baseDir = filepath.Clean(baseDir)
	fullPath := filepath.Join(baseDir, cleanFilename)
	relativePath, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to validate upload path: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid filename: path traversal is not allowed")
	}
	return fullPath, nil
}

func (m *manager) ensureDirectory(dir string) error {
	_, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Info("uploads directory does not exist, creating it", zap.String("path", dir))
			if err := os.MkdirAll(dir, 0755); err != nil {
				m.logger.Error("failed to create uploads directory", zap.String("path", dir), zap.Error(err))
				return err
			}
			m.logger.Info("uploads directory created successfully", zap.String("path", dir))
			return nil
		}
	}
	return nil
}
