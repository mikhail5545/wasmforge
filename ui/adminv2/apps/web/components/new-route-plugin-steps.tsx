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
import { Plugin } from "types/Plugin"
import { PaginatedData } from "@/hooks/use-paginated-data"
import React from "react"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Spinner } from "@workspace/ui/components/spinner"
import { RoutesListControls } from "@/components/routes-list-controls"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { RouteOff } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { ScrollArea } from "@workspace/ui/components/scroll-area"
import { motion } from "motion/react"
import { PluginsListControls } from "@/components/plugins-list-controls"

interface RoutePluginStep1Props {
  selectedRoute: Route | null
  setSelectedRoute: React.Dispatch<React.SetStateAction<Route | null>>
  routesData: PaginatedData<Route>
  orderField: string
  setOrderField: React.Dispatch<React.SetStateAction<string>>
  orderDirection: string
  setOrderDirection: React.Dispatch<React.SetStateAction<string>>
}

interface RoutePluginStep2Props {
  selectedPlugin: Plugin | null
  setSelectedPlugin: React.Dispatch<React.SetStateAction<Plugin | null>>
  pluginsData: PaginatedData<Plugin>
  orderField: string
  setOrderField: React.Dispatch<React.SetStateAction<string>>
  orderDirection: string
  setOrderDirection: React.Dispatch<React.SetStateAction<string>>
}

