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

package fs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mikhail5545/wasmforge/internal/storage/core"
	"go.uber.org/zap"
)

type Config struct {
	DataRoot     string
	MaxSizeBytes int64
}

type ObjectStore struct {
	cfg    Config
	logger *zap.Logger
}

func (s *ObjectStore) Put(_ context.Context, ref core.ObjectRef, r io.Reader) (core.ObjectInfo, error) {
	dest, err := os.Create(ref.Key)
	if err != nil {
		return core.ObjectInfo{}, err
	}
	defer func() {
		if err := dest.Close(); err != nil {
			s.logger.Error("failed to close file", zap.Error(err))
		}
	}()

	hasher := sha256.New()
	s.logger.Debug("starting to stream object into storage", zap.String("key", ref.Key))
	written, err := io.Copy(io.MultiWriter(dest, hasher), io.LimitReader(r, s.cfg.MaxSizeBytes))
	if err != nil {
		s.logger.Error("failed to write object", zap.Error(err))
		return core.ObjectInfo{}, fmt.Errorf("failed to write object: %w", err)
	}
	if written > s.cfg.MaxSizeBytes {
		s.logger.Warn("data exceeded size limit while streaming", zap.Int64("max_size_bytes", s.cfg.MaxSizeBytes), zap.Int64("written_bytes", written))
		if removeErr := os.Remove(ref.Key); removeErr != nil {
			s.logger.Error("failed to remove oversized destination file", zap.Error(removeErr))
		}
		return core.ObjectInfo{}, core.NewSizeLimitExceededError("written file exceeds maximum allowed size")
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	s.logger.Debug("object saved successfully", zap.Int64("bytes_written", written), zap.String("key", ref.Key))
	return core.ObjectInfo{
		Ref:       ref,
		SizeBytes: written,
		Checksum:  hash,
	}, nil
}

func (s *ObjectStore) Get(_ context.Context, ref core.ObjectRef) (io.ReadCloser, core.ObjectInfo, error) {
	file, err := os.Open(ref.Key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, core.ObjectInfo{}, core.NewObjectNotFoundError(ref.Key)
		}
		s.logger.Error("failed to open file", zap.Error(err))
		return nil, core.ObjectInfo{}, fmt.Errorf("failed to open file: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		s.logger.Error("failed to stat file", zap.Error(err))
	}
	return file, core.ObjectInfo{
		Ref:       ref,
		SizeBytes: stat.Size(),
	}, nil
}

func (s *ObjectStore) Delete(_ context.Context, ref core.ObjectRef) error {
	if err := os.Remove(ref.Key); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.logger.Warn("file does not exist, nothing to delete", zap.String("key", ref.Key))
			return core.NewObjectNotFoundError(ref.Key)
		}
		s.logger.Error("failed to remove file", zap.Error(err))
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

func (s *ObjectStore) Exists(_ context.Context, ref core.ObjectRef) (bool, error) {
	_, err := os.Stat(ref.Key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		s.logger.Error("failed to stat file", zap.Error(err))
		return false, fmt.Errorf("failed to stat file: %w", err)
	}
	return true, nil
}
