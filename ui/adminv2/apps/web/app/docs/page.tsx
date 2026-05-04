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
import { DocSection, InfoCard, TocCrossPage } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"

const docsToc = [
  {
    label: "Authentication",
    link: "/docs/auth",
    sub: [
      ["overview", "Authentication Overview"],
      ["auth-config", "Authentication Config"],
      ["configuration-policies", "Configuration Policies"],
      ["key-backends", "Key Backends"],
      ["key-encryption", "Private Key Encryption"],
      ["plugin-auth-context", "Plugin Auth Context"],
      ["considerations", "Security Considerations"],
      ["troubleshooting", "Troubleshooting"],
    ],
  },
  {
    label: "Routes",
    link: "/docs/routes",
    sub: [
      ["model", "Route Model"],
      ["lifecycle", "Lifecycle"],
      ["create", "Create and Update"],
      ["methods", "Method Policies"],
      ["runtime", "Runtime Behavior"],
    ],
  },
  {
    label: "Plugins",
    link: "/docs/plugins",
    sub: [
      ["concepts", "Core Concepts"],
      ["upload", "Upload a Plugin"],
      ["binding", "Bind to a Route"],
      ["versioning", "Versioning and Auto-Switch"],
    ],
  },
  {
    label: "Statistics",
    link: "/docs/stats",
    sub: [
      ["scopes", "Metric Scopes"],
      ["collection", "Collection Pipeline"],
      ["queries", "Query Stats"],
      ["responses", "Response Fields"],
    ],
  },
]

const firstRouteExample = `curl -X POST http://localhost:8080/api/routes \\
  -H "Content-Type: application/json" \\
  -d '{
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
  }'

curl -X POST http://localhost:8080/api/routes/<route-id>/enable`

const pluginFlowExample = `cargo build --release --target wasm32-wasip1

curl -X POST http://localhost:8080/api/plugins \\
  -F "wasm_file=@./target/wasm32-wasip1/release/header_filter.wasm" \\
  -F 'metadata={"name":"header_filter","version":"1.0.0","filename":"header_filter.wasm"}'

curl -X POST http://localhost:8080/api/route-plugins \\
  -H "Content-Type: application/json" \\
  -d '{
    "route_id": "<route-id>",
    "plugin_id": "<plugin-id>",
    "version_constraint": ">=1.0.0 <2.0.0",
    "execution_order": 10,
    "config": "{\\"header\\":\\"x-gateway\\",\\"value\\":\\"wasmforge\\"}"
  }'`

export default function DocsPage() {
  return (
    <SidebarLayout page_title={"Documentation"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"}>WasmForge Gateway</Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Documentation</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Configure routes, attach WASM plugins, and add native JWT auth
            without giving plugins access to signing keys or private material.
          </p>
        </header>

        <TocCrossPage data={docsToc} />

        <DocSection
          id={"getting-started"}
          title={"Getting Started"}
          description={"Build the gateway, embedded adminv2 UI, and Go binary."}
        >
          <p>
            Install <strong>Go 1.25+</strong> and <strong>Node 22+</strong>.
            From the repository root, run one build command. The active admin UI
            is{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              ui/adminv2
            </code>
            ; the old{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              ui/admin-ui
            </code>{" "}
            directory is deprecated.
          </p>
          <CodeBlock
            tabs={[
              { label: "Make", code: "make build" },
              { label: "Bash", code: "bash ./scripts/build.sh" },
              { label: "PowerShell", code: "powershell ./scripts/build.ps1" },
            ]}
          />
          <CodeBlock
            tabs={[
              { label: "Bash", code: "./bin/wasmforge" },
              { label: "PowerShell", code: "./bin/wasmforge.exe" },
            ]}
          />
          <p>
            The admin panel and admin API are served from{" "}
            <strong>http://localhost:8080</strong>. Admin API routes are under{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>/api</code>.
          </p>
        </DocSection>

        <DocSection
          id={"admin-ui"}
          title={"Admin UI Workflow"}
          description={
            "The browser UI is a control surface for the same admin API documented in this section."
          }
        >
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Routes"}>
              Create path-to-upstream mappings, tune transport settings, manage
              allowed methods, and enable or disable live proxy handlers.
            </InfoCard>
            <InfoCard title={"Plugins"}>
              Upload WASM modules, review versions, and bind plugins to routes
              in a deterministic execution order with route-specific JSON
              configuration.
            </InfoCard>
            <InfoCard title={"Settings"}>
              Start, stop, or restart the proxy, update the listen port and
              read-header timeout, and manage TLS certificates.
            </InfoCard>
          </div>
          <p>
            Changes that affect enabled routes are applied through the proxy
            assembly path. Route enablement, route disablement, plugin binding
            changes, and compatible plugin upgrades can update live traffic
            without editing the upstream service.
          </p>
        </DocSection>

        <DocSection
          id={"quick-route"}
          title={"Create a Route"}
          description={
            "A route matches an incoming path prefix and forwards traffic to one upstream service."
          }
        >
          <p>
            Create routes while disabled, finish method policies, plugins, and
            auth settings, then enable the route. The <code>path</code> must
            start with <code>/</code>; <code>target_url</code> must be a full
            URL including scheme.
          </p>
          <CodeBlock code={firstRouteExample} />
        </DocSection>

        <DocSection
          id={"quick-plugin"}
          title={"Attach a Plugin"}
          description={
            "WASM plugins run before the request is proxied and can inspect or mutate request data."
          }
        >
          <p>
            A plugin upload creates a reusable versioned artifact. A
            route-plugin binding selects an artifact by <code>plugin_id</code>,
            constrains acceptable versions, sets <code>execution_order</code>,
            and provides an optional JSON config string to the guest module.
          </p>
          <CodeBlock code={pluginFlowExample} />
        </DocSection>
      </main>
    </SidebarLayout>
  )
}