const RoutePluginStep1: React.FC<RoutePluginStep1Props> = ({
  selectedRoute,
  setSelectedRoute,
  routesData,
  orderField,
  setOrderField,
  orderDirection,
  setOrderDirection,
}) => {
  return (
    <motion.div
      key={"route-select"}
      initial={{ opacity: 0, x: -100 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 100 }}
      transition={{ type: "spring", stiffness: 300, damping: 30 }}
      className={"flex w-full flex-col gap-5"}
    >
      <p className={"text-xl"}>Select Route</p>
      <div className={"flex flex-col gap-5 lg:flex-row"}>
        <div className={"w-full lg:w-1/3"}>
          <Card className={"w-full"}>
            <CardHeader>
              <CardTitle>Selected Route</CardTitle>
              <CardDescription>
                Plugin will be attached to this route and will modify requests
                accepted by it.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {selectedRoute ? (
                <div className={"flex w-full flex-row items-center"}>
                  <div
                    className={
                      "flex w-1/3 flex-col gap-4 text-muted-foreground"
                    }
                  >
                    <span>Path</span>
                    <span>Target URL</span>
                    <span>Created At</span>
                  </div>
                  <div className={"flex w-2/3 flex-col gap-4 truncate"}>
                    <span>{selectedRoute.path}</span>
                    <span>{selectedRoute.target_url}</span>
                    <span>
                      {new Date(selectedRoute.created_at).toLocaleString()}
                    </span>
                  </div>
                </div>
              ) : (
                <div
                  className={"flex flex-col items-center justify-center py-20"}
                >
                  <p className={"text-lg"}>No route selected</p>
                  <p className={"text-sm text-muted-foreground"}>
                    Please select a route to continue
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
        <div className={"w-full lg:w-2/3"}>
          {routesData.loading ? (
            <div className={"flex items-center justify-center py-20"}>
              <Spinner className={"h-8 w-8"} />
            </div>
          ) : (
            <div className={"flex w-full flex-col gap-5"}>
              <RoutesListControls
                orderField={orderField}
                setOrderField={setOrderField}
                orderDirection={orderDirection}
                setOrderDirection={setOrderDirection}
                showCreateButton={false}
                routesData={routesData}
              />
              {routesData.data.length === 0 ? (
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant={"icon"}>
                      <RouteOff />
                    </EmptyMedia>
                    <EmptyTitle>No Routes Available</EmptyTitle>
                    <EmptyDescription>
                      You haven&#39;t created any routes yet. Get started by
                      creating a new route first.
                    </EmptyDescription>
                  </EmptyHeader>
                  <EmptyContent className={"flex-row justify-center"}>
                    <Button asChild>
                      <a href={"/routes/new"}>Create Route</a>
                    </Button>
                  </EmptyContent>
                </Empty>
              ) : (
                <ScrollArea className={"h-96 w-full rounded-xl"}>
                  <div className={"mr-2 flex flex-col gap-5"}>
                    {routesData.data.map((route) => (
                      <div
                        key={route.id}
                        className={`flex w-full cursor-pointer flex-row gap-5 rounded-xl p-5 transition-colors duration-200 ${
                          selectedRoute?.id === route.id
                            ? "bg-primary text-primary-foreground"
                            : "bg-card text-card-foreground hover:bg-accent"
                        }`}
                        onClick={() => setSelectedRoute(route)}
                      >
                        <div className={"flex w-1/4 flex-col gap-2"}>
                          <p className={"text-sm opacity-70"}>Path</p>
                          <p className={"truncate font-semibold"}>
                            {route.path}
                          </p>
                        </div>
                        <div className={"flex w-2/4 flex-col gap-2"}>
                          <p className={"text-sm opacity-70"}>Target URL</p>
                          <p className={"truncate font-semibold"}>
                            {route.target_url}
                          </p>
                        </div>
                        <div className={"flex w-1/4 flex-col gap-2"}>
                          <p className={"text-sm opacity-70"}>Status</p>
                          <p className={"truncate font-semibold"}>
                            {route.enabled ? "Enabled" : "Disabled"}
                          </p>
                        </div>
                      </div>
                    ))}
                    {routesData.nextPageToken && (
                      <div
                        className={"flex flex-row items-center justify-center"}
                      >
                        <Button
                          onClick={() =>
                            routesData.nextPage(routesData.nextPageToken)
                          }
                          variant={"ghost"}
                          size={"sm"}
                        >
                          Load More
                        </Button>
                      </div>
                    )}
                  </div>
                </ScrollArea>
              )}
            </div>
          )}
        </div>
      </div>
    </motion.div>
  )
}

const RoutePluginStep2: React.FC<RoutePluginStep2Props> = ({
  selectedPlugin,
  setSelectedPlugin,
  pluginsData,
  orderField,
  setOrderField,
  orderDirection,
  setOrderDirection
}) => {
  return (
    <motion.div
      key={"route-select"}
      initial={{ opacity: 0, x: -100 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 100 }}
      transition={{ type: "spring", stiffness: 300, damping: 30 }}
      className={"flex w-full flex-col gap-5"}
    >
      <p className={"text-xl"}>Select Plugin</p>
      <div className={"flex flex-col gap-5 lg:flex-row"}>
        <div className={"w-full lg:w-1/3"}>
          <Card className={"w-full"}>
            <CardHeader>
              <CardTitle>Selected Plugin</CardTitle>
              <CardDescription>
                This plugin will be attached to selected route
              </CardDescription>
            </CardHeader>
            <CardContent>
              {selectedPlugin ? (
                <div className={"flex w-full flex-row items-center"}>
                  <div
                    className={
                      "flex w-1/3 flex-col gap-4 text-muted-foreground"
                    }
                  >
                    <span>Name</span>
                    <span>Filename</span>
                    <span>Version</span>
                    <span>Created At</span>
                  </div>
                  <div className={"flex w-2/3 flex-col gap-4 truncate"}>
                    <span>{selectedPlugin.name}</span>
                    <span>{selectedPlugin.filename}</span>
                    <span>{`v${selectedPlugin.version}`}</span>
                    <span>
                      {new Date(selectedPlugin.created_at).toLocaleString()}
                    </span>
                  </div>
                </div>
              ) : (
                <div
                  className={"flex flex-col items-center justify-center py-20"}
                >
                  <p className={"text-lg"}>No plugin selected</p>
                  <p className={"text-sm text-muted-foreground"}>
                    Please select a plugin to continue
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
        <div className={"w-full lg:w-2/3"}>
          {pluginsData.loading ? (
            <div className={"flex items-center justify-center py-20"}>
              <Spinner className={"h-8 w-8"} />
            </div>
          ) : (
            <div className={"flex w-full flex-col gap-5"}>
              <PluginsListControls
                orderField={orderField}
                setOrderField={setOrderField}
                orderDirection={orderDirection}
                setOrderDirection={setOrderDirection}
                showCreateButton={false}
                pluginsData={pluginsData}
              />
              {pluginsData.data.length === 0 ? (
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant={"icon"}>
                      <RouteOff />
                    </EmptyMedia>
                    <EmptyTitle>No Plugins Available</EmptyTitle>
                    <EmptyDescription>
                      You haven&#39;t created any plugins yet. Get started by
                      creating a new plugin first.
                    </EmptyDescription>
                  </EmptyHeader>
                  <EmptyContent className={"flex-row justify-center"}>
                    <Button asChild>
                      <a href={"/plugins/new"}>Create Plugin</a>
                    </Button>
                  </EmptyContent>
                </Empty>
              ) : (
                <ScrollArea className={"h-96 w-full rounded-xl"}>
                  <div className={"mr-2 flex flex-col gap-5"}>
                    {pluginsData.data.map((plugin) => (
                      <div
                        key={plugin.id}
                        className={`flex w-full cursor-pointer flex-row gap-5 rounded-xl p-5 transition-colors duration-200 ${
                          selectedPlugin?.id === plugin.id
                            ? "bg-primary text-primary-foreground"
                            : "bg-card text-card-foreground hover:bg-accent"
                        }`}
                        onClick={() => setSelectedPlugin(plugin)}
                      >
                        <div className={"flex w-1/4 flex-col gap-2"}>
                          <p className={"text-sm opacity-70"}>Name</p>
                          <p className={"truncate font-semibold"}>
                            {plugin.name}
                          </p>
                        </div>
                        <div className={"flex w-2/4 flex-col gap-2"}>
                          <p className={"text-sm opacity-70"}>Filename</p>
                          <p className={"truncate font-semibold"}>
                            {plugin.filename}
                          </p>
                        </div>
                        <div className={"flex w-1/4 flex-col gap-2"}>
                          <p className={"text-sm opacity-70"}>Version</p>
                          <p className={"truncate font-semibold"}>
                            {`v${plugin.version}`}
                          </p>
                        </div>
                      </div>
                    ))}
                    {pluginsData.nextPageToken && (
                      <div
                        className={"flex flex-row items-center justify-center"}
                      >
                        <Button
                          onClick={() =>
                            pluginsData.nextPage(pluginsData.nextPageToken)
                          }
                          variant={"ghost"}
                          size={"sm"}
                        >
                          Load More
                        </Button>
                      </div>
                    )}
                  </div>
                </ScrollArea>
              )}
            </div>
          )}
        </div>
      </div>
    </motion.div>
  )
}

export {
  RoutePluginStep1,
  RoutePluginStep2,
}