[CmdletBinding()]
param(
    [switch]$SkipNpmInstall
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command "mockgen" -ErrorAction SilentlyContinue)) {
    throw "Required command 'mockgen' not found in PATH."
}

# Repository (database) packages
mockgen -destination=./internal/mocks/database/route/repository.go -package=route source=./internal/database/route/repository.go Repository
mockgen -destination=./internal/mocks/database/plugin/repository.go -package=plugin source=./internal/database/plugin/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/config/repository.go -package=config source=./internal/database/proxy/config/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/route/plugin/repository.go -package=plugin source=./internal/database/proxy/route/plugin/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/route/method/repository.go -package=method source=./internal/database/proxy/route/method/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/auth/audit/repository.go -package=audit source=./internal/database/proxy/auth/audit/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/auth/key/repository.go -package=key source=./internal/database/proxy/auth/key/repository.go Repository
mockgen -destination=./internal/mocks/database/proxy/auth/config/repository.go -package=config source=./internal/database/proxy/auth/config/repository.go Repository

# Proxy packages
mockgen -destination=./internal/mocks/proxy/factory.go -package=proxy source=./internal/proxy/factory.go Factory
mockgen -destination=./internal/mocks/proxy/builder.go -package=proxy source=./internal/proxy/builder.go Builder
mockgen -destination=./internal/mocks/proxy/middleware/factory.go -package=middleware source=./internal/proxy/middleware/factory.go Factory

# Services packages
mockgen -destination=./internal/mocks/services/auth/issuer.go -package=auth source=./internal/services/auth/issuer.go TokenIssuer
mockgen -destination=./internal/mocks/services/auth/key_manager.go -package=auth source=./internal/services/auth/key_maanger.go KeyManager
mockgen -destination=./internal/mocks/services/auth/validator.go -package=auth source=./internal/services/auth/validator.go TokenValidator

mockgen -destination=./internal/mocks/uploads/manager.go -package=uploads source=./internal/uploads/manager.go Manager
