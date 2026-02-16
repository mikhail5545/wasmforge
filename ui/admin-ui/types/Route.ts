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

namespace WasmForge {
    export interface Route {
        readonly id: string;
        created_at: string;
        path: string;
        target_url: string;
        enabled: boolean;
        idle_conn_timeout: number;
        tls_handshake_timeout: number;
        expect_continue_timeout: number;
        max_idle_cons?: number;
        max_idle_cons_per_host?: number;
        response_header_timeout?: number;
        max_cons_per_host?: number;
        plugins?: WasmForge.RoutePlugin[];
    }
}
