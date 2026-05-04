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
import { DocSection, InfoCard, PayloadTable, TocInPage } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"

const toc = [
  ["concepts", "Core Concepts"],
  ["upload", "Upload a Plugin"],
  ["binding", "Bind to a Route"],
  ["versioning", "Versioning and Auto-Switch"],
]

const uploadExample = `curl -X POST http://localhost:8080/api/plugins \\
  -F "wasm_file=@./target/wasm32-wasip1/release/header_filter.wasm" \\
  -F 'metadata={"name":"header_filter","version":"1.0.0","filename":"header_filter.wasm"}'`

const bindingExample = `{
  "route_id": "018f0000-0000-7000-8000-000000000001",
  "plugin_id": "018f0000-0000-7000-8000-000000000010",
  "version_constraint": ">=1.0.0 <2.0.0",
  "execution_order": 10,
  "config": "{\\"header\\":\\"x-gateway\\",\\"value\\":\\"wasmforge\\"}"
}`

const rustPluginExample = `use wasmforge_sdk::{proxy, proxy_handler};

fn handle_request() {
    let path = proxy::get_path().unwrap_or_default();

    if path == "/api/orders/internal" {
        proxy::send_response(403, "forbidden");
        return;
    }

    proxy::set_header("x-wasmforge-plugin", "header_filter");
    proxy::log(proxy::LogLevel::INFO, "request accepted");
}

proxy_handler!(handle_request);`

export default function PluginsDocsPage() {
  return (
    <SidebarLayout page_title={"Plugins Overview"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className={"w-fit"}>
            WasmForge Gateway
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Plugins</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Plugins are WebAssembly modules that run inside the gateway request
            chain. They can read request data, set request headers, inspect
            route-specific JSON config, log through the host, and return an
            early response before the upstream service is called.
          </p>
        </header>

        <TocInPage description={"Create, upload, bind to route and configure auto-switching"} data={toc} />

        <DocSection
          id={"concepts"}
          title={"Core Concepts"}
          description={
            "Plugin artifacts are reusable; route-plugin bindings decide where and how they run."
          }
        >
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Artifact"}>
              A plugin record points to an uploaded <code>.wasm</code> file and
              is identified by <code>name</code>, <code>version</code>, and
              <code> filename</code>.
            </InfoCard>
            <InfoCard title={"Binding"}>
              A route-plugin binding attaches one plugin to one route, sets a
              semantic version constraint, and provides optional config.
            </InfoCard>
            <InfoCard title={"Execution"}>
              Bindings execute by ascending <code>execution_order</code>. Lower
              values run earlier and can short-circuit later plugins.
            </InfoCard>
          </div>
        </DocSection>

        <DocSection
          id={"upload"}
          title={"Upload a Plugin"}
          description={
            "Upload uses multipart form data: the WASM file plus JSON metadata."
          }
        >
          <CodeBlock code={uploadExample} />
          <PayloadTable
            data={[
              {
                property: "wasm_file",
                type: "file",
                description:
                  "Required multipart file field containing the compiled module.",
              },
              {
                property: "metadata.name",
                type: "string",
                description:
                  "Required plugin name. Use lowercase words separated by underscores, such as header_filter.",
              },
              {
                property: "metadata.version",
                type: "string",
                description: "Required semantic version, such as 1.0.0.",
              },
              {
                property: "metadata.filename",
                type: "string",
                description:
                  "Required stored WASM filename. Must end with .wasm.",
              },
            ]}
          />
          <p>
            Upload responses return <code>{'{ "plugin": ... }'}</code>. The
            stored module is hashed during upload, and the metadata is used by
            route-plugin version resolution.
          </p>
        </DocSection>

        <DocSection
          id={"binding"}
          title={"Bind to a Route"}
          description={
            "Bindings make plugin behavior route-specific without duplicating the WASM artifact."
          }
        >
          <CodeBlock language={"json"} code={bindingExample} />
          <PayloadTable
            data={[
              {
                property: "route_id",
                type: "string",
                description: "Required UUIDv7 route ID.",
              },
              {
                property: "plugin_id",
                type: "string",
                description: "Required UUIDv7 plugin artifact ID.",
              },
              {
                property: "version_constraint",
                type: "string",
                description:
                  "Semantic version constraint used to resolve compatible plugin versions.",
              },
              {
                property: "execution_order",
                type: "number",
                description:
                  "Ascending order in the route chain. Use gaps like 10, 20, 30 for easier inserts.",
              },
              {
                property: "config",
                type: "string",
                description:
                  "Optional JSON string exposed through host_get_json_config.",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"versioning"}
          title={"Versioning and Auto-Switch"}
          description={
            "Version constraints let routes pick up compatible plugin releases."
          }
        >
          <p>
            When a new plugin version is uploaded, WasmForge can update
            compatible route-plugin bindings to the new artifact and reassemble
            affected enabled routes. Use fixed versions for high-risk routes,
            patch ranges for conservative maintenance, and wider ranges only
            when the plugin has stable compatibility guarantees.
          </p>
          <CodeBlock
            language={"text"}
            code={`1.2.3              exact version
~1.2.0             patch updates within 1.2.x
^1.2.0             compatible minor and patch updates
>=1.0.0 <2.0.0     explicit supported major range`}
          />
        </DocSection>

        <DocSection
          id={"guest-code"}
          title={"Guest Code Shape"}
          description={
            "A plugin must export on_request. The Rust SDK macro wraps that export for you."
          }
        >
          <CodeBlock language={"rust"} code={rustPluginExample} />
          <p>
            Plugins should treat host calls as fallible. Buffer-returning host
            functions return <code>0xFFFFFFFF</code> when data is missing or an
            access violation occurs, and return the number of bytes written on
            success.
          </p>
        </DocSection>
      </main>
    </SidebarLayout>
  )
}
