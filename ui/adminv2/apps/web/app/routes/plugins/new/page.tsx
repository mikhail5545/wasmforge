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
import { Plugin } from "types/Plugin"
import { Route } from "types/route"
import { RoutePlugin } from "@/types/RoutePlugin"
import React, { useCallback } from "react"
import { ErrorResponse } from "@/types/ErrorResponse"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { AnimatePresence, motion } from "motion/react"
import { usePaginatedData } from "@/hooks/use-paginated-data"
import { Spinner } from "@workspace/ui/components/spinner"
import {
  ArrowLeft,
  ArrowRight,
  FileBraces,
  RotateCcw,
  CircleAlert,
  CircleCheck,
} from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { Progress } from "@workspace/ui/components/progress"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { useMutation } from "@/hooks/use-mutation"
import {
  RoutePluginStep1,
  RoutePluginStep2,
} from "@/components/new-route-plugin-steps"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupText,
  InputGroupTextarea,
} from "@workspace/ui/components/input-group"
import { AlertModal } from "@/components/dialog/alert-modal"

const initialFormState: Omit<RoutePlugin, "id" | "created_at" | "plugin"> = {
  route_id: "",
  plugin_id: "",
  version_constraint: "*",
  execution_order: 1,
  config: null,
}

export default function NewRoutePluginPage() {
  const params = useSearchParams()
  const pluginId = params.get("pluginId")
  const routeId = params.get("routeId")

  const routePath = routeId
    ? `http://localhost:8080/api/routes/${routeId}`
    : null
  const pluginPath = pluginId
    ? `http://localhost:8080/api/plugins/${pluginId}`
    : null

  const routeData = useData<Route>(routePath, "route")
  const pluginData = useData<Plugin>(pluginPath, "plugin")

  const router = useRouter()

  const [routesOrderField, setRoutesOrderField] = React.useState("created_at")
  const [routesOrderDirection, setRoutesOrderDirection] = React.useState("asc")

  const routesData = usePaginatedData<Route>(
    "/api/routes",
    "routes",
    10,
    routesOrderField,
    routesOrderDirection as "asc" | "desc",
    { preload: true }
  )

  const [pluginsOrderField, setPluginsOrderField] = React.useState("created_at")
  const [pluginsOrderDirection, setPluginsOrderDirection] =
    React.useState("asc")

  const pluginsData = usePaginatedData<Plugin>(
    "/api/plugins",
    "plugins",
    10,
    pluginsOrderField,
    pluginsOrderDirection as "asc" | "desc",
    { preload: true }
  )

  const [selectedRoute, setSelectedRoute] = React.useState<Route | null>(null)
  const [selectedPlugin, setSelectedPlugin] = React.useState<Plugin | null>(
    null
  )
  const [formState, setFormState] =
    React.useState<typeof initialFormState>(initialFormState)
  const [currentStep, setCurrentStep] = React.useState<number>(1)
  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [showErrorAlert, setShowErrorAlert] = React.useState(false)
  const [validationErrors, setValidationErrors] = React.useState<Record<
    string,
    string
  > | null>(null)
  const exampleConfigJSON = `{\n  "key": "value"\n}`
  const [jsonConfigString, setJSONConfigString] =
    React.useState<string>(exampleConfigJSON)
  const [jsonConfigValid, setJSONConfigValid] = React.useState(true)

  React.useEffect(() => {
    if (!routeId) {
      return
    }

    if (routeData.error) {
      setShowErrorAlert(true)
      setSelectedRoute(null)
      return
    } else if (!routeData.loading && routeData.data) {
      setSelectedRoute(routeData.data)
      setFormState((prev) => ({ ...prev, route_id: routeData.data!.id }))
    }
  }, [routeData.data, routeData.error, routeData.loading, routeId])

  React.useEffect(() => {
    if (!pluginId) {
      return
    }

    if (pluginData.error) {
      setShowErrorAlert(true)
      setSelectedPlugin(null)
      return
    } else if (!pluginData.loading && pluginData.data) {
      setSelectedPlugin(pluginData.data)
      setFormState((prev) => ({ ...prev, plugin_id: pluginData.data!.id }))
    }
  }, [pluginData.data, pluginData.error, pluginData.loading, pluginId])

  const validateConfig = (config: string) => {
    try {
      JSON.parse(config)
      setJSONConfigValid(true)
      return
    } catch {
      setJSONConfigValid(false)
    }
  }

  const resolveError = useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
    } else {
      setValidationErrors(null)
      setShowErrorAlert(true)
    }
  }, [])

  const submit = React.useCallback(async () => {
    if (formState === initialFormState) return

    if (jsonConfigValid && jsonConfigString !== exampleConfigJSON) {
      setFormState((prev) => ({ ...prev, config: jsonConfigString }))
    }

    const result = await mutate(
      "http://localhost:8080/api/route-plugins",
      "POST",
      JSON.stringify(formState),
      { "Content-Type": "application/json" }
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
  }, [
    exampleConfigJSON,
    formState,
    jsonConfigString,
    jsonConfigValid,
    mutate,
    resolveError,
  ])

  return (
    <SidebarLayout page_title={"New Route Plugin"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={showErrorAlert || !!routesData.error || !!pluginsData.error}
        title={"Unexpected error occurred"}
        description={
          routesData.error?.message ||
          pluginsData.error?.message ||
          routeData.error?.message ||
          pluginData.error?.message ||
          error?.details ||
          "No additional details available. Page will be automatically reloaded in 5 seconds."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          if (routesData.error) void routesData.refetch()
          if (pluginsData.error) void pluginsData.refetch()
          if (routeData.error) void routeData.refetch()
          if (pluginData.error) void pluginData.refetch()
          if (error) reset()
          setShowErrorAlert(false)
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Plugin created successfully!"}
        description={
          "Plugin created successfully! You will be redirected to plugin page in 5 seconds."
        }
        icon={<CircleCheck size={15} />}
        onClose={() => {
          setShowSuccess(false)
          router.push("/routes")
        }}
      />
      <div className={"flex flex-col items-center justify-center p-6"}>
        <div
          className={"flex w-full flex-col items-center justify-center gap-3"}
        >
          <p className={"text-3xl"}>Creating a new Route Plugin</p>
          <Progress
            value={(currentStep * 100) / 3}
            className={"w-full lg:w-1/3"}
          />
        </div>
        <AnimatePresence mode={"wait"}>
          {currentStep === 1 && (
            <RoutePluginStep1
              selectedRoute={selectedRoute}
              setSelectedRoute={setSelectedRoute}
              routesData={routesData}
              orderField={routesOrderField}
              setOrderField={setRoutesOrderField}
              orderDirection={routesOrderDirection}
              setOrderDirection={setRoutesOrderDirection}
            />
          )}
          {currentStep === 2 && (
            <RoutePluginStep2
              selectedPlugin={selectedPlugin}
              setSelectedPlugin={setSelectedPlugin}
              pluginsData={pluginsData}
              orderField={pluginsOrderField}
              setOrderField={setPluginsOrderField}
              orderDirection={pluginsOrderDirection}
              setOrderDirection={setPluginsOrderDirection}
            />
          )}
          {currentStep === 3 && (
            <motion.div
              key={"route-select"}
              initial={{ opacity: 0, x: -100 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 100 }}
              transition={{ type: "spring", stiffness: 300, damping: 30 }}
              className={"flex w-full flex-col gap-5"}
            >
              <p className={"text-xl"}>Fill Information</p>
              <div className={"flex flex-col gap-5 lg:flex-row"}>
                <div className={"w-full lg:w-1/3"}>
                  <Card className={"w-full"}>
                    <CardHeader>
                      <CardTitle>Route Plugin Information</CardTitle>
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
                                  execution_order:
                                    parseInt(e.target.value) || 1,
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
                                validationErrors["version_constraint"] !=
                                  "" && (
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
                        This JSON Configuration is optional. It will be
                        available for your plugin via host function. You can
                        fill it and use this information inside of your WASM
                        plugin.
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
                            placeholder={exampleConfigJSON}
                            className={"font-mono text-sm"}
                            value={jsonConfigString}
                            onChange={(e) => {
                              setJSONConfigString(e.target.value)
                              validateConfig(e.target.value)
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
                                setJSONConfigString(exampleConfigJSON)
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
            </motion.div>
          )}
        </AnimatePresence>
        <div
          className={
            "mt-10 flex w-full flex-row items-center justify-end gap-5"
          }
        >
          <Button
            variant={"outline"}
            onClick={() => setCurrentStep((prev) => Math.max(prev - 1, 1))}
            disabled={currentStep === 1}
          >
            <ArrowLeft />
            Back
          </Button>
          {currentStep <= 2 && (
            <Button
              disabled={
                (currentStep === 1 && !selectedRoute) ||
                (currentStep === 2 && !selectedPlugin)
              }
              onClick={() => {
                setCurrentStep((prev) => prev + 1)
                if (currentStep === 1) {
                  setFormState((prev) => ({
                    ...prev,
                    route_id: selectedRoute!.id,
                  }))
                }
                if (currentStep === 2) {
                  setFormState((prev) => ({
                    ...prev,
                    plugin_id: selectedPlugin!.id,
                  }))
                }
              }}
            >
              Next
              <ArrowRight />
            </Button>
          )}
          {currentStep === 3 && (
            <Button
              variant={"outline"}
              onClick={() => setFormState(initialFormState)}
            >
              Clear
            </Button>
          )}
          {currentStep === 3 && (
            <Button
              disabled={loading || !jsonConfigValid || showErrorAlert}
              onClick={submit}
            >
              {loading && <Spinner />}
              Create
            </Button>
          )}
        </div>
      </div>
    </SidebarLayout>
  )
}
