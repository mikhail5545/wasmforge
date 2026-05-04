/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import type { ReactNode } from "react"

import { CodeBlock } from "@/components/code-block"
import { DocSection, Endpoint, InfoCard, TocCrossPage } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Separator } from "@workspace/ui/components/separator"

const authConfigPayload = `{
  "validate_tokens": true,
  "issue_tokens": true,
  "key_backend_type": "database",
  "token_ttl_seconds": 3600,
  "issuer": "wasmforge",
  "audience": "orders-api",
  "allowed_algorithms": ["RS256"],
  "required_claims": ["sub"],
  "claims_mapping": {},
  "metadata": {
    "upstream_auth_header": "Authorization"
  }
}`

const routeMethodPayload = `{
  "methods": [
    {
      "method": "POST",
      "max_request_payload_bytes": 1048576,
      "request_timeout_ms": 3000,
      "response_timeout_ms": 5000,
      "rate_limit_per_minute": 120,
      "require_authentication": true,
      "allowed_auth_schemes": ["Bearer"],
      "metadata": {
        "policy": "write-orders"
      }
    }
  ]
}`

const pluginUploadExample = `curl -X POST http://localhost:8080/api/plugins \\
  -F 'wasm_file=@./target/wasm32-wasip1/release/filter.wasm' \\
  -F 'metadata={"name":"header_filter","version":"1.0.0","filename":"header_filter.wasm"}'`

const rustImportsExample = `#[link(wasm_import_module = "env")]
unsafe extern "C" {
    fn host_get_header(k_ptr: *const u8, k_len: i32, buf_ptr: *const u8, buf_len: i32) -> i32;
    fn host_set_header(k_ptr: *const u8, k_len: i32, v_ptr: *const u8, v_len: i32);
    fn host_get_json_config(buf_ptr: *const u8, buf_len: i32) -> i32;
    fn host_auth_is_authenticated() -> i32;
    fn host_auth_subject(buf_ptr: *const u8, buf_len: i32) -> i32;
    fn host_auth_claim(k_ptr: *const u8, k_len: i32, buf_ptr: *const u8, buf_len: i32) -> i32;
}

#[no_mangle]
pub extern "C" fn on_request() {
    // inspect request, mutate headers, or send an early response
}`

const toc = [
  { label: "Proxy API", link: "/docs/api-reference/admin-api/proxy", sub: [] },
  { label: "Routes", link: "/docs/api-reference/admin-api/routes", sub: [] },
  { label: "Plugins", link: "/docs/api-reference/admin-api/plugins", sub: [] },
  { label: "Auth", link: "/docs/api-reference/admin-api/auth", sub: [] },
  { label: "WASM Exports", link: "/docs/api-reference/wasm-exports", sub: [] },
]

export default function ApiReferencePage() {
  return (
    <SidebarLayout page_title={"API Reference"}>
      <main
        className={
          "mx-auto flex w-full max-w-6xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"}>Reference</Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>API Reference</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Admin API endpoints and WASM host imports used by adminv2,
            automation, and gateway plugins. Base URL defaults to{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              http://localhost:8080
            </code>
            .
          </p>
        </header>

        <TocCrossPage data={toc}/>

        <Card>
          <CardHeader>
            <CardTitle>Conventions</CardTitle>
            <CardDescription>
              All admin API routes are rooted at{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>/api</code>.
            </CardDescription>
          </CardHeader>
          <CardContent
            className={"grid gap-4 text-sm leading-7 md:grid-cols-3"}
          >
            <InfoCard title={"Responses"}>
              Object responses are wrapped by resource name, for example{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                {'{ "route": ... }'}
              </code>
              . Paginated lists include{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                next_page_token
              </code>
              .
            </InfoCard>
            <InfoCard title={"IDs"}>
              IDs are UUIDv7 strings. Some get endpoints also accept natural
              identifiers such as route path, plugin name, or plugin filename.
            </InfoCard>
            <InfoCard title={"Pagination"}>
              List endpoints use{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>ps</code> for
              page size and{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>pt</code> for
              page token.
            </InfoCard>
          </CardContent>
        </Card>

        <DocSection
          id={"overview"}
          title={"Overview"}
          description={"Overview on HTTP admin API and WASM host exports"}
        >
          <p>
            WasmForge exposes two different interaction plains: HTTP admin API and WASM host exports.
            You can use admin API to manipulate routes, plugins, auth and statistics while WASM exports
            allow your plugins to interact with requests that go through your proxy, extract configuration,
            and perform authentication and authorization checks. WASM exports also allow plugins
            to use internal logger, intercept requests and immediately send response with modified body, headers or status code.
          </p>
          <p>
            For easier integration, WasmForge offers small SDK wrappers for your C++ and Rust plugins. They import
            the underlying WASM host functions and expose a more ergonomic API. You can find examples of how to use them in the{" "}
            <a href={"/docs/api-reference/wasm-exports"} className={"underline"}>
              exports reference
            </a>
          </p>
        </DocSection>

      </main>
    </SidebarLayout>
  )
}