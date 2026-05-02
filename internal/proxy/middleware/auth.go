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

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	auditrepo "github.com/mikhail5545/wasmforge/internal/database/auth/audit"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	auditmodel "github.com/mikhail5545/wasmforge/internal/models/auth/audit"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	authsvc "github.com/mikhail5545/wasmforge/internal/services/auth"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
	"go.uber.org/zap"
)

type authMiddleware struct {
	configRepo configrepo.Repository
	validator  authsvc.TokenValidator
	issuer     authsvc.TokenIssuer
	auditRepo  auditrepo.Repository
	logger     *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware that validates and manages JWTs.
func NewAuthMiddleware(
	configRepo configrepo.Repository,
	validator authsvc.TokenValidator,
	issuer authsvc.TokenIssuer,
	auditRepo auditrepo.Repository,
	logger *zap.Logger,
) func(http.Handler) http.Handler {
	am := &authMiddleware{
		configRepo: configRepo,
		validator:  validator,
		issuer:     issuer,
		auditRepo:  auditRepo,
		logger:     logger.With(zap.String("component", "auth-middleware")),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			am.ServeHTTP(w, r, next)
		})
	}
}

func (am *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	ctx := r.Context()
	state := reqctx.RequestStateFromContext(ctx)
	logger := reqctx.LoggerFromContext(ctx)

	// Initialize auth context
	state.AuthContext = &reqctx.AuthContext{
		IsAuthenticated: false,
	}

	// Get RouteID from context
	routeID, ok := reqctx.RouteIDFromContext(ctx)
	if !ok || routeID == uuid.Nil {
		logger.Warn("RouteID not found in context, skipping auth middleware")
		next.ServeHTTP(w, r)
		return
	}

	// Load auth config for this route
	authConfig, err := am.configRepo.Get(ctx, configrepo.WithRouteIDs(routeID))
	if err != nil {
		logger.Error("failed to get auth config for route", zap.String("route_id", routeID.String()), zap.Error(err))
		am.respondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	// If auth config doesn't exist or is disabled, continue to next handler
	if authConfig == nil || !authConfig.Enabled {
		logger.Debug("auth config not found or disabled for route", zap.String("route_id", routeID.String()))
		next.ServeHTTP(w, r)
		return
	}

	state.AuthContext.AuthConfig = authConfig

	// If token validation is enabled, validate the bearer token
	if authConfig.ValidateTokens {
		token := am.extractBearerToken(r)
		if token == "" {
			logger.Warn("missing bearer token for protected route", zap.String("route_id", routeID.String()))
			am.respondError(w, http.StatusUnauthorized, "missing_token", "Missing or invalid authorization header")
			return
		}

		validatedToken, err := am.validator.ValidateToken(ctx, token, authConfig)
		if err != nil {
			am.handleValidationError(w, logger, routeID, authConfig, err, ctx)
			return
		}

		state.AuthContext.ValidatedToken = validatedToken
		state.AuthContext.Subject = validatedToken.Subject
		state.AuthContext.IsAuthenticated = true

		logger.Debug("token validated successfully",
			zap.String("route_id", routeID.String()),
			zap.String("subject", validatedToken.Subject))

		// Log successful validation to audit
		am.logAudit(ctx, logger, routeID, authConfig, auditmodel.ActionValidate, auditmodel.ResultSuccess, validatedToken.Subject, "")
	}

	// If token issuance is enabled, optionally issue a token for upstream
	if authConfig.IssueTokens {
		claims := make(map[string]interface{})
		if state.AuthContext.Subject != "" {
			claims["sub"] = state.AuthContext.Subject
		}

		issuedToken, err := am.issuer.IssueToken(ctx, claims, authConfig)
		if err != nil {
			logger.Warn("failed to issue token",
				zap.String("route_id", routeID.String()),
				zap.Error(err))
			am.logAudit(ctx, logger, routeID, authConfig, auditmodel.ActionIssue, auditmodel.ResultFailure, state.AuthContext.Subject, err.Error())
			// Don't fail the request if token issuance fails, just log it
		} else {
			headerName := metadata.UpstreamAuthHeader(authConfig)
			headerValue := issuedToken
			if strings.EqualFold(headerName, "Authorization") {
				headerValue = fmt.Sprintf("Bearer %s", issuedToken)
			}
			r.Header.Set(headerName, headerValue)
			logger.Debug("token issued successfully", zap.String("route_id", routeID.String()), zap.String("upstream_header", headerName))
			am.logAudit(ctx, logger, routeID, authConfig, auditmodel.ActionIssue, auditmodel.ResultSuccess, state.AuthContext.Subject, "")
		}
	}

	next.ServeHTTP(w, r)
}

func (am *authMiddleware) extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func (am *authMiddleware) handleValidationError(
	w http.ResponseWriter,
	logger *zap.Logger,
	routeID uuid.UUID,
	authConfig *configmodel.AuthConfig,
	err error,
	ctx context.Context,
) {
	logger.Warn("token validation failed",
		zap.String("route_id", routeID.String()),
		zap.Error(err))

	// Log failed validation to audit
	am.logAudit(ctx, logger, routeID, authConfig, auditmodel.ActionValidate, auditmodel.ResultFailure, "", err.Error())

	// Determine appropriate error code and message
	if errors.Is(err, authsvc.ErrTokenExpired) {
		am.respondError(w, http.StatusUnauthorized, "token_expired", "Token has expired")
	} else if errors.Is(err, authsvc.ErrTokenInvalid) || errors.Is(err, authsvc.ErrInvalidSignature) {
		am.respondError(w, http.StatusUnauthorized, "invalid_token", "Invalid token")
	} else if errors.Is(err, authsvc.ErrTokenMalformed) {
		am.respondError(w, http.StatusUnauthorized, "malformed_token", "Token is malformed")
	} else if errors.Is(err, authsvc.ErrMissingClaims) {
		am.respondError(w, http.StatusForbidden, "missing_claims", "Token is missing required claims")
	} else if errors.Is(err, authsvc.ErrInvalidAudience) {
		am.respondError(w, http.StatusForbidden, "invalid_audience", "Token audience is not valid for this service")
	} else if errors.Is(err, authsvc.ErrInvalidIssuer) {
		am.respondError(w, http.StatusForbidden, "invalid_issuer", "Token issuer is not trusted")
	} else {
		am.respondError(w, http.StatusUnauthorized, "authentication_failed", "Authentication failed")
	}
}

func (am *authMiddleware) respondError(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error": message,
		"code":  code,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		am.logger.Error("failed to encode error response", zap.Error(err))
	}
}

func (am *authMiddleware) logAudit(
	ctx context.Context,
	logger *zap.Logger,
	routeID uuid.UUID,
	authConfig *configmodel.AuthConfig,
	action auditmodel.Action,
	result auditmodel.Result,
	subject string,
	errorMsg string,
) {
	audit := &auditmodel.AuthAudit{
		RouteID:      routeID,
		AuthConfigID: authConfig.ID,
		Action:       action,
		Result:       result,
		Subject:      subject,
		ErrorMessage: errorMsg,
	}

	if err := am.auditRepo.Create(ctx, audit); err != nil {
		logger.Warn("failed to create audit log", zap.Error(err))
	}
}
