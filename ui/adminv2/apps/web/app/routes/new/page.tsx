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

'use client';

import { ArrowLeft, CircleAlert, CircleCheck } from "lucide-react"
import { Input } from "@workspace/ui/components/input"
import { Button } from "@workspace/ui/components/button"
import {Route} from "@/types/route"
import { ErrorResponse } from "@/types/ErrorResponse"
import { useCallback, useState } from "react"
import { useMutation } from "@/hooks/use-mutation"
import { Spinner } from "@workspace/ui/components/spinner"
import { useRouter } from "next/navigation"
import { AlertModal } from "@/components/dialog/alert-modal"
import { SidebarLayout } from "@/components/navigation/sidebar"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@workspace/ui/components/field"

const initialRouteFormState: Omit<Route, "id" | "created_at" | "enabled"> = {
  path: '',
  target_url: '',
  idle_conn_timeout: 10,
  tls_handshake_timeout: 10,
  expect_continue_timeout: 10,
}

export default function NewRoutePage() {

  const router = useRouter()

  const [formState, setFormState] = useState<typeof initialRouteFormState>(initialRouteFormState)
  const { loading, error, mutate, reset } = useMutation()
  const [createdPath, setCreatedPath] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<Record<string, string> | null>(null)
  const [showErrorAlert, setShowErrorAlert] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)

  const resolveError = useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
    } else {
      setValidationErrors(null)
      setShowErrorAlert(true)
    }
  }, [])

  const submit = useCallback(async () => {
    if (formState === initialRouteFormState) {
      return
    }

    const result = await mutate(
      'http://localhost:8080/api/routes',
      'POST',
      JSON.stringify(formState),
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
      const created: { "route": Route } = await result.response.json()
      if (created && created["route"]) {
        setCreatedPath(created.route.path)
      }
    }
  }, [formState, mutate, resolveError])

  return (
    <SidebarLayout pageTitle={"Create a new route"}>
      <div className={"flex flex-col p-6"}>
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
          title={"Route created successfully!"}
          description={
            "Now you can enable it, attach plugins, or edit properties." + (createdPath && ' You will be redirected to route page in 5 seconds')
          }
          icon={<CircleCheck size={15} />}
          onClose={() => {
            setShowSuccess(false)
            router.push(
              createdPath ? `/routes/route?path=${createdPath}` : `/routes`
            )
          }}
        />
        <div className={"flex flex-col"}>
          <div className={"flex items-center justify-center"}>
            <div
              className={
                "flex w-full flex-row items-center justify-between rounded-xl p-5"
              }
            ></div>
          </div>
          <div className={"flex flex-col gap-5 py-10"}>
            <div className={"flex flex-col gap-5 lg:flex-row"}>
              <div className={"flex w-full lg:w-1/2"}>
                <div className={"flex w-full rounded-lg bg-card p-5"}>
                  <FieldSet>
                    <FieldLegend>Required Information</FieldLegend>
                    <FieldDescription>
                      This information is necessary to create a new route.
                      Please, fill this fields carefully
                    </FieldDescription>
                    <FieldGroup>
                      <Field>
                        <FieldLabel htmlFor={"path"}>Path</FieldLabel>
                        <Input
                          aria-label={"path"}
                          required={true}
                          autoComplete="off"
                          type={"text"}
                          value={formState.path}
                          placeholder={"/api/example"}
                          onChange={(e) =>
                            setFormState((prev) => ({
                              ...prev,
                              path: e.target.value,
                            }))
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
                        <FieldLabel htmlFor={"target_url"}>
                          Target URL
                        </FieldLabel>
                        <Input
                          aria-label={"target_url"}
                          required={true}
                          type={"url"}
                          autoComplete="off"
                          value={formState.target_url}
                          placeholder={"http://localhost:3000/path"}
                          onChange={(e) =>
                            setFormState((prev) => ({
                              ...prev,
                              target_url: e.target.value,
                            }))
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
                      <div className={"flex flex-row gap-2"}>
                        <Field className={"w-1/2"}>
                          <FieldLabel htmlFor={"idle_conn_timeout"}>
                            Idle connection timeout
                          </FieldLabel>
                          <Input
                            aria-label={"idle connection timeout"}
                            required={true}
                            autoComplete="off"
                            value={formState.idle_conn_timeout}
                            type={"number"}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                idle_conn_timeout: parseInt(e.target.value),
                              }))
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
                          <FieldLabel htmlFor={"tls_handshake_timeout"}>
                            TLS handshake timeout
                          </FieldLabel>
                          <Input
                            aria-label={"tls handshake timeout"}
                            type={"number"}
                            required={true}
                            autoComplete="off"
                            value={formState.tls_handshake_timeout}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                tls_handshake_timeout: parseInt(e.target.value),
                              }))
                            }
                          />
                          <FieldError>
                            {validationErrors != null &&
                              validationErrors["tls_handshake_timeout"] !=
                                "" && (
                                <p className={"text-sm"}>
                                  {validationErrors["tls_handshake_timeout"]}
                                </p>
                              )}
                          </FieldError>
                        </Field>
                      </div>
                      <div className={"flex flex-row gap-2"}>
                        <Field className={"w-1/2"}>
                          <FieldLabel htmlFor={"expect_continue_timeout"}>
                            Expect continue timeout
                          </FieldLabel>
                          <Input
                            aria-label={"expect continue timeout"}
                            type={"number"}
                            required={true}
                            autoComplete="off"
                            value={formState.expect_continue_timeout}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                expect_continue_timeout: parseInt(
                                  e.target.value
                                ),
                              }))
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
                      </div>
                    </FieldGroup>
                  </FieldSet>
                </div>
              </div>
              <div className={"flex w-full lg:w-1/2"}>
                <div className={"flex w-full rounded-lg bg-card p-5"}>
                  <FieldSet>
                    <FieldLegend>Additional Timings Configuration</FieldLegend>
                    <FieldDescription>
                      These fields are optional and have default values. You can
                      adjust them according to your needs, but it&#39;s
                      recommended to keep default values unless you have
                      specific requirements.
                    </FieldDescription>
                    <FieldGroup>
                      <div className={"flex flex-row gap-2"}>
                        <Field className={"w-1/2"}>
                          <FieldLabel htmlFor={"expect_continue_timeout"}>
                            Max idle connections
                          </FieldLabel>
                          <Input
                            aria-label={"max idle connections"}
                            type={"number"}
                            required={false}
                            autoComplete="off"
                            value={formState.max_idle_cons}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                max_idle_cons: parseInt(e.target.value),
                              }))
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
                          <FieldLabel htmlFor={"max_idle_cons_per_host"}>
                            Max idle connections per host
                          </FieldLabel>
                          <Input
                            aria-label={"max idle connections per host"}
                            type={"number"}
                            required={false}
                            autoComplete="off"
                            value={formState.max_idle_cons_per_host}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                max_idle_cons_per_host: parseInt(
                                  e.target.value
                                ),
                              }))
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
                      </div>
                      <div className={"flex flex-row gap-2"}>
                        <Field className={"w-1/2"}>
                          <FieldLabel htmlFor={"response_header_timeout"}>
                            Response header timeout
                          </FieldLabel>
                          <Input
                            aria-label={"response header timeout"}
                            type={"number"}
                            required={false}
                            autoComplete="off"
                            value={formState.response_header_timeout}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                response_header_timeout: parseInt(
                                  e.target.value
                                ),
                              }))
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
                          <FieldLabel htmlFor={"max_cons_per_host"}>
                            Max connections per host
                          </FieldLabel>
                          <Input
                            aria-label={"max idle connections per host"}
                            type={"number"}
                            required={false}
                            autoComplete="off"
                            value={formState.max_cons_per_host}
                            placeholder={"5"}
                            onChange={(e) =>
                              setFormState((prev) => ({
                                ...prev,
                                max_cons_per_host: parseInt(e.target.value),
                              }))
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
                      </div>
                    </FieldGroup>
                  </FieldSet>
                </div>
              </div>
            </div>
            <div
              className={"flex flex-row justify-between rounded-xl bg-card p-5"}
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
              <div className={"flex flex-row gap-2"}>
                <Button
                  onClick={() => {
                    setFormState(initialRouteFormState)
                    reset()
                  }}
                  disabled={loading}
                  variant={"outline"}
                >
                  Cancel
                </Button>
                <Button
                  onClick={submit}
                  disabled={loading || !!error}
                  variant={"default"}
                >
                  {loading && <Spinner />}
                  Create
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </SidebarLayout>
  )
}