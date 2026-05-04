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

const pluginUploadExample = `curl -X POST http://localhost:8080/api/plugins \\
  -F 'wasm_file=@./target/wasm32-wasip1/release/filter.wasm' \\
  -F 'metadata={"name":"header_filter","version":"1.0.0","filename":"header_filter.wasm"}'`

export default function PluginsApiReferencePage() {
  return (
    <SidebarLayout page_title={"Plugins API Reference"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className="w-fit">
            Reference
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Plugins API</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Admin API endpoints for uploading versioned WASM modules and binding
            them to routes with semantic version constraints.
          </p>
        </header>

        <ReferenceSection
          title={"Plugins"}
          description={"Upload and manage versioned WASM modules."}
        >
          <EndpointGrid>
            <Endpoint method={"GET"} path={"/api/plugins/:id"}>
              Get by UUID, plugin name, or filename. Optional{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>v</code>{" "}
              selects version.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={"/api/plugins?ids=&n=&v=&fn=&of=&od=&ps=&pt="}
            >
              List plugins by ID, name, version, filename, and sort order.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/plugins"}>
              Multipart upload with{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                wasm_file
              </code>{" "}
              and JSON{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                metadata
              </code>
              .
            </Endpoint>
            <Endpoint method={"DELETE"} path={"/api/plugins/:id"}>
              Delete plugin metadata and uploaded module file.
            </Endpoint>
          </EndpointGrid>
          <CodeBlock code={pluginUploadExample} />
        </ReferenceSection>

        <ReferenceSection
          title={"Route Plugins"}
          description={
            "Attach plugins to routes with version constraints and JSON config."
          }
        >
          <EndpointGrid>
            <Endpoint method={"GET"} path={"/api/route-plugins/:id"}>
              Get a route-plugin attachment.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={"/api/route-plugins?ids=&r_ids=&p_ds=&of=&od=&ps=&pt="}
            >
              List attachments by route, plugin, and order.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/route-plugins"}>
              Create an attachment.
            </Endpoint>
            <Endpoint method={"PATCH"} path={"/api/route-plugins/:id"}>
              Update plugin, version constraint, execution order, or config.
            </Endpoint>
            <Endpoint method={"DELETE"} path={"/api/route-plugins/:id"}>
              Remove an attachment from the route chain.
            </Endpoint>
          </EndpointGrid>
          <CodeBlock
            language={"json"}
            code={`{
  "route_id": "<route-id>",
  "plugin_id": "<plugin-id>",
  "version_constraint": ">=1.0.0 <2.0.0",
  "execution_order": 10,
  "config": "{\\"header\\":\\"x-gateway\\",\\"value\\":\\"wasmforge\\"}"
}`}
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
