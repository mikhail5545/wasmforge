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
import { useData } from "@/hooks/use-data"
import { ProxyServerStatus } from "@/types/ProxyServerStatus"
import { Spinner } from "@workspace/ui/components/spinner"
import React, { useCallback, useState } from "react"
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardDescription,
  CardFooter,
} from "@workspace/ui/components/card"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import { ErrorResponse } from "@/types/ErrorResponse"
import { useMutation } from "@/hooks/use-mutation"
import { Button } from "@workspace/ui/components/button"
import { AlertModal } from "@/components/dialog/alert-modal"
import { CircleAlert, CircleCheck } from "lucide-react"
import { useRouter } from "next/navigation"

type configForm = {
  listen_port: number
  read_header_timeout: number
}

export default function EditProxyServerConfigPage() {
  const proxyServerStatus = useData<ProxyServerStatus>(
    "/api/proxy/config",
    "status"
  )
  const router = useRouter()
  const [configFormState, setConfigFormState] = React.useState<configForm | null>(null)
  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [showErrorAlert, setShowErrorAlert] = React.useState(false)
  const [validationErrors, setValidationErrors] = useState<Record<
    string,
    string
  > | null>(null)

  const resolveError = useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
    } else {
      setValidationErrors(null)
      setShowErrorAlert(true)
    }
  }, [])

  const submit = React.useCallback(async () => {
    if (!configFormState) return

    const response = await mutate(
      '/api/proxy/config',
      'PATCH',
      JSON.stringify(configFormState),
      { 'Content-Type': 'application/json' },
    )

    if (!response.success) {
      resolveError(response.error)
      return
    } else {
      setShowSuccess(true)
    }

    if (response.response) {
      setValidationErrors(null)
    }
  }, [configFormState, mutate, resolveError])

  React.useEffect(() => {
    if (proxyServerStatus.data?.config && configFormState === null) {
      setConfigFormState(proxyServerStatus.data.config)
    }
  }, [configFormState, proxyServerStatus.data])

  return (
    <SidebarLayout page_title={"Edit Proxy Server Configuration"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={showErrorAlert || !!proxyServerStatus.error}
        title={"Unexpected error occurred"}
        description={
          error?.message ||
          proxyServerStatus.error?.message ||
          "No additional details available. Page will be automatically reloaded in 5 seconds."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          if (error) reset()
          if (proxyServerStatus.error) void proxyServerStatus.refetch()
          setShowErrorAlert(false)
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Proxy Server Configuration edited successfully!"}
        description={
          "The configuration was edited successfully. You will be redirected to the dashboard page in 5 seconds."
        }
        icon={<CircleCheck size={15} />}
        onClose={() => {
          setShowSuccess(false)
          router.push("/")
        }}
      />
      <div className={"flex flex-col p-6"}>
        {proxyServerStatus.loading ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-8"} />
          </div>
        ) : (
          <div
            className={"flex flex-col items-center justify-center gap-5 py-10"}
          >
            <Card className={"max-w-lg min-w-md"}>
              <CardHeader>
                <CardTitle>Edit Proxy Server Configuration</CardTitle>
                <CardDescription>
                  After editing the configuration, make sure to restart the
                  proxy server for the changes to take effect.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <FieldSet>
                  <FieldGroup>
                    <Field>
                      <FieldLabel>Listen Port</FieldLabel>
                      <Input
                        aria-invalid={
                          validationErrors != null &&
                          validationErrors["listen_port"] != ""
                        }
                        aria-label={"listen port"}
                        type={"number"}
                        value={configFormState?.listen_port || ""}
                        onChange={(e) =>
                          setConfigFormState((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  listen_port: parseInt(e.target.value),
                                }
                              : null
                          )
                        }
                      />
                      <FieldError>
                        {validationErrors != null &&
                          validationErrors["listen_port"] != "" && (
                            <p className={"text-sm"}>
                              {validationErrors["listen_port"]}
                            </p>
                          )}
                      </FieldError>
                    </Field>
                    <Field>
                      <FieldLabel>Read Header Timeout</FieldLabel>
                      <Input
                        aria-invalid={
                          validationErrors != null &&
                          validationErrors["read_header_timeout"] != ""
                        }
                        aria-label={"read header timeout"}
                        type={"number"}
                        value={configFormState?.read_header_timeout || ""}
                        onChange={(e) =>
                          setConfigFormState((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  read_header_timeout: parseInt(e.target.value),
                                }
                              : null
                          )
                        }
                      />
                      <FieldError>
                        {validationErrors != null &&
                          validationErrors["read_header_timeout"] != "" && (
                            <p className={"text-sm"}>
                              {validationErrors["read_header_timeout"]}
                            </p>
                          )}
                      </FieldError>
                    </Field>
                  </FieldGroup>
                </FieldSet>
              </CardContent>
              <CardFooter>
                <div className={"flex flex-row items-center justify-end gap-5"}>
                  <Button
                    variant={"outline"}
                    disabled={loading || !!proxyServerStatus.error}
                    onClick={() =>
                      setConfigFormState(proxyServerStatus.data?.config || null)
                    }
                  >
                    Revert
                  </Button>
                  <Button
                    disabled={loading || !!error || !!proxyServerStatus.error}
                    onClick={submit}
                  >
                    {loading && <Spinner />}
                    Submit
                  </Button>
                </div>
              </CardFooter>
            </Card>
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}