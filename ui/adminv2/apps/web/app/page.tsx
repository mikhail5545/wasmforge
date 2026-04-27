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

import { useData } from "@/hooks/use-data"
import { ProxyServerStatus } from "@/types/ProxyServerStatus"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Spinner } from "@workspace/ui/components/spinner"
import React from "react"
import { Badge } from "@workspace/ui/components/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemMedia,
  ItemTitle,
} from "@workspace/ui/components/item"
import {
  BadgeCheck,
  ChevronRight,
  Database,
  MoreHorizontal,
  Power,
  PowerOff, RotateCcw,
  Settings,
  TriangleAlert,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { StatusCountsRadarChart, StatusPercentagesBarChart } from "@/components/status-codes-charts"
import { format } from "date-fns"
import { TimeSeriesAreaChart, TimeSeriesStatusCodeLineChart } from "@/components/time-series-charts"
import { NormalizeTimeSeriesPointsStatusCodes } from "@/lib/normalize-stats"
import { useRange } from "@/hooks/use-range"
import { Separator } from "@workspace/ui/components/separator"
import {
  Popover,
  PopoverContent,
  PopoverHeader,
  PopoverTrigger,
  PopoverTitle,
  PopoverDescription
} from "@workspace/ui/components/popover"
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { DateTimeInput } from "@/components/date-time-input"
import {
  OverviewResponse,
  TimeseriesResponse,
} from "@/types/ProxyServerStatistics"
import { useMutation } from "@/hooks/use-mutation"
import { AlertModal } from "@/components/dialog/alert-modal"

export default function Page() {

  const proxyServerStatus = useData<ProxyServerStatus>(
    "http://localhost:8080/api/proxy/config",
    "status"
  );

  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")

  const overviewRange = useRange()
  const timeSeriesRange = useRange()

  const overViewData = useData<OverviewResponse>(
    `http://localhost:8080/api/proxy/stats/overview?from=${overviewRange.range.from.date.toISOString()}&to=${overviewRange.range.to.date.toISOString()}`,
    'overview',
  )
  const timeSeriesData = useData<TimeseriesResponse>(
    `http://localhost:8080/api/proxy/stats/timeseries?from=${overviewRange.range.from.date.toISOString()}&to=${overviewRange.range.to.date.toISOString()}`,
    "timeseries"
  )

  const toggleRun = React.useCallback(async() => {
    if (!proxyServerStatus.data) return

    const callUrl = `http://localhost:8080/api/proxy/server/${proxyServerStatus.data.running ? "stop" : "start"}`

    const response = await mutate(
      callUrl,
      'POST'
    )

    if (response.success){
      setShowSuccess(true)
      setSuccessMessage(proxyServerStatus.data.running ? "Proxy server stopped successfully" : "Proxy server started successfully")
      await proxyServerStatus.refetch()
    }

  }, [mutate, proxyServerStatus])

  const restart = React.useCallback(async () => {
    if (!proxyServerStatus.data) return

    const response = await mutate(
      'http://localhost:8080/api/proxy/server/restart',
      'POST'
    )

    if (response.success) {
      setShowSuccess(true)
      setSuccessMessage("Proxy server restarted successfully")
    }
    await proxyServerStatus.refetch()
  }, [mutate, proxyServerStatus])

  const normalizedTimeSeriesData = NormalizeTimeSeriesPointsStatusCodes(timeSeriesData.data?.points)

  return (
    <SidebarLayout page_title={"Dashboard"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={
          !!proxyServerStatus.error ||
          !!error ||
          !!overViewData.error ||
          !!timeSeriesData.error
        }
        description={
          proxyServerStatus?.error?.message ||
          error?.message ||
          overViewData?.error?.message ||
          timeSeriesData?.error?.message ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={() => {
          if (proxyServerStatus.error) void proxyServerStatus.refetch()
          if (overViewData.error) void overViewData.refetch()
          if (timeSeriesData.error) void timeSeriesData.refetch()
          if (error) reset()
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
        }}
      />
      <div className={"flex flex-col p-6"}>
        {proxyServerStatus.loading ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-8"} />
          </div>
        ) : (
          <div className={"flex flex-col gap-5"}>
            <div className={"flex flex-col gap-5 lg:flex-row"}>
              <div className={"w-full lg:w-1/3"}>
                <Card className={"w-full"}>
                  <CardHeader
                    className={"flex flex-row items-center justify-between"}
                  >
                    <div className={"flex flex-col gap-1"}>
                      <CardTitle>Proxy Server</CardTitle>
                      <CardDescription>{`::${proxyServerStatus.data?.config.listen_port}`}</CardDescription>
                    </div>
                    <div
                      className={
                        "flex flex-row items-center justify-center gap-2"
                      }
                    >
                      <Badge
                        className={
                          proxyServerStatus.data?.running
                            ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
                            : "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
                        }
                      >
                        {proxyServerStatus.data?.running
                          ? "Running"
                          : "Stopped"}
                      </Badge>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant={"ghost"} size={"icon"}>
                            <MoreHorizontal />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent>
                          <DropdownMenuItem
                            onClick={toggleRun}
                            disabled={loading || !!error}
                          >
                            {proxyServerStatus.data?.running ? (
                              <>
                                <PowerOff />
                                Stop
                              </>
                            ) : (
                              <>
                                <Power />
                                Start
                              </>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuItem asChild>
                            <a href={"/settings"}>
                              <Settings />
                              Settings
                            </a>
                          </DropdownMenuItem>
                          {proxyServerStatus.data?.running && (
                            <DropdownMenuItem
                              onClick={restart}
                              disabled={loading || !!error}
                            >
                              <RotateCcw />
                              Restart
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </CardHeader>
                  <CardContent className={"flex flex-col gap-5"}>
                    <Item variant={"outline"} size={"sm"} asChild>
                      <a href={"/settings"}>
                        <ItemMedia>
                          {proxyServerStatus.data?.config.tls_enabled ? (
                            <BadgeCheck />
                          ) : (
                            <TriangleAlert />
                          )}
                        </ItemMedia>
                        <ItemContent>
                          <ItemTitle>
                            {proxyServerStatus.data?.config.tls_enabled
                              ? "TLS is configured"
                              : "TLS is not configured"}
                          </ItemTitle>
                          <ItemDescription>
                            {proxyServerStatus.data?.config.tls_enabled
                              ? "Server is configured to use TLS for secure communication"
                              : "Server is not configured to use TLS. Communication may not be secure"}
                          </ItemDescription>
                        </ItemContent>
                        <ItemActions>
                          <ChevronRight />
                        </ItemActions>
                      </a>
                    </Item>
                  </CardContent>
                </Card>
              </div>
              <div className={"w-full lg:w-2/3"}>
                <div className={"flex flex-col gap-5"}>
                  <div className={"flex w-full flex-row gap-5"}>
                    <Card className={"w-1/4 text-center"}>
                      <CardHeader>
                        <CardDescription>Total Requests</CardDescription>
                        <CardTitle className={"text-4xl"}>
                          {overViewData.loading ? (
                            <Spinner />
                          ) : (
                            overViewData.data?.total_requests
                          )}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card className={"w-1/4 text-center"}>
                      <CardHeader>
                        <CardDescription>Average RPS</CardDescription>
                        <CardTitle className={"text-4xl"}>
                          {overViewData.loading ? (
                            <Spinner />
                          ) : (
                            overViewData.data?.avg_rps
                          )}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card className={"w-1/4 text-center"}>
                      <CardHeader>
                        <CardDescription>Average Latency ms</CardDescription>
                        <CardTitle className={"text-4xl"}>
                          {overViewData.loading ? (
                            <Spinner />
                          ) : (
                            overViewData.data?.avg_latency_ms
                          )}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                    <Card className={"w-1/4 text-center"}>
                      <CardHeader>
                        <CardDescription>Dropped Events</CardDescription>
                        <CardTitle className={"text-4xl"}>
                          {overViewData.loading ? (
                            <Spinner />
                          ) : (
                            overViewData.data?.dropped_events
                          )}
                        </CardTitle>
                      </CardHeader>
                    </Card>
                  </div>
                </div>
              </div>
            </div>
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
                        overViewData.data?.status_code_percentages || {}
                      }
                      title={"Status Code Percentages"}
                      description={`Showing status code percentages in a period from ${format(overviewRange.range.from.date, "PPP-HH:mm")} to ${format(overviewRange.range.to.date, "PPP-HH:mm")}`}
                    />
                  </div>
                  <div className={"w-full lg:w-1/2"}>
                    <StatusCountsRadarChart
                      counts={overViewData.data?.status_code_counts || {}}
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
                          Not enough data to show time series. Try expanding the
                          time range or check back later
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
                      {normalizedTimeSeriesData.map((series) => (
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
    </SidebarLayout>
  )
}
