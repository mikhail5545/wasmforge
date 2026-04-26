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

'use client'

import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useRouter, useSearchParams } from "next/navigation"
import { useData } from "@/hooks/use-data"
import { RoutePlugin } from "@/types/RoutePlugin"
import { Plugin } from "@/types/Plugin"
import React, { useCallback } from "react"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupText,
  InputGroupTextarea,
} from "@workspace/ui/components/input-group"
import { Progress } from "@workspace/ui/components/progress"
import { AnimatePresence, motion } from "motion/react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { usePaginatedData } from "@/hooks/use-paginated-data"
import { Spinner } from "@workspace/ui/components/spinner"
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@workspace/ui/components/empty"
import {
  ChevronDownIcon,
  ChevronLeft,
  ChevronRight, CircleAlert, CircleCheck,
  FileBraces, RotateCcw,
  ToyBrick,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { Input } from "@workspace/ui/components/input"
import { Checkbox } from "@workspace/ui/components/checkbox"
import {
  DropdownMenu,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import { DropdownMenuContent } from "@radix-ui/react-dropdown-menu"
import { PluginsListControls } from "@/components/plugins-list-controls"
import { useMutation } from "@/hooks/use-mutation"
import { ErrorResponse } from "@/types/ErrorResponse"
import { AlertModal } from "@/components/dialog/alert-modal"

export default function EditRoutePluginPage() {
  const params = useSearchParams()
  const routePluginId = params.get("pluginId") ?? ""
  const routePluginData = useData<RoutePlugin>(
    `http://localhost:8080/api/route-plugins/${routePluginId}`,
    'route_plugin'
  )
  const router = useRouter()

  const [orderField, setOrderField] = React.useState('created_at')
  const [orderDirection, setOrderDirection] = React.useState("asc")
  const [perPage, setPerPage] = React.useState('10')
  const pluginsData = usePaginatedData<Plugin>(
    '/api/plugins',
    'plugins',
    Number(perPage),
    orderField,
    orderDirection as 'asc' | 'desc',
    { preload: true }
  )

  const [editableRoutePlugin, setEditableRoutePlugin] = React.useState<Omit<RoutePlugin, "id" | "route_id" | "resolved_plugin_version" | "created_at" | "plugin"> | null>(null)

  React.useEffect(() => {
    if (routePluginData.data && editableRoutePlugin === null) {
      setEditableRoutePlugin(routePluginData.data)
    }
  }, [editableRoutePlugin, routePluginData.data])


  const pluginPath = routePluginData.data?.plugin_id ? `http://localhost:8080/api/plugins/${routePluginData.data.plugin_id}` : null
  const pluginData = useData<Plugin>(pluginPath, "plugin")
  const [selectedPlugin, setSelectedPlugin] = React.useState<Plugin | null>(
    pluginData.data
  )
  const [showAlert, setShowAlert] = React.useState(false)
  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [validationErrors, setValidationErrors] = React.useState<Record<
    string,
    string
  > | null>(null)
  const [jsonConfigString, setJSONConfigString] =
    React.useState<string>(editableRoutePlugin?.config ?? "")
  const [jsonConfigValid, setJSONConfigValid] = React.useState(true)

  const validateConfig = (config: string) => {
    try {
      JSON.parse(config)
      setJSONConfigValid(true)
      return
    } catch {
      setJSONConfigValid(false)
    }
  }

  React.useEffect(() => {
    if (!routePluginData.data) return

    if (pluginData.error) {
      setShowAlert(true)
      setSelectedPlugin(null)
      return
    } else if (!pluginData.loading && routePluginData.data) {
      setSelectedPlugin(pluginData.data)
    }
  }, [pluginData.data, pluginData.error, pluginData.loading, routePluginData.data])

  const resolveError = useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
    } else {
      setValidationErrors(null)
      setShowAlert(true)
    }
  }, [])

  const submit = React.useCallback(async () => {
    if (!editableRoutePlugin || !routePluginData.data) return

    if (jsonConfigValid && jsonConfigString !== '' && jsonConfigString !== routePluginData.data.config) {
      setEditableRoutePlugin(prev => prev ? { ...prev, config: jsonConfigString } : prev)
    }

    const result = await mutate(
      `http://localhost:8080/api/route-plugins/${routePluginData.data.plugin_id}`,
      'PATCH',
      JSON.stringify(editableRoutePlugin),
      { 'Content-Type': 'application/json' },
    )

    if (!result.success) {
      resolveError(result.error)
      return
    } else {
      setShowSuccess(true)
    }

    if (result.response) {
      setValidationErrors(null)
    }
  }, [editableRoutePlugin, jsonConfigString, jsonConfigValid, mutate, resolveError, routePluginData.data])

  return (
    <SidebarLayout page_title={"Edit Route Plugin"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={
          showAlert ||
          !!pluginsData.error ||
          !!routePluginData.error ||
          !!pluginData.error
        }
        title={"Unexpected error occurred"}
        description={
          pluginsData.error?.details ||
          routePluginData.error?.details ||
          pluginData.error?.details ||
          error?.details ||
          "No additional details available. Trying to refresh in 5 seconds."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          if (routePluginData.error) void routePluginData.refetch()
          if (pluginsData.error) void pluginsData.refetch()
          if (pluginData.error) void pluginData.refetch()
          if (error) reset()
          setShowAlert(false)
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Plugin edited successfully!"}
        description={
          "All info has been saved! You will be redirected to plugin page in 5 seconds."
        }
        icon={<CircleCheck size={15} />}
        onClose={() => {
          setShowSuccess(false)
          router.push(`/routes/plugins/plugin?pluginId=${routePluginData.data?.id}`)
        }}
      />
      <div className={"flex flex-col gap-5 p-6"}>
        <div className={"flex flex-col gap-2"}>
          <p className={"text-2xl"}>1. Selected Plugin</p>
          <p className={"text-md"}>
            If you want to change currently attached plugin, you can simply
            chose new one. This action will detach currently attached plugin and
            it&#39;s configuration from the route plugin.
          </p>
        </div>
        <div className={"flex flex-col gap-5 lg:flex-row"}>
          <div className={"w-full lg:w-1/3"}>
            <Card>
              <CardHeader>
                <CardTitle>
                  {selectedPlugin?.id === routePluginData.data?.plugin_id
                    ? "Currently attached plugin"
                    : "This Plugin will be attached"}
                </CardTitle>
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
                    className={
                      "flex flex-col items-center justify-center py-20"
                    }
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
                                    setEditableRoutePlugin((prev) =>
                                      prev
                                        ? { ...prev, plugin_id: plugin.id }
                                        : prev
                                    )
                                  }
                                }}
                              />
                            </TableCell>
                            <TableCell>{plugin.name}</TableCell>
                            <TableCell>{plugin.filename}</TableCell>
                            <TableCell>{plugin.version}</TableCell>
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
                className={"flex flex-row items-center justify-center gap-2"}
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
                className={"flex flex-row items-center justify-center gap-2"}
              >
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
        <div className={"flex flex-col gap-2"}>
          <p className={"text-2xl"}>2. Edit Route Plugin</p>
          <p className={"text-md"}>
            You can change any information listed here safely. If route on which
            this plugin is attached to receives a request while you are editing
            this plugin, old configuration will be used on that request. Once
            you confirm changes, route&#39;s middleware will be reassembled and
            hot-swapped.
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
                      <FieldLabel>Execution Order</FieldLabel>
                      <Input
                        aria-invalid={
                          validationErrors != null &&
                          validationErrors["execution_order"] != ""
                        }
                        aria-label={"execution order"}
                        type={"number"}
                        value={editableRoutePlugin?.execution_order ?? ""}
                        onChange={(e) =>
                          setEditableRoutePlugin((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  execution_order: parseInt(e.target.value),
                                }
                              : null
                          )
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
                      <FieldLabel>Version Constraint</FieldLabel>
                      <Input
                        aria-invalid={
                          validationErrors != null &&
                          validationErrors["version_constraint"] != ""
                        }
                        aria-label={"version constraint"}
                        type={"text"}
                        value={editableRoutePlugin?.version_constraint ?? ""}
                        onChange={(e) =>
                          setEditableRoutePlugin((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  version_constraint: e.target.value,
                                }
                              : null
                          )
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
                      value={jsonConfigString}
                      aria-invalid={
                        !jsonConfigValid ||
                        (validationErrors != null &&
                          validationErrors["config"] != "")
                      }
                      placeholder={
                        "You did not specify config for this plugin yet."
                      }
                      className={"font-mono text-sm"}
                      onChange={(e) => {
                        validateConfig(e.target.value)
                        setJSONConfigString(e.target.value)
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
                          setJSONConfigString(editableRoutePlugin?.config ?? "")
                          setJSONConfigValid(true)
                        }}
                      >
                        <RotateCcw />
                      </InputGroupButton>
                    </InputGroupAddon>
                  </InputGroup>
                  <FieldError>
                    {validationErrors != null &&
                      validationErrors["config"] != "" && (
                        <p className={"text-sm"}>
                          {validationErrors["config"]}
                        </p>
                      )}
                  </FieldError>
                </Field>
              </CardContent>
            </Card>
          </div>
        </div>
        <div className={"flex flex-row items-center justify-end gap-4"}>
          <Button
            variant={"outline"}
            disabled={loading}
            onClick={() => {
              setEditableRoutePlugin(routePluginData.data)
              setJSONConfigString(routePluginData.data?.config ?? "")
              setJSONConfigValid(true)
            }}
          >
            Revert
          </Button>
          <Button
            onClick={submit}
            disabled={
              loading ||
              !selectedPlugin ||
              !editableRoutePlugin ||
              !!error ||
              !!pluginsData.error ||
              !!routePluginData.error
            }
          >
            {loading && <Spinner />}
            Submit
          </Button>
        </div>
      </div>
    </SidebarLayout>
  )
}