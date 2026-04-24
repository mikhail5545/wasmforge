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
  Ban,
  ChevronDownIcon,
  Grid2X2,
  HardDriveUpload,
  Plus,
  Search,
  Settings2,
  Sheet,
  TextAlignJustify,
  Trash2,
  Wrench,
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
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
import { useState } from "react"
import { usePaginatedData } from "@/hooks/use-paginated-data"
import {Route} from "@/types/route"
import { Input } from "@workspace/ui/components/input"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useRouter } from "next/navigation"

export default function RoutesPage() {
  const router = useRouter()

  const [orderField, setOrderField] = useState("created_at");
  const [orderDirection, setOrderDirection] = useState("asc");
  const [searchQuery, setSearchQuery] = useState("");
  const [searchBy, setSearchBy] = useState("path");
  const [viewMode, setViewMode] = useState('table');

  const routeData = usePaginatedData<Route>(
    '/api/routes',
    'routes',
    10,
    orderField,
    orderDirection as 'asc' | 'desc',
    { preload: true },
  );

  const refetchWithQuery = async () => {
    const trimmedQuery = searchQuery.trim()

    if (trimmedQuery === "") {
      routeData.setQueryParams({})
      await routeData.refetch()
      return
    }

    routeData.setQueryParams(
      searchBy === "path" ? { paths: trimmedQuery } : { turls: trimmedQuery }
    )
    await routeData.refetch()
  }

  return (
    <SidebarLayout page_title={"Routes"}>
      <div className={"flex flex-col p-6 pb-40"}>
        <div className={"flex flex-row gap-2"}>
          <div className={"flex min-w-100 rounded-lg border"}>
            <div className={"relative inline-flex min-w-30"}>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <motion.button
                    className={
                      "flex-row items-center justify-between rounded-lg px-3 py-2 text-sm"
                    }
                  >
                    Search by
                    <ChevronDownIcon
                      size={10}
                      className={"ml-3 inline-block"}
                    />
                  </motion.button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuRadioGroup
                    value={searchBy}
                    onValueChange={setSearchBy}
                  >
                    <DropdownMenuRadioItem value={"path"}>
                      Path
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value={"target_url"}>
                      Target URL
                    </DropdownMenuRadioItem>
                  </DropdownMenuRadioGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
            <Input
              aria-label={"search a route"}
              aria-labelledby={"routes-search-input"}
              className={"h-full rounded-l-none"}
              placeholder={
                searchBy === "path"
                  ? "/api/resource"
                  : "https://example.com/api/resource"
              }
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  void refetchWithQuery()
                }
              }}
            />
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className={
                "flex-row items-center justify-between rounded-lg p-3 text-sm"
              }
              onClick={() => void refetchWithQuery()}
            >
              <Search size={15} />
            </motion.button>
          </div>
          <motion.a
            whileTap={{ scale: 0.95 }}
            className={
              "flex flex-row items-center justify-center gap-2 rounded-lg border border-background bg-secondary px-3 py-1 text-sm text-foreground transition-colors duration-200 hover:bg-accent"
            }
            href={"/routes/new"}
          >
            <Plus size={15} />
            New
          </motion.a>
          <motion.button
            whileTap={{ scale: 0.95 }}
            className={
              "rounded-lg border border-background bg-secondary px-3 py-3 text-foreground transition-colors duration-200 hover:bg-accent"
            }
            onClick={() => setViewMode(viewMode === "table" ? "grid" : "table")}
          >
            <AnimatePresence mode={"wait"}>
              {viewMode === "grid" && <Grid2X2 size={15} />}
              {viewMode === "table" && <Sheet size={15} />}
            </AnimatePresence>
          </motion.button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <motion.button
                whileTap={{ scale: 0.95 }}
                className={
                  "rounded-lg border border-background bg-secondary px-3 py-3 text-foreground transition-colors duration-200 hover:bg-accent"
                }
              >
                <TextAlignJustify size={15} />
              </motion.button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem>
                <Plus size={10} />
                <span>New</span>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuGroup>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger inset>
                    Order by
                  </DropdownMenuSubTrigger>
                  <DropdownMenuPortal>
                    <DropdownMenuSubContent>
                      <DropdownMenuRadioGroup
                        value={orderField}
                        onValueChange={setOrderField}
                      >
                        <DropdownMenuRadioItem value={"created_at"}>
                          Created Date
                        </DropdownMenuRadioItem>
                        <DropdownMenuRadioItem value={"updated_at"}>
                          Updated Date
                        </DropdownMenuRadioItem>
                        <DropdownMenuRadioItem value={"route"}>
                          Route
                        </DropdownMenuRadioItem>
                      </DropdownMenuRadioGroup>
                    </DropdownMenuSubContent>
                  </DropdownMenuPortal>
                </DropdownMenuSub>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger inset>
                    Direction
                  </DropdownMenuSubTrigger>
                  <DropdownMenuPortal>
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
                  </DropdownMenuPortal>
                </DropdownMenuSub>
                <DropdownMenuSub>
                  <DropdownMenuSubTrigger inset>
                    View mode
                  </DropdownMenuSubTrigger>
                  <DropdownMenuPortal>
                    <DropdownMenuSubContent>
                      <DropdownMenuRadioGroup
                        value={viewMode}
                        onValueChange={setViewMode}
                      >
                        <DropdownMenuRadioItem value={"table"}>
                          Table
                        </DropdownMenuRadioItem>
                        <DropdownMenuRadioItem value={"grid"}>
                          Grid
                        </DropdownMenuRadioItem>
                      </DropdownMenuRadioGroup>
                    </DropdownMenuSubContent>
                  </DropdownMenuPortal>
                </DropdownMenuSub>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
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
              <div className={"flex items-center justify-center py-20"}>
                <div className={"flex flex-col items-center justify-center"}>
                  <p className={"text-xl font-semibold"}>
                    There is no routes to show you
                  </p>
                  <p className={"text-md"}>
                    Maybe you want to{" "}
                    <a className={"underline"} href={"/routes/new"}>
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
                  {viewMode == "table" && (
                    <motion.div
                      initial={{ opacity: 0, y: 100 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -100 }}
                      transition={{ duration: 0.5, ease: "easeIn" }}
                      className={"flex w-full flex-col gap-2"}
                    >
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Path</TableHead>
                            <TableHead>Target URL</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Created at</TableHead>
                            <TableHead>Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {routeData.data.map((route) => (
                            <TableRow key={route.id}>
                              <TableCell>{route.path}</TableCell>
                              <TableCell>{route.target_url}</TableCell>
                              <TableCell>
                                {route.enabled ? "enabled" : "disabled"}
                              </TableCell>
                              <TableCell>
                                {new Date(route.created_at).toLocaleString()}
                              </TableCell>
                              <TableCell>
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <button className={"p-3"}>
                                      <Settings2 size={12} />
                                    </button>
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
                      {routeData.nextPageToken && (
                        <div
                          className={"flex items-center justify-center py-5"}
                        >
                          <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            className={"px-4 py-2 text-xs underline"}
                          >
                            Load More
                          </motion.button>
                        </div>
                      )}
                    </motion.div>
                  )}
                  {viewMode == "grid" && (
                    <motion.div
                      initial={{ opacity: 0, y: 100 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -100 }}
                      transition={{ duration: 0.5, ease: "easeIn" }}
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
                          className={
                            "col-span-1 rounded-lg border border-border bg-card p-5"
                          }
                        >
                          <div className={"flex flex-col gap-5"}>
                            <div
                              className={
                                "flex flex-row items-center justify-between"
                              }
                            >
                              <div className={"flex flex-col gap-1"}>
                                <p
                                  className={
                                    "text-xs text-secondary-foreground"
                                  }
                                >
                                  Path
                                </p>
                                <p className={"text-md truncate font-semibold"}>
                                  {route.path}
                                </p>
                              </div>
                              <div
                                className={"flex items-center justify-center"}
                              >
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <button className={"p-3"}>
                                      <Settings2 size={12} />
                                    </button>
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
                                    {route.enabled ? (
                                      <DropdownMenuItem>
                                        <Ban size={10} />
                                        Disable
                                      </DropdownMenuItem>
                                    ) : (
                                      <DropdownMenuItem>
                                        <HardDriveUpload size={10} />
                                        Enable
                                      </DropdownMenuItem>
                                    )}
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                      className={"text-destructive"}
                                    >
                                      <Trash2 size={10} />
                                      Delete
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </div>
                            </div>
                            <div className={"flex flex-col gap-1"}>
                              <p
                                className={"text-sm text-secondary-foreground"}
                              >
                                Target URL
                              </p>
                              <p className={"text-md truncate font-semibold"}>
                                {route.target_url}
                              </p>
                            </div>
                            <div className={"flex flex-col gap-1"}>
                              <p
                                className={"text-sm text-secondary-foreground"}
                              >
                                Status
                              </p>
                              <p className={"text-md truncate font-semibold"}>
                                {route.enabled
                                  ? "Active route"
                                  : "Inactive route"}
                              </p>
                            </div>
                            <div
                              className={
                                "flex flex-row items-center justify-end"
                              }
                            >
                              <motion.a
                                whileHover={{ scale: 1.05 }}
                                whileTap={{ scale: 0.95 }}
                                href={`/routes/route?path=${route.path}`}
                                className={
                                  "flex flex-row items-center justify-center rounded-xl bg-primary px-4 py-2 text-center text-sm text-primary-foreground transition-opacity duration-200 hover:opacity-850"
                                }
                              >
                                View Details
                              </motion.a>
                            </div>
                          </div>
                        </motion.div>
                      ))}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}