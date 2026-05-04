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
import { useData } from "@/hooks/use-data"
import { Route } from "@/types/route"
import { useRouter, useSearchParams } from "next/navigation"
import React, { useCallback, useState } from "react"
import { Spinner } from "@workspace/ui/components/spinner"
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
} from "@workspace/ui/components/card"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { Button } from "@workspace/ui/components/button"
import { ArrowLeft, CircleAlert, CircleCheck } from "lucide-react"
import { useMutation } from "@/hooks/use-mutation"
import { ErrorResponse } from "@/types/ErrorResponse"
import { AlertModal } from "@/components/dialog/alert-modal"

function EditRoutePageContent() {
  const router = useRouter()
  const params = useSearchParams()
  const path = params.get("path") ?? ""
  const routeData = useData<Route>(
    `/api/routes/${encodeURIComponent(path)}`,
    "route"
  )

  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [showErrorAlert, setShowErrorAlert] = React.useState(false)
  const [validationErrors, setValidationErrors] = useState<Record<
    string,
    string
  > | null>(null)
  const [editableRoute, setEditableRoute] = React.useState<Omit<
    Route,
    "id" | "created_at" | "enabled"
  > | null>(null)

  React.useEffect(() => {
    if (routeData.data && editableRoute === null) {
      setEditableRoute(routeData.data)
    }
  }, [editableRoute, routeData.data])

  const resolveError = useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
    } else {
      setValidationErrors(null)
      setShowErrorAlert(true)
    }
  }, [])

  const submit = React.useCallback(async () => {
    if (!editableRoute) {
      return
    }
    if (!routeData.data) {
      return
    }
    const result = await mutate(
      `/api/routes/${routeData.data.id}`,
      "PATCH",
      JSON.stringify(editableRoute),
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
  }, [editableRoute, mutate, resolveError, routeData.data])

  return (
    <SidebarLayout page_title={"Edit Route"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={showErrorAlert}
        title={"Unexpected error occurred"}
        description={
          error?.details ??
          "No additional details available. Page will be automatically reloaded in 5 seconds."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          setShowErrorAlert(false)
          router.refresh()
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Route edited successfully!"}
        description={
          "The route has been edited successfully. You will be redirected to the route page in 5 seconds."
        }
        icon={<CircleCheck size={15} />}
        onClose={() => {
          setShowSuccess(false)
          router.push(`/routes/route?path=${path}`)
        }}
      />
      <div className={"flex flex-col p-6"}>
        {routeData.loading ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"h-10 w-10"} />
          </div>
        ) : (
          <div className={"flex flex-col gap-5 py-10"}>
            <div className={"flex flex-col gap-5 lg:flex-row"}>
              <Card className={"w-full"}>
                <CardHeader>
                  <CardTitle className={"text-xl font-semibold"}>
                    Main Information
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <FieldSet>
                    <FieldGroup>
                      <Field>
                        <FieldLabel>Path</FieldLabel>
                        <Input
                          aria-label={"path"}
                          type={"text"}
                          value={editableRoute?.path ?? ""}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev ? { ...prev, path: e.target.value } : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["path"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["path"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                      <Field>
                        <FieldLabel>Target URL</FieldLabel>
                        <Input
                          aria-label={"target url"}
                          type={"url"}
                          value={editableRoute?.target_url ?? ""}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? { ...prev, target_url: e.target.value }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["target_url"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["target_url"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                    </FieldGroup>
                    <FieldGroup className={"flex flex-row"}>
                      <Field className={"w-1/2"}>
                        <FieldLabel>Idle connection timeout</FieldLabel>
                        <Input
                          aria-label={"idle connection timeout"}
                          type={"number"}
                          value={editableRoute?.idle_conn_timeout ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    idle_conn_timeout: parseInt(e.target.value),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["idle_conn_timeout"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["idle_conn_timeout"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                      <Field className={"w-1/2"}>
                        <FieldLabel>TLS handshake timeout</FieldLabel>
                        <Input
                          aria-label={"tls handshake timeout"}
                          type={"number"}
                          value={editableRoute?.tls_handshake_timeout ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    tls_handshake_timeout: parseInt(
                                      e.target.value
                                    ),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["tls_handshake_timeout"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["tls_handshake_timeout"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                    </FieldGroup>
                    <FieldGroup className={"flex flex-row"}>
                      <Field className={"w-1/2"}>
                        <FieldLabel>Expect continue timeout</FieldLabel>
                        <Input
                          aria-label={"expect continue timeout"}
                          type={"number"}
                          value={editableRoute?.expect_continue_timeout ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    expect_continue_timeout: parseInt(
                                      e.target.value
                                    ),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["expect_continue_timeout"] !=
                              "" && (
                              <p className={"text-sm"}>
                                {validationErrors["expect_continue_timeout"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                    </FieldGroup>
                  </FieldSet>
                </CardContent>
              </Card>
              <Card className={"w-full"}>
                <CardHeader>
                  <CardTitle className={"text-xl font-semibold"}>
                    Additional Timings Configuration
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <FieldSet>
                    <FieldGroup className={"flex flex-row"}>
                      <Field className={"w-1/2"}>
                        <FieldLabel>Max idle connections</FieldLabel>
                        <Input
                          aria-label={"max idle connections"}
                          type={"number"}
                          value={editableRoute?.max_idle_cons ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    max_idle_cons: parseInt(e.target.value),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["max_idle_cons"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["max_idle_cons"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                      <Field className={"w-1/2"}>
                        <FieldLabel>Max idle connections per host</FieldLabel>
                        <Input
                          aria-label={"max idle connections per host"}
                          type={"number"}
                          value={editableRoute?.max_idle_cons_per_host ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    max_idle_cons_per_host: parseInt(
                                      e.target.value
                                    ),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["max_idle_cons_per_host"] !=
                              "" && (
                              <p className={"text-sm"}>
                                {validationErrors["max_idle_cons_per_host"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                    </FieldGroup>
                    <FieldGroup className={"flex flex-row"}>
                      <Field className={"w-1/2"}>
                        <FieldLabel>Response header timeout</FieldLabel>
                        <Input
                          aria-label={"response header timeout"}
                          type={"number"}
                          value={editableRoute?.response_header_timeout ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    response_header_timeout: parseInt(
                                      e.target.value
                                    ),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["response_header_timeout"] !=
                              "" && (
                              <p className={"text-sm"}>
                                {validationErrors["response_header_timeout"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                      <Field className={"w-1/2"}>
                        <FieldLabel>Max connections per host</FieldLabel>
                        <Input
                          aria-label={"max connections per host"}
                          type={"number"}
                          value={editableRoute?.max_cons_per_host ?? 0}
                          onChange={(e) =>
                            setEditableRoute((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    max_cons_per_host: parseInt(e.target.value),
                                  }
                                : null
                            )
                          }
                        />
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["max_cons_per_host"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["max_cons_per_host"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                    </FieldGroup>
                  </FieldSet>
                </CardContent>
              </Card>
            </div>
            <div
              className={
                "flex flex-row items-center justify-between rounded-xl bg-card p-5"
              }
            >
              <Button
                variant={"outline"}
                aria-label={"back"}
                onClick={() => router.back()}
                disabled={loading}
              >
                <ArrowLeft size={15} />
                Back
              </Button>
              <div
                className={"flex flex-row items-center justify-center gap-2"}
              >
                <Button
                  variant={"outline"}
                  aria-label={"cancel"}
                  disabled={loading}
                  onClick={() => {
                    setEditableRoute(routeData.data)
                    setValidationErrors(null)
                  }}
                >
                  Cancel
                </Button>
                <Button
                  aria-label={"submit"}
                  onClick={submit}
                  disabled={loading || showErrorAlert || !!routeData.error}
                >
                  {loading && <Spinner />}
                  Submit
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}

export default function EditRoutePage() {
  return (
    <React.Suspense
      fallback={
        <SidebarLayout page_title={"Edit Route"}>
          <div className={"flex items-center justify-center p-6"}>
            <Spinner className={"h-10 w-10"} />
          </div>
        </SidebarLayout>
      }
    >
      <EditRoutePageContent />
    </React.Suspense>
  )
}
