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

import {useState, useCallback, Dispatch, SetStateAction} from "react";
import {makeFallbackErrorResponse,parseErrorResponse} from "@/lib/ErrorResponse";

type Method = "POST" | "PUT" | "PATCH" | "DELETE";

interface UseMutation {
    loading: boolean;
    error: WasmForge.ErrorResponse | null;
    setError: Dispatch<SetStateAction<WasmForge.ErrorResponse | null>>;
    mutate: (
        path: string,
        method: Method,
        payload?: BodyInit,
        extraHeaders?: Record<string, string>,
    ) => Promise<{ success: boolean, response?: Response }>;
}

export function useMutation(): UseMutation {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<WasmForge.ErrorResponse | null>(null);
    const [success, setSuccess] = useState(false);

    const makeRequest = useCallback(
        async(path: string, method: Method, payload?: BodyInit, extraHeaders: Record<string, string> = {}) => {
            setLoading(true);
            setError(null);
            setSuccess(false);

            try{
                const headers: Record<string, string> = { ...extraHeaders };
                const isFormData = typeof FormData !== "undefined" && payload instanceof FormData;
                if (!isFormData) {
                    headers["Content-Type"] = "application/json";
                }

                const response = await fetch(path, {
                    method,
                    headers,
                    body: payload ?? null,
                });

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
                    return {success: false}
                }

                setSuccess(true);
                return { success: true, response };
            } catch (err: unknown) {
                setError(makeFallbackErrorResponse(500, "Unexpected error", err instanceof Error ? err.message : String(err)));
                return { success: false };
            } finally {
                setLoading(false);
            }
        }, []
    );
    return { loading, error, setError, mutate: makeRequest };
}