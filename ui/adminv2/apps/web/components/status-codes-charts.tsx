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
  NormalizeStatusCounts,
  NormalizeStatusPercentages,
} from "@/lib/normalize-stats"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@workspace/ui/components/chart"
import { Bar, BarChart, CartesianGrid, LabelList, PolarAngleAxis, PolarGrid, Radar, RadarChart, XAxis } from "recharts"
import React from "react"
import { cn } from "@workspace/ui/lib/utils"

interface StatusPercentagesBarChartProps {
  percentages: Record<string, number>
  title?: string
  description?: string
  className?: string
}

interface StatusCountsRadarChartProps {
  counts: Record<string, number>
  title?: string,
  description?: string,
  className?: string
}

const StatusPercentagesBarChart = ({
  percentages,
  title,
  description,
  className
}: StatusPercentagesBarChartProps) => {
  const normalizedData = NormalizeStatusPercentages(percentages)
  const config = {
    percentage: {
      label: "Percentage",
      color: "var(--chart-4)",
    },
    label: {
      color: "var(--background)",
    },
  } satisfies ChartConfig

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        {normalizedData.length === 0 ? (
          <div className={"py-10 text-center text-xl text-muted-foreground"}>
            <p>Not enough data</p>
          </div>
        ) : (
        <ChartContainer config={config}>
          <BarChart
            accessibilityLayer
            data={normalizedData}
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
          )}
      </CardContent>
    </Card>
  )
}

const StatusCountsRadarChart = ({
  counts,
  title,
  description,
  className,
}: StatusCountsRadarChartProps) => {
  const normalizedData = NormalizeStatusCounts(counts)
  const config = {
    count: {
      label: "Count",
      color: "var(--chart-3)",
    },
    label: {
      color: "var(--background)",
    },
  } satisfies ChartConfig

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        {normalizedData.length === 0 ? (
          <div className={"py-10 text-center text-xl text-muted-foreground"}>
            <p>Not enough data</p>
          </div>
        ) : (
          <ChartContainer config={config} className={"ms-auto aspect-square"}>
            <RadarChart data={normalizedData}>
              <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
              <PolarAngleAxis dataKey={"code"} />
              <PolarGrid />
              <Radar
                dataKey={"count"}
                fill={"var(--color-count)"}
                fillOpacity={0.6}
              />
            </RadarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}

export {
  StatusPercentagesBarChart,
  StatusCountsRadarChart,
}