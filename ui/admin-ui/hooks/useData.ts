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

import {useEffect, useState, useCallback} from "react";
import {parseErrorResponse, makeFallbackErrorResponse} from "@/lib/ErrorResponse";

interface Data<T>{
    data: T;
    loading: boolean;
    error: WasmForge.ErrorResponse | null;
    refetch: () => Promise<void>;
}

export function useData<T>(path: string | null, key: string) : Data<T> {
    const [data, setData] = useState<T>(null as any);
    const [loading, setLoading] = useState<boolean>(() => !!path);
    const [error, setError] = useState<WasmForge.ErrorResponse | null>(null);

    const fetchData = useCallback(
        async() => {
            // If no path is provided, don't attempt to fetch.
            if (!path) {
                setLoading(false);
                setError(null);
                setData(null as any);
                return;
            }

            setLoading(true);
            setError(null);

            try{
                const response = await fetch(path, { method: "GET" });
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

                // Successful response. Try to extract requested key or use entire body.
                if (parsed && typeof parsed === 'object' && key in parsed) {
                    setData(parsed[key] as T);
                } else {
                    // If parsed is an object with a single root matching 'data' or similar, try to use parsed.
                    setData(parsed as T);
                }
            } catch (err: unknown) {
                setError(makeFallbackErrorResponse(500, "Unexpected error", err instanceof Error ? err.message : String(err)));
            } finally {
                setLoading(false);
            }
        }, [path, key]
    );

    useEffect(() => {
        // If there's no path, ensure state reflects that (no loading, no data)
        if (!path) {
            setData(null as any);
            setLoading(false);
            setError(null);
            return;
        }
        fetchData();
    }, [fetchData, path]);

    return { data, loading, error, refetch: fetchData };
}