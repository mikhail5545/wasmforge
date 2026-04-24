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
  OverviewResponse,
  RouteSummary,
  RouteSummaryResponse,
  RoutesResponse,
  TimeseriesPoint,
  TimeseriesResponse,
} from "@/types/ProxyServerStatistics"

type StatusRatios = Array<[string, number]>

type RouteProfile = {
  route_path: string
  share: number
  base_latency_ms: number
  jitter_ms: number
  status_ratios: StatusRatios
}

const DEFAULT_BUCKET_SECONDS = 60
const DEFAULT_SERIES_POINTS = 60

const ROUTE_PROFILES: RouteProfile[] = [
  {
    route_path: "/api/products",
    share: 0.2,
    base_latency_ms: 92,
    jitter_ms: 22,
    status_ratios: [
      ["200", 0.94],
      ["201", 0.01],
      ["400", 0.02],
      ["404", 0.02],
      ["500", 0.01],
    ],
  },
  {
    route_path: "/api/orders",
    share: 0.18,
    base_latency_ms: 138,
    jitter_ms: 30,
    status_ratios: [
      ["200", 0.89],
      ["201", 0.03],
      ["400", 0.03],
      ["404", 0.01],
      ["500", 0.04],
    ],
  },
  {
    route_path: "/api/search",
    share: 0.16,
    base_latency_ms: 76,
    jitter_ms: 18,
    status_ratios: [
      ["200", 0.91],
      ["201", 0.01],
      ["400", 0.05],
      ["404", 0.02],
      ["500", 0.01],
    ],
  },
  {
    route_path: "/api/cart",
    share: 0.12,
    base_latency_ms: 110,
    jitter_ms: 25,
    status_ratios: [
      ["200", 0.88],
      ["201", 0.02],
      ["400", 0.04],
      ["404", 0.01],
      ["500", 0.05],
    ],
  },
  {
    route_path: "/api/login",
    share: 0.11,
    base_latency_ms: 158,
    jitter_ms: 35,
    status_ratios: [
      ["200", 0.84],
      ["201", 0.01],
      ["400", 0.04],
      ["401", 0.08],
      ["500", 0.03],
    ],
  },
  {
    route_path: "/api/checkout",
    share: 0.1,
    base_latency_ms: 182,
    jitter_ms: 40,
    status_ratios: [
      ["200", 0.82],
      ["201", 0.02],
      ["400", 0.03],
      ["409", 0.03],
      ["500", 0.1],
    ],
  },
  {
    route_path: "/api/users",
    share: 0.08,
    base_latency_ms: 124,
    jitter_ms: 24,
    status_ratios: [
      ["200", 0.9],
      ["201", 0.02],
      ["400", 0.03],
      ["404", 0.02],
      ["500", 0.03],
    ],
  },
  {
    route_path: "/healthz",
    share: 0.05,
    base_latency_ms: 8,
    jitter_ms: 4,
    status_ratios: [
      ["200", 0.995],
      ["500", 0.005],
    ],
  },
]

function round2(value: number): number {
  return Number(value.toFixed(2))
}

function clampMin(value: number, min: number): number {
  return value < min ? min : value
}

function pseudoNoise(seed: number): number {
  const x = Math.sin(seed * 12.9898) * 43758.5453
  return x - Math.floor(x)
}

function floorToBucket(date: Date, bucketSeconds: number): Date {
  const bucketMs = bucketSeconds * 1000
  return new Date(Math.floor(date.getTime() / bucketMs) * bucketMs)
}

function buildStatusCounts(
  totalRequests: number,
  ratios: StatusRatios
): Record<string, number> {
  const counts: Record<string, number> = {}
  let remaining = totalRequests

  ratios.forEach(([statusCode, ratio], index) => {
    if (index === ratios.length - 1) {
      counts[statusCode] = clampMin(remaining, 0)
      return
    }

    const count = Math.max(0, Math.floor(totalRequests * ratio))
    counts[statusCode] = count
    remaining -= count
  })

  return counts
}

function buildPercentages(
  counts: Record<string, number>,
  totalRequests: number
): Record<string, number> {
  if (totalRequests <= 0) {
    return {}
  }

  return Object.fromEntries(
    Object.entries(counts).map(([statusCode, count]) => [
      statusCode,
      round2((count / totalRequests) * 100),
    ])
  )
}

function sumCounts(counts: Record<string, number>): number {
  return Object.values(counts).reduce((acc, value) => acc + value, 0)
}

