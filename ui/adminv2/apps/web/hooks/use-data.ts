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

import { useEffect, useState, useCallback, useRef } from "react"
import {
  parseErrorResponse,
  makeFallbackErrorResponse,
} from "@/lib/ErrorResponse"
import { ErrorResponse } from "@/types/ErrorResponse"
import { getApiBaseUrl } from "@/config"

interface Data<T> {
  data: T | null
  loading: boolean
  error: ErrorResponse | null
  refetch: () => Promise<void>
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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}

export function useData<T>(path: string | null, key: string): Data<T> {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState<boolean>(() => !!path)
  const [error, setError] = useState<ErrorResponse | null>(null)
  const isMountedRef = useRef(true)
  const abortControllerRef = useRef<AbortController | null>(null)
  const pendingRequestsRef = useRef(0)
  const requestIdRef = useRef(0)

  useEffect(() => {
    return () => {
      isMountedRef.current = false
      abortControllerRef.current?.abort()
    }
  }, [])

  const fetchData = useCallback(async () => {
    const requestId = ++requestIdRef.current

    // If no path is provided, don't attempt to fetch.
    if (!path) {
      abortControllerRef.current?.abort()
      pendingRequestsRef.current = 0

      if (!isMountedRef.current) {
        return
      }

      setLoading(false)
      setError(null)
      setData(null)
      return
    }

    abortControllerRef.current?.abort()
    const controller = new AbortController()
    abortControllerRef.current = controller

    pendingRequestsRef.current += 1

    setLoading(true)
    setError(null)

    const url = `${getApiBaseUrl()}${path}`
    try {
      const response = await fetch(url, { method: "GET", signal: controller.signal })
      const text = await response.text()
      const parsed = parseResponseBody(text)

      if (!isMountedRef.current || requestId !== requestIdRef.current) {
        return
      }

      if (!response.ok) {
        const parsedError = parseErrorResponse(parsed)
        if (parsedError) {
          setError(parsedError)
        } else {
          setError(
            makeFallbackErrorResponse(
              response.status,
              response.statusText,
              parsed
            )
          )
        }
        return
      }

      // Successful response. Try to extract requested key or use entire body.
      if (isRecord(parsed) && key in parsed) {
        setData(parsed[key] as T)
      } else {
        setData(parsed as T)
      }
    } catch (err: unknown) {
      if (err instanceof DOMException && err.name === "AbortError") {
        return
      }

      if (!isMountedRef.current || requestId !== requestIdRef.current) {
        return
      }

      setError(
        makeFallbackErrorResponse(
          500,
          "Unexpected error",
          err instanceof Error ? err.message : String(err)
        )
      )
    } finally {
      pendingRequestsRef.current = Math.max(0, pendingRequestsRef.current - 1)

      if (isMountedRef.current) {
        setLoading(pendingRequestsRef.current > 0)
      }
    }
  }, [path, key])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}