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

package validation

import (
	"fmt"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

// pathRegexp matches strings that start with a forward slash and can contain any characters after it.
var pathRegexp = regexp.MustCompile("^/.*$")

// wasmFilenameRegexp matches strings that consist of a valid filename (letters, numbers, underscores, or hyphens) followed by the .wasm extension.
var wasmFilenameRegexp = regexp.MustCompile("^([a-zA-Z0-9_-]+)\\.(wasm)$")
var pluginNameRegexp = regexp.MustCompile(`^[a-z0-9]+(?:_[&?a-z0-9]+)*$`)

func composeRules(required bool, additional ...validation.Rule) []validation.Rule {
	rules := additional
	if required {
		rules = append([]validation.Rule{validation.Required}, rules...)
	}
	return rules
}

func extractValue[T any](dest *T, value any) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case T:
		*dest = v
	case *T:
		if v == nil {
			return nil
		}
		*dest = *v
	default:
		return fmt.Errorf("must be a %T or *%T", *dest, *dest)
	}
	return nil
}

func isValidStringUUIDv7(strID string) error {
	if strID == "" {
		return nil // empty value is allowed
	}
	uid, err := uuid.Parse(strID)
	if err != nil {
		return fmt.Errorf("must be a valid UUIDv7: %w", err)
	}
	if uid.Version() != uuid.Version(7) {
		return fmt.Errorf("must be a valid UUIDv7")
	}
	return nil
}

func IsValidUUIDv7(value any) error {
	if value == nil {
		return nil
	}
	var strID string
	var uuidVal uuid.UUID
	strErr := extractValue(&strID, value)
	uuidErr := extractValue(&uuidVal, value)
	if strErr != nil && uuidErr != nil {
		return fmt.Errorf("must be a string UUIDv7 or uuid.UUID: %v; %v", strErr, uuidErr)
	}
	if strErr == nil {
		return isValidStringUUIDv7(strID)
	}
	if uuidVal == uuid.Nil {
		return nil
	}
	if uuidVal.Version() != uuid.Version(7) {
		return fmt.Errorf("must be a valid UUIDv7")
	}
	return nil
}

func IsValidPath(value any) error {
	if value == nil {
		return nil
	}
	var path string
	if err := extractValue(&path, value); err != nil {
		return fmt.Errorf("must be a string: %w", err)
	}
	if !pathRegexp.MatchString(path) {
		return fmt.Errorf("must start with '/'")
	}
	return nil
}

func IsValidWasmFilename(value any) error {
	if value == nil {
		return nil
	}
	var filename string
	if err := extractValue(&filename, value); err != nil {
		return fmt.Errorf("must be a string: %w", err)
	}
	if !wasmFilenameRegexp.MatchString(filename) {
		return fmt.Errorf("must match the pattern 'name.wasm' where name can contain letters, numbers, underscores, or hyphens")
	}
	return nil
}

func IsValidPluginName(value any) error {
	if value == nil {
		return nil
	}
	var name string
	if err := extractValue(&name, value); err != nil {
		return fmt.Errorf("must be a string: %w", err)
	}
	if !pluginNameRegexp.MatchString(name) {
		return fmt.Errorf("must match the pattern 'name' or 'name_part1_name_part2' where name and parts can contain lowercase letters, numbers, and underscores")
	}
	return nil
}

// UUIDRule returns the ozzo-validation rules for a UUID.
func UUIDRule(required bool) []validation.Rule {
	return composeRules(required, validation.By(IsValidUUIDv7))
}

func PathRule(required bool) []validation.Rule {
	return composeRules(required, validation.By(IsValidPath))
}

func WasmFilenameRule(required bool) []validation.Rule {
	return composeRules(required, validation.By(IsValidWasmFilename))
}

func PluginNameRule(required bool) []validation.Rule {
	return composeRules(required, validation.By(IsValidPluginName))
}
