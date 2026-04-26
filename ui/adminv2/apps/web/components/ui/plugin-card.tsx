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
import { Plugin } from "types/Plugin"
import { Button } from "@workspace/ui/components/button"
import React from "react"
import { Copy, Ellipsis, Paperclip, Pencil, Trash } from "lucide-react"

interface PluginCardProps {
  plugin: Plugin
  className?: string
  onDelete?: () => void
}

const PluginCard: React.FC<PluginCardProps> = ({
  plugin,
  className,
  onDelete,
}) => {
  return (
    <Card className={className}>
      <CardHeader className={"flex flex-row items-center justify-between"}>
        <CardTitle className={"text-2xl"}>{plugin.name}</CardTitle>
        <div className={"flex flex-row items-center justify-center gap-2"}>
          <Badge>{`v${plugin.version}`}</Badge>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant={"ghost"} size={"icon"}>
                <Ellipsis />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuGroup>
                <DropdownMenuItem asChild>
                  <a
                    href={`/plugins/plugin/edit?name=${encodeURIComponent(plugin.name)}&version=${encodeURIComponent(plugin.version)}`}
                  >
                    <Pencil />
                    <span>Edit</span>
                  </a>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <a href={`/routes/plugins/new?pluginId=${plugin.id}`}>
                    <Paperclip/>
                    Attach
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
        <div className={"flex w-full flex-row items-center"}>
          <div className={"flex w-1/3 flex-col gap-4 text-muted-foreground"}>
            <span>Filename</span>
            <span>Version</span>
            <span>File Checksum</span>
            <span>Created at</span>
          </div>
          <div className={"flex w-2/3 flex-col gap-4 truncate"}>
            <span>{plugin.filename}</span>
            <span>{`v${plugin.version}`}</span>
            <div className={"flex flex-row items-center justify-center"}>
              <p className={"truncate"}>{plugin.checksum}</p>
              <Button
                variant={"ghost"}
                size={"icon"}
                onClick={() => {
                  void navigator.clipboard.writeText(plugin.checksum)
                }}
              >
                <Copy />
              </Button>
            </div>
            <span>{new Date(plugin.created_at).toLocaleString()}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export { PluginCard }