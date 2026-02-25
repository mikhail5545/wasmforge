#!/usr/bin/env bash

#
# Copyright (c) 2026. Mikhail Kulik.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Generate all mocks for tests using mockgen tool. This script should be run from the project root.

if ! command -v mockgen &>/dev/null; then
    echo "mockgen could not be found. Please install it first."
    exit 1
fi

# Repository (database) packages
mockgen -destination=./internal/mocks/database/route/repository.go -package=route source=./internal/database/route/repository.go Repository
mockgen -destination=./internal/mocks/database/plugin/repository.go -package=plugin source=./internal/database/plugin/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/config/repository.go -package=config source=./internal/database/proxy/config/repository.go Repository

# Proxy packages
mockgen -destination=./internal/mocks/proxy/factory.go -package=proxy source=./internal/proxy/factory.go Factory
mockgen -destination=./internal/mocks/proxy/builder.go -package=proxy source=./internal/proxy/builder.go Builder
mockgen -destination=./internal/mocks/proxy/middleware/factory.go -package=middleware source=./internal/proxy/middleware/factory.go Factory

mockgen -destination=./internal/mocks/uploads/manager.go -package=uploads source=./internal/uploads/manager.go Manager