function averageLatency(totalLatencyMs: number, totalRequests: number): number {
  return totalRequests > 0 ? round2(totalLatencyMs / totalRequests) : 0
}

function hashRoutePath(routePath: string): number {
  let hash = 0

  for (let index = 0; index < routePath.length; index += 1) {
    hash = (hash * 31 + routePath.charCodeAt(index)) >>> 0
  }

  return hash
}

function normalizeRoutePath(routePath: string): string {
  const trimmed = routePath.trim()

  if (!trimmed) {
    return "/"
  }

  try {
    return decodeURIComponent(trimmed)
  } catch {
    return trimmed
  }
}

function buildRouteSummaryEntry(
  profile: RouteProfile,
  count: number,
  now: Date,
  noiseSeed = 0
): RouteSummary {
  const noise = pseudoNoise(now.getTime() / 1000 + noiseSeed * 37)
  const avgLatencyMs = round2(
    clampMin(profile.base_latency_ms + (noise - 0.5) * profile.jitter_ms, 1)
  )
  const counts = buildStatusCounts(count, profile.status_ratios)

  return {
    route_path: profile.route_path,
    total_requests: count,
    avg_rps: round2(count / 3600),
    avg_latency_ms: avgLatencyMs,
    status_code_counts: counts,
    status_code_percentages: buildPercentages(counts, count),
    dropped_events: 0,
  }
}

function buildFallbackRouteProfile(routePath: string): RouteProfile {
  const normalized = normalizeRoutePath(routePath)
  const lowerCasePath = normalized.toLowerCase()
  const seed = hashRoutePath(normalized)
  const baseLatencyMs = round2(50 + pseudoNoise(seed + 1) * 180)
  const jitterMs = round2(10 + pseudoNoise(seed + 2) * 40)

  if (/auth|login|signin|session/.test(lowerCasePath)) {
    return {
      route_path: normalized,
      share: 0.08 + pseudoNoise(seed) * 0.04,
      base_latency_ms: baseLatencyMs,
      jitter_ms: jitterMs,
      status_ratios: [
        ["200", 0.8],
        ["400", 0.05],
        ["401", 0.12],
        ["500", 0.03],
      ],
    }
  }

  if (/create|update|edit|delete|checkout|order|cart|payment|submit/.test(lowerCasePath)) {
    return {
      route_path: normalized,
      share: 0.1 + pseudoNoise(seed) * 0.05,
      base_latency_ms: baseLatencyMs,
      jitter_ms: jitterMs,
      status_ratios: [
        ["200", 0.84],
        ["201", 0.04],
        ["400", 0.04],
        ["409", 0.04],
        ["500", 0.04],
      ],
    }
  }

  if (/search|query|filter|lookup/.test(lowerCasePath)) {
    return {
      route_path: normalized,
      share: 0.07 + pseudoNoise(seed) * 0.05,
      base_latency_ms: baseLatencyMs,
      jitter_ms: jitterMs,
      status_ratios: [
        ["200", 0.9],
        ["400", 0.05],
        ["404", 0.03],
        ["500", 0.02],
      ],
    }
  }

  return {
    route_path: normalized,
    share: 0.05 + pseudoNoise(seed) * 0.08,
    base_latency_ms: baseLatencyMs,
    jitter_ms: jitterMs,
    status_ratios: [
      ["200", 0.92],
      ["201", 0.02],
      ["400", 0.03],
      ["404", 0.02],
      ["500", 0.01],
    ],
  }
}

function resolveRouteProfile(routePath: string): RouteProfile {
  const normalized = normalizeRoutePath(routePath)
  return ROUTE_PROFILES.find((profile) => profile.route_path === normalized) ??
    buildFallbackRouteProfile(normalized)
}

function buildTimeseries(
  now: Date,
  bucketSeconds = DEFAULT_BUCKET_SECONDS,
  points = DEFAULT_SERIES_POINTS
) {
  const bucketEnd = floorToBucket(now, bucketSeconds)
  const series: TimeseriesPoint[] = []
  let totalRequests = 0
  let totalLatencyMs = 0
  const statusTotals: Record<string, number> = {}

  for (let i = points - 1; i >= 0; i--) {
    const bucketStart = new Date(bucketEnd.getTime() - i * bucketSeconds * 1000)
    const noise = pseudoNoise(bucketStart.getTime() / 1000)
    const wave =
      Math.sin((points - i) / 7) * 0.25 + Math.cos((points - i) / 5) * 0.15
    const count = clampMin(Math.round(70 + wave * 35 + noise * 20), 12)
    const latencyMs = round2(
      clampMin(86 + Math.sin((points - i) / 9) * 18 + noise * 26, 8)
    )
    const counts = buildStatusCounts(count, [
      ["200", 0.95],
      ["404", 0.03],
      ["500", 0.02],
    ])

    series.push({
      bucket_start: bucketStart.toISOString(),
      total_requests: count,
      avg_latency_ms: latencyMs,
      status_code_counts: counts,
    })

    totalRequests += count
    totalLatencyMs += latencyMs * count

    for (const [statusCode, statusCount] of Object.entries(counts)) {
      statusTotals[statusCode] = (statusTotals[statusCode] ?? 0) + statusCount
    }
  }

  return { series, totalRequests, totalLatencyMs, statusTotals }
}

function buildRoutes(
  totalRequests: number,
  now: Date
): RouteSummary[] {
  let remaining = totalRequests

  return ROUTE_PROFILES.map((profile, index) => {
    const isLast = index === ROUTE_PROFILES.length - 1
    const idealCount = Math.round(totalRequests * profile.share)
    const count = isLast
      ? remaining
      : Math.max(
          1,
          Math.min(idealCount, remaining - (ROUTE_PROFILES.length - index - 1))
        )
    remaining -= count

    return buildRouteSummaryEntry(profile, count, now, index)
  }).sort((a, b) => b.total_requests - a.total_requests)
}

export type MockProxyStatsBundle = {
  overview: OverviewResponse
  routes: RoutesResponse
  timeseries: TimeseriesResponse
}

export function isProxyStatsMockEnabled(): boolean {
  const flag = process.env.NEXT_PUBLIC_PROXY_STATS_MOCK
  return flag === "1" || flag === "true"
}

export function buildMockProxyStats(now = new Date()): MockProxyStatsBundle {
  const timeseries = buildTimeseries(now)
  const routes = buildRoutes(timeseries.totalRequests, now)

  const overviewCounts = timeseries.statusTotals
  const overviewTotal = sumCounts(overviewCounts)
  const overviewLatency = averageLatency(
    timeseries.totalLatencyMs,
    timeseries.totalRequests
  )
  const from = timeseries.series[0]?.bucket_start ?? now.toISOString()
  const to =
    timeseries.series.length > 0
      ? timeseries.series[timeseries.series.length - 1]?.bucket_start
      : now.toISOString()

  return {
    overview: {
      from,
      to: to?? now.toISOString(),
      scope: "overall",
      total_requests: overviewTotal,
      avg_rps: round2(
        overviewTotal / (DEFAULT_BUCKET_SECONDS * DEFAULT_SERIES_POINTS)
      ),
      avg_latency_ms: overviewLatency,
      status_code_counts: overviewCounts,
      status_code_percentages: buildPercentages(overviewCounts, overviewTotal),
      dropped_events: Math.round(overviewTotal * 0.002),
    },
    routes: {
      from: from,
      to: to ?? now.toISOString(),
      routes: routes,
    },
    timeseries: {
      from: from,
      scope: 'overall',
      bucket_seconds: DEFAULT_BUCKET_SECONDS,
      to: to ?? now.toISOString(),
      points: timeseries.series,
    }
  }
}

export function buildMockRouteSummaryResponse(
  routePath: string,
  now = new Date()
): RouteSummaryResponse {
  const timeseries = buildTimeseries(now)
  const normalizedRoutePath = normalizeRoutePath(routePath)
  const from = timeseries.series[0]?.bucket_start ?? now.toISOString()
  const to =
    timeseries.series.length > 0
      ? timeseries.series[timeseries.series.length - 1]?.bucket_start ?? now.toISOString()
      : now.toISOString()
  const routes = buildRoutes(timeseries.totalRequests, now)
  const matchedRoute = routes.find(
    (route) => route.route_path === normalizedRoutePath
  )
  const fallbackProfile = resolveRouteProfile(normalizedRoutePath)

  return {
    from,
    to,
    summary:
      matchedRoute ??
      buildRouteSummaryEntry(
        fallbackProfile,
        Math.max(1, Math.round(timeseries.totalRequests * fallbackProfile.share)),
        now,
        hashRoutePath(normalizedRoutePath)
      ),
  }
}

