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

import {
  useState,
  useCallback,
  Dispatch,
  SetStateAction,
  useRef,
  useEffect,
} from "react"
import {
  makeFallbackErrorResponse,
  parseErrorResponse,
} from "@/lib/ErrorResponse"
import {ErrorResponse} from "@/types/ErrorResponse";
import { getApiBaseUrl } from "@/config"

type Method = "POST" | "PUT" | "PATCH" | "DELETE"
type MutationPayload = BodyInit | Record<string, unknown> | null
type MutationResult =
  | { success: true; response: Response; error: null }
  | { success: false; response?: undefined; error: ErrorResponse }

interface UseMutation {
  loading: boolean
  error: ErrorResponse | null
  setError: Dispatch<SetStateAction<ErrorResponse | null>>
  reset: () => void
  mutate: (
    path: string,
    method: Method,
    payload?: MutationPayload,
    extraHeaders?: Record<string, string>
  ) => Promise<MutationResult>
}

function parseResponseBody(text: string): unknown {
  if (!text) {
    return null
  }

  try {
    return JSON.parse(text) as unknown
  } catch {
    return text
  }
}

function isBodyInit(value: unknown): value is BodyInit {
  if (typeof value === "string") {
    return true
  }

  if (typeof Blob !== "undefined" && value instanceof Blob) {
    return true
  }

  if (typeof FormData !== "undefined" && value instanceof FormData) {
    return true
  }

  if (typeof URLSearchParams !== "undefined" && value instanceof URLSearchParams) {
    return true
  }

  if (typeof ReadableStream !== "undefined" && value instanceof ReadableStream) {
    return true
  }

  return value instanceof ArrayBuffer || ArrayBuffer.isView(value)
}

export function useMutation(): UseMutation {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<ErrorResponse | null>(null)
  const pendingRequestsRef = useRef(0)
  const isMountedRef = useRef(true)

  useEffect(() => {
    return () => {
      isMountedRef.current = false
    }
  }, [])

  const makeRequest = useCallback(
    async (
      path: string,
      method: Method,
      payload?: MutationPayload,
      extraHeaders: Record<string, string> = {}
    ): Promise<MutationResult> => {
      pendingRequestsRef.current += 1
      if (isMountedRef.current) {
        setLoading(true)
        setError(null)
      }

      try {
        const headers: Record<string, string> = { ...extraHeaders }
        let body: BodyInit | null = null

        if (payload !== undefined && payload !== null) {
          if (isBodyInit(payload)) {
            body = payload
          } else {
            body = JSON.stringify(payload)
            if (!headers["Content-Type"]) {
              headers["Content-Type"] = "application/json"
            }
          }
        }

        const url = `${getApiBaseUrl()}${path}`
        const response = await fetch(url, {
          method,
          headers,
          body,
        })
        const copy = response.clone()

        const text = await response.text()
        const parsed = parseResponseBody(text)

        if (!response.ok) {
          const parsedError = parseErrorResponse(parsed)
          const resolvedError =
            parsedError ??
            makeFallbackErrorResponse(
              response.status,
              response.statusText,
              parsed
            )

          if (isMountedRef.current) {
            setError(resolvedError)
          }
          return { success: false, error: resolvedError }
        }

        return { success: true, response: copy, error: null }
      } catch (err: unknown) {
        const resolvedError = makeFallbackErrorResponse(
          500,
          "Unexpected error",
          err instanceof Error ? err.message : String(err)
        )

        if (isMountedRef.current) {
          setError(resolvedError)
        }

        return { success: false, error: resolvedError }
      } finally {
        pendingRequestsRef.current = Math.max(0, pendingRequestsRef.current - 1)
        if (isMountedRef.current) {
          setLoading(pendingRequestsRef.current > 0)
        }
      }
    },
    []
  )

  const reset = useCallback(() => {
    if (!isMountedRef.current) {
      return
    }

    setError(null)
    if (pendingRequestsRef.current === 0) {
      setLoading(false)
    }
  }, [])

  return { loading, error, setError, reset, mutate: makeRequest }
}