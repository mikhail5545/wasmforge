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
  DialogTrigger,
  DialogClose,
} from "@workspace/ui/components/dialog"
import { Button } from "@workspace/ui/components/button"
import { RouteCard } from "@/components/ui/route-card"
import React from "react"
import { format } from "date-fns"
import { DateTimeInput } from "@/components/date-time-input"
import { RouteSummaryResponse } from "@/types/ProxyServerStatistics"
import {
  StatusCountsRadarChart,
  StatusPercentagesBarChart,
} from "@/components/status-codes-charts"
import { useMutation } from "@/hooks/use-mutation"

type DateRangeState = {
  from: {
    date: Date
    time: string
  }
  to: {
    date: Date
    time: string
  }
}

export default function RoutePage() {
  const params = useSearchParams()
  const path = params.get("path") ?? ""
  const routeData = useData<Route>(
    `http://localhost:8080/api/routes/${encodeURIComponent(path)}`,
    "route"
  )
  const router = useRouter()

  const [statsConfigDialogOpen, setStatsConfigDialogOpen] =
    React.useState<boolean>(false)
  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")
  const [successRedirect, setSuccessRedirect] = React.useState<string | null>(
    null
  )
  const [range, setRange] = React.useState<DateRangeState>(() => {
    const now = new Date()
    const hourAgo = new Date(now.getTime() - 60 * 60 * 1000)
    const fromTime = format(hourAgo, "HH:mm")
    const toTime = format(now, "HH:mm")

    return {
      from: {
        date: hourAgo,
        time: fromTime,
      },
      to: {
        date: now,
        time: toTime,
      },
    }
  })

  const updateFrom = (date: Date) => {
    setRange((prev) => ({ ...prev, from: { ...prev.from, date: date } }))
  }
  const updateTo = (date: Date) => {
    setRange((prev) => ({ ...prev, to: { ...prev.to, date: date } }))
  }
  const updateFromTime = (time: string) => {
    setRange((prev) => {
      const next = new Date(prev.from.date)
      const [hours, minutes] = time.split(":").map(Number)

      if (hours != undefined && !Number.isNaN(hours)) next.setHours(hours)
      if (minutes != undefined && !Number.isNaN(minutes))
        next.setMinutes(minutes)

      return { ...prev, from: { date: next, time: time } }
    })
  }
  const updateToTime = (time: string) => {
    setRange((prev) => {
      const next = new Date(prev.to.date)
      const [hours, minutes] = time.split(":").map(Number)

      if (hours != undefined && !Number.isNaN(hours)) next.setHours(hours)
      if (minutes != undefined && !Number.isNaN(minutes))
        next.setMinutes(minutes)

      return { ...prev, to: { date: next, time: time } }
    })
  }

  const routeSummaryData = useData<RouteSummaryResponse>(
    `http://localhost:8080/api/proxy/stats/route?path=${encodeURIComponent(path)}&from=${range.from.date.toISOString()}&to=${range.to.date.toISOString()}`,
    "route"
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

  return (
    <SidebarLayout page_title={"Route Details"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!routeData.error || !!error || !!routeSummaryData.error}
        description={
          routeData?.error?.message ||
          error?.message ||
          routeSummaryData?.error?.message ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={async () => {
          if (routeData.error) {
            await routeData.refetch()
          }
          if (routeSummaryData.error) {
            await routeSummaryData.refetch()
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
      <div className={"flex flex-col p-6"}>
        {(routeData.loading && routeData.data === null) ||
        routeSummaryData.loading ? (
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
                        date={range.from.date}
                        setDate={updateFrom}
                        time={range.from.time}
                        setTime={updateFromTime}
                        layout={"row"}
                      />
                      <p className={"mt-4"}>Show stats to</p>
                      <DateTimeInput
                        date={range.to.date}
                        setDate={updateTo}
                        time={range.to.time}
                        setTime={updateToTime}
                        layout={"row"}
                      />
                      <DialogFooter>
                        <DialogClose asChild>
                          <Button variant={"secondary"}>Close</Button>
                        </DialogClose>
                        <Button
                          onClick={async () => {
                            await routeSummaryData.refetch()
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
                      onEnableToggle={toggleEnable}
                      enabling={loading}
                      onDelete={deleteRoute}
                    />
                  </div>
                  <div className={"flex w-2/3 flex-col gap-5"}>
                    <div className={"flex w-full flex-row gap-5"}>
                      <Card className={"w-1/3 text-center"}>
                        <CardHeader>
                          <CardDescription>Average RPS</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {routeSummaryData.data?.summary.avg_rps ===
                            undefined
                              ? "N/A"
                              : routeSummaryData.data?.summary.avg_rps}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/3 text-center"}>
                        <CardHeader>
                          <CardDescription>Average Latency ms</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {routeSummaryData.data?.summary.avg_latency_ms ===
                            undefined
                              ? "N/A"
                              : routeSummaryData.data?.summary.avg_latency_ms}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                      <Card className={"w-1/3 text-center"}>
                        <CardHeader>
                          <CardDescription>Total Requests</CardDescription>
                          <CardTitle className={"text-4xl"}>
                            {routeSummaryData.data?.summary.total_requests ===
                            undefined
                              ? "N/A"
                              : routeSummaryData.data?.summary.total_requests}
                          </CardTitle>
                        </CardHeader>
                      </Card>
                    </div>
                    <div className={"flex w-full flex-row gap-5"}>
                      <div className={"w-full"}>
                        <StatusPercentagesBarChart
                          percentages={
                            routeSummaryData.data?.summary
                              .status_code_percentages || {}
                          }
                          title={"Status Code Percentages"}
                          description={`Showing status code percentages in a period from ${format(range.from.date, "PPP-HH:mm")} to ${format(range.to.date, "PPP-HH:mm")}`}
                        />
                      </div>

                      <StatusCountsRadarChart
                        counts={
                          routeSummaryData.data?.summary.status_code_counts ||
                          {}
                        }
                        title={"Status Code Counts"}
                        description={`Showing status code counts in a period from ${format(range.from.date, "PPP-HH:mm")} to ${format(range.to.date, "PPP-HH:mm")}`}
                      />
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
