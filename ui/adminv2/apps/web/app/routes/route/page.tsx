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

'use client';

import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useSearchParams } from "next/navigation"
import { useData } from "@/hooks/use-data"
import { Route } from "types/route"
import { Spinner } from "@workspace/ui/components/spinner"
import { AlertModal } from "@/components/dialog/alert-modal"
import { buildMockRouteSummaryResponse } from "@/lib/mock-stats"
import {
  Card, CardContent,
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
  DialogTrigger,
  DialogClose
} from "@workspace/ui/components/dialog"
import { Button } from "@workspace/ui/components/button"
import { RouteCard } from "@/components/ui/route-card"
import React, { useCallback } from "react"
import { format } from "date-fns"
import { DateTimeInput } from "@/components/date-time-input"
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@workspace/ui/components/chart"
import { Bar, BarChart, CartesianGrid, LabelList, XAxis } from "recharts"
import { NormalizeStatusPercentages } from "@/lib/normalize-stats"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { ErrorResponse } from "@/types/ErrorResponse"

export default function RoutePage() {
  const params = useSearchParams()
  const path = params.get('path') ?? ""
  const routeData = useData<Route>(`http://localhost:8080/api/routes/${encodeURIComponent(path)}`, 'route')

  const mockRouteStats = buildMockRouteSummaryResponse("/api/users")

  const [from, setFrom] = React.useState<Date>(() => new Date())
  const [fromTime, setFromTime] = React.useState<string>("")
  const [to, setTo] = React.useState<Date>(() => new Date())
  const [toTime, setToTime] = React.useState<string>("")
  const [statsConfigDialogOpen, setStatsConfigDialogOpen] = React.useState<boolean>(false)
  const [editDialogOpen, setEditDialogOpen] = React.useState<boolean>(false)
  const [editableRouteData, setEditableRouteData] = React.useState<Route | null>(null)

  React.useEffect(() => {
    const id = setInterval(() => {
      const now = new Date().getTime()
      const hourAgo = new Date(now - 60 * 60 * 1000)
      setFrom(new Date(now))
      setFromTime(format(now, "HH:mm"))
      setTo(hourAgo)
      setToTime(format(hourAgo, "HH:mm"))
    }, 60 * 60 * 1000) // updates every one hour

    return () => clearInterval(id)
  }, [])

  const submitStatsTimestamp = React.useCallback(async () => {
    const [fromHours, fromMinutes] = fromTime.split(":").map(Number)
    const [toHours, toMinutes] = toTime.split(":").map(Number)

    if (fromHours) {
      from.setHours(fromHours)
    }
    if (fromMinutes) {
      to.setMinutes(fromMinutes)
    }
    if (toHours) {
      to.setHours(toHours)
    }
    if (toMinutes) {
      to.setMinutes(toMinutes)
    }


  }, [from, fromTime, to, toTime])

  const openEditDialog = (route: Route) => {
    setEditableRouteData(route)
    setEditDialogOpen(true)
  }

  const closeEditDialog = () => {
    setEditDialogOpen(false)
    if (!routeData.data) {
      return
    }
    setEditableRouteData(routeData.data)
  }

  const chartConfig = {
    percentage: {
      label: "Percentage",
      color: "var(--chart-4)",
    },
    label: {
      color: "var(--background)"
    }
  } satisfies ChartConfig

  const chartData = NormalizeStatusPercentages(mockRouteStats.summary.status_code_percentages)

  return (
    <SidebarLayout page_title={"Route Details"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!routeData.error}
        description={
          routeData?.error?.message ??
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={async () => {
          await routeData.refetch()
        }}
      />
      <div className={"flex flex-col p-6"}>
        {routeData.loading && routeData.data == null ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <div className={"flex flex-col"}>
            {routeData.error ? (
              <div></div>
            ) : (
              <>
                <div className={"flex flex-row items-center justify-end pb-5"}>
                  <Dialog
                    open={statsConfigDialogOpen}
                    onOpenChange={setStatsConfigDialogOpen}
                  >
                    <DialogTrigger asChild>
                      <Button variant={"ghost"}>Configure stats</Button>
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>Pick timestamp</DialogTitle>
                      </DialogHeader>
                      <DialogDescription>
                        Stats will be showed according to this timestamp.
                      </DialogDescription>
                      <p>Show stats from</p>
                      <DateTimeInput
                        date={from}
                        setDate={setFrom}
                        time={fromTime}
                        setTime={setFromTime}
                        layout={"row"}
                      />
                      <p className={"mt-4"}>Show stats to</p>
                      <DateTimeInput
                        date={to}
                        setDate={setTo}
                        time={toTime}
                        setTime={setToTime}
                        layout={"row"}
                      />
                      <DialogFooter>
                        <DialogClose asChild>
                          <Button variant={"secondary"}>Close</Button>
                        </DialogClose>
                        <Button
                          onClick={async () => {
                            await submitStatsTimestamp()
                            setStatsConfigDialogOpen(false)
                          }}
                        >
                          Submit
                        </Button>
                      </DialogFooter>
                    </DialogContent>
                  </Dialog>
                </div>
                <div className={"flex flex-col gap-5 lg:flex-row"}>
                  <div>
                    <RouteCard
                      route={routeData.data!}
                    />
                  </div>
                  <div className={"flex w-2/3 flex-col gap-5"}>
                    <div className={"flex w-full flex-row gap-5"}>
                      <Card className={"w-1/4 text-center"}>
                        <CardHeader>
                          <CardDescription>Average RPS</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {mockRouteStats.summary.avg_rps}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/4 text-center"}>
                        <CardHeader>
                          <CardDescription>Average Latency ms</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {mockRouteStats.summary.avg_latency_ms}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/4 text-center"}>
                        <CardHeader>
                          <CardDescription>Dropped Events</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {mockRouteStats.summary.dropped_events}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/4 text-center"}>
                        <CardHeader>
                          <CardDescription>Total Requests</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {mockRouteStats.summary.total_requests}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                    </div>
                    <div className={"flex w-full flex-row gap-5"}>
                      <Card className={"w-full"}>
                        <CardHeader>
                          <CardTitle>Status Codes</CardTitle>
                          <CardDescription>{`Status codes percentages from ${format(from, "PPP-HH:mm")} to ${format(to, "PPP-HH:mm")}`}</CardDescription>
                        </CardHeader>
                        <CardContent>
                          <ChartContainer config={chartConfig}>
                            <BarChart
                              accessibilityLayer
                              data={chartData}
                              margin={{ top: 20 }}
                            >
                              <CartesianGrid vertical={false} />
                              <XAxis
                                dataKey={"code"}
                                tickLine={false}
                                tickMargin={10}
                                axisLine={false}
                                tickFormatter={(value) => value.slice(0, 3)}
                              />
                              <ChartTooltip
                                cursor={false}
                                content={<ChartTooltipContent hideLabel />}
                              />
                              <Bar
                                dataKey={"percentage"}
                                fill={"var(--color-percentage)"}
                                radius={8}
                              >
                                <LabelList
                                  position={"top"}
                                  offset={12}
                                  className={"fill-foreground"}
                                  fontSize={12}
                                />
                              </Bar>
                            </BarChart>
                          </ChartContainer>
                        </CardContent>
                      </Card>
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}