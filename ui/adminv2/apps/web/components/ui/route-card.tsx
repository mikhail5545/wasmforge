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

import { Route } from "types/route"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Badge } from "@workspace/ui/components/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger } from "@workspace/ui/components/dropdown-menu"
import { Button } from "@workspace/ui/components/button"
import { ChevronsUpDown, Ellipsis, Pencil, Power, PowerOff, Trash } from "lucide-react"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@workspace/ui/components/collapsible"

interface RouteCardProps {
  route: Route
  className?: string
}

const RouteCard = ({ route, className }: RouteCardProps) => {
  return (
    <Card className={className}>
      <CardHeader className={"flex flex-row items-center justify-between"}>
        <CardTitle className={"text-2xl"}>Route Details</CardTitle>
        <div className={"flex flex-row items-center justify-center gap-2"}>
          <Badge
            className={
              route.enabled
                ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
                : "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
            }
          >
            {route.enabled ? "Running" : "Stopped"}
          </Badge>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant={"ghost"} size={"icon"}>
                <Ellipsis />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuGroup>
                <DropdownMenuItem asChild>
                  <a href={`/routes/edit?path=${route.path}`}>
                    <Pencil />
                    <span>Edit</span>
                  </a>
                </DropdownMenuItem>
                <DropdownMenuItem>
                  {route.enabled ? (
                    <>
                      <PowerOff />
                      <span>Stop</span>
                    </>
                  ) : (
                    <>
                      <Power />
                      <span>Start</span>
                    </>
                  )}
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem variant={"destructive"}>
                <Trash />
                <span>Delete</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent className={"mt-5"}>
        <div className={"flex w-full flex-row items-center"}>
          <div className={"flex w-1/3 flex-col gap-4 text-muted-foreground"}>
            <span>Path</span>
            <span>Target URL</span>
          </div>
          <div className={"flex w-2/3 flex-col gap-4 truncate"}>
            <span>{route?.path || "N/A"}</span>
            <span>{route?.target_url || "N/A"}</span>
          </div>
        </div>
        <div className={"mt-10 grid grid-cols-2 gap-2"}>
          <div className={"flex flex-col gap-4 rounded-xl bg-background p-5"}>
            <p className={"text-center text-muted-foreground"}>
              Idle Connection Timeout
            </p>
            <p
              className={"text-center text-xl"}
            >{`${route?.idle_conn_timeout} sec`}</p>
          </div>
          <div className={"flex flex-col gap-4 rounded-xl bg-background p-5"}>
            <p className={"text-center text-muted-foreground"}>
              TLS handshake timeout
            </p>
            <p
              className={"text-center text-xl"}
            >{`${route?.idle_conn_timeout} sec`}</p>
          </div>
          <div className={"flex flex-col gap-4 rounded-xl bg-background p-5"}>
            <p className={"text-center text-muted-foreground"}>
              Expect continue timeout
            </p>
            <p
              className={"text-center text-xl"}
            >{`${route.expect_continue_timeout} sec`}</p>
          </div>
        </div>
        <Collapsible className={"mt-10 flex w-full flex-col gap-2"}>
          <div className={"flex items-center justify-between gap-4"}>
            <p className={"text-sm font-semibold"}>
              Optional Timeouts and Limits
            </p>
            <CollapsibleTrigger asChild>
              <Button variant={"ghost"} size={"icon"}>
                <ChevronsUpDown />
                <span className={"sr-only"}>
                  Toggle optional timeouts and limits
                </span>
              </Button>
            </CollapsibleTrigger>
          </div>
          <CollapsibleContent className={"flex flex-row"}>
            <div className={"flex w-2/3 flex-col gap-4 text-muted-foreground"}>
              <span>Max idle connections</span>
              <span>Max idle connections per host</span>
              <span>Response header timeout</span>
              <span>Max connections per host</span>
            </div>
            <div className={"flex w-1/3 flex-col gap-4"}>
              <span>
                {route?.max_idle_cons ? route.max_idle_cons : "default"}
              </span>
              <span>
                {route?.max_idle_cons_per_host
                  ? route.max_idle_cons_per_host
                  : "default"}
              </span>
              <span>
                {route?.response_header_timeout
                  ? `${route.response_header_timeout} sec`
                  : "default"}
              </span>
              <span>
                {route?.max_cons_per_host ? route.max_cons_per_host : "default"}
              </span>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </CardContent>
    </Card>
  )
}

export { RouteCard }