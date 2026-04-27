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
import {
  ChevronDownIcon,
  ChevronLeft,
  ChevronRight,
  FileBraces,
  RotateCcw,
  ToyBrick,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { PaginatedData } from "@/hooks/use-paginated-data"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { cn } from "@workspace/ui/lib/utils"
import { Checkbox } from "@workspace/ui/components/checkbox"
import { RoutePlugin } from "@/types/RoutePlugin"
import { Badge } from "@workspace/ui/components/badge"
import {
  DropdownMenu,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import { DropdownMenuContent } from "@radix-ui/react-dropdown-menu"
import { PluginsListControls } from "@/components/plugins-list-controls"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupText,
  InputGroupTextarea,
} from "@workspace/ui/components/input-group"

interface NewRoutePluginSectionSelectRouteProps {
  selectedRoute: Route | null
  setSelectedRoute: React.Dispatch<React.SetStateAction<Route | null>>
  orderField: string
  setOrderField: React.Dispatch<React.SetStateAction<string>>
  orderDirection: string
  setOrderDirection: React.Dispatch<React.SetStateAction<string>>
  routesData: PaginatedData<Route>
  setFormState: React.Dispatch<
    React.SetStateAction<Omit<RoutePlugin, "id" | "created_at" | "plugin">>
  >
  perPage: string
  setPerPage: React.Dispatch<React.SetStateAction<string>>
  className?: string
}

interface NewRoutePluginSectionSelectPluginProps {
  selectedPlugin: Plugin | null
  pluginsData: PaginatedData<Plugin>
  orderField: string
  orderDirection: string
  perPage: string
  setOrderField: React.Dispatch<React.SetStateAction<string>>
  setOrderDirection: React.Dispatch<React.SetStateAction<string>>
  setFormState: React.Dispatch<
    React.SetStateAction<Omit<RoutePlugin, "id" | "created_at" | "plugin">>
  >
  setSelectedPlugin: React.Dispatch<React.SetStateAction<Plugin | null>>
  setPerPage: React.Dispatch<React.SetStateAction<string>>
  className?: string
}

const NewRoutePluginSectionSelectRoute: React.FC<
  NewRoutePluginSectionSelectRouteProps
> = ({
  selectedRoute,
  setSelectedRoute,
  orderField,
  setOrderField,
  orderDirection,
  setOrderDirection,
  routesData,
  setFormState,
  perPage,
  setPerPage,
  className,
}) => {
  return (
    <div className={cn("flex flex-col gap-5 p-6", className)}>
      <div className={"flex flex-col gap-2"}>
        <p className={"text-2xl"}>1. Select a Route</p>
        <p className={"text-md"}>
          Plugin will be attached to this route. It will work in the sandboxed
          WASM runtime inside of the middleware and be able to modify requests
          that go through this route.
        </p>
      </div>
      <div className={"flex flex-col gap-5 lg:flex-row"}>
        <div className={"w-full lg:w-1/3"}>
          <Card>
            <CardHeader>
              <CardTitle>Selected Route</CardTitle>
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
        <div className={"w-full gap-5 lg:w-2/3"}>
          {routesData.loading ? (
            <div className={"flex items-center justify-center py-20"}>
              <Spinner className={"h-8 w-8"} />
            </div>
          ) : (
            <div className={"flex w-full flex-col gap-5"}>
              <p className={"text-xl"}>Available Routes</p>
              <RoutesListControls
                orderField={orderField}
                setOrderField={setOrderField}
                orderDirection={orderDirection}
                setOrderDirection={setOrderDirection}
                showCreateButton={true}
                routesData={routesData}
                className={"justify-between"}
              />
              {routesData.data.length === 0 ? (
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant={"icon"}>
                      <ToyBrick />
                    </EmptyMedia>
                    <EmptyTitle>No Routes</EmptyTitle>
                    <EmptyDescription>
                      You haven&#39;t created any routes yet. Start with
                      creating one.
                    </EmptyDescription>
                  </EmptyHeader>
                  <EmptyContent className={"flex-row justify-center gap-4"}>
                    <Button size={"sm"} asChild>
                      <a href={"/routes/new"}>Create Route</a>
                    </Button>
                  </EmptyContent>
                </Empty>
              ) : (
                <div className={"overflow-hidden rounded-lg border"}>
                  <Table>
                    <TableHeader className={"sticky top-0 z-10 bg-muted"}>
                      <TableRow>
                        <TableCell></TableCell>
                        <TableHead>Path</TableHead>
                        <TableHead>Target URL</TableHead>
                        <TableHead>Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {routesData.data.map((route) => (
                        <TableRow key={route.id}>
                          <TableCell>
                            <Checkbox
                              checked={route.id === selectedRoute?.id}
                              onCheckedChange={() => {
                                if (route.id === selectedRoute?.id) {
                                  setSelectedRoute(null)
                                } else {
                                  setSelectedRoute(route)
                                  setFormState((prev) => ({
                                    ...prev,
                                    route_id: route.id,
                                  }))
                                }
                              }}
                            />
                          </TableCell>
                          <TableCell>{route.path}</TableCell>
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
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>
          )}
          <div className={"mt-5 flex flex-row justify-end gap-5"}>
            <div className={"flex flex-row items-center justify-center gap-2"}>
              <p className={"text-sm font-semibold"}>Rows per page</p>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant={"outline"} disabled={routesData.loading}>
                    {perPage}
                    <ChevronDownIcon />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuRadioGroup
                    value={perPage}
                    onValueChange={setPerPage}
                  >
                    <DropdownMenuRadioItem value={"5"}>5</DropdownMenuRadioItem>
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
            <div className={"flex flex-row items-center justify-center gap-2"}>
              <Button
                variant={"outline"}
                size={"icon"}
                disabled={
                  routesData.loading || routesData.previousPageToken === ""
                }
                onClick={() => routesData.previousPage()}
              >
                <ChevronLeft />
              </Button>
              <Button
                variant={"outline"}
                size={"icon"}
                disabled={routesData.loading || routesData.nextPageToken === ""}
                onClick={() => routesData.nextPage()}
              >
                <ChevronRight />
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

const NewRoutePluginSectionSelectPlugin: React.FC<
  NewRoutePluginSectionSelectPluginProps
> = ({
  pluginsData,
  orderField,
  setOrderField,
  orderDirection,
  setOrderDirection,
  setFormState,
  perPage,
  selectedPlugin,
  setSelectedPlugin,
  setPerPage,
  className,
}) => {
  return (
    <div className={cn("flex flex-col gap-5 p-6", className)}>
      <div className={"flex flex-col gap-2"}>
        <p className={"text-2xl"}>2. Select a Plugin</p>
        <p className={"text-md"}>
          This plugin will be attached to the selected route.
        </p>
      </div>
      <div className={"flex flex-col gap-5 lg:flex-row"}>
        <div className={"w-full lg:w-1/3"}>
          <Card>
            <CardHeader>
              <CardTitle>Selected Plugin</CardTitle>
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
        <div className={"w-full gap-5 lg:w-2/3"}>
          {pluginsData.loading ? (
            <div className={"flex items-center justify-center py-20"}>
              <Spinner className={"h-8 w-8"} />
            </div>
          ) : (
            <div className={"flex w-full flex-col gap-5"}>
              <p className={"text-xl"}>Available Plugins</p>
              <PluginsListControls
                orderField={orderField}
                setOrderField={setOrderField}
                orderDirection={orderDirection}
                setOrderDirection={setOrderDirection}
                showCreateButton={true}
                pluginsData={pluginsData}
                className={"justify-between"}
              />
              {pluginsData.data.length === 0 ? (
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant={"icon"}>
                      <ToyBrick />
                    </EmptyMedia>
                    <EmptyTitle>No Plugins</EmptyTitle>
                    <EmptyDescription>
                      You haven&#39;t created any plugins yet. Start with
                      creating one.
                    </EmptyDescription>
                  </EmptyHeader>
                  <EmptyContent className={"flex-row justify-center gap-4"}>
                    <Button size={"sm"} asChild>
                      <a href={"/plugins/new"}>Create Plugin</a>
                    </Button>
                  </EmptyContent>
                </Empty>
              ) : (
                <div className={"overflow-hidden rounded-lg border"}>
                  <Table>
                    <TableHeader className={"sticky top-0 z-10 bg-muted"}>
                      <TableRow>
                        <TableCell></TableCell>
                        <TableHead>Name</TableHead>
                        <TableHead>Filename</TableHead>
                        <TableHead>Version</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {pluginsData.data.map((plugin) => (
                        <TableRow key={plugin.id}>
                          <TableCell>
                            <Checkbox
                              checked={plugin.id === selectedPlugin?.id}
                              onCheckedChange={() => {
                                if (plugin.id === selectedPlugin?.id) {
                                  setSelectedPlugin(null)
                                } else {
                                  setSelectedPlugin(plugin)
                                  setFormState((prev) => ({
                                    ...prev,
                                    plugin_id: plugin.id,
                                  }))
                                }
                              }}
                            />
                          </TableCell>
                          <TableCell>{plugin.name}</TableCell>
                          <TableCell>{plugin.filename}</TableCell>
                          <TableCell>
                            <Badge>{`v${plugin.version}`}</Badge>
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
            <div className={"flex flex-row items-center justify-center gap-2"}>
              <p className={"text-sm font-semibold"}>Rows per page</p>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant={"outline"} disabled={pluginsData.loading}>
                    {perPage}
                    <ChevronDownIcon />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuRadioGroup
                    value={perPage}
                    onValueChange={setPerPage}
                  >
                    <DropdownMenuRadioItem value={"5"}>5</DropdownMenuRadioItem>
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
            <div className={"flex flex-row items-center justify-center gap-2"}>
              <Button
                variant={"outline"}
                size={"icon"}
                disabled={
                  pluginsData.loading || pluginsData.previousPageToken === ""
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
      </div>
    </div>
  )
}

interface NewRoutePluginSectionRoutePluginFormProps {
  formState: Omit<RoutePlugin, "id" | "created_at" | "plugin">
  setFormState: React.Dispatch<
    React.SetStateAction<Omit<RoutePlugin, "id" | "created_at" | "plugin">>
  >
  validationErrors: Record<string, string> | null
  jsonConfigString: string
  setJsonConfigString: React.Dispatch<React.SetStateAction<string>>
  validateJsonConfig: (config: string) => void
  jsonConfigValid: boolean
  setJsonConfigValid: React.Dispatch<React.SetStateAction<boolean>>
  exampleJsonConfig: string
  className?: string
}

const NewRoutePluginSectionPluginForm: React.FC<
  NewRoutePluginSectionRoutePluginFormProps
> = ({
  formState,
  setFormState,
  validationErrors,
  jsonConfigString,
  setJsonConfigString,
  validateJsonConfig,
  jsonConfigValid,
  setJsonConfigValid,
  exampleJsonConfig,
  className,
}) => {
  return (
    <div className={cn("flex flex-col gap-5 p-6", className)}>
      <div className={"flex flex-col gap-2"}>
        <p className={"text-2xl"}>3. Create Route Plugin</p>
        <p className={"text-md"}>
          Fill this information to create a route plugin. It will be attached to
          the selected route and use the selected plugin. You can modify the
          configuration JSON as you want, and your plugin will be able to access
          it through host functions.{" "}
          <a className={"underline"} href={"/docs#plugin-json-config"}>
            Learn more
          </a>{" "}
          about plugin JSON configuration. After creating a new route plugin,
          you will be able to edit this information.
        </p>
      </div>
      <div className={"flex flex-col gap-5 lg:flex-row"}>
        <div className={"w-full lg:w-1/3"}>
          <Card className={"w-full"}>
            <CardHeader>
              <CardTitle>Main Information</CardTitle>
            </CardHeader>
            <CardContent>
              <FieldSet>
                <FieldGroup>
                  <Field>
                    <FieldLabel>Route ID</FieldLabel>
                    <Input
                      disabled
                      aria-label={"route id"}
                      type={"text"}
                      value={formState.route_id}
                    />
                  </Field>
                  <Field>
                    <FieldLabel>Plugin ID</FieldLabel>
                    <Input
                      disabled
                      aria-label={"route id"}
                      type={"text"}
                      value={formState.plugin_id}
                    />
                  </Field>
                </FieldGroup>
                <FieldGroup>
                  <Field>
                    <FieldLabel>
                      Execution Order
                      <span className={"text-destructive"}>*</span>
                    </FieldLabel>
                    <Input
                      aria-invalid={
                        validationErrors != null &&
                        validationErrors["execution_order"] != ""
                      }
                      aria-label={"execution order"}
                      type={"number"}
                      value={formState.execution_order}
                      onChange={(e) =>
                        setFormState((prev) => ({
                          ...prev,
                          execution_order: parseInt(e.target.value) || 1,
                        }))
                      }
                    />
                    <FieldError>
                      {validationErrors != null &&
                        validationErrors["execution_order"] != "" && (
                          <p className={"text-sm"}>
                            {validationErrors["execution_order"]}
                          </p>
                        )}
                    </FieldError>
                  </Field>
                  <Field>
                    <FieldLabel>
                      Version Constraint
                      <span className={"text-destructive"}>*</span>
                    </FieldLabel>
                    <Input
                      aria-invalid={
                        validationErrors != null &&
                        validationErrors["version_constraint"] != ""
                      }
                      aria-label={"version constraint"}
                      type={"text"}
                      value={formState.version_constraint}
                      onChange={(e) =>
                        setFormState((prev) => ({
                          ...prev,
                          version_constraint: e.target.value,
                        }))
                      }
                    />
                    <FieldError>
                      {validationErrors != null &&
                        validationErrors["version_constraint"] != "" && (
                          <p className={"text-sm"}>
                            {validationErrors["version_constraint"]}
                          </p>
                        )}
                    </FieldError>
                  </Field>
                </FieldGroup>
              </FieldSet>
            </CardContent>
          </Card>
        </div>
        <div className={"w-full lg:w-2/3"}>
          <Card className={"w-full"}>
            <CardHeader>
              <CardTitle>JSON Configuration</CardTitle>
              <CardDescription>
                This JSON Configuration is available for the plugin via host
                function.{" "}
                <a className={"underline"} href={"/docs#plugin-json-config"}>
                  Learn more
                </a>{" "}
                about plugin JSON configuration.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Field>
                <FieldLabel>Config</FieldLabel>
                <InputGroup>
                  <InputGroupTextarea
                    aria-invalid={
                      !jsonConfigValid ||
                      (validationErrors != null &&
                        validationErrors["config"] != "")
                    }
                    aria-label={"plugin config"}
                    placeholder={exampleJsonConfig}
                    className={"font-mono text-sm"}
                    value={jsonConfigString}
                    onChange={(e) => {
                      setJsonConfigString(e.target.value)
                      validateJsonConfig(e.target.value)
                    }}
                  />
                  <InputGroupAddon align={"block-start"}>
                    <FileBraces className={"text-muted-foreground"} />
                    <InputGroupText className={"font-mono"}>
                      config.json
                    </InputGroupText>
                    <InputGroupButton
                      className={"ml-auto"}
                      size={"icon-sm"}
                      onClick={() => {
                        setJsonConfigString(exampleJsonConfig)
                        setJsonConfigValid(true)
                      }}
                    >
                      <RotateCcw />
                    </InputGroupButton>
                  </InputGroupAddon>
                </InputGroup>
                <FieldError>
                  {validationErrors != null &&
                    validationErrors["config"] != "" && (
                      <p className={"text-sm"}>{validationErrors["config"]}</p>
                    )}
                </FieldError>
              </Field>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export {
  NewRoutePluginSectionSelectRoute,
  NewRoutePluginSectionSelectPlugin,
  NewRoutePluginSectionPluginForm,
}
