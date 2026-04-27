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
import { format } from "date-fns"

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

export interface Range {
  range: DateRangeState
  setRange: React.Dispatch<React.SetStateAction<DateRangeState>>
  updateTo: (date: Date) => void
  updateFrom: (date: Date) => void
  updateToTime: (time: string) => void
  updateFromTime: (time: string) => void
}

export function useRange(): Range {
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

  return {
    range,
    setRange,
    updateFrom,
    updateTo,
    updateFromTime,
    updateToTime,
  }
}