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
import { usePaginatedData } from "@/hooks/use-paginated-data"
import { Plugin } from "types/Plugin"
import React from "react"
import { Spinner } from "@workspace/ui/components/spinner"
import { AnimatePresence, motion } from "motion/react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem, DropdownMenuRadioGroup, DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import { ChevronDownIcon, ChevronLeft, ChevronRight, MoreHorizontal, MoreVertical, ToyBrick, Wrench } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { Badge } from "@workspace/ui/components/badge"
import { AlertModal } from "@/components/dialog/alert-modal"
import { PluginsListControls } from "@/components/plugins-list-controls"


export default function PluginsPage() {
  const [orderField, setOrderField] = React.useState("created_at")
  const [orderDirection, setOrderDirection] = React.useState("asc")
  const [viewMode, setViewMode] = React.useState("table")
  const [perPage, setPerPage] = React.useState('10')

  const pluginsData = usePaginatedData<Plugin>(
    "/api/plugins",
    "plugins",
    Number(perPage),
    orderField,
    orderDirection as "asc" | "desc",
    { preload: true }
  )

  return (
    <SidebarLayout page_title={"Plugins"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!pluginsData.error}
        description={
          pluginsData.error?.message ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={() => {
          void pluginsData.refetch()
        }}
      />
      <div className={"flex flex-col p-6 gap-5"}>
        <PluginsListControls
          orderField={orderField}
          setOrderField={setOrderField}
          orderDirection={orderDirection}
          setOrderDirection={setOrderDirection}
          showCreateButton={true}
          pluginsData={pluginsData}
          className={"justify-between"}
          viewMode={viewMode}
          setViewMode={setViewMode}
        />
        {pluginsData.loading ? (
          <div className={"flex items-center justify-center py-20"}>
            <Spinner className={"h-10 w-10"} />
          </div>
        ) : (
          <div>
            {pluginsData.data.length === 0 ? (
              <Empty>
                <EmptyHeader>
                  <EmptyMedia variant={"icon"}>
                    <ToyBrick />
                  </EmptyMedia>
                  <EmptyTitle>No Plugins</EmptyTitle>
                  <EmptyDescription>
                    You haven&#39;t created any plugins yet. Please create a
                    plugin to get started.
                  </EmptyDescription>
                </EmptyHeader>
                <EmptyContent className={"flex-row justify-center gap-4"}>
                  <Button size={"sm"} asChild>
                    <a href={"/plugins/new"}>Create Plugin</a>
                  </Button>
                </EmptyContent>
              </Empty>
            ) : (
              <div
                className={
                  "flex w-full flex-col"
                }
              >
                <AnimatePresence mode={"wait"}>
                  {viewMode === "table" && (
                    <motion.div
                      key={"table"}
                      initial={{ opacity: 0, y: 100 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -100 }}
                      transition={{ duration: 0.5, ease: "easeIn" }}
                      className={"flex w-full flex-col gap-2"}
                    >
                      <div className={"overflow-hidden rounded-lg border"}>
                        <Table>
                          <TableHeader className={"sticky top-0 z-10 bg-muted"}>
                            <TableRow>
                              <TableHead className={"pr-4"}>Filename</TableHead>
                              <TableHead>Name</TableHead>
                              <TableHead>Version</TableHead>
                              <TableHead>Created at</TableHead>
                              <TableHead>Actions</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {pluginsData.data.map((plugin) => (
                              <TableRow key={plugin.id}>
                                <TableCell className={"pr-4"}>
                                  {plugin.filename}
                                </TableCell>
                                <TableCell>{plugin.name}</TableCell>
                                <TableCell>
                                  <Badge>{`v${plugin.version}`}</Badge>
                                </TableCell>
                                <TableCell>
                                  {new Date(plugin.created_at).toLocaleString()}
                                </TableCell>
                                <TableCell>
                                  <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                      <Button variant={"ghost"} size={"icon"}>
                                        <MoreVertical />
                                      </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent>
                                      <DropdownMenuItem asChild>
                                        <a
                                          href={`/plugins/plugin?name=${encodeURIComponent(plugin.name)}&version=${encodeURIComponent(plugin.version)}`}
                                        >
                                          <Wrench size={10} />
                                          <span>Details</span>
                                        </a>
                                      </DropdownMenuItem>
                                    </DropdownMenuContent>
                                  </DropdownMenu>
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </div>
                    </motion.div>
                  )}
                  {viewMode === "grid" && (
                    <motion.div
                      key={"grid"}
                      className={
                        "grid w-full grid-cols-1 gap-2 md:grid-cols-3 lg:grid-cols-4"
                      }
                    >
                      {pluginsData.data.map((plugin, idx) => (
                        <motion.div
                          key={plugin.id}
                          initial={{ opacity: 0, y: 50 }}
                          animate={{ opacity: 1, y: 0 }}
                          exit={{ opacity: 0, y: -50 }}
                          transition={{ duration: 0.3, delay: idx * 0.1 }}
                        >
                          <Card>
                            <CardHeader
                              className={
                                "flex flex-row items-center justify-between"
                              }
                            >
                              <CardTitle>{plugin.name}</CardTitle>
                              <div
                                className={
                                  "flex flex-row items-center justify-center gap-2"
                                }
                              >
                                <Badge>{plugin.version}</Badge>
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button variant={"ghost"} size={"icon"}>
                                      <MoreHorizontal />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent>
                                    <DropdownMenuItem asChild>
                                      <a
                                        href={`/plugins/plugin?name=${encodeURIComponent(plugin.name)}&version=${encodeURIComponent(plugin.version)}`}
                                      >
                                        <Wrench size={10} />
                                        <span>Details</span>
                                      </a>
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </div>
                            </CardHeader>
                            <CardContent>
                              <div className={"flex flex-col gap-4"}>
                                <div className={"flex flex-col gap-1"}>
                                  <p className={"text-muted-foreground"}>
                                    Filename
                                  </p>
                                  <p className={"truncate"}>
                                    {plugin.filename}
                                  </p>
                                </div>
                                <div className={"flex flex-col gap-1"}>
                                  <p className={"text-muted-foreground"}>
                                    Created At
                                  </p>
                                  <p className={"truncate"}>
                                    {new Date(
                                      plugin.created_at
                                    ).toLocaleString()}
                                  </p>
                                </div>
                              </div>
                            </CardContent>
                            <CardFooter>
                              <Button asChild>
                                <a
                                  href={`/plugins/plugin?name=${encodeURIComponent(plugin.name)}&version=${encodeURIComponent(plugin.version)}`}
                                >
                                  Vew Details
                                </a>
                              </Button>
                            </CardFooter>
                          </Card>
                        </motion.div>
                      ))}
                    </motion.div>
                  )}
                </AnimatePresence>
                <div className={"mt-5 flex flex-row justify-end gap-5"}>
                  <div
                    className={
                      "flex flex-row items-center justify-center gap-2"
                    }
                  >
                    <p className={"text-sm font-semibold"}>Entries per page</p>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant={"outline"}>
                          {perPage}
                          <ChevronDownIcon />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <DropdownMenuRadioGroup
                          value={perPage}
                          onValueChange={setPerPage}
                        >
                          <DropdownMenuRadioItem value={"5"}>
                            5
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value={"10"}>
                            10
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value={"20"}>
                            20
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value={"30"}>
                            30
                          </DropdownMenuRadioItem>
                          <DropdownMenuRadioItem value={"40"}>
                            40
                          </DropdownMenuRadioItem>
                        </DropdownMenuRadioGroup>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                  <div
                    className={
                      "flex flex-row items-center justify-center gap-2"
                    }
                  >
                    <Button
                      variant={"outline"}
                      size={"icon"}
                      disabled={
                        pluginsData.loading ||
                        pluginsData.previousPageToken === ""
                      }
                      onClick={() => pluginsData.previousPage()}
                    >
                      <ChevronLeft />
                    </Button>
                    <Button
                      variant={"outline"}
                      size={"icon"}
                      disabled={
                        pluginsData.loading || pluginsData.nextPageToken === ""
                      }
                      onClick={() => pluginsData.nextPage()}
                    >
                      <ChevronRight />
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}