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
import { Endpoint } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"
import { Separator } from "@workspace/ui/components/separator"

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

export default function RoutesApiReferencePage() {
  return (
    <SidebarLayout page_title={"Routes API Reference"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className="w-fit">
            Reference
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Routes API</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Admin API endpoints for managing proxy routes, configuring
            method-specific policies, and controlling runtime enablement.
          </p>
        </header>

        <ReferenceSection
          title={"Routes"}
          description={"Reverse proxy route CRUD and runtime enablement."}
        >
          <EndpointGrid>
            <Endpoint method={"GET"} path={"/api/routes/:id"}>
              Get a route by UUID or path.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={
                "/api/routes?ids=&paths=&pids=&turls=&enabled=&of=&od=&ps=&pt="
              }
            >
              List routes with filters and cursor pagination.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/routes"}>
              Create a route.
            </Endpoint>
            <Endpoint method={"PATCH"} path={"/api/routes/:id"}>
              Update path, target URL, transport settings, or allowed methods.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/routes/:id/enable"}>
              Enable a route in the proxy builder.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/routes/:id/disable"}>
              Disable a route.
            </Endpoint>
            <Endpoint method={"DELETE"} path={"/api/routes/:id"}>
              Delete a route and remove it from runtime routing.
            </Endpoint>
          </EndpointGrid>
          <CodeBlock
            language={"json"}
            code={`{
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
}`}
          />
        </ReferenceSection>

        <ReferenceSection
          title={"Route Methods"}
          description={"Per-method runtime policy endpoints."}
        >
          <EndpointGrid>
            <Endpoint method={"GET"} path={"/api/routes/:route_id/methods"}>
              List policies for a route.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={"/api/routes/:route_id/methods/:method"}
            >
              Get one method policy.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/routes/:route_id/methods/"}>
              Replace the route method policy set.
            </Endpoint>
            <Endpoint
              method={"DELETE"}
              path={"/api/routes/:route_id/methods/:method"}
            >
              Delete one method policy.
            </Endpoint>
          </EndpointGrid>
          <CodeBlock language={"json"} code={routeMethodPayload} />
          <FieldList
            fields={[
              [
                "max_request_payload_bytes",
                "Caps request body size with http.MaxBytesReader.",
              ],
              [
                "request_timeout_ms",
                "Adds a request context timeout for downstream middleware and proxy work.",
              ],
              [
                "response_timeout_ms",
                "Buffers downstream output and returns 504 when the response is too slow.",
              ],
              [
                "rate_limit_per_minute",
                "Per-method token bucket rate limit. Exceeded requests return 429.",
              ],
              [
                "require_authentication",
                "Requires an Authorization header before continuing.",
              ],
              [
                "allowed_auth_schemes",
                "Optional schemes such as Bearer. Empty means any non-empty scheme/token pair.",
              ],
              [
                "metadata",
                "Free-form JSON copied into request context for downstream runtime logic.",
              ],
            ]}
          />
        </ReferenceSection>
      </main>
    </SidebarLayout>
  )
}

function ReferenceSection({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <section className={"scroll-mt-20"}>
      <Separator className={"mb-8"} />
      <div className={"flex flex-col gap-5"}>
        <div className={"flex flex-col gap-2"}>
          <h2 className={"text-2xl font-bold tracking-tight"}>{title}</h2>
          <p className={"text-muted-foreground"}>{description}</p>
        </div>
        <div className={"flex flex-col gap-5"}>{children}</div>
      </div>
    </section>
  )
}

function EndpointGrid({ children }: { children: ReactNode }) {
  return <div className={"grid gap-3"}>{children}</div>
}

function FieldList({ fields }: { fields: [string, string][] }) {
  return (
    <div className={"grid gap-2"}>
      {fields.map(([name, description]) => (
        <div
          key={name}
          className={"rounded-lg border px-3 py-2 text-sm leading-6"}
        >
          <code className={"font-mono"}>{name}</code>
          <span className={"text-muted-foreground"}> - {description}</span>
        </div>
      ))}
    </div>
  )
}
