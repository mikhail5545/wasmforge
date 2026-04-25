/*
 * Copyright (c) 2026. Mikhail Kulik.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {ErrorResponse} from "@/types/ErrorResponse";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

function parseValidationDetails(details: string): Record<string, string> | null {
  const normalized = details.trim()
  if (!normalized) {
    return null
  }

  const withoutPrefix = normalized.replace(/^validation failed:\s*/i, "")
  const fieldErrors: Record<string, string> = {}

  for (const segment of withoutPrefix.split(";")) {
    const entry = segment.trim()
    if (!entry) {
      continue
    }

    const separatorIndex = entry.indexOf(":")
    if (separatorIndex <= 0) {
      continue
    }

    const field = entry.slice(0, separatorIndex).trim()
    const message = entry.slice(separatorIndex + 1).trim()
    if (!field || !message) {
      continue
    }

    fieldErrors[field] = message
  }

  return Object.keys(fieldErrors).length > 0 ? fieldErrors : null
}

function withValidationErrors(error: ErrorResponse): ErrorResponse {
  if (error.code !== "VALIDATION_FAILED") {
    return error
  }

  const validationErrors = parseValidationDetails(error.details)
  if (!validationErrors) {
    return error
  }

  return {
    ...error,
    validationErrors,
  }
}

export function isErrorResponse(obj: unknown): obj is ErrorResponse {
  if (!isRecord(obj)) {
    return false
  }

  return (
    typeof obj.code === "string" &&
    typeof obj.message === "string" &&
    typeof obj.details === "string"
  )
}

export function makeFallbackErrorResponse(
  status?: number | string,
  statusText?: string,
  body?: unknown
): ErrorResponse {
  const details = (() => {
    if (body === null || body === undefined) {
      return "No additional details"
    }

    if (typeof body === "string") {
      return body
    }

    try {
      return JSON.stringify(body)
    } catch {
      return String(body)
    }
  })()

  return {
    code: `HTTP_${status ?? "UNKNOWN"}`,
    message: statusText ?? "Unknown error",
    details,
  }
}

export function parseErrorResponse(obj: unknown): ErrorResponse | null {
  if (!obj) return null
  if (isErrorResponse(obj)) return withValidationErrors(obj)

  if (isRecord(obj) && isErrorResponse(obj.error)) {
    return withValidationErrors(obj.error)
  }

  return null
}