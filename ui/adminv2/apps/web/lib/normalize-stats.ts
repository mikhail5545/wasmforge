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


import { TimeseriesPoint } from "@/types/ProxyServerStatistics"

export function NormalizeStatusPercentages(percentages: Record<string, number>) {
  const normalized: {code: string, percentage: number}[] = []
  Object.entries(percentages).forEach(([name, percentage]) => {
    normalized.push({ code: name, percentage: percentage })
  })
  return normalized
}

export function NormalizeStatusCounts(counts: Record<string, number>) {
  const normalized: {code: string, count: number}[] = []
  Object.entries(counts).forEach(([name, count]) => {
    normalized.push({ code: name, count: count })
  })
  return normalized
}

export function NormalizeTimeSeriesPointsLatencyAndRequests(points: TimeseriesPoint[]) {
  const normalized: { bucket_start: string; total_requests: number; avg_latency_ms: number }[] = []
  points.map((point: TimeseriesPoint) => {
    normalized.push({
      bucket_start: point.bucket_start,
      total_requests: point.total_requests,
      avg_latency_ms: point.avg_latency_ms
    })
  })
  return normalized
}

export function NormalizeTimeSeriesPointsStatusCodeCounts(points: TimeseriesPoint[]) {
  const normalized: { bucket_start: string, codes: {code: string, count: number}[] }[] = []
  points.map((point: TimeseriesPoint) => {
    const codes: {code: string, count: number}[] = []
    Object.entries(point.status_code_counts).forEach(([name, count]) => {
      codes.push({ code: name, count: count })
    })
    normalized.push({ bucket_start: point.bucket_start, codes: codes })
  })
  return normalized
}

export function NormalizeTimeSeriesPointsStatusCodes(points: TimeseriesPoint[]) {
  const byCode = new Map<string, { bucket_start: string, count: number }[]>()

  points.forEach((point) => {
    Object.entries(point.status_code_counts).forEach(([code, count]) => {
      if (!byCode.has(code)) {
        byCode.set(code, [])
      }

      byCode.get(code)!.push({
        bucket_start: point.bucket_start,
        count,
      })
    })
  })
  const normalized: {
    code: string
    buckets: { bucket_start: string; count: number }[]
  }[] = []
  byCode.forEach((buckets, code) => {
    normalized.push({ code, buckets })
  })

  return normalized
}