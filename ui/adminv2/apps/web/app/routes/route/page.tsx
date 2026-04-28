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

"use client"

import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useRouter, useSearchParams } from "next/navigation"
import { useData } from "@/hooks/use-data"
import { Route } from "types/route"
import { Spinner } from "@workspace/ui/components/spinner"
import { AlertModal } from "@/components/dialog/alert-modal"
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"
import { Button } from "@workspace/ui/components/button"
import { RouteCard } from "@/components/ui/route-card"
import React from "react"
import { format } from "date-fns"
import { DateTimeInput } from "@/components/date-time-input"
import {
  RoutePluginsResponse,
  RouteSummaryResponse,
  TimeseriesResponse,
} from "@/types/ProxyServerStatistics"
import {
  StatusCountsRadarChart,
  StatusPercentagesBarChart,
} from "@/components/status-codes-charts"
import { useMutation } from "@/hooks/use-mutation"
import { usePaginatedData } from "@/hooks/use-paginated-data"
import { RoutePlugin } from "@/types/RoutePlugin"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import {
  ChevronDownIcon,
  ChevronLeft,
  ChevronRight, Database,
  MoreVertical,
  PencilIcon,
  RouteOff,
  TrashIcon,
  WrenchIcon,
} from "lucide-react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuGroup,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from "@workspace/ui/components/dropdown-menu"
import { DropdownMenuContent } from "@radix-ui/react-dropdown-menu"
import { RoutePluginsListControls } from "@/components/route-plugins-list-controls"
import { Separator } from "@workspace/ui/components/separator"
import { useRange } from "@/hooks/use-range"
import {
  Popover,
  PopoverContent,
  PopoverDescription,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from "@workspace/ui/components/popover"
import { NormalizeTimeSeriesPointsStatusCodes } from "@/lib/normalize-stats"
import { TimeSeriesAreaChart, TimeSeriesStatusCodeLineChart } from "@/components/time-series-charts"
import { RouteFlowRuntime } from "@/components/route-flow-runtime"

function RoutePageContent() {
  const params = useSearchParams()
  const path = params.get("path") ?? ""
  const routeData = useData<Route>(
    `http://localhost:8080/api/routes/${encodeURIComponent(path)}`,
    "route"
  )
  const router = useRouter()

  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")
  const [successRedirect, setSuccessRedirect] = React.useState<string | null>(
    null
  )

  const overviewRange = useRange()
  const timeSeriesRange = useRange()

  const overViewData = useData<RouteSummaryResponse>(
    `http://localhost:8080/api/proxy/stats/route?path=${encodeURIComponent(path)}&from=${overviewRange.range.from.date.toISOString()}&to=${overviewRange.range.to.date.toISOString()}`,
    "route"
  )

  const timeSeriesData = useData<TimeseriesResponse>(
    `http://localhost:8080/api/proxy/stats/timeseries?route=${encodeURIComponent(path)}&from=${timeSeriesRange.range.from.date.toISOString()}&to=${timeSeriesRange.range.to.date.toISOString()}`,
    "timeseries"
  )
  const routePluginMetricsData = useData<RoutePluginsResponse>(
    `http://localhost:8080/api/proxy/stats/route/plugins?path=${encodeURIComponent(path)}&from=${overviewRange.range.from.date.toISOString()}&to=${overviewRange.range.to.date.toISOString()}`,
    "route_plugins",
  )

  const [orderDirection, setOrderDirection] = React.useState<string>("asc")
  const [orderField, setOrderField] = React.useState<string>("created_at")
  const [perPage, setPerPage] = React.useState<string>("10")
  const [showDeleteConfirmation, setShowDeleteConfirmation] =
    React.useState(false)

  const routePluginsData = usePaginatedData<RoutePlugin>(
    `/api/route-plugins?r_ids=${routeData.data?.id ?? ""}`,
    "route_plugins",
    parseInt(perPage),
    orderField,
    orderDirection as "asc" | "desc",
    { preload: true }
  )

  const toggleEnable = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }

    const enabled = routeData.data.enabled
    const callUrl = `http://localhost:8080/api/routes/${routeData.data.id}/${enabled ? "disable" : "enable"}`

    const response = await mutate(callUrl, "POST")

    if (response.success) {
      setShowSuccess(true)
      setSuccessMessage(
        `Route ${enabled ? "disabled" : "enabled"} successfully`
      )
      await routeData.refetch()
    }
  }, [mutate, routeData])

  const deleteRoute = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }

    const response = await mutate(
      `http://localhost:8080/api/routes/${routeData.data.id}`,
      "DELETE"
    )

    if (response.success) {
      setShowSuccess(true)
      setSuccessMessage(
        "Route deleted successfully. You will be redirected to routes list in 5 seconds."
      )
      setSuccessRedirect("/routes")
    }
  }, [mutate, routeData.data])

  const normalizedTimeSeriesPoints = NormalizeTimeSeriesPointsStatusCodes(timeSeriesData.data?.points)

  return (
    <SidebarLayout page_title={"Route Details"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!routeData.error || !!error || !!overViewData.error || !!routePluginMetricsData.error || !!routePluginsData.error}
        description={
          routeData?.error?.message ||
          error?.message ||
          overViewData?.error?.message ||
          routePluginMetricsData?.error?.message ||
          routePluginsData?.error?.message ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={async () => {
          if (routeData.error) {
            await routeData.refetch()
          }
          if (overViewData.error) {
            await overViewData.refetch()
          }
          if (routePluginMetricsData.error) {
            await routePluginMetricsData.refetch()
          }
          if (routePluginsData.error) {
            await routePluginsData.refetch()
          }
          if (error) {
            reset()
          }
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        title={"Success"}
        visible={showSuccess}
        description={successMessage}
        onClose={() => {
          setShowSuccess(false)
          if (successRedirect) router.push(successRedirect)
        }}
      />
      <Dialog
        open={showDeleteConfirmation}
        onOpenChange={setShowDeleteConfirmation}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Are you sure?</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            This action cannot be undone. This will permanently delete route and
            remove all associated route plugins.
          </DialogDescription>
          <DialogFooter>
            <Button
              variant={"outline"}
              onClick={() => setShowDeleteConfirmation(false)}
            >
              Cancel
            </Button>
            <Button
              variant={"destructive"}
              onClick={deleteRoute}
              disabled={loading || !!error}
            >
              {loading && <Spinner />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <div className={"flex flex-col p-6"}>
        {(routeData.loading && routeData.data === null) ||
        overViewData.loading ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <div className={"flex flex-col"}>
            {routeData.error ? (
              <div></div>
            ) : (
              <div className={"flex flex-col gap-5"}>
                <div className={"flex flex-col gap-5 lg:flex-row"}>
                  <div>
                    <RouteCard
                      route={routeData.data!}
                      onEnableToggle={toggleEnable}
                      enabling={loading}
                      onDelete={() => setShowDeleteConfirmation(true)}
                    />
                  </div>
                  <div className={"flex w-2/3 flex-col gap-5"}>
                    <div className={"flex w-full flex-row gap-5"}>
                      <Card className={"w-1/3 text-center"}>
                        <CardHeader>
                          <CardDescription>Average RPS</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {overViewData.data?.summary.avg_rps === undefined
                              ? "N/A"
                              : overViewData.data?.summary.avg_rps}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/3 text-center"}>
                        <CardHeader>
                          <CardDescription>Average Latency ms</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {overViewData.data?.summary.avg_latency_ms ===
                            undefined
                              ? "N/A"
                              : overViewData.data?.summary.avg_latency_ms}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/3 text-center"}>
                        <CardHeader>
                          <CardDescription>Total Requests</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {overViewData.data?.summary.total_requests ===
                            undefined
                              ? "N/A"
                              : overViewData.data?.summary.total_requests}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                    </div>
                    {routePluginsData.loading ? (
                      <div className={"flex items-center justify-center py-20"}>
                        <Spinner className={"h-8 w-8"} />
                      </div>
                    ) : (
                      <div className={"flex w-full flex-col gap-5"}>
                        <div className={"flex w-full flex-col gap-1"}>
                          <p className={"text-xl"}>Associated Plugins</p>
                          <p className={"text-muted-foreground"}>
                            This plugins are working on this route as middleware
                            in WASM runtime.
                          </p>
                        </div>
                        {routePluginsData.data.length === 0 ? (
                          <Empty>
                            <EmptyHeader>
                              <EmptyMedia variant={"icon"}>
                                <RouteOff />
                              </EmptyMedia>
                              <EmptyTitle>No Associated Plugins</EmptyTitle>
                              <EmptyDescription>
                                You haven&#39;t attached any plugins to this
                                route yet. Please create a new plugin and attach
                                it to this route, or attach an existing one.
                              </EmptyDescription>
                            </EmptyHeader>
                            <EmptyContent
                              className={"flex-row justify-center gap-4"}
                            >
                              <Button size={"sm"} variant={"outline"} asChild>
                                <a href={"/plugins/new"}>Create Plugin</a>
                              </Button>
                              <Button size={"sm"} asChild>
                                <a
                                  href={`/routes/plugins/new?routeId=${routeData.data?.id}`}
                                >
                                  Attach Plugin
                                </a>
                              </Button>
                            </EmptyContent>
                          </Empty>
                        ) : (
                          <div className={"overflow-hidden rounded-lg border"}>
                            <Table>
                              <TableHeader
                                className={"sticky top-0 z-10 bg-muted"}
                              >
                                <TableRow>
                                  <TableHead>Name</TableHead>
                                  <TableHead>Version Constraint</TableHead>
                                  <TableHead>Execution Order</TableHead>
                                  <TableHead>Resolved Plugin Version</TableHead>
                                  <TableHead></TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {routePluginsData.data.map((plugin) => (
                                  <TableRow key={plugin.id}>
                                    <TableCell>{plugin.plugin?.name}</TableCell>
                                    <TableCell>
                                      {plugin.version_constraint}
                                    </TableCell>
                                    <TableCell>
                                      {plugin.execution_order}
                                    </TableCell>
                                    <TableCell>
                                      {plugin.resolved_plugin_version}
                                    </TableCell>
                                    <TableCell>
                                      <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                          <Button
                                            variant={"ghost"}
                                            size={"icon"}
                                          >
                                            <MoreVertical />
                                          </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent>
                                          <DropdownMenuGroup>
                                            <DropdownMenuItem asChild>
                                              <a
                                                href={`/routes/plugins/plugin?pluginId=${plugin.id}`}
                                              >
                                                <WrenchIcon />
                                                Details
                                              </a>
                                            </DropdownMenuItem>
                                            <DropdownMenuItem asChild>
                                              <a
                                                href={`/routes/plugins/edit?pluginId=${plugin.id}`}
                                              >
                                                <PencilIcon />
                                                Edit
                                              </a>
                                            </DropdownMenuItem>
                                          </DropdownMenuGroup>
                                          <DropdownMenuSeparator />
                                          <DropdownMenuItem
                                            variant={"destructive"}
                                          >
                                            <TrashIcon />
                                            Delete
                                          </DropdownMenuItem>
                                        </DropdownMenuContent>
                                      </DropdownMenu>
                                    </TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          </div>
                        )}
                      </div>
                    )}
                    <div className={"flex flex-row justify-end gap-5"}>
                      <div
                        className={
                          "flex flex-row items-center justify-center gap-2"
                        }
                      >
                        <p className={"text-sm font-semibold"}>Rows per page</p>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant={"outline"}>
                              {perPage}
                              <ChevronDownIcon />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent>
                            <DropdownMenuRadioGroup
                              value={perPage}
                              onValueChange={setPerPage}
                            >
                              <DropdownMenuRadioItem value={"5"}>
                                5
                              </DropdownMenuRadioItem>
                              <DropdownMenuRadioItem value={"10"}>
                                10
                              </DropdownMenuRadioItem>
                              <DropdownMenuRadioItem value={"20"}>
                                20
                              </DropdownMenuRadioItem>
                              <DropdownMenuRadioItem value={"30"}>
                                30
                              </DropdownMenuRadioItem>
                              <DropdownMenuRadioItem value={"40"}>
                                40
                              </DropdownMenuRadioItem>
                            </DropdownMenuRadioGroup>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                      <div
                        className={
                          "flex flex-row items-center justify-center gap-2"
                        }
                      >
                        <Button
                          variant={"outline"}
                          size={"icon"}
                          disabled={
                            routePluginsData.loading ||
                            routePluginsData.previousPageToken === ""
                          }
                          onClick={() => routePluginsData.previousPage()}
                        >
                          <ChevronLeft />
                        </Button>
                        <Button
                          variant={"outline"}
                          size={"icon"}
                          disabled={
                            routePluginsData.loading ||
                            routePluginsData.nextPageToken === ""
                          }
                          onClick={() => routePluginsData.nextPage()}
                        >
                          <ChevronRight />
                        </Button>
                      </div>
                      <RoutePluginsListControls
                        orderDirection={orderDirection}
                        setOrderDirection={setOrderDirection}
                        orderField={orderField}
                        setOrderField={setOrderField}
                        showCreateButton={false}
                      />
                    </div>
                  </div>
                </div>
                <Separator />
                <RouteFlowRuntime
                  routePath={routeData.data!.path}
                  targetURL={routeData.data!.target_url}
                  routePlugins={routePluginsData.data}
                  pluginMetrics={routePluginMetricsData.data?.plugins ?? []}
                  pluginsLoading={routePluginsData.loading}
                  metricsLoading={routePluginMetricsData.loading}
                  metricsError={routePluginMetricsData.error?.message}
                />
                <Separator />
                <div className={"flex flex-col gap-5"}>
                  <div className={"flex flex-row items-center gap-5"}>
                    <p className={"text-2xl"}>Status Codes Statistics</p>
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button variant={"outline"}>Set Time Range</Button>
                      </PopoverTrigger>
                      <PopoverContent className={"w-80"}>
                        <PopoverHeader>
                          <PopoverTitle>Select a time range</PopoverTitle>
                          <PopoverDescription>
                            Overview will be shown accordingly
                          </PopoverDescription>
                        </PopoverHeader>
                        <div className={"flex flex-col gap-4"}>
                          <Separator />
                          <div className={"flex flex-col gap-1"}>
                            <p className={"text-lg font-semibold"}>Show from</p>
                            <DateTimeInput
                              date={overviewRange.range.from.date}
                              setDate={overviewRange.updateFrom}
                              time={overviewRange.range.from.time}
                              setTime={overviewRange.updateFromTime}
                              layout={"row"}
                            />
                          </div>
                          <Separator />
                          <div className={"flex flex-col gap-1"}>
                            <p className={"text-lg font-semibold"}>Show To</p>
                            <DateTimeInput
                              date={overviewRange.range.to.date}
                              setDate={overviewRange.updateTo}
                              time={overviewRange.range.to.time}
                              setTime={overviewRange.updateToTime}
                              layout={"row"}
                            />
                          </div>
                        </div>
                      </PopoverContent>
                    </Popover>
                  </div>
                  {overViewData.loading ? (
                    <div className={"flex items-center justify-center py-20"}>
                      <Spinner className={"size-8"} />
                    </div>
                  ) : (
                    <div className={"flex flex-col gap-5 lg:flex-row"}>
                      <div className={"w-full lg:w-1/2"}>
                        <StatusPercentagesBarChart
                          percentages={
                            overViewData.data?.summary
                              .status_code_percentages || {}
                          }
                          title={"Status Code Percentages"}
                          description={`Showing status code percentages in a period from ${format(overviewRange.range.from.date, "PPP-HH:mm")} to ${format(overviewRange.range.to.date, "PPP-HH:mm")}`}
                        />
                      </div>
                      <div className={"w-full lg:w-1/2"}>
                        <StatusCountsRadarChart
                          counts={
                            overViewData.data?.summary.status_code_counts || {}
                          }
                          title={"Status Code Counts"}
                          description={`Showing status code counts in a period from ${format(overviewRange.range.from.date, "PPP-HH:mm")} to ${format(overviewRange.range.to.date, "PPP-HH:mm")}`}
                        />
                      </div>
                    </div>
                  )}
                </div>
                <Separator />
                <div className={"flex flex-col gap-5"}>
                  <div className={"flex flex-row items-center gap-5"}>
                    <p className={"text-2xl"}>Time Series</p>
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button variant={"outline"}>Set Time Range</Button>
                      </PopoverTrigger>
                      <PopoverContent className={"w-80"}>
                        <PopoverHeader>
                          <PopoverTitle>Select a time range</PopoverTitle>
                          <PopoverDescription>
                            Time Series will be shown accordingly
                          </PopoverDescription>
                        </PopoverHeader>
                        <div className={"flex flex-col gap-4"}>
                          <Separator />
                          <div className={"flex flex-col gap-1"}>
                            <p className={"text-lg font-semibold"}>Show from</p>
                            <DateTimeInput
                              date={timeSeriesRange.range.from.date}
                              setDate={timeSeriesRange.updateFrom}
                              time={timeSeriesRange.range.from.time}
                              setTime={timeSeriesRange.updateFromTime}
                              layout={"row"}
                            />
                          </div>
                          <Separator />
                          <div className={"flex flex-col gap-1"}>
                            <p className={"text-lg font-semibold"}>Show To</p>
                            <DateTimeInput
                              date={timeSeriesRange.range.to.date}
                              setDate={timeSeriesRange.updateTo}
                              time={timeSeriesRange.range.to.time}
                              setTime={timeSeriesRange.updateToTime}
                              layout={"row"}
                            />
                          </div>
                        </div>
                      </PopoverContent>
                    </Popover>
                  </div>
                  {timeSeriesData.loading ? (
                    <div className={"flex items-center justify-center py-20"}>
                      <Spinner className={"size-8"} />
                    </div>
                  ) : (
                    <>
                      {timeSeriesData.data?.points.length === 0 ||
                      !timeSeriesData.data ? (
                        <Empty>
                          <EmptyHeader>
                            <EmptyMedia variant={"icon"}>
                              <Database />
                            </EmptyMedia>
                            <EmptyTitle>Not Enough Data</EmptyTitle>
                            <EmptyDescription>
                              Not enough data to show time series. Try expanding
                              the time range or check back later
                            </EmptyDescription>
                          </EmptyHeader>
                        </Empty>
                      ) : (
                        <>
                          <TimeSeriesAreaChart
                            title={"Total Requests Over Time"}
                            description={`Showing how many requests were sent in a period from ${format(timeSeriesRange.range.from.date, "PPP-HH:mm")} to ${format(timeSeriesRange.range.to.date, "PPP-HH:mm")}`}
                            timeSeries={timeSeriesData.data.points}
                          />
                          {normalizedTimeSeriesPoints.map((series) => (
                            <TimeSeriesStatusCodeLineChart
                              color={`var(--chart-1)`}
                              key={series.code}
                              buckets={series}
                              title={`Status code ${series.code} over time`}
                              description={`Showing how many responses with status code ${series.code} were sent in a period from ${format(timeSeriesRange.range.from.date, "PPP-HH:mm")} to ${format(timeSeriesRange.range.to.date, "PPP-HH:mm")}`}
                            />
                          ))}
                        </>
                      )}
                    </>
                  )}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}

export default function RoutePage() {
  return (
    <React.Suspense
      fallback={
        <SidebarLayout page_title={"Route Details"}>
          <div className={"flex items-center justify-center p-6"}>
            <Spinner className={"size-10"} />
          </div>
        </SidebarLayout>
      }
    >
      <RoutePageContent />
    </React.Suspense>
  )
}
