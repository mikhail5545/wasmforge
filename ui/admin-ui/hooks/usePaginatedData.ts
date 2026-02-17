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

import {useCallback, useEffect, useState} from "react";
import {makeFallbackErrorResponse, parseErrorResponse} from "@/lib/ErrorResponse";

const BACKEND_URL = "http://localhost:8080";

interface CacheEntry {
    data: any;
    nextPageToken: string;
    timestamp: number;
}

const cache = new Map<string, CacheEntry>();

interface PaginatedData<T> {
    pageSize: number;
    nextPageToken: string;
    loading: boolean;
    error: WasmForge.ErrorResponse | null;
    data: T[];
    nextPage: (token?: string, options?: { append?: boolean, force?: boolean }) => Promise<void>;
    refetch: (token?: string) => Promise<void>;
}

interface FetchPageOptions {
    preload?: boolean;
    maxAge?: number; // in milliseconds
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
    path:string,
    key: string,
    pageSize: number = 10,
    orderField: string = "created_at",
    orderDirection: "asc" | "desc" = "desc",
    { preload = false, maxAge = 5*60*1000 }: FetchPageOptions = {},
): PaginatedData<T> {
    const [data, setData] = useState<T[]>([]);
    const [nextPageToken, setNextPageToken] = useState<string>("");
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<WasmForge.ErrorResponse | null>(null);

    const buildUrl = useCallback(
        (orderField: string, orderDirection: "asc" | "desc", token?: string) => {
            const url = new URL(path, BACKEND_URL);
            url.searchParams.set("of", orderField);
            url.searchParams.set("od", orderDirection);
            url.searchParams.set("ps", pageSize.toString());
            if (token) {
                url.searchParams.set("pt", token);
            }
            return url.toString();
        }, [path, pageSize]
    );

    const fetchPage = useCallback(
        async(token: string = "", { append = false, force = false } = {}) => {
            const url = buildUrl(orderField, orderDirection, token);

            if (!force) {
                const cachedEntry = cache.get(url);
                if (cachedEntry && (Date.now() - cachedEntry.timestamp < maxAge)) {
                    setData((prev) => (append ? [...prev, ...cachedEntry.data[key]] : cachedEntry.data[key] || []));
                    setNextPageToken(cachedEntry.nextPageToken);
                    setLoading(false);
                    setError(null);
                    return;
                }
            }

            setLoading(true);
            setError(null);

            try {
                const response = await fetch(url, { method: "GET" });
                const text = await response.text();

                // Try to parse JSON if possible
                let parsed: any;
                try{
                    parsed = text ? JSON.parse(text) : null;
                } catch (e) {
                    parsed = text;
                }

                if (!response.ok) {
                    const parsedError = parseErrorResponse(parsed);
                    if (parsedError) {
                        setError(parsedError);
                    } else {
                        setError(makeFallbackErrorResponse(response.status, response.statusText, parsed));
                    }
                    return;
                }

                const responseData = parsed;

                const nextToken = responseData["next_page_token"] || "";
                cache.set(url, { data: responseData, timestamp: Date.now(), nextPageToken: nextToken });

                const dataArray = responseData[key] || [];
                setData((prev) => (append ? [...prev, ...dataArray] : dataArray));
                setNextPageToken(nextToken);
            } catch (err: unknown) {
                setError(makeFallbackErrorResponse(500, "Unexpected error", err instanceof Error ? err.message : String(err)));
            } finally {
                setLoading(false);
            }
    }, [buildUrl, key, maxAge, orderField, orderDirection]);

    const refetch = useCallback((token: string = "") => fetchPage(token, { force: true }), [fetchPage]);

    useEffect(() => {
        if (preload && !loading) {
            fetchPage();
        }
    }, [fetchPage, orderField, orderDirection, path, preload]);

    return { pageSize, nextPageToken, loading, error, data, nextPage: fetchPage, refetch };
}
