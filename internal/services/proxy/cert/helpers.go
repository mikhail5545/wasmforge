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

package cert

import (
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/uploads"
	"go.uber.org/zap"
)

func (s *Service) deleteCertFiles(certPath, keyPath *string) error {
	switch {
	case certPath != nil:
		if err := s.uploadManager.Delete(*certPath, uploads.CertUpload); err != nil {
			s.logger.Error("failed to remove cert file from storage", zap.Error(err))
			return fmt.Errorf("failed to remove cert file from storage: %w", err)
		}
	case keyPath != nil:
		if err := s.uploadManager.Delete(*keyPath, uploads.CertUpload); err != nil {
			s.logger.Error("failed to remove key file from storage", zap.Error(err))
			return fmt.Errorf("failed to remove key file from storage: %w", err)
		}
	default:
		return fmt.Errorf("both cert path and key path are nil, nothing to delete")
	}
	return nil
}
