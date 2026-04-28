/*
 * Copyright (c) 2026. Mikhail Kulik.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from "react"
import { RoutePlugin } from "@/types/RoutePlugin"
import { RoutePluginSummary } from "@/types/ProxyServerStatistics"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@workspace/ui/components/card"
import { Badge } from "@workspace/ui/components/badge"
import { AlertTriangle, ArrowRight, Gauge, Link2, Router } from "lucide-react"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@workspace/ui/components/table"
import { Spinner } from "@workspace/ui/components/spinner"

interface RouteFlowRuntimeProps {
  routePath: string
  targetURL: string
  routePlugins: readonly RoutePlugin[]
  pluginMetrics: readonly RoutePluginSummary[]
  pluginsLoading?: boolean
  metricsLoading: boolean
  metricsError?: string | null
}

function formatErrorRate(statusCodePercentages: Record<string, number>): string {
  let errorRate = 0
  for (const [statusCode, percentage] of Object.entries(statusCodePercentages)) {
    if ((statusCode.startsWith("4") || statusCode.startsWith("5")) && Number.isFinite(percentage)) {
      errorRate += percentage
    }
  }
  return `${errorRate.toFixed(1)}%`
}

interface RuntimePluginStage {
  key: string
  routePluginID: string
  pluginID: string
  pluginName: string
  executionOrder: number
  metric: RoutePluginSummary | null
  metadataAvailable: boolean
}

const orderPluginStages = (left: RuntimePluginStage, right: RuntimePluginStage): number => {
  if (left.executionOrder !== right.executionOrder) {
    return left.executionOrder - right.executionOrder
  }
  return left.key.localeCompare(right.key)
}

export const RouteFlowRuntime: React.FC<RouteFlowRuntimeProps> = ({
  routePath,
  targetURL,
  routePlugins,
  pluginMetrics,
  pluginsLoading = false,
  metricsLoading,
  metricsError = null,
}) => {
  const metricsByRoutePluginID = new Map(pluginMetrics.map((metric) => [metric.route_plugin_id, metric]))
  const routePluginIDs = new Set(routePlugins.map((plugin) => plugin.id))

  const pluginStages: RuntimePluginStage[] = routePlugins
    .map((plugin) => {
      const metric = metricsByRoutePluginID.get(plugin.id) ?? null
      return {
        key: plugin.id,
        routePluginID: plugin.id,
        pluginID: plugin.plugin_id,
        pluginName: plugin.plugin?.name ?? metric?.plugin_name ?? plugin.plugin_id,
        executionOrder: plugin.execution_order,
        metric,
        metadataAvailable: !!plugin.plugin,
      }
    })
    .concat(
      pluginMetrics
        .filter((metric) => !routePluginIDs.has(metric.route_plugin_id))
        .map((metric) => ({
          key: `metric-${metric.route_plugin_id}`,
          routePluginID: metric.route_plugin_id,
          pluginID: metric.plugin_id,
          pluginName: metric.plugin_name || metric.plugin_id,
          executionOrder: metric.execution_order,
          metric,
          metadataAvailable: false,
        }))
    )
    .sort(orderPluginStages)

  const hasPluginStages = pluginStages.length > 0
  const showLoadingState = (pluginsLoading || metricsLoading) && !hasPluginStages

  return (
    <Card>
      <CardHeader>
        <CardTitle className={"flex items-center gap-2"}>
          <Gauge className={"size-4"} />
          <span>Flow & Runtime</span>
        </CardTitle>
        <CardDescription>
          Visual execution chain for this route with live plugin metrics.
        </CardDescription>
      </CardHeader>
      <CardContent className={"flex flex-col gap-6"}>
        <div className={"flex flex-wrap items-center gap-2 rounded-lg border p-4"}>
          <Badge variant={"outline"}>Gateway</Badge>
          <ArrowRight className={"size-4 text-muted-foreground"} />
          <Badge variant={"secondary"} className={"font-mono"}>{routePath}</Badge>
          {pluginStages.map((stage) => (
            <React.Fragment key={stage.key}>
              <ArrowRight className={"size-4 text-muted-foreground"} />
              <Badge variant={"outline"}>
                {stage.pluginName} #{stage.executionOrder}
              </Badge>
            </React.Fragment>
          ))}
          <ArrowRight className={"size-4 text-muted-foreground"} />
          <Badge className={"max-w-full truncate"}>{targetURL}</Badge>
        </div>

        <div className={"overflow-hidden rounded-lg border"}>
          <Table>
            <TableHeader className={"sticky top-0 z-10 bg-muted"}>
              <TableRow>
                <TableHead>Plugin</TableHead>
                <TableHead>Order</TableHead>
                <TableHead>Requests</TableHead>
                <TableHead>Avg Latency ms</TableHead>
                <TableHead>Error Rate</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {showLoadingState ? (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className={"flex items-center justify-center py-6"}>
                      <Spinner className={"size-5"} />
                    </div>
                  </TableCell>
                </TableRow>
              ) : metricsError ? (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className={"flex items-center justify-center gap-2 py-6 text-muted-foreground"}>
                      <AlertTriangle className={"size-4"} />
                      <span>{metricsError}</span>
                    </div>
                  </TableCell>
                </TableRow>
              ) : !hasPluginStages ? (
                <TableRow>
                  <TableCell colSpan={5}>
                    <div className={"flex items-center justify-center gap-2 py-6 text-muted-foreground"}>
                      <Router className={"size-4"} />
                      <span>No plugins attached to this route.</span>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                pluginStages.map((stage) => {
                  const metric = stage.metric
                  return (
                    <TableRow key={stage.key}>
                      <TableCell className={"flex items-center gap-2"}>
                        <span>{stage.pluginName}</span>
                        {!stage.metadataAvailable && (
                          <Badge variant={"secondary"}>metrics-only</Badge>
                        )}
                      </TableCell>
                      <TableCell>{stage.executionOrder}</TableCell>
                      <TableCell>{metric?.total_requests ?? "—"}</TableCell>
                      <TableCell>{metric?.avg_latency_ms?.toFixed(2) ?? "—"}</TableCell>
                      <TableCell>
                        {metric ? formatErrorRate(metric.status_code_percentages) : "—"}
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>
        <p className={"flex items-center gap-2 text-sm text-muted-foreground"}>
          <Link2 className={"size-4"} />
          Plugin metrics are measured per route-plugin execution stage.
          {metricsLoading && hasPluginStages && <Spinner className={"size-4"} />}
        </p>
      </CardContent>
    </Card>
  )
}

