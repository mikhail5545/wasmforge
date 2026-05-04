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
import { ProxyServerStatus } from "@/types/ProxyServerStatus"
import { Spinner } from "@workspace/ui/components/spinner"
import React, { useCallback } from "react"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
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
import { Input } from "@workspace/ui/components/input"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { Button } from "@workspace/ui/components/button"
import { useMutation } from "@/hooks/use-mutation"
import { ErrorResponse } from "@/types/ErrorResponse"
import { AlertModal } from "@/components/dialog/alert-modal"
import { CircleAlert, CircleCheck } from "lucide-react"
import { useRouter } from "next/navigation"

const initialTLSGenerateFormState: {
  common_name: string
  valid_days: number
  rsa_bits: 2048 | 4096
} = {
  common_name: "example",
  valid_days: 356,
  rsa_bits: 2048,
}

export default function TLSConfigurationPage() {
  const proxyServerStatus = useData<ProxyServerStatus>(
    "/api/proxy/config",
    "status"
  )

  const router = useRouter()
  const [certFile, setCertFile] = React.useState<File | null>(null)
  const [keyFile, setKeyFile] = React.useState<File | null>(null)
  const [tlsGenerateFormState, setTLSGenerateFormState] = React.useState<
    typeof initialTLSGenerateFormState
  >(initialTLSGenerateFormState)

  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")
  const [showAlert, setShowAlert] = React.useState(false)
  const [alertMessage, setAlertMessage] = React.useState("")
  const [validationErrors, setValidationErrors] = React.useState<Record<
    string,
    string
  > | null>(null)

  const resolveError = useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
    } else {
      setValidationErrors(null)
      setShowAlert(true)
    }
  }, [])

  const generateCertificates = React.useCallback(async () => {
    if (tlsGenerateFormState === initialTLSGenerateFormState) return

    const result = await mutate(
      '/api/proxy/certs/generate',
      'POST',
      JSON.stringify(tlsGenerateFormState),
      { 'Content-Type': 'application/json' },
    )

    if (!result.success) {
      resolveError(result.error)
      return
    } else {
      setShowSuccess(true)
      setSuccessMessage("Certificates generated successfully! You will be redirected to proxy server settings page in 5 seconds")
    }

    if (result.response) {
      setValidationErrors(null)
    }
  }, [mutate, resolveError, tlsGenerateFormState])

  const uploadCertificates = React.useCallback(async () => {
    if (!certFile || !keyFile) {
      setAlertMessage("Please select both certificate and key files")
      setShowAlert(true)
      return
    }

    const formData = new FormData()
    formData.append("cert_file", certFile)
    formData.append("key_file", keyFile)

    const result = await mutate(
      '/api/proxy/certs/upload',
      'POST',
      formData
    )

    if (!result.success) {
      resolveError(result.error)
      return
    } else {
      setShowSuccess(true)
      setSuccessMessage(
        "Certificates uploaded successfully! You will be redirected to proxy server settings page in 5 seconds"
      )
    }

    if (result.response) {
      setValidationErrors(null)
    }
  }, [certFile, keyFile, mutate, resolveError])

  return (
    <SidebarLayout page_title={"TLS Configuration"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={showAlert || !!proxyServerStatus.error || !!error}
        title={"Unexpected error occurred"}
        description={
          proxyServerStatus.error?.message ||
          alertMessage ||
          "No additional details available. Trying to refresh in 5 seconds."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          if (proxyServerStatus.error) void proxyServerStatus.refetch()
          if (error) reset()
          setShowAlert(false)
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"TLS Configured successfully!"}
        description={
          successMessage ||
          "All info has been saved! You will be redirected to plugin page in 5 seconds."
        }
        icon={<CircleCheck size={15} />}
        onClose={() => {
          setShowSuccess(false)
          router.push("/settings")
        }}
      />
      <div className={"flex flex-col p-6"}>
        {proxyServerStatus.loading ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-8"} />
          </div>
        ) : (
          <div className={"flex flex-col gap-5"}>
            <div className={"flex flex-col gap-2"}>
              <p className={"text-2xl"}>TLS Setup</p>
              <p className={"text-md"}>
                You can either upload your TLS certificates or generate
                self-signed certificates. You can always reset TLS configuration
                or edit it.
              </p>
            </div>
            <div className={"flex flex-col gap-5 lg:flex-row"}>
              <div className={"w-full lg:w-1/2"}>
                <Card className={"w-full"}>
                  <CardHeader>
                    <CardTitle>Upload Certificates</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <FieldSet>
                      <FieldGroup>
                        <Field>
                          <FieldLabel>
                            TLS Certificate
                            <span className={"text-destructive"}>*</span>
                          </FieldLabel>
                          <Input
                            aria-label={"upload tls certificate"}
                            type={"file"}
                            accept={".crt,.pem"}
                            onChange={(e) =>
                              setCertFile(
                                e.target.files && e.target.files[0]
                                  ? e.target.files[0]
                                  : null
                              )
                            }
                          />
                        </Field>
                        <Field>
                          <FieldLabel>
                            TLS Key
                            <span className={"text-destructive"}>*</span>
                          </FieldLabel>
                          <Input
                            aria-label={"upload tls certificate"}
                            type={"file"}
                            accept={".key,.pem"}
                            onChange={(e) =>
                              setCertFile(
                                e.target.files && e.target.files[0]
                                  ? e.target.files[0]
                                  : null
                              )
                            }
                          />
                        </Field>
                      </FieldGroup>
                    </FieldSet>
                  </CardContent>
                  <CardFooter>
                    <div
                      className={"flex flex-row items-center justify-end gap-2"}
                    >
                      <Button
                        variant={"outline"}
                        onClick={() => {
                          setCertFile(null)
                          setKeyFile(null)
                        }}
                        disabled={!certFile || !keyFile || loading || !!error}
                      >
                        Revert
                      </Button>
                      <Button
                        onClick={uploadCertificates}
                        disabled={!certFile || !keyFile || loading || !!error}
                      >
                        {loading && <Spinner />}
                        Upload
                      </Button>
                    </div>
                  </CardFooter>
                </Card>
              </div>
              <div className={"w-full lg:w-1/2"}>
                <Card className={"w-full"}>
                  <CardHeader>
                    <CardTitle>Generate Certificates</CardTitle>
                    <CardDescription>
                      This setup will generate a self-signed certificates and
                      assign them to Proxy Server configuration
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <FieldSet>
                      <FieldGroup>
                        <Field>
                          <FieldLabel>
                            Common Name (CN)
                            <span className={"text-destructive"}>*</span>
                          </FieldLabel>
                          <Input
                            aria-invalid={
                              validationErrors != null &&
                              validationErrors["common_name"] != ""
                            }
                            aria-label={"common Name"}
                            type={"text"}
                            value={tlsGenerateFormState.common_name}
                            onChange={(e) =>
                              setTLSGenerateFormState((prev) => ({
                                ...prev,
                                common_name: e.target.value,
                              }))
                            }
                          />
                          <FieldError>
                            {validationErrors != null &&
                              validationErrors["common_name"] != "" && (
                                <p className={"text-sm"}>
                                  {validationErrors["common_name"]}
                                </p>
                              )}
                          </FieldError>
                        </Field>
                        <Field>
                          <FieldLabel>
                            Valid Days
                            <span className={"text-destructive"}>*</span>
                          </FieldLabel>
                          <Input
                            aria-invalid={
                              validationErrors != null &&
                              validationErrors["valid_days"] != ""
                            }
                            aria-label={"valid days"}
                            type={"number"}
                            value={tlsGenerateFormState.valid_days}
                            onChange={(e) =>
                              setTLSGenerateFormState((prev) => ({
                                ...prev,
                                valid_days: parseInt(e.target.value),
                              }))
                            }
                          />
                          <FieldError>
                            {validationErrors != null &&
                              validationErrors["valid_days"] != "" && (
                                <p className={"text-sm"}>
                                  {validationErrors["valid_days"]}
                                </p>
                              )}
                          </FieldError>
                        </Field>
                      </FieldGroup>
                      <Field>
                        <FieldLabel>
                          RSA bits<span className={"text-destructive"}>*</span>
                        </FieldLabel>
                        <Select
                          aria-invalid={
                            validationErrors != null &&
                            validationErrors["rsa_bits"] != ""
                          }
                          onValueChange={(value) =>
                            setTLSGenerateFormState((prev) => ({
                              ...prev,
                              rsa_bits: parseInt(value) as 2048 | 4096,
                            }))
                          }
                        >
                          <SelectTrigger className={"w-full max-w-48"}>
                            <SelectValue placeholder={"RSA bits"} />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectGroup>
                              <SelectLabel>RSA bits</SelectLabel>
                              <SelectItem value={"2048"}>2048</SelectItem>
                              <SelectItem value={"4096"}>4096</SelectItem>
                            </SelectGroup>
                          </SelectContent>
                        </Select>
                        <FieldError>
                          {validationErrors != null &&
                            validationErrors["rsa_bits"] != "" && (
                              <p className={"text-sm"}>
                                {validationErrors["rsa_bits"]}
                              </p>
                            )}
                        </FieldError>
                      </Field>
                    </FieldSet>
                  </CardContent>
                  <CardFooter>
                    <div
                      className={"flex flex-row items-center justify-end gap-2"}
                    >
                      <Button
                        variant={"outline"}
                        onClick={() =>
                          setTLSGenerateFormState(initialTLSGenerateFormState)
                        }
                        disabled={
                          tlsGenerateFormState ===
                            initialTLSGenerateFormState ||
                          loading ||
                          !!error
                        }
                      >
                        Revert
                      </Button>
                      <Button
                        onClick={generateCertificates}
                        disabled={
                          tlsGenerateFormState ===
                            initialTLSGenerateFormState ||
                          loading ||
                          !!error
                        }
                      >
                        {loading && <Spinner />}
                        Generate
                      </Button>
                    </div>
                  </CardFooter>
                </Card>
              </div>
            </div>
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}
