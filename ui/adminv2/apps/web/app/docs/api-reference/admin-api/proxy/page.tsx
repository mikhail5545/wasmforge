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

import { Endpoint, ReferenceSection, EndpointGrid } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"

export default function ProxyApiReferencePage() {
  return (
    <SidebarLayout page_title={"Proxy API Reference"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className="w-fit">
            Reference
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Proxy API</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Admin API endpoints for managing proxy lifecycle, TLS certificates,
            global configuration, and extracting telemetry statistics.
          </p>
        </header>

        <ReferenceSection
          title={"Health"}
          description={"Admin server readiness."}
        >
          <EndpointGrid>
            <Endpoint method={"GET"} path={"/api/health"}>
              Returns{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                {'{ "status": "healthy" }'}
              </code>
              .
            </Endpoint>
          </EndpointGrid>
        </ReferenceSection>

        <ReferenceSection
          title={"Server Lifecycle & Config"}
          description={"Proxy startup, shutdown, and global limits."}
        >
          <EndpointGrid>
            <Endpoint method={"POST"} path={"/api/proxy/server/start"}>
              Start the proxy server.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/proxy/server/stop"}>
              Stop the proxy server.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/proxy/server/restart"}>
              Restart the proxy server.
            </Endpoint>
            <Endpoint method={"GET"} path={"/api/proxy/config"}>
              Read listen port and read-header timeout.
            </Endpoint>
            <Endpoint method={"PATCH"} path={"/api/proxy/config"}>
              Update{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                listen_port
              </code>{" "}
              or{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                read_header_timeout
              </code>
              .
            </Endpoint>
          </EndpointGrid>
        </ReferenceSection>

        <ReferenceSection
          title={"TLS Certificates"}
          description={"Manage TLS certificates for the proxy."}
        >
          <EndpointGrid>
            <Endpoint method={"POST"} path={"/api/proxy/certs"}>
              Upload multipart{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                cert_file
              </code>{" "}
              and{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                key_file
              </code>
              .
            </Endpoint>
            <Endpoint method={"DELETE"} path={"/api/proxy/certs"}>
              Remove configured TLS certificate files.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/proxy/certs/generate"}>
              Generate a self-signed cert with{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                common_name
              </code>
              ,{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                valid_days
              </code>
              , and{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                rsa_bits
              </code>
              .
            </Endpoint>
          </EndpointGrid>
        </ReferenceSection>

        <ReferenceSection
          title={"Statistics"}
          description={"Extract telemetry aggregates and timeseries data."}
        >
          <EndpointGrid>
            <Endpoint
              method={"GET"}
              path={"/api/proxy/stats/overview?from=&to=&route="}
            >
              Gateway-wide or route-scoped totals, RPS, latency, and status
              codes.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={"/api/proxy/stats/routes?from=&to=&limit="}
            >
              Ranked route summaries.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={
                "/api/proxy/stats/timeseries?from=&to=&route=&bucket_seconds="
              }
            >
              Bucketed request metrics.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={"/api/proxy/stats/route?path=&from=&to="}
            >
              Single route summary.
            </Endpoint>
            <Endpoint
              method={"GET"}
              path={"/api/proxy/stats/route/plugins?path=&from=&to="}
            >
              Per-plugin metrics for a route.
            </Endpoint>
          </EndpointGrid>
        </ReferenceSection>
      </main>
    </SidebarLayout>
  )
}
