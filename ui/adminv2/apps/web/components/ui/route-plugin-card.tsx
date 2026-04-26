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
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Badge } from "@workspace/ui/components/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import { RoutePlugin } from "@/types/RoutePlugin"
import React from "react"
import { Button } from "@workspace/ui/components/button"
import { Ellipsis, Paperclip, Pencil, Trash } from "lucide-react"

interface RoutePluginCardProps {
  plugin: RoutePlugin
  className?: string
  onDelete?: () => void
}

const RoutePluginCard: React.FC<RoutePluginCardProps> = ({
  plugin,
  className,
  onDelete,
}) => {
  return (
    <Card className={className}>
      <CardHeader className={"flex flex-row items-center justify-between"}>
        <CardTitle className={"text-2xl"}>{plugin.plugin?.name}</CardTitle>
        <div className={"flex flex-row items-center justify-center gap-2"}>
          <Badge>{`v${plugin.plugin?.version}`}</Badge>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant={"ghost"} size={"icon"}>
                <Ellipsis />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuGroup>
                <DropdownMenuItem asChild>
                  <a href={`/routes/plugins/edit?pluginId=${plugin.id}`}>
                    <Pencil />
                    Edit
                  </a>
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={onDelete} variant={"destructive"}>
                <Trash />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      <CardContent className={"mt-5"}>
        <div className={"flex w-full flex-col gap-4"}>
          <div className={"flex flex-row items-center"}>
            <p className={"w-1/3 text-muted-foreground"}>Filename</p>
            <p className={"w-2/3 truncate"}>{plugin.plugin?.filename}</p>
          </div>
          <div className={"flex flex-row items-center"}>
            <p className={"w-1/3 text-muted-foreground"}>Version</p>
            <p className={"w-2/3 truncate"}>{plugin.plugin?.version}</p>
          </div>
          <div className={"flex flex-row items-center"}>
            <p className={"w-1/3 text-muted-foreground"}>Resolved Version</p>
            <p className={"w-2/3 truncate"}>{plugin.resolved_plugin_version}</p>
          </div>
          <div className={"flex flex-row items-center"}>
            <p className={"w-1/3 text-muted-foreground"}>Version Constraint</p>
            <p className={"w-2/3 truncate"}>{plugin.version_constraint}</p>
          </div>
          <div className={"flex flex-row items-center"}>
            <p className={"w-1/3 text-muted-foreground"}>Execution Order</p>
            <p className={"w-2/3 truncate"}>{plugin.execution_order}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export { RoutePluginCard }