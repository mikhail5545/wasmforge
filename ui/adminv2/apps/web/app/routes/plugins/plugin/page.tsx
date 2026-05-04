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
import { Spinner } from "@workspace/ui/components/spinner"
import React from "react"
import { Button } from "@workspace/ui/components/button"
import { ArrowLeft, FileBraces } from "lucide-react"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupText,
  InputGroupTextarea,
} from "@workspace/ui/components/input-group"
import {
  Field,
  FieldLabel,
} from "@workspace/ui/components/field"
import { RoutePluginCard} from "@/components/ui/route-plugin-card"
import { useMutation } from "@/hooks/use-mutation"
import { AlertModal } from "@/components/dialog/alert-modal"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@workspace/ui/components/dialog"

function RoutePluginPageContent() {
  const router = useRouter()
  const params = useSearchParams()
  const pluginId = params.get("pluginId") ?? ""
  const pluginData = useData<RoutePlugin>(
    `/api/route-plugins/${pluginId}`,
    'route_plugin',
  )

  const [showDeleteConfirmation, setShowDeleteConfirmation] = React.useState(false)
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successRedirect, setSuccessRedirect] = React.useState<string | null>(null)
  const [successMessage, setSuccessMessage] = React.useState("")

  const { loading, error, mutate, reset } = useMutation()

  const deletePlugin = React.useCallback(async () => {
    if (!pluginData.data) return

    const result = await mutate(
      `/api/route-plugins/${pluginId}`,
      "DELETE"
    )

    if (result.success) {
      setShowSuccess(true)
      setSuccessMessage(
        "Plugin deleted successfully. You will be redirected to plugins list in 5 seconds."
      )
      setSuccessRedirect("/plugins")
    }
  }, [mutate, pluginData.data, pluginId])

  return (
    <SidebarLayout page_title={"Route Plugin Details"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!pluginData.error || !!error}
        description={
          pluginData.error?.details ||
          error?.details ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={() => {
          if (pluginData.error) void pluginData.refetch()
          if (error) reset()
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        title={"Success"}
        visible={showSuccess}
        description={successMessage}
        onClose={() => {
          setShowSuccess(false)
          if (successRedirect) router.push(successRedirect)
        }}
      />
      <Dialog
        open={showDeleteConfirmation}
        onOpenChange={setShowDeleteConfirmation}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Are you sure?</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            This action cannot be undone. This will permanently disassociate this
            plugin from the route it&#39;s attached to. The plugin
            itself will remain untouched.
          </DialogDescription>
          <DialogFooter>
            <Button
              variant={"outline"}
              onClick={() => setShowDeleteConfirmation(false)}
            >
              Cancel
            </Button>
            <Button
              variant={"destructive"}
              onClick={deletePlugin}
              disabled={loading || !!error}
            >
              {loading && <Spinner />}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <div className={"flex flex-col p-6"}>
        {pluginData.loading ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <div className={"flex flex-col"}>
            {pluginData.error ? (
              <div></div>
            ) : (
              <>
                <div
                  className={"flex flex-row items-center justify-start pb-5"}
                >
                  <Button
                    variant={"ghost"}
                    size={"icon"}
                    onClick={() => router.back()}
                  >
                    <ArrowLeft />
                  </Button>
                </div>
                <div className={"flex flex-col gap-5 lg:flex-row"}>
                  <div className={"w-full lg:w-1/3"}>
                    <RoutePluginCard plugin={pluginData.data!} onDelete={() => setShowDeleteConfirmation(true)} />
                  </div>
                  <div className={"w-full lg:w-2/3"}>
                    <Card className={"w-full"}>
                      <CardHeader>
                        <CardTitle>JSON Configuration</CardTitle>
                        <CardDescription>
                          This JSON Configuration is available for the plugin
                          via host function.{" "}
                          <a
                            className={"underline"}
                            href={"/docs#plugin-json-config"}
                          >
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
                              value={
                                pluginData.data?.config ??
                                "\nYou did not specify a config for this plugin.\n"
                              }
                              readOnly
                              className={"font-mono text-sm"}
                            />
                            <InputGroupAddon align={"block-start"}>
                              <FileBraces className={"text-muted-foreground"} />
                              <InputGroupText className={"font-mono"}>
                                config.json
                              </InputGroupText>
                            </InputGroupAddon>
                          </InputGroup>
                        </Field>
                      </CardContent>
                    </Card>
                  </div>
                </div>
              </>
            )}
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}

export default function RoutePluginPage() {
  return (
    <React.Suspense
      fallback={
        <SidebarLayout page_title={"Route Plugin Details"}>
          <div className={"flex items-center justify-center p-6"}>
            <Spinner className={"size-10"} />
          </div>
        </SidebarLayout>
      }
    >
      <RoutePluginPageContent />
    </React.Suspense>
  )
}
