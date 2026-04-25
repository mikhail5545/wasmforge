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
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
} from "@workspace/ui/components/dropdown-menu"
import {
  ChevronDown,
  Ellipsis,
  Grid2X2,
  Pencil,
  Plus,
  Search,
  Sheet,
  Wrench,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Field, FieldGroup } from "@workspace/ui/components/field"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from "@workspace/ui/components/input-group"
import { Badge } from "@workspace/ui/components/badge"
import { AlertModal } from "@/components/dialog/alert-modal"


export default function PluginsPage() {
  const [orderField, setOrderField] = React.useState("created_at")
  const [orderDirection, setOrderDirection] = React.useState("asc")
  const [searchQuery, setSearchQuery] = React.useState("")
  const [searchBy, setSearchBy] = React.useState("name")
  const [viewMode, setViewMode] = React.useState("table")

  const pluginsData = usePaginatedData<Plugin>(
    "/api/plugins",
    "plugins",
    10,
    orderField,
    orderDirection as "asc" | "desc",
    { preload: true }
  )

  const refetchWithQuery = React.useCallback(async () => {
    const trimmedQuery = searchQuery.trim()

    if (trimmedQuery === "") {
      pluginsData.setQueryParams({})
      await pluginsData.refetch()
      return
    }

    pluginsData.setQueryParams(
      searchBy === 'name' ? { n: trimmedQuery }
        : searchBy === 'filename' ? { fn: trimmedQuery }
          : { v: trimmedQuery }
    )

    await pluginsData.refetch()
  }, [pluginsData, searchBy, searchQuery])

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
      <div className={"flex flex-col p-6 pb-40"}>
        <div className={"flex flex-row justify-between gap-2"}>
          <FieldGroup className={"flex max-w-sm flex-row"}>
            <Field>
              <InputGroup>
                <InputGroupInput
                  aria-label={"search plugins"}
                  type={"text"}
                  value={searchQuery}
                  placeholder={
                    searchBy === "name"
                      ? "my_plugin"
                      : searchBy === "filename"
                        ? "my_plugin.wasm"
                        : "0.0.3"
                  }
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      void refetchWithQuery()
                    }
                  }}
                />
                <InputGroupAddon className={"block-start"}>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <InputGroupButton>
                        <span>Search by</span>
                        <ChevronDown className={"mt-1 inline-block"} />
                      </InputGroupButton>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent>
                      <DropdownMenuRadioGroup
                        value={searchBy}
                        onValueChange={setSearchBy}
                      >
                        <DropdownMenuRadioItem value={"name"}>
                          Name
                        </DropdownMenuRadioItem>
                        <DropdownMenuRadioItem value={"filename"}>
                          Filename
                        </DropdownMenuRadioItem>
                        <DropdownMenuRadioItem value={"version"}>
                          Version
                        </DropdownMenuRadioItem>
                      </DropdownMenuRadioGroup>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </InputGroupAddon>
              </InputGroup>
            </Field>
            <Button
              variant={"outline"}
              size={"icon"}
              onClick={() => { void refetchWithQuery() }}
              disabled={pluginsData.loading || !!pluginsData.error}
            >
              <Search />
            </Button>
          </FieldGroup>
          <div className={"flex flex-row gap-2"}>
            <Button variant={"outline"} asChild>
              <a href={"/plugins/new"}>
                <Plus />
                New
              </a>
            </Button>
            <Button
              variant={"outline"}
              size={"icon"}
              onClick={() =>
                setViewMode(viewMode === "table" ? "grid" : "table")
              }
            >
              {viewMode === "table" ? <Grid2X2 /> : <Sheet />}
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant={"outline"} size={"icon"}>
                  <Ellipsis />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem asChild>
                  <a href={"/plugins/new"}>
                    <Plus />
                    New
                  </a>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger inset>
                    Order by
                  </DropdownMenuSubTrigger>
                  <DropdownMenuSubContent>
                    <DropdownMenuRadioGroup
                      value={orderField}
                      onValueChange={setOrderField}
                    >
                      <DropdownMenuRadioItem value={"name"}>
                        Name
                      </DropdownMenuRadioItem>
                      <DropdownMenuRadioItem value={"filename"}>
                        Filename
                      </DropdownMenuRadioItem>
                      <DropdownMenuRadioItem value={"version"}>
                        Version
                      </DropdownMenuRadioItem>
                      <DropdownMenuRadioItem value={"created_at"}>
                        Created at
                      </DropdownMenuRadioItem>
                    </DropdownMenuRadioGroup>
                  </DropdownMenuSubContent>
                </DropdownMenuSub>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger inset>
                    Direction
                  </DropdownMenuSubTrigger>
                  <DropdownMenuSubContent>
                    <DropdownMenuRadioGroup
                      value={orderDirection}
                      onValueChange={setOrderDirection}
                    >
                      <DropdownMenuRadioItem value={"asc"}>
                        Ascending
                      </DropdownMenuRadioItem>
                      <DropdownMenuRadioItem value={"desc"}>
                        Descending
                      </DropdownMenuRadioItem>
                    </DropdownMenuRadioGroup>
                  </DropdownMenuSubContent>
                </DropdownMenuSub>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        {pluginsData.loading ? (
          <div className={"flex items-center justify-center py-20"}>
            <Spinner className={"h-10 w-10"} />
          </div>
        ) : (
          <div>
            {pluginsData.data.length === 0 ? (
              <div className={"flex items-center justify-center py-20"}>
                <div className={"flex flex-col items-center justify-center"}>
                  <p className={"text-xl font-semibold"}>
                    There is no plugins to show you
                  </p>
                  <p className={"text-md"}>
                    Maybe you want to{" "}
                    <a className={"underline"} href={"/plugins/new"}>
                      create one?
                    </a>
                  </p>
                </div>
              </div>
            ) : (
              <div
                className={
                  "flex w-full flex-col items-center justify-center py-10"
                }
              >
                <AnimatePresence mode={"wait"}>
                  {viewMode === "table" && (
                    <motion.div
                      initial={{ opacity: 0, x: 100 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: -100 }}
                      transition={{ duration: 0.5, ease: "easeIn" }}
                      className={"flex w-full flex-col gap-2"}
                    >
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Filename</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>Version</TableHead>
                            <TableHead>Created at</TableHead>
                            <TableHead>Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {pluginsData.data.map((plugin) => (
                            <TableRow key={plugin.id}>
                              <TableCell>{plugin.filename}</TableCell>
                              <TableCell>{plugin.name}</TableCell>
                              <TableCell>{plugin.version}</TableCell>
                              <TableCell>
                                {new Date(plugin.created_at).toLocaleString()}
                              </TableCell>
                              <TableCell>
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <button className={"p-3"}>
                                      <Ellipsis size={12} />
                                    </button>
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
                                    <DropdownMenuItem asChild>
                                      <a
                                        href={`/plugins/plugin/edit?name=${encodeURIComponent(plugin.name)}&version=${encodeURIComponent(plugin.version)}`}
                                      >
                                        <Pencil />
                                        <span>Edit</span>
                                      </a>
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </motion.div>
                  )}
                  {viewMode === "grid" && (
                    <motion.div
                      initial={{ opacity: 0, y: 100 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -100 }}
                      transition={{ duration: 0.5, ease: "easeIn" }}
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
                                      <Ellipsis />
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
                                    <DropdownMenuItem asChild>
                                      <a
                                        href={`/plugins/plugin/edit?name=${encodeURIComponent(plugin.name)}&version=${encodeURIComponent(plugin.version)}`}
                                      >
                                        <Pencil />
                                        <span>Edit</span>
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
                {pluginsData.nextPageToken && (
                  <div className={"flex items-center justify-center py-5"}>
                    <Button
                      size={"sm"}
                      variant={"ghost"}
                      onClick={() =>
                        pluginsData.nextPage(pluginsData.nextPageToken)
                      }
                    >
                      Load more
                    </Button>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}