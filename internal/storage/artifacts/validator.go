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

package artifacts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	runtime "github.com/mikhail5545/wasmforge/internal/runtime/core"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
	"go.uber.org/zap"
)

type (
	ValidationInput struct {
		Name     string
		SizeHint int64
		R        io.Reader
		Ref      core.ObjectRef
	}

	ValidationResult struct {
		ChecksumSHA256Hex string
		Metadata          Metadata
	}

	Metadata struct {
		ABI        string             `json:"abi"`
		ModuleName string             `json:"module_name"`
		Imports    []FunctionMetadata `json:"imports"`
		Exports    []FunctionMetadata `json:"exports"`
	}

	FunctionMetadata struct {
		Name        string   `json:"name"`
		ParamTypes  []string `json:"param_types"`
		ParamNames  []string `json:"param_names"`
		ReturnTypes []string `json:"return_types"`
	}
)

type Validator struct {
	runtime runtime.Runtime
	logger  *zap.Logger
}

func NewValidator(runtime runtime.Runtime, logger *zap.Logger) *Validator {
	return &Validator{
		runtime: runtime,
		logger:  logger.With(zap.String("domain", "storage"), zap.String("component", "artifact_validator")),
	}
}

func (v *Validator) Validate(ctx context.Context, input ValidationInput) (ValidationResult, error) {
	v.logger.Debug("validating artifact", zap.String("name", input.Name), zap.String("sizeHint", fmt.Sprint(input.SizeHint)))
	bytes, err := io.ReadAll(input.R)
	if err != nil {
		v.logger.Error("failed to read artifact content", zap.Error(err))
		return ValidationResult{}, fmt.Errorf("failed to read artifact bytes: %w", err)
	}

	compiled, err := v.runtime.CompileModule(ctx, bytes)
	if err != nil {
		v.logger.Error("failed to compile artifact", zap.Error(err))
		return ValidationResult{}, core.NewInvalidObjectFormatError("invalid artifact")
	}

	metadata := Metadata{
		ModuleName: compiled.Name(),
		ABI:        "wasmforge:v1",
		Imports:    make([]FunctionMetadata, 0, len(compiled.ImportedFunctions())),
		Exports:    make([]FunctionMetadata, 0, len(compiled.ExportedFunctions())),
	}
	for _, def := range compiled.ImportedFunctions() {
		paramTypes := make([]string, 0, len(def.ParamTypes()))
		returnTypes := make([]string, 0, len(def.ResultTypes()))
		for _, paramType := range def.ParamTypes() {
			paramTypes = append(paramTypes, string(paramType))
		}
		for _, returnType := range def.ResultTypes() {
			returnTypes = append(returnTypes, string(returnType))
		}

		metadata.Imports = append(metadata.Imports, FunctionMetadata{
			Name:        def.Name(),
			ParamNames:  def.ParamNames(),
			ParamTypes:  paramTypes,
			ReturnTypes: returnTypes,
		})
	}

	checksum := sha256Hex(bytes)

	return ValidationResult{
		ChecksumSHA256Hex: checksum,
		Metadata:          metadata,
	}, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
