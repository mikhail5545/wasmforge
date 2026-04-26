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

import { useCallback, useEffect, useRef, useState } from "react"
import {
  makeFallbackErrorResponse,
  parseErrorResponse,
} from "@/lib/ErrorResponse"
import { ErrorResponse } from "@/types/ErrorResponse";

const BACKEND_URL = "http://localhost:8080"

interface CacheEntry {
  items: unknown[]
  nextPageToken: string
  timestamp: number
}

const cache = new Map<string, CacheEntry>()

type QueryParamValue = string | number | boolean | null | undefined
type QueryParams = Record<string, QueryParamValue>

interface FetchPageRequestOptions {
  append?: boolean
  force?: boolean
  queryParams?: QueryParams
}

export interface PaginatedData<T> {
  pageSize: number
  nextPageToken: string
  previousPageToken: string
  loading: boolean
  error: ErrorResponse | null
  data: T[]
  nextPage: (
    token?: string,
    options?: FetchPageRequestOptions
  ) => Promise<void>
  previousPage: (
    options?: Omit<FetchPageRequestOptions, "append">
  ) => Promise<void>
  refetch: (token?: string) => Promise<void>
  setQueryParams: (queryParams?: QueryParams) => void
}

interface FetchPageOptions {
  preload?: boolean
  maxAge?: number // in milliseconds
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

function normalizeQueryParams(queryParams?: QueryParams): Record<string, string> {
  if (!queryParams) {
    return {}
  }

  return Object.keys(queryParams)
    .sort()
    .reduce<Record<string, string>>((acc, key) => {
      const value = queryParams[key]
      if (value === null || value === undefined || value === "") {
        return acc
      }

      acc[key] = String(value)
      return acc
    }, {})
}

/**
 * Custom hook to fetch paginated data from a given API endpoint with caching and error handling.
 * @param path - API endpoint to fetch data from
 * @param key - Key in the API response that contains the array of items
 * @param pageSize - Number of items to fetch per page (default: 10)
 * @param orderField - Field to order the results by (default: "created_at")
 * @param orderDirection - Direction to order the results ("asc" or "desc", default: "desc")
 * @param preload - Whether to automatically fetch the first page on mount (default: false)
 * @param maxAge - Maximum age of cached data in milliseconds (default: 5 minutes)
 * @returns An object containing the paginated data, loading state, error state, and functions to fetch the next page and refetch data
 */
export function usePaginatedData<T>(
  path: string,
  key: string,
  pageSize: number = 10,
  orderField: string = "created_at",
  orderDirection: "asc" | "desc" = "desc",
  { preload = false, maxAge = 5 * 60 * 1000 }: FetchPageOptions = {}
): PaginatedData<T> {
  const [data, setData] = useState<T[]>([])
  const [nextPageToken, setNextPageToken] = useState<string>("")
  const [previousPageToken, setPreviousPageToken] = useState<string>("")
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<ErrorResponse | null>(null)
  const isMountedRef = useRef(true)
  const abortControllerRef = useRef<AbortController | null>(null)
  const requestIdRef = useRef(0)
  const activeQueryParamsRef = useRef<Record<string, string>>({})
  const previousTokenByCurrentTokenRef = useRef<Map<string, string>>(new Map())
  const orderSignatureRef = useRef<string>(`${orderField}:${orderDirection}`)

  const resetPaginationTokens = useCallback(() => {
    previousTokenByCurrentTokenRef.current.clear()
    setNextPageToken("")
    setPreviousPageToken("")
  }, [])

  const syncPaginationTokens = useCallback(
    (currentToken: string, nextToken: string) => {
      if (nextToken) {
        previousTokenByCurrentTokenRef.current.set(nextToken, currentToken)
      }

      setPreviousPageToken(
        previousTokenByCurrentTokenRef.current.get(currentToken) ?? ""
      )
    },
    []
  )

  useEffect(() => {
    return () => {
      isMountedRef.current = false
      abortControllerRef.current?.abort()
    }
  }, [])

  useEffect(() => {
    const currentOrderSignature = `${orderField}:${orderDirection}`
    if (orderSignatureRef.current === currentOrderSignature) {
      return
    }

    orderSignatureRef.current = currentOrderSignature
    resetPaginationTokens()
  }, [orderField, orderDirection, resetPaginationTokens])

  const buildUrl = useCallback(
    (
      orderField: string,
      orderDirection: "asc" | "desc",
      token?: string,
      queryParams?: QueryParams
    ) => {
      const url = new URL(path, BACKEND_URL)
      url.searchParams.set("of", orderField)
      url.searchParams.set("od", orderDirection)
      url.searchParams.set("ps", pageSize.toString())
      if (token) {
        url.searchParams.set("pt", token)
      }

      const normalizedQueryParams = normalizeQueryParams(queryParams)
      for (const [queryKey, queryValue] of Object.entries(normalizedQueryParams)) {
        url.searchParams.set(queryKey, queryValue)
      }

      return url.toString()
    },
    [path, pageSize]
  )

  const fetchPage = useCallback(
    async (
      token: string = "",
      { append = false, force = false, queryParams }: FetchPageRequestOptions = {}
    ) => {
      const resolvedQueryParams =
        queryParams !== undefined
          ? normalizeQueryParams(queryParams)
          : activeQueryParamsRef.current

      if (queryParams !== undefined) {
        activeQueryParamsRef.current = resolvedQueryParams
      }

      const url = buildUrl(orderField, orderDirection, token, resolvedQueryParams)
      const requestId = ++requestIdRef.current

      if (!force) {
        const cachedEntry = cache.get(url)
        if (cachedEntry && Date.now() - cachedEntry.timestamp < maxAge) {
          if (!isMountedRef.current || requestId !== requestIdRef.current) {
            return
          }

          const cachedItems = cachedEntry.items as T[]
          setData((prev) => (append ? [...prev, ...cachedItems] : cachedItems))
          setNextPageToken(cachedEntry.nextPageToken)
          syncPaginationTokens(token, cachedEntry.nextPageToken)
          setLoading(false)
          setError(null)
          return
        }
      }

      abortControllerRef.current?.abort()
      const controller = new AbortController()
      abortControllerRef.current = controller

      setLoading(true)
      setError(null)

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

        if (!isRecord(parsed)) {
          setError(
            makeFallbackErrorResponse(
              response.status,
              "Invalid response shape",
              parsed
            )
          )
          return
        }

        const rawNextToken = parsed["next_page_token"]
        const nextToken = typeof rawNextToken === "string" ? rawNextToken : ""
        const rawItems = parsed[key]
        const dataArray = Array.isArray(rawItems) ? (rawItems as T[]) : []
        cache.set(url, {
          items: dataArray,
          timestamp: Date.now(),
          nextPageToken: nextToken,
        })

        setData((prev) => (append ? [...prev, ...dataArray] : dataArray))
        setNextPageToken(nextToken)
        syncPaginationTokens(token, nextToken)
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
        if (isMountedRef.current && requestId === requestIdRef.current) {
          setLoading(false)
        }
      }
    },
    [buildUrl, key, maxAge, orderField, orderDirection, syncPaginationTokens]
  )

  const previousPage = useCallback(
    (options: Omit<FetchPageRequestOptions, "append"> = {}) =>
      fetchPage(previousPageToken, {
        ...options,
        append: false,
      }),
    [fetchPage, previousPageToken]
  )

  const refetch = useCallback(
    (token: string = "") =>
      fetchPage(token, {
        force: true,
        append: false,
      }),
    [fetchPage]
  )

  const setQueryParams = useCallback((queryParams?: QueryParams) => {
    activeQueryParamsRef.current = normalizeQueryParams(queryParams)
    resetPaginationTokens()
  }, [resetPaginationTokens])

  useEffect(() => {
    if (!preload) {
      return
    }

    void fetchPage()
  }, [fetchPage, preload, path, orderField, orderDirection])

  return {
    pageSize,
    nextPageToken,
    previousPageToken,
    loading,
    error,
    data,
    nextPage: fetchPage,
    previousPage,
    refetch,
    setQueryParams,
  }
}