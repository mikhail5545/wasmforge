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

import React from "react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  ArrowLeft,
  CircleAlert,
  CircleCheck,
  Plus,
  ShieldCheck,
  Trash2,
} from "lucide-react"

import { AlertModal } from "@/components/dialog/alert-modal"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useData } from "@/hooks/use-data"
import { useMutation } from "@/hooks/use-mutation"
import { ErrorResponse } from "@/types/ErrorResponse"
import { HttpMethod, Route, RouteMethod, RouteMethodSpec } from "@/types/route"
import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Checkbox } from "@workspace/ui/components/checkbox"
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { Separator } from "@workspace/ui/components/separator"
import { Spinner } from "@workspace/ui/components/spinner"
import { Switch } from "@workspace/ui/components/switch"
import { Textarea } from "@workspace/ui/components/textarea"

const httpMethods: HttpMethod[] = [
  "GET",
  "POST",
  "PUT",
  "DELETE",
  "PATCH",
  "HEAD",
  "OPTIONS",
  "TRACE",
  "CONNECT",
]

type EditableRouteMethod = {
  method: HttpMethod
  enabled: boolean
  max_request_payload_bytes: string
  request_timeout_ms: string
  response_timeout_ms: string
  rate_limit_per_minute: string
  require_authentication: boolean
  allowed_auth_schemes: string
  metadata: string
}

function emptyMethod(method: HttpMethod, enabled = false): EditableRouteMethod {
  return {
    method,
    enabled,
    max_request_payload_bytes: "",
    request_timeout_ms: "",
    response_timeout_ms: "",
    rate_limit_per_minute: "",
    require_authentication: false,
    allowed_auth_schemes: "",
    metadata: "",
  }
}

function parseJSON<T>(value: string | undefined, fallback: T): T {
  if (!value?.trim()) {
    return fallback
  }
  try {
    return JSON.parse(value) as T
  } catch {
    return fallback
  }
}

function formatJSON(value: string | undefined): string {
  const parsed = parseJSON<unknown>(value, null)
  return parsed === null ? "" : JSON.stringify(parsed, null, 2)
}

function fromRouteMethod(method: RouteMethod): EditableRouteMethod {
  return {
    method: method.method,
    enabled: true,
    max_request_payload_bytes: method.max_request_payload_bytes?.toString() ?? "",
    request_timeout_ms: method.request_timeout_ms?.toString() ?? "",
    response_timeout_ms: method.response_timeout_ms?.toString() ?? "",
    rate_limit_per_minute: method.rate_limit_per_minute?.toString() ?? "",
    require_authentication: method.require_authentication,
    allowed_auth_schemes: parseJSON<string[]>(
      method.allowed_auth_schemes,
      []
    ).join(", "),
    metadata: formatJSON(method.metadata),
  }
}

function optionalNumber(value: string): number | undefined {
  if (value.trim() === "") {
    return undefined
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function parseCSV(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
}

function buildPayload(methods: EditableRouteMethod[]): RouteMethodSpec[] {
  return methods
    .filter((method) => method.enabled)
    .map((method) => {
      const metadata = parseJSON<Record<string, unknown>>(method.metadata, {})
      return {
        method: method.method,
        max_request_payload_bytes: optionalNumber(
          method.max_request_payload_bytes
        ),
        request_timeout_ms: optionalNumber(method.request_timeout_ms),
        response_timeout_ms: optionalNumber(method.response_timeout_ms),
        rate_limit_per_minute: optionalNumber(method.rate_limit_per_minute),
        require_authentication: method.require_authentication,
        allowed_auth_schemes: parseCSV(method.allowed_auth_schemes),
        metadata: Object.keys(metadata).length > 0 ? metadata : undefined,
      }
    })
}

function EditRouteMethodsPageContent() {
  const router = useRouter()
  const params = useSearchParams()
  const path = params.get("path") ?? ""
  const routeData = useData<Route>(
    path ? `http://localhost:8080/api/routes/${encodeURIComponent(path)}` : null,
    "route"
  )
  const methodsData = useData<RouteMethod[]>(
    routeData.data
      ? `http://localhost:8080/api/routes/${routeData.data.id}/methods`
      : null,
    "methods"
  )
  const { loading, error, mutate, reset } = useMutation()
  const [methods, setMethods] = React.useState<EditableRouteMethod[]>([])
  const [validationErrors, setValidationErrors] = React.useState<Record<
    string,
    string
  > | null>(null)
  const [showErrorAlert, setShowErrorAlert] = React.useState(false)
  const [showSuccess, setShowSuccess] = React.useState(false)

  React.useEffect(() => {
    if (!routeData.data || !methodsData.data) {
      return
    }
    const existing = new Map(
      methodsData.data.map((method) => [method.method, fromRouteMethod(method)])
    )
    const allowed = new Set(routeData.data.allowed_methods ?? [])
    setMethods(
      httpMethods.map((method) => {
        const existingMethod = existing.get(method)
        if (existingMethod) {
          return existingMethod
        }
        return emptyMethod(method, allowed.size === 0 || allowed.has(method))
      })
    )
  }, [methodsData.data, routeData.data])

  const resolveError = React.useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
      return
    }
    setValidationErrors(null)
    setShowErrorAlert(true)
  }, [])

  const updateMethod = React.useCallback(
    (method: HttpMethod, patch: Partial<EditableRouteMethod>) => {
      setMethods((prev) =>
        prev.map((item) =>
          item.method === method ? { ...item, ...patch } : item
        )
      )
    },
    []
  )

  const submit = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }
    const invalidMetadata = methods.find((method) => {
      if (!method.enabled || !method.metadata.trim()) {
        return false
      }
      try {
        JSON.parse(method.metadata)
        return false
      } catch {
        return true
      }
    })
    if (invalidMetadata) {
      setValidationErrors({
        metadata: `${invalidMetadata.method} metadata must be valid JSON object`,
      })
      return
    }

    const result = await mutate(
      `http://localhost:8080/api/routes/${routeData.data.id}/methods/`,
      "POST",
      { methods: buildPayload(methods) }
    )
    if (!result.success) {
      resolveError(result.error)
      return
    }
    setValidationErrors(null)
    setShowSuccess(true)
    await methodsData.refetch()
    await routeData.refetch()
  }, [methods, methodsData, mutate, resolveError, routeData])

  const visibleError =
    routeData.error?.message ||
    methodsData.error?.message ||
    validationErrors?.metadata ||
    error?.message

  return (
    <SidebarLayout page_title={"Edit Route Methods"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={showErrorAlert || !!routeData.error || !!methodsData.error}
        title={"Unable to save route methods"}
        description={visibleError ?? "No additional details available."}
        icon={<CircleAlert size={15} />}
        onClose={() => {
          setShowErrorAlert(false)
          reset()
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Route methods updated"}
        description={"The proxy will use these method policies on this route."}
        icon={<CircleCheck size={15} />}
        onClose={() => setShowSuccess(false)}
      />
      <div className={"flex flex-col gap-5 p-6"}>
        {routeData.loading || methodsData.loading ? (
          <div className={"flex items-center justify-center py-40"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <>
            <Card>
              <CardHeader className={"flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"}>
                <div>
                  <CardTitle>Method Runtime Policy</CardTitle>
                  <CardDescription>
                    Configure allowed methods, payload limits, rate limits,
                    auth scheme gates, request timeout, response timeout, and
                    method metadata for{" "}
                    <span className={"font-mono"}>{routeData.data?.path}</span>.
                  </CardDescription>
                </div>
                <div className={"flex flex-wrap gap-2"}>
                  <Button
                    variant={"outline"}
                    onClick={() =>
                      setMethods((prev) =>
                        prev.map((method) => ({ ...method, enabled: true }))
                      )
                    }
                  >
                    <Plus size={15} />
                    Enable all
                  </Button>
                  <Button
                    variant={"outline"}
                    onClick={() =>
                      setMethods((prev) =>
                        prev.map((method) => ({ ...method, enabled: false }))
                      )
                    }
                  >
                    <Trash2 size={15} />
                    Disable all
                  </Button>
                </div>
              </CardHeader>
            </Card>
            <div className={"grid gap-5 xl:grid-cols-2"}>
              {methods.map((method) => (
                <Card
                  key={method.method}
                  className={method.enabled ? "" : "opacity-70"}
                >
                  <CardHeader>
                    <div className={"flex items-center justify-between gap-4"}>
                      <div className={"flex items-center gap-3"}>
                        <Badge variant={method.enabled ? "default" : "outline"}>
                          {method.method}
                        </Badge>
                        <CardTitle className={"text-lg"}>
                          {method.enabled ? "Enabled" : "Disabled"}
                        </CardTitle>
                      </div>
                      <Switch
                        checked={method.enabled}
                        onCheckedChange={(checked) =>
                          updateMethod(method.method, { enabled: checked })
                        }
                      />
                    </div>
                    <CardDescription>
                      Disabled methods are omitted from the route method set and
                      rejected by the proxy.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className={"flex flex-col gap-5"}>
                    <FieldSet disabled={!method.enabled}>
                      <FieldGroup className={"grid gap-4 md:grid-cols-2"}>
                        <Field>
                          <FieldLabel>Max payload bytes</FieldLabel>
                          <Input
                            type={"number"}
                            min={0}
                            value={method.max_request_payload_bytes}
                            placeholder={"1048576"}
                            onChange={(event) =>
                              updateMethod(method.method, {
                                max_request_payload_bytes: event.target.value,
                              })
                            }
                          />
                        </Field>
                        <Field>
                          <FieldLabel>Rate limit / minute</FieldLabel>
                          <Input
                            type={"number"}
                            min={0}
                            value={method.rate_limit_per_minute}
                            placeholder={"120"}
                            onChange={(event) =>
                              updateMethod(method.method, {
                                rate_limit_per_minute: event.target.value,
                              })
                            }
                          />
                        </Field>
                        <Field>
                          <FieldLabel>Request timeout ms</FieldLabel>
                          <Input
                            type={"number"}
                            min={0}
                            value={method.request_timeout_ms}
                            placeholder={"2000"}
                            onChange={(event) =>
                              updateMethod(method.method, {
                                request_timeout_ms: event.target.value,
                              })
                            }
                          />
                        </Field>
                        <Field>
                          <FieldLabel>Response timeout ms</FieldLabel>
                          <Input
                            type={"number"}
                            min={0}
                            value={method.response_timeout_ms}
                            placeholder={"5000"}
                            onChange={(event) =>
                              updateMethod(method.method, {
                                response_timeout_ms: event.target.value,
                              })
                            }
                          />
                        </Field>
                      </FieldGroup>
                      <Separator className={"my-5"} />
                      <Field orientation={"horizontal"}>
                        <Checkbox
                          checked={method.require_authentication}
                          onCheckedChange={(checked) =>
                            updateMethod(method.method, {
                              require_authentication: checked === true,
                            })
                          }
                        />
                        <FieldContent>
                          <FieldLabel className={"flex items-center gap-2"}>
                            <ShieldCheck size={15} />
                            Require Authorization header
                          </FieldLabel>
                          <FieldDescription>
                            The proxy rejects this method with 401 before WASM
                            plugins if the header is missing or uses a blocked
                            scheme.
                          </FieldDescription>
                        </FieldContent>
                      </Field>
                      <Field className={"mt-4"}>
                        <FieldLabel>Allowed auth schemes</FieldLabel>
                        <Input
                          value={method.allowed_auth_schemes}
                          placeholder={"Bearer, Basic"}
                          onChange={(event) =>
                            updateMethod(method.method, {
                              allowed_auth_schemes: event.target.value,
                            })
                          }
                        />
                        <FieldDescription>
                          Comma-separated. Leave empty to allow any scheme when
                          authentication is required.
                        </FieldDescription>
                      </Field>
                      <Field className={"mt-4"}>
                        <FieldLabel>Method metadata JSON</FieldLabel>
                        <Textarea
                          className={"min-h-24 font-mono text-xs"}
                          value={method.metadata}
                          placeholder={'{"tier":"partner","audit":true}'}
                          onChange={(event) =>
                            updateMethod(method.method, {
                              metadata: event.target.value,
                            })
                          }
                        />
                        <FieldDescription>
                          Parsed once and attached to the request context for
                          downstream middleware/plugins.
                        </FieldDescription>
                        <FieldError>
                          {validationErrors?.metadata?.includes(method.method)
                            ? validationErrors.metadata
                            : ""}
                        </FieldError>
                      </Field>
                    </FieldSet>
                  </CardContent>
                </Card>
              ))}
            </div>
            <div className={"flex items-center justify-between rounded-xl bg-card p-5"}>
              <Button
                variant={"outline"}
                onClick={() => router.back()}
                disabled={loading}
              >
                <ArrowLeft size={15} />
                Back
              </Button>
              <Button
                onClick={submit}
                disabled={loading || !!routeData.error || !!methodsData.error}
              >
                {loading && <Spinner />}
                Save Method Policies
              </Button>
            </div>
          </>
        )}
      </div>
    </SidebarLayout>
  )
}

export default function EditRouteMethodsPage() {
  return (
    <React.Suspense
      fallback={
        <SidebarLayout page_title={"Edit Route Methods"}>
          <div className={"flex items-center justify-center p-6"}>
            <Spinner className={"h-10 w-10"} />
          </div>
        </SidebarLayout>
      }
    >
      <EditRouteMethodsPageContent />
    </React.Suspense>
  )
}
