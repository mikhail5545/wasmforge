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
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardDescription,
  CardFooter,
} from "@workspace/ui/components/card"
import {
  Field, FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {Plugin} from "types/Plugin"
import { useRouter } from "next/navigation"
import React, { useCallback } from "react"
import { useMutation } from "@/hooks/use-mutation"
import { ErrorResponse } from "@/types/ErrorResponse"
import { AlertModal } from "@/components/dialog/alert-modal"
import { ArrowLeft, CircleAlert, CircleCheck } from "lucide-react"
import {Button} from "@workspace/ui/components/button"
import { Spinner } from "@workspace/ui/components/spinner"

const initialFormState: Omit<Plugin, "id" | "created_at" | "checksum"> = {
  name: "bearer_checker",
  filename: "bearer_checker.wasm",
  version: "0.0.1",
}

export default function NewPluginPage(){
  const router = useRouter()

  const [formState, setFormState] = React.useState(initialFormState)
  const [file, setFile] = React.useState<File | null>(null)

  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [showErrorAlert, setShowErrorAlert] = React.useState(false)
  const [validationErrors, setValidationErrors] = React.useState<Record<
    string,
    string
  > | null>(null)

  const createdPluginPage = `/plugins/plugin?name=${encodeURIComponent(formState.name)}&version=${encodeURIComponent(formState.version)}`

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
    if (!file) return

    const formData = new FormData()
    formData.append("wasm_file", file)
    formData.append("metadata", JSON.stringify(formState))

    const result = await mutate(
      'http://localhost:8080/api/plugins',
      'POST',
      formData,
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
  }, [file, formState, mutate, resolveError])

  return (
    <SidebarLayout page_title={"Create a new Plugin"}>
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
          reset()
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
          router.push(createdPluginPage)
        }}
      />
      <div className={"flex flex-col p-6"}>
        <div
          className={"flex flex-col items-center justify-center gap-5 py-10"}
        >
          <Card className={"max-w-lg min-w-md"}>
            <CardHeader>
              <CardTitle className={"text-xl"}>Create a new Plugin</CardTitle>
              <CardDescription>
                You need to create a plugin with and filename. Filename will be
                used to rewrite the name of the original plugin file.
                Combination of name and version must be unique.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <FieldSet>
                <FieldGroup>
                  <Field>
                    <FieldLabel>
                      Name<span className={"text-destructive"}>*</span>
                    </FieldLabel>
                    <Input
                      required
                      aria-invalid={
                        validationErrors != null &&
                        validationErrors["name"] != ""
                      }
                      aria-label={"plugin name"}
                      value={formState.name}
                      type={"text"}
                      onChange={(e) =>
                        setFormState((prev) => ({
                          ...prev,
                          name: e.target.value,
                        }))
                      }
                    />
                    <FieldError>
                      {validationErrors != null &&
                        validationErrors["name"] != "" && (
                          <p className={"text-sm"}>
                            {validationErrors["name"]}
                          </p>
                        )}
                    </FieldError>
                  </Field>
                  <Field>
                    <FieldLabel>
                      Filename<span className={"text-destructive"}>*</span>
                    </FieldLabel>
                    <Input
                      required
                      aria-invalid={
                        validationErrors != null &&
                        validationErrors["filename"] != ""
                      }
                      aria-label={"plugin filename"}
                      value={formState.filename}
                      type={"text"}
                      onChange={(e) =>
                        setFormState((prev) => ({
                          ...prev,
                          filename: e.target.value,
                        }))
                      }
                    />
                    <FieldError>
                      {validationErrors != null &&
                        validationErrors["filename"] != "" && (
                          <p className={"text-sm"}>
                            {validationErrors["filename"]}
                          </p>
                        )}
                    </FieldError>
                  </Field>
                </FieldGroup>
                <Field>
                  <FieldLabel>
                    WASM File<span className={"text-destructive"}>*</span>
                  </FieldLabel>
                  <Input
                    required
                    aria-label={"plugin wasm file"}
                    type={"file"}
                    accept={".wasm"}
                    onChange={(e) =>
                      setFile(
                        e.target.files && e.target.files[0]
                          ? e.target.files[0]
                          : null
                      )
                    }
                  />
                  <FieldDescription>
                    Select a plugin file (.wasm)
                  </FieldDescription>
                </Field>
              </FieldSet>
            </CardContent>
            <CardFooter className={"flex flex-row justify-between"}>
              <Button
                variant={"outline"}
                disabled={loading}
                onClick={() => router.back()}
              >
                <ArrowLeft />
                Back
              </Button>
              <div className={"flex flex-row gap-3"}>
                <Button
                  variant={"secondary"}
                  onClick={() => setFormState(initialFormState)}
                >
                  Cancel
                </Button>
                <Button disabled={loading || !!error} onClick={submit}>
                  {loading && <Spinner />}
                  Create
                </Button>
              </div>
            </CardFooter>
          </Card>
        </div>
      </div>
    </SidebarLayout>
  )
}