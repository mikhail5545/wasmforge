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
import { useRouter, useSearchParams } from "next/navigation"
import { useData } from "@/hooks/use-data"
import { Plugin } from "@/types/Plugin"
import { Route } from "@/types/route"
import { Spinner } from "@workspace/ui/components/spinner"
import React from "react"
import { PluginCard } from "@/components/ui/plugin-card"
import { usePaginatedData } from "@/hooks/use-paginated-data"

import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { ChevronDownIcon, ChevronLeft, ChevronRight, HardDriveUpload, MoreVertical, RouteOff, Trash2, Wrench } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { ScrollArea } from "@workspace/ui/components/scroll-area"
import { RoutesListControls } from "@/components/routes-list-controls"
import { useMutation } from "@/hooks/use-mutation"
import { AlertModal } from "@/components/dialog/alert-modal"
import {
  Dialog,
  DialogContent,
  DialogDescription, DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { Badge } from "@workspace/ui/components/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@workspace/ui/components/dropdown-menu"

export default function PluginPage() {
  const params = useSearchParams()
  const name = params.get("name") ?? ""
  const version = params.get("version") ?? ""
  const pluginData = useData<Plugin>(
    `http://localhost:8080/api/plugins/${encodeURIComponent(name)}?version=${encodeURIComponent(version)}`,
    "plugin"
  )

  const router = useRouter()

  const [pluginId, setPluginId] = React.useState<string>("")
  const [orderDirection, setOrderDirection] = React.useState("asc")
  const [orderField, setOrderField] = React.useState<string>("created_at")
  const [perPage, setPerPage] = React.useState('10')
  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")
  const [successRedirect, setSuccessRedirect] = React.useState<string | null>(
    null
  )
  const [showDeleteConfirmation, setShowDeleteConfirmation] = React.useState(false)


  React.useEffect(() => {
    if (pluginData.data && pluginId === "") {
      setPluginId(pluginData.data.id)
    }
  }, [pluginData.data, pluginId])

  const routesData = usePaginatedData<Route>(
    `/api/routes?pids=${pluginId}`,
    "routes",
    Number(perPage),
    orderField,
    orderDirection as "asc" | "desc",
    { preload: true }
  )

  const deletePlugin = React.useCallback(async () => {
    if (!pluginData.data) return

    const result = await mutate(
      `http://localhost:8080/api/plugins/${pluginId}`,
      'DELETE'
    )

    if (result.success) {
      setShowSuccess(true)
      setSuccessMessage(
        "Plugin deleted successfully. You will be redirected to plugins list in 5 seconds."
      )
      setSuccessRedirect("/plugins")
    }
  }, [mutate, pluginData.data, pluginId])

  return (
    <SidebarLayout page_title={"Plugin"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!pluginData.error || !!error || !!routesData.error}
        description={
          pluginData.error?.details ||
          error?.details ||
          routesData.error?.details ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={() => {
          if (pluginData.error) void pluginData.refetch()
          if (routesData.error) void routesData.refetch()
          if (error) reset()
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        title={"Success"}
        visible={showSuccess}
        description={successMessage}
        onClose={() => {
          setShowSuccess(false)
          if (successRedirect) router.push(successRedirect)
        }}
      />
      <Dialog
        open={showDeleteConfirmation}
        onOpenChange={setShowDeleteConfirmation}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Are you sure?</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            This action cannot be undone. This will permanently delete plugin
            and remove it&#39;s binary file.
          </DialogDescription>
          <DialogFooter>
            <Button
              variant={"outline"}
              onClick={() => setShowDeleteConfirmation(false)}
            >
              Cancel
            </Button>
            <Button
              variant={"destructive"}
              onClick={deletePlugin}
              disabled={loading || !!error}
            >
              {loading && <Spinner />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <div className={"flex flex-col p-6"}>
        {pluginData.loading && pluginData.data === null ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <div className={"flex flex-col"}>
            {pluginData.error ? (
              <div className={"flex items-center justify-center py-50"}>
                <p className={"text-xl text-muted-foreground"}>
                  Something went wrong
                </p>
              </div>
            ) : (
              <>
                <div
                  className={"flex flex-row items-center justify-end pb-5"}
                ></div>
                <div className={"flex flex-col gap-5 lg:flex-row"}>
                  <div className={"w-full lg:w-1/3"}>
                    <PluginCard
                      plugin={pluginData.data!}
                      className={"w-full"}
                      onDelete={() => setShowDeleteConfirmation(true)}
                    />
                  </div>
                  <div className={"w-full lg:w-2/3"}>
                    {routesData.loading ? (
                      <div className={"flex items-center justify-center py-20"}>
                        <Spinner className={"h-8 w-8"} />
                      </div>
                    ) : (
                      <div className={"flex w-full flex-col gap-5"}>
                        <div className={"flex w-full flex-col gap-1"}>
                          <p className={"text-xl"}>Associated Routes</p>
                          <p className={"text-muted-foreground"}>
                            This plugin is attached to this routes and working
                            in the middleware WASM runtime.
                          </p>
                        </div>
                        <RoutesListControls
                          orderField={orderField}
                          setOrderField={setOrderField}
                          orderDirection={orderDirection}
                          setOrderDirection={setOrderDirection}
                          showCreateButton={true}
                          createUrlOverride={`/routes/plugins/new?pluginId=${pluginData.data?.id}`}
                          routesData={routesData}
                          className={"justify-between"}
                        />
                        {routesData.data.length === 0 ? (
                          <Empty>
                            <EmptyHeader>
                              <EmptyMedia variant={"icon"}>
                                <RouteOff />
                              </EmptyMedia>
                              <EmptyTitle>No Associated Routes</EmptyTitle>
                              <EmptyDescription>
                                You haven&#39;t attached this plugin to any
                                route yet. Please create a route and attach this
                                plugin to it, or attach this plugin to existing
                                routes.
                              </EmptyDescription>
                            </EmptyHeader>
                            <EmptyContent
                              className={"flex-row justify-center gap-4"}
                            >
                              <Button size={"sm"} variant={"outline"} asChild>
                                <a href={"/routes/new"}>Create Route</a>
                              </Button>
                              <Button size={"sm"} asChild>
                                <a
                                  href={`/routes/plugins/new?pluginId=${pluginData.data?.id}`}
                                >
                                  Attach Route
                                </a>
                              </Button>
                            </EmptyContent>
                          </Empty>
                        ) : (
                          <div className={"overflow-hidden rounded-lg border"}>
                            <Table>
                              <TableHeader
                                className={"sticky top-0 z-10 bg-muted"}
                              >
                                <TableRow>
                                  <TableHead className={"pl-4"}>Path</TableHead>
                                  <TableHead>Target URL</TableHead>
                                  <TableHead>Status</TableHead>
                                  <TableHead>Created at</TableHead>
                                  <TableHead>Actions</TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {routesData.data.map((route) => (
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
                                      {new Date(
                                        route.created_at
                                      ).toLocaleString()}
                                    </TableCell>
                                    <TableCell>
                                      <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                          <Button
                                            variant={"ghost"}
                                            size={"icon"}
                                          >
                                            <MoreVertical />
                                          </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent>
                                          <DropdownMenuItem asChild>
                                            <a href={`/routes/${route.id}`}>
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
                        )}
                      </div>
                    )}
                    <div className={"mt-5 flex flex-row justify-end gap-5"}>
                      <div
                        className={
                          "flex flex-row items-center justify-center gap-2"
                        }
                      >
                        <p className={"text-sm font-semibold"}>Rows per page</p>
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
                            routesData.loading ||
                            routesData.previousPageToken === ""
                          }
                          onClick={() => routesData.previousPage()}
                        >
                          <ChevronLeft />
                        </Button>
                        <Button
                          variant={"outline"}
                          size={"icon"}
                          disabled={
                            routesData.loading ||
                            routesData.nextPageToken === ""
                          }
                          onClick={() => routesData.nextPage()}
                        >
                          <ChevronRight />
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}
