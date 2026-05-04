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
  ["scopes", "Metric Scopes"],
  ["collection", "Collection Pipeline"],
  ["queries", "Query Stats"],
  ["responses", "Response Fields"],
]

const queryExamples = `curl "http://localhost:8080/api/proxy/stats/overview?from=2026-05-03T00:00:00Z&to=2026-05-03T01:00:00Z"

curl "http://localhost:8080/api/proxy/stats/timeseries?route=/api/orders&bucket_seconds=60"

curl "http://localhost:8080/api/proxy/stats/route/plugins?path=/api/orders"`

export default function StatsDocsPage() {
  return (
    <SidebarLayout page_title={"Stats Collection"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className={"w-fit"}>
            WasmForge Gateway
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Stats</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            WasmForge records gateway telemetry for overall traffic, individual
            routes, and plugins inside a route chain. The admin UI uses these
            aggregates to show throughput, latency, status code distribution,
            and route/plugin hotspots.
          </p>
        </header>

        <TocInPage description={"Learn how to use stats collectors and query data."} data={toc}/>

        <DocSection
          id={"scopes"}
          title={"Metric Scopes"}
          description={
            "Choose the narrowest scope that answers the operational question."
          }
        >
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Overall"}>
              Gateway-wide request count, average RPS, average latency, status
              code counts, and dropped telemetry events.
            </InfoCard>
            <InfoCard title={"Route"}>
              Per-route request volume and latency. Use this when one upstream
              is slow or returning unexpected status codes.
            </InfoCard>
            <InfoCard title={"Plugin"}>
              Per-route-plugin totals and latency. Use this to identify slow or
              failing WASM middleware in an enabled route chain.
            </InfoCard>
          </div>
        </DocSection>

        <DocSection
          id={"collection"}
          title={"Collection Pipeline"}
          description={
            "Request telemetry is collected asynchronously so stats do not dominate request latency."
          }
        >
          <p>
            Proxy middleware emits telemetry events with route, plugin, status
            code, and duration data. A background collector aggregates events
            into time buckets, periodically flushes them to storage, and exposes
            query-oriented summaries through <code>/api/proxy/stats/*</code>.
          </p>
          <p>
            If the telemetry channel is saturated, events can be dropped instead
            of blocking request handling. The overview response includes
            <code> dropped_events</code> so operators can detect when collection
            capacity is too low for current traffic.
          </p>
        </DocSection>

        <DocSection
          id={"queries"}
          title={"Query Stats"}
          description={
            "Stats endpoints accept optional ISO-8601 windows. Missing windows use service defaults."
          }
        >
          <CodeBlock code={queryExamples} />
          <PayloadTable
            data={[
              {
                property: "from",
                type: "string",
                description:
                  "Optional start timestamp, usually RFC3339/ISO-8601.",
              },
              {
                property: "to",
                type: "string",
                description: "Optional end timestamp.",
              },
              {
                property: "route",
                type: "string",
                description:
                  "Optional route path for overview and timeseries endpoints.",
              },
              {
                property: "path",
                type: "string",
                description:
                  "Required route path for route summary and route plugin breakdowns.",
              },
              {
                property: "limit",
                type: "number",
                description:
                  "Routes ranking limit. Valid range is 1 through 200.",
              },
              {
                property: "bucket_seconds",
                type: "number",
                description:
                  "Timeseries bucket width. Valid range is 1 through 3600.",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"responses"}
          title={"Response Fields"}
          description={
            "All stats responses include the resolved time window and scope-specific data."
          }
        >
          <PayloadTable
            data={[
              {
                property: "total_requests",
                type: "number",
                description:
                  "Number of requests counted in the selected window.",
              },
              {
                property: "avg_rps",
                type: "number",
                description:
                  "Average requests per second over the selected window.",
              },
              {
                property: "avg_latency_ms",
                type: "number",
                description: "Average request latency in milliseconds.",
              },
              {
                property: "status_code_counts",
                type: "object",
                description: "Map of HTTP status code to request count.",
              },
              {
                property: "status_code_percentages",
                type: "object",
                description:
                  "Map of HTTP status code to percentage of total requests.",
              },
              {
                property: "points",
                type: "array",
                description:
                  "Timeseries buckets with bucket_start, total_requests, avg_latency_ms, and status_code_counts.",
              },
            ]}
          />
        </DocSection>
      </main>
    </SidebarLayout>
  )
}
