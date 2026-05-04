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
import { usePaginatedData } from "@/hooks/use-paginated-data"
import { Spinner } from "@workspace/ui/components/spinner"
import { CircleAlert, CircleCheck } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import { useMutation } from "@/hooks/use-mutation"
import { AlertModal } from "@/components/dialog/alert-modal"
import {
  NewRoutePluginSectionPluginForm,
  NewRoutePluginSectionSelectPlugin,
  NewRoutePluginSectionSelectRoute,
} from "@/components/new-route-plugin-sections"

const initialFormState: Omit<RoutePlugin, "id" | "created_at" | "plugin"> = {
  route_id: "",
  plugin_id: "",
  version_constraint: "*",
  execution_order: 1,
  config: null,
}

function NewRoutePluginPageContent() {
  const params = useSearchParams()
  const pluginId = params.get("pluginId")
  const routeId = params.get("routeId")

  const routePath = routeId
    ? `/api/routes/${routeId}`
    : null
  const pluginPath = pluginId
    ? `/api/plugins/${pluginId}`
    : null

  const routeData = useData<Route>(routePath, "route")
  const pluginData = useData<Plugin>(pluginPath, "plugin")

  const router = useRouter()

  const [routesOrderField, setRoutesOrderField] = React.useState("created_at")
  const [routesOrderDirection, setRoutesOrderDirection] = React.useState("asc")
  const [routesPerPage, setRoutesPerPage] = React.useState("10")

  const routesData = usePaginatedData<Route>(
    "/api/routes",
    "routes",
    parseInt(routesPerPage),
    routesOrderField,
    routesOrderDirection as "asc" | "desc",
    { preload: true }
  )

  const [pluginsOrderField, setPluginsOrderField] = React.useState("created_at")
  const [pluginsOrderDirection, setPluginsOrderDirection] =
    React.useState("asc")
  const [pluginsPerPage, setPluginsPerPage] = React.useState("10")

  const pluginsData = usePaginatedData<Plugin>(
    "/api/plugins",
    "plugins",
    parseInt(pluginsPerPage),
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
      "/api/route-plugins",
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
      <NewRoutePluginSectionSelectRoute
        selectedRoute={selectedRoute}
        setSelectedRoute={setSelectedRoute}
        orderField={routesOrderField}
        setOrderField={setRoutesOrderField}
        orderDirection={routesOrderDirection}
        setOrderDirection={setRoutesOrderDirection}
        routesData={routesData}
        setFormState={setFormState}
        perPage={routesPerPage}
        setPerPage={setRoutesPerPage}
      />
      <NewRoutePluginSectionSelectPlugin
        selectedPlugin={selectedPlugin}
        pluginsData={pluginsData}
        orderField={pluginsOrderField}
        orderDirection={pluginsOrderDirection}
        perPage={pluginsPerPage}
        setOrderField={setPluginsPerPage}
        setOrderDirection={setPluginsOrderDirection}
        setFormState={setFormState}
        setSelectedPlugin={setSelectedPlugin}
        setPerPage={setPluginsPerPage}
      />
      <NewRoutePluginSectionPluginForm
        formState={formState}
        setFormState={setFormState}
        validationErrors={validationErrors}
        jsonConfigString={jsonConfigString}
        setJsonConfigString={setJSONConfigString}
        validateJsonConfig={validateConfig}
        jsonConfigValid={jsonConfigValid}
        setJsonConfigValid={setJSONConfigValid}
        exampleJsonConfig={exampleConfigJSON}
      />
      <div className={"flex flex-row items-center justify-end gap-5 p-6"}>
        <Button
          variant={"outline"}
          disabled={loading}
          onClick={() => {
            setSelectedRoute(null)
            setSelectedPlugin(null)
            setFormState(initialFormState)
            setJSONConfigString(exampleConfigJSON)
          }}
        >
          Reset all
        </Button>
        <Button
          disabled={
            loading ||
            !!error ||
            !!pluginsData.error ||
            !!routesData.error ||
            !selectedPlugin ||
            !selectedRoute
          }
          onClick={submit}
        >
          {loading && <Spinner />}
          Create
        </Button>
      </div>
    </SidebarLayout>
  )
}

export default function NewRoutePluginPage() {
  return (
    <React.Suspense
      fallback={
        <SidebarLayout page_title={"Attach Plugin"}>
          <div className={"flex items-center justify-center p-6"}>
            <Spinner className={"size-10"} />
          </div>
        </SidebarLayout>
      }
    >
      <NewRoutePluginPageContent />
    </React.Suspense>
  )
}
