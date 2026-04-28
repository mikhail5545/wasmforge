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

import {
  ChevronDownIcon, ChevronLeft, ChevronRight, CircleAlert,
  HardDriveUpload,
  MoreHorizontal,
  MoreVertical,
  RouteOff,
  Trash2,
  Wrench,
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { AnimatePresence, motion } from "motion/react"
import React, { useState } from "react"
import { usePaginatedData } from "@/hooks/use-paginated-data"
import {Route} from "@/types/route"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useRouter } from "next/navigation"
import { RoutesListControls } from "@/components/routes-list-controls"
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@workspace/ui/components/empty"
import { Button } from "@workspace/ui/components/button"
import { Badge } from "@workspace/ui/components/badge"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { AlertModal } from "@/components/dialog/alert-modal"

function RoutesPageContent() {
  const router = useRouter()

  const [orderField, setOrderField] = useState("created_at");
  const [orderDirection, setOrderDirection] = useState("asc");
  const [viewMode, setViewMode] = useState('table');
  const [perPage, setPerPage] = useState('10')

  const routeData = usePaginatedData<Route>(
    '/api/routes',
    'routes',
    parseInt(perPage),
    orderField,
    orderDirection as 'asc' | 'desc',
    { preload: true },
  );

  return (
    <>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={!!routeData.error}
        title={routeData.error?.message ?? "Unexpected error occurred"}
        description={
          routeData.error?.details ??
          "No additional details available. Page will be automatically reloaded in 5 seconds."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          void routeData.refetch()
        }}
      />
      <div className={"flex flex-col gap-5 p-6"}>
        <RoutesListControls
          orderField={orderField}
          setOrderField={setOrderField}
          orderDirection={orderDirection}
          setOrderDirection={setOrderDirection}
          showCreateButton={true}
          routesData={routeData}
          viewMode={viewMode}
          setViewMode={setViewMode}
          className={"justify-between"}
        />
        {routeData.loading ? (
          <div className={"flex items-center justify-center py-20"}>
            <div
              className={
                "h-10 w-10 animate-spin rounded-full border-3 border-t-foreground"
              }
            />
          </div>
        ) : (
          <div>
            {routeData.data.length === 0 ? (
              <Empty>
                <EmptyHeader>
                  <EmptyMedia variant={"icon"}>
                    <RouteOff />
                  </EmptyMedia>
                  <EmptyTitle>No Routes</EmptyTitle>
                  <EmptyDescription>
                    You haven&#39;t created any routes yet. Please create a
                    route to get started.
                  </EmptyDescription>
                </EmptyHeader>
                <EmptyContent className={"flex-row justify-center gap-4"}>
                  <Button size={"sm"} asChild>
                    <a href={"/routes/new"}>Create Route</a>
                  </Button>
                </EmptyContent>
              </Empty>
            ) : (
              <div className={"flex w-full flex-col"}>
                <AnimatePresence mode={"wait"}>
                  {viewMode == "table" && (
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
                              <TableHead className={"pl-4"}>Path</TableHead>
                              <TableHead>Target URL</TableHead>
                              <TableHead>Status</TableHead>
                              <TableHead>Created at</TableHead>
                              <TableHead>Actions</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {routeData.data.map((route) => (
                              <TableRow key={route.id}>
                                <TableCell className={"pl-4"}>
                                  {route.path}
                                </TableCell>
                                <TableCell>{route.target_url}</TableCell>
                                <TableCell>
                                  <Badge
                                    className={
                                      route.enabled
                                        ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
                                        : "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
                                    }
                                  >
                                    {route.enabled ? "Active" : "Stopped"}
                                  </Badge>
                                </TableCell>
                                <TableCell>
                                  {new Date(route.created_at).toLocaleString()}
                                </TableCell>
                                <TableCell>
                                  <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                      <Button variant={"ghost"} size={"icon"}>
                                        <MoreVertical />
                                      </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent>
                                      <DropdownMenuItem
                                        onClick={() =>
                                          router.push(
                                            `/routes/route?path=${route.path}`
                                          )
                                        }
                                      >
                                        <Wrench size={10} />
                                        <span>Details</span>
                                      </DropdownMenuItem>
                                      <DropdownMenuItem>
                                        <HardDriveUpload size={10} />
                                        Enable
                                      </DropdownMenuItem>
                                      <DropdownMenuSeparator />
                                      <DropdownMenuItem
                                        className={"text-destructive"}
                                      >
                                        <Trash2 size={10} />
                                        Delete
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
                  {viewMode == "grid" && (
                    <motion.div
                      key={"grid"}
                      className={
                        "grid w-full grid-cols-1 gap-2 md:grid-cols-3 lg:grid-cols-4"
                      }
                    >
                      {routeData.data.map((route, idx) => (
                        <motion.div
                          key={route.id}
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
                              <CardTitle className={"truncate"}>
                                {route.path}
                              </CardTitle>
                              <div
                                className={
                                  "flex flex-row items-center justify-center gap-2"
                                }
                              >
                                <Badge
                                  className={
                                    route.enabled
                                      ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
                                      : "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
                                  }
                                >
                                  {route.enabled ? "Active" : "Stopped"}
                                </Badge>
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button variant={"ghost"} size={"icon"}>
                                      <MoreHorizontal />
                                    </Button>
                                  </DropdownMenuTrigger>
                                </DropdownMenu>
                              </div>
                            </CardHeader>
                            <CardContent>
                              <div className={"flex flex-col gap-4"}>
                                <div className={"flex flex-col gap-1"}>
                                  <p className={"text-muted-foreground"}>
                                    Target URL
                                  </p>
                                  <p className={"truncate"}>
                                    {route.target_url}
                                  </p>
                                </div>
                                <div className={"flex flex-col gap-1"}>
                                  <p className={"text-muted-foreground"}>
                                    Created at
                                  </p>
                                  <p className={"truncate"}>
                                    {new Date(
                                      route.created_at
                                    ).toLocaleString()}
                                  </p>
                                </div>
                              </div>
                            </CardContent>
                            <CardFooter>
                              <Button asChild>
                                <a href={`/routes/route?path=${route.path}`}>
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
                        routeData.loading || routeData.previousPageToken === ""
                      }
                      onClick={() => routeData.previousPage()}
                    >
                      <ChevronLeft />
                    </Button>
                    <Button
                      variant={"outline"}
                      size={"icon"}
                      disabled={
                        routeData.loading || routeData.nextPageToken === ""
                      }
                      onClick={() => routeData.nextPage()}
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
    </>
  )
}

export default function RoutesPage() {

  return (
    <SidebarLayout page_title={"Routes"}>
      <RoutesPageContent />
    </SidebarLayout>
  )
}