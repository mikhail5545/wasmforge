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
import React from "react"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Badge } from "@workspace/ui/components/badge"
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemMedia,
  ItemTitle,
} from "@workspace/ui/components/item"
import {
  BadgeCheck,
  CircleCheck,
  Copy,
  Database,
  MoreHorizontal,
  Pencil, Power, PowerOff, RotateCcw,
  Trash2,
  TriangleAlert,
} from "lucide-react"
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@workspace/ui/components/empty"
import { Button } from "@workspace/ui/components/button"
import { AlertModal } from "@/components/dialog/alert-modal"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuGroup
} from "@workspace/ui/components/dropdown-menu"
import { useMutation } from "@/hooks/use-mutation"

export default function SettingsPage() {

  const proxyServerStatus = useData<ProxyServerStatus>(
    "/api/proxy/config",
    "status"
  )

  const { loading, error, mutate, reset } = useMutation()
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")

  const resetTLSConfiguration = React.useCallback(async () => {
    if (!proxyServerStatus.data) return

    const response = await mutate(
      "/api/proxy/certs",
      "DELETE"
    )

    if (response.success){
      setShowSuccess(true)
      setSuccessMessage("TLS configuration reset successfully")
    }
    await proxyServerStatus.refetch()
  }, [mutate, proxyServerStatus])

  const toggleRun = React.useCallback(async () => {
    if (!proxyServerStatus.data) return

    const callUrl = `/api/proxy/server/${proxyServerStatus.data.running ? "stop" : "start"}`

    const response = await mutate(callUrl, "POST")

    if (response.success) {
      setShowSuccess(true)
      setSuccessMessage(
        proxyServerStatus.data.running
          ? "Proxy server stopped successfully"
          : "Proxy server started successfully"
      )
      await proxyServerStatus.refetch()
    }
  }, [mutate, proxyServerStatus])

  const restart = React.useCallback(async () => {
    if (!proxyServerStatus.data) return

    const response = await mutate(
      '/api/proxy/server/restart',
      'POST'
    )

    if (response.success) {
      setShowSuccess(true)
      setSuccessMessage("Proxy server restarted successfully")
    }
    await proxyServerStatus.refetch()
  }, [mutate, proxyServerStatus])

  return (
    <SidebarLayout page_title={"Proxy Server Settings"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        title={"Unexpected error occurred"}
        visible={!!proxyServerStatus.error || !!error}
        description={
          proxyServerStatus?.error?.message ||
          error?.message ||
          "No additional information available. Retrying in 5 seconds."
        }
        onClose={() => {
          if (proxyServerStatus.error) void proxyServerStatus.refetch()
          if (error) reset()
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Action successful!"}
        description={
          successMessage ||
          "All info has been saved. This dialog will be closed in 5 seconds"
        }
        icon={<CircleCheck size={15} />}
        onClose={() => {
          setShowSuccess(false)
        }}
      />
      <div className={"flex flex-col p-6"}>
        {proxyServerStatus.loading || !proxyServerStatus.data?.config ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-8"} />
          </div>
        ) : (
          <div className={"flex flex-col gap-5"}>
            <div className={"flex flex-col gap-5 lg:flex-row"}>
              <div className={"w-full lg:w-1/3"}>
                <Card className={"w-full"}>
                  <CardHeader
                    className={
                      "between flex flex-row items-center justify-between"
                    }
                  >
                    <CardTitle>Proxy Server</CardTitle>
                    <div
                      className={
                        "flex flex-row items-center justify-between gap-2"
                      }
                    >
                      <Badge
                        className={
                          proxyServerStatus.data?.running
                            ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
                            : "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300"
                        }
                      >
                        {proxyServerStatus.data?.running
                          ? "Running"
                          : "Stopped"}
                      </Badge>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant={"outline"} size={"icon"}>
                            <MoreHorizontal />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent>
                          <DropdownMenuItem asChild>
                            <a href={"/settings/edit"}>
                              <Pencil />
                              Edit
                            </a>
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            disabled={loading || !!error}
                            onClick={toggleRun}
                          >
                            {proxyServerStatus.data?.running ? (
                              <>
                                <PowerOff />
                                Stop
                              </>
                            ) : (
                              <>
                                <Power />
                                Start
                              </>
                            )}
                          </DropdownMenuItem>
                          {proxyServerStatus.data?.running && (
                            <DropdownMenuItem
                              onClick={restart}
                              disabled={loading || !!error}
                            >
                              <RotateCcw />
                              Restart
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className={"flex flex-col gap-4"}>
                      <div className={"flex flex-row"}>
                        <p className={"w-1/3 text-sm text-muted-foreground"}>
                          Port
                        </p>
                        <p className={"ml-3 w-2/3 text-sm font-semibold"}>
                          {proxyServerStatus.data?.config.listen_port}
                        </p>
                      </div>
                      <div className={"flex flex-row"}>
                        <p className={"w-1/3 text-sm text-muted-foreground"}>
                          Read Header Timeout
                        </p>
                        <p
                          className={
                            "ml-3 w-2/3 truncate text-sm font-semibold"
                          }
                        >
                          {`${proxyServerStatus.data?.config.read_header_timeout} sec`}
                        </p>
                      </div>
                      <Item variant={"outline"} size={"sm"} className={"mt-5"}>
                        <ItemMedia>
                          {proxyServerStatus.data?.config.tls_enabled ? (
                            <BadgeCheck />
                          ) : (
                            <TriangleAlert />
                          )}
                        </ItemMedia>
                        <ItemContent>
                          <ItemTitle>
                            {proxyServerStatus.data?.config.tls_enabled
                              ? "TLS is configured"
                              : "TLS is not configured"}
                          </ItemTitle>
                          <ItemDescription>
                            {proxyServerStatus.data?.config.tls_enabled
                              ? "Server is configured to use TLS for secure communication"
                              : "Server is not configured to use TLS. Communication may not be secure"}
                          </ItemDescription>
                        </ItemContent>
                      </Item>
                    </div>
                  </CardContent>
                </Card>
              </div>
              <div className={"w-full lg:w-2/3"}>
                <Card className={"w-full"}>
                  <CardHeader
                    className={"flex flex-row items-center justify-between"}
                  >
                    <CardTitle>TLS Configuration</CardTitle>
                    <div
                      className={
                        "flex flex-row items-center justify-between gap-2"
                      }
                    >
                      <Badge
                        className={
                          proxyServerStatus.data?.config.tls_enabled
                            ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
                            : "bg-yellow-50 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-300"
                        }
                      >
                        {proxyServerStatus.data?.config.tls_enabled ? (
                          <>
                            {" "}
                            <BadgeCheck /> <span>configured</span>{" "}
                          </>
                        ) : (
                          <>
                            {" "}
                            <TriangleAlert /> <span>not configured</span>{" "}
                          </>
                        )}
                      </Badge>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant={"outline"} size={"icon"}>
                            <MoreHorizontal />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent>
                          <DropdownMenuGroup>
                            <DropdownMenuItem asChild>
                              <a href={"/settings/tls"}>
                                <Pencil />
                                Edit
                              </a>
                            </DropdownMenuItem>
                            {proxyServerStatus.data?.config.tls_enabled && (
                              <DropdownMenuItem
                                variant={"destructive"}
                                onClick={resetTLSConfiguration}
                                disabled={loading || !!error}
                              >
                                <Trash2 />
                                Reset
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuGroup>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {proxyServerStatus.data?.config.tls_enabled ? (
                      <div className={"flex flex-col gap-4"}>
                        <div className={"flex flex-row"}>
                          <p className={"w-1/3 text-sm text-muted-foreground"}>
                            Certificate
                          </p>
                          <p className={"ml-3 w-2/3 text-sm font-semibold"}>
                            {proxyServerStatus.data?.config.tls_cert_path}
                          </p>
                        </div>
                        <div className={"flex flex-row"}>
                          <p className={"w-1/3 text-sm text-muted-foreground"}>
                            Key
                          </p>
                          <p className={"ml-3 w-2/3 text-sm font-semibold"}>
                            {proxyServerStatus.data?.config.tls_key_path}
                          </p>
                        </div>
                        <div className={"flex flex-col gap-4"}>
                          <div
                            className={
                              "flex w-full flex-col items-center justify-center gap-2 rounded-xl bg-background p-5"
                            }
                          >
                            <p className={"text-sm text-muted-foreground"}>
                              Certificate Checksum
                            </p>
                            <div
                              className={
                                "flex flex-row items-center justify-between gap-5"
                              }
                            >
                              <p className={"truncate text-sm font-semibold"}>
                                {`SHA-256:${proxyServerStatus.data?.config.tls_cert_hash}`}
                              </p>
                              <Button
                                variant={"ghost"}
                                size={"icon"}
                                onClick={() => {
                                  void navigator.clipboard.writeText(
                                    proxyServerStatus.data?.config
                                      .tls_cert_hash || ""
                                  )
                                }}
                              >
                                <Copy />
                              </Button>
                            </div>
                          </div>
                          <div
                            className={
                              "flex w-full flex-col items-center justify-center gap-2 rounded-xl bg-background p-5"
                            }
                          >
                            <p className={"text-sm text-muted-foreground"}>
                              Key Checksum
                            </p>
                            <div
                              className={
                                "flex flex-row items-center justify-between gap-5"
                              }
                            >
                              <p className={"truncate text-sm font-semibold"}>
                                {`SHA-256:${proxyServerStatus.data?.config.tls_key_hash}`}
                              </p>
                              <Button
                                variant={"ghost"}
                                size={"icon"}
                                onClick={() => {
                                  void navigator.clipboard.writeText(
                                    proxyServerStatus.data?.config
                                      .tls_key_hash || ""
                                  )
                                }}
                              >
                                <Copy />
                              </Button>
                            </div>
                          </div>
                        </div>
                      </div>
                    ) : (
                      <Empty>
                        <EmptyHeader>
                          <EmptyMedia variant={"icon"}>
                            <Database />
                          </EmptyMedia>
                          <EmptyTitle>Not Configured</EmptyTitle>
                          <EmptyDescription>
                            Configure TLS to view and edit configuration
                          </EmptyDescription>
                        </EmptyHeader>
                        <EmptyContent
                          className={"flex-row justify-center gap-4"}
                        >
                          <Button size={"sm"} asChild>
                            <a href={`/settings/tls`}>Configure TLS</a>
                          </Button>
                        </EmptyContent>
                      </Empty>
                    )}
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        )}
      </div>
    </SidebarLayout>
  )
}