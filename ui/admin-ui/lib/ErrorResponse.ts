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

export function isErrorResponse(obj: any): obj is WasmForge.ErrorResponse {
    return !!obj && typeof obj === "object"
        && typeof obj.code === "string"
        && typeof obj.message === "string"
        && typeof obj.details === "string";
}

export function makeFallbackErrorResponse(status?: number | string, statusText?: string, body?: any): WasmForge.ErrorResponse {
    return {
        code: `HTTP_${status ?? "UNKNOWN"}`,
        message: statusText ?? "Unknown error",
        details: body ? (typeof body === 'string' ? body : JSON.stringify(body)) : "No additional details"
    };
}

export function parseErrorResponse(obj: any): WasmForge.ErrorResponse | null {
    if (!obj) return null;
    if (isErrorResponse(obj)) return obj;
    if (isErrorResponse(obj.error)) return obj.error;
    return null;
}