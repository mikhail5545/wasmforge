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
import { Field, FieldLabel } from "@workspace/ui/components/field"
import { Calendar } from "@workspace/ui/components/calendar"
import { Input } from "@workspace/ui/components/input"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@workspace/ui/components/popover"
import { Button } from "@workspace/ui/components/button"
import { ChevronDown, Clock } from "lucide-react"

interface DateTimeInputProps {
  date: Date
  setDate: (date: Date) => void
  time: string
  setTime: (time: string) => void
  layout: "column" | "row"
}

export const DateTimeInput = ({ date, setDate, time, setTime, layout }: DateTimeInputProps) => {
  const [open, setOpen] = React.useState(false)

  return (
    <div className={layout === 'column' ? 'flex flex-col gap-2' : 'flex flex-row gap-2'}>
      {/* Date Picker */}
      <Field>
        <FieldLabel>Date</FieldLabel>
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button
              variant={"outline"}
              className={"w-32 justify-between font-normal"}
            >
              {format(date, "PPP")}
              <ChevronDown />
            </Button>
          </PopoverTrigger>
          <PopoverContent
            className={"w-auto overflow-hidden p-0"}
            align={"start"}
          >
            <Calendar
              mode={"single"}
              selected={date}
              captionLayout={"dropdown"}
              defaultMonth={date}
              onSelect={(date) => {
                if (date) {
                  setDate(date)
                }
                setOpen(false)
              }}
            />
          </PopoverContent>
        </Popover>
      </Field>

      {/* Time Input */}
      <Field>
        <FieldLabel>Time</FieldLabel>
        <div className={"relative w-[250px]"}>
          <Input
            id={"time"}
            type={"time"}
            aria-labelledby={"time input"}
            value={time}
            step={"60"} // one minute
            onChange={(e) => setTime(e.target.value)}
          />
        </div>
      </Field>
    </div>
  )
}