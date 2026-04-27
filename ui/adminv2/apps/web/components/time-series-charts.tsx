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

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { TimeseriesPoint } from "@/types/ProxyServerStatistics"
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@workspace/ui/components/chart"
import { Area, AreaChart, CartesianGrid, Line, LineChart, XAxis } from "recharts"
import { NormalizeTimeSeriesPointsLatencyAndRequests } from "@/lib/normalize-stats"
import {format} from "date-fns"
import React from "react"

interface TimeSeriesAreaChartProps {
  title?: string
  description?: string
  timeSeries: TimeseriesPoint[]
  className?: string
}

const TimeSeriesAreaChart: React.FC<TimeSeriesAreaChartProps> = ({
  title,
  description,
  timeSeries,
  className,
}) => {
  const config = {
    amount: {
      label: "Amount",
    },
    total_requests: {
      label: "Total requests",
      color: "var(--chart-1)"
    },
    avg_latency_ms: {
      label: "Average latency ms",
      color: "var(--chart-2)",
    }
  } satisfies ChartConfig

  const normalized = NormalizeTimeSeriesPointsLatencyAndRequests(timeSeries)

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className={"px-2 pt-4 sm:px-6 sm:pt-6"}>
        <ChartContainer
          config={config}
          className={"aspect-auto h-[250px] w-full"}
        >
          <AreaChart data={normalized}>
            <defs>
              <linearGradient
                id={"color-requests"}
                x1={"0"}
                y1={"0"}
                x2={"0"}
                y2={"1"}
              >
                <stop
                  offset={"5%"}
                  stopColor={"var(--color-total_requests)"}
                  stopOpacity={0.8}
                />
                <stop
                  offset={"95%"}
                  stopColor={"var(--color-total_requests)"}
                  stopOpacity={0.1}
                />
              </linearGradient>
              <linearGradient
                id={"color-latency"}
                x1={"0"}
                y1={"0"}
                x2={"0"}
                y2={"1"}
              >
                <stop
                  offset={"5%"}
                  stopColor={"var(--color-avg_latency_ms)"}
                  stopOpacity={0.8}
                />
                <stop
                  offset={"95%"}
                  stopColor={"var(--color-avg_latency_ms)"}
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey={"bucket_start"}
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
              tickFormatter={(value) => {
                const date = new Date(value)
                return format(date, "PPP-HH:mm")
              }}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  labelFormatter={(value) => {
                    const date = new Date(value)
                    return format(date, "PPP-HH:mm")
                  }}
                  indicator={"dot"}
                  className={"w-[210px]"}
                />
              }
            />
            <Area
              dataKey={"total_requests"}
              type={"natural"}
              fill={"url(#color-requests)"}
              stroke={"var(--color-total_requests)"}
              strokeWidth={2}
              stackId={"a"}
            />
            <Area
              dataKey={"avg_latency_ms"}
              type={"natural"}
              fill={"url(#color-latency)"}
              stroke={"var(--color-avg_latency_ms)"}
              strokeWidth={2}
              stackId={"a"}
            />
            <ChartLegend content={<ChartLegendContent />} />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}

interface TimeSeriesStatusCodeLineChartProps {
  title?: string
  description?: string
  className?: string
  buckets: { code: string; buckets: {bucket_start: string; count: number}[] }
  color?: string
}

const TimeSeriesStatusCodeLineChart: React.FC<
  TimeSeriesStatusCodeLineChartProps
> = ({ title, description, className, buckets, color }) => {
  const config = {
    code: {
      label: "Count",
      color: color ?? 'var(--chart-1)',
    }
  } satisfies ChartConfig

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className={"px-2 pt-4 sm:px-6 sm:pt-6"}>
        <ChartContainer
          config={config}
          className={"aspect-auto h-[300px] w-full"}
        >
          <LineChart data={buckets.buckets}>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey={"bucket_start"}
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => {
                const date = new Date(value)
                return format(date, "PPP-HH:mm")
              }}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  labelFormatter={(value) => {
                    const date = new Date(value)
                    return format(date, "PPP-HH:mm")
                  }}
                  indicator={"dot"}
                  className={"w-[210px]"}
                />
              }
            />
            <Line
              dataKey={"count"}
              type={"step"}
              stroke={"var(--color-code)"}
              strokeWidth={2}
              dot={false}
            />
          </LineChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}

export { TimeSeriesAreaChart, TimeSeriesStatusCodeLineChart }