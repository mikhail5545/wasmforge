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

import { CodeBlock } from "@/components/code-block"
import { DocSection, InfoCard, PayloadTable } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"

const toc = [
  ["model", "Route Model"],
  ["lifecycle", "Lifecycle"],
  ["create", "Create and Update"],
  ["methods", "Method Policies"],
  ["runtime", "Runtime Behavior"],
]

const routePayload = `{
  "path": "/api/orders",
  "target_url": "https://orders.internal",
  "allowed_methods": ["GET", "POST"],
  "idle_conn_timeout": 90,
  "tls_handshake_timeout": 10,
  "expect_continue_timeout": 1,
  "response_header_timeout": 30,
  "max_idle_conns": 100,
  "max_idle_conns_per_host": 10,
  "max_conns_per_host": 50
}`

const methodPayload = `{
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

export default function RoutesDocsPage() {
  return (
    <SidebarLayout page_title={"Routes Overview"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className={"w-fit"}>
            WasmForge Gateway
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Routes</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Routes are path-prefix mappings from external HTTP traffic to an
            upstream service. A route owns transport settings, method policy,
            authentication configuration, and the ordered WASM plugin chain used
            for matching requests.
          </p>
        </header>

        <Card>
          <CardHeader>
            <CardTitle>Table of Contents</CardTitle>
            <CardDescription>
              Start with the route model, then configure runtime policy.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <nav className={"grid gap-2 sm:grid-cols-2 lg:grid-cols-5"}>
              {toc.map(([id, label]) => (
                <a
                  key={id}
                  href={`#${id}`}
                  className={
                    "rounded-lg border px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
                  }
                >
                  {label}
                </a>
              ))}
            </nav>
          </CardContent>
        </Card>

        <DocSection
          id={"model"}
          title={"Route Model"}
          description={
            "A route records what to match, where to proxy, and how the upstream transport should behave."
          }
        >
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Match"}>
              <code>path</code> is the public path prefix. For example,
              <code> /api/orders</code> matches <code>/api/orders/123</code>.
            </InfoCard>
            <InfoCard title={"Proxy"}>
              <code>target_url</code> is the upstream origin. It must include
              the scheme, such as <code>https://orders.internal</code>.
            </InfoCard>
            <InfoCard title={"Policy"}>
              Route methods, auth config, and route plugins add request limits,
              authentication rules, and custom WASM middleware.
            </InfoCard>
          </div>
          <PayloadTable
            data={[
              {
                property: "path",
                type: "string",
                description:
                  "Required. Public route prefix. Must start with /.",
              },
              {
                property: "target_url",
                type: "string",
                description:
                  "Required. Full upstream URL including http or https.",
              },
              {
                property: "allowed_methods",
                type: "string[]",
                description:
                  "Optional coarse method allow-list. Empty means all methods are accepted before method policies run.",
              },
              {
                property: "idle_conn_timeout",
                type: "number",
                description:
                  "Seconds an idle upstream connection may remain open.",
              },
              {
                property: "tls_handshake_timeout",
                type: "number",
                description:
                  "Seconds allowed for TLS handshakes to the upstream.",
              },
              {
                property: "expect_continue_timeout",
                type: "number",
                description:
                  "Seconds to wait for an upstream 100-continue response.",
              },
              {
                property: "response_header_timeout",
                type: "number",
                description: "Seconds to wait for upstream response headers.",
              },
              {
                property: "max_idle_conns",
                type: "number",
                description:
                  "Total idle upstream connections retained by the transport.",
              },
              {
                property: "max_idle_conns_per_host",
                type: "number",
                description: "Idle upstream connections retained for one host.",
              },
              {
                property: "max_conns_per_host",
                type: "number",
                description:
                  "Maximum idle plus active connections to one upstream host.",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"lifecycle"}
          title={"Lifecycle"}
          description={
            "Routes are stored first, then explicitly enabled in the live proxy."
          }
        >
          <ol className={"mx-5 list-decimal space-y-2"}>
            <li>
              Create the route with <code>POST /api/routes</code>. New routes
              are not assumed to be ready for traffic.
            </li>
            <li>
              Configure method policies, route plugins, and optional route auth.
            </li>
            <li>
              Enable the route with <code>POST /api/routes/:id/enable</code>.
              The proxy assembles the live handler chain.
            </li>
            <li>
              Update or disable the route when policy changes are needed.
              Disable removes the route from runtime routing without deleting
              stored configuration.
            </li>
            <li>Delete the route only when the mapping is no longer needed.</li>
          </ol>
        </DocSection>

        <DocSection
          id={"create"}
          title={"Create and Update"}
          description={
            "Use the admin UI for manual changes or the admin API for repeatable configuration."
          }
        >
          <CodeBlock
            tabs={[
              { label: "payload", language: "json", code: routePayload },
              {
                label: "curl",
                code: `curl -X POST http://localhost:8080/api/routes \\
  -H "Content-Type: application/json" \\
  -d @route.json`,
              },
              {
                label: "patch",
                code: `curl -X PATCH http://localhost:8080/api/routes/<route-id> \\
  -H "Content-Type: application/json" \\
  -d '{"target_url":"https://orders-v2.internal","response_header_timeout":15}'`,
              },
            ]}
          />
          <p>
            <code>GET /api/routes/:id</code> accepts either a UUIDv7 route ID or
            a URL-escaped route path. List endpoints use <code>ps</code> for
            page size, <code>pt</code> for the next-page token, <code>of</code>
            for order field, and <code>od</code> for <code>asc</code> or
            <code> desc</code>.
          </p>
        </DocSection>

        <DocSection
          id={"methods"}
          title={"Method Policies"}
          description={
            "Method policies add request controls for individual HTTP methods."
          }
        >
          <p>
            Use method policies when <code>GET</code>, <code>POST</code>, or
            another method needs different payload limits, rate limits,
            authentication requirements, or metadata. Calling
            <code> POST /api/routes/:route_id/methods</code> replaces the
            configured method policy set for the route.
          </p>
          <CodeBlock language={"json"} code={methodPayload} />
          <PayloadTable
            data={[
              {
                property: "method",
                type: "string",
                description:
                  "One of GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT.",
              },
              {
                property: "max_request_payload_bytes",
                type: "number",
                description:
                  "Rejects larger bodies with 413 Payload Too Large.",
              },
              {
                property: "request_timeout_ms",
                type: "number",
                description:
                  "Adds a request context timeout for middleware and proxy work.",
              },
              {
                property: "response_timeout_ms",
                type: "number",
                description:
                  "Returns 504 if downstream processing takes too long.",
              },
              {
                property: "rate_limit_per_minute",
                type: "number",
                description: "Per-method limit. Exceeded requests return 429.",
              },
              {
                property: "require_authentication",
                type: "boolean",
                description:
                  "Requires an Authorization header and auth context before continuing.",
              },
              {
                property: "allowed_auth_schemes",
                type: "string[]",
                description: "Optional accepted schemes such as Bearer.",
              },
              {
                property: "metadata",
                type: "object",
                description:
                  "Free-form JSON reserved for policy data and downstream consumers.",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"runtime"}
          title={"Runtime Behavior"}
          description={
            "Enabled routes are assembled into a deterministic request path."
          }
        >
          <p>
            For a matched request, WasmForge applies route-level method checks,
            native auth when configured, route plugins in ascending
            <code> execution_order</code>, and finally the reverse proxy. A
            plugin can stop the chain by calling <code>host_send_response</code>
            .
          </p>
          <p>
            Keep routes small and explicit. Prefer one route per upstream API
            boundary, use method policies for method-specific controls, and keep
            route-plugin config JSON focused on the plugin behavior for that
            route.
          </p>
        </DocSection>
      </main>
    </SidebarLayout>
  )
}
