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
import { useSearchParams } from "next/navigation"
import { useData } from "@/hooks/use-data"
import { Plugin } from "@/types/Plugin"
import { Route } from "@/types/route"
import { Spinner } from "@workspace/ui/components/spinner"
import React from "react"
import { PluginCard } from "@/components/ui/plugin-card"
import { usePaginatedData } from "@/hooks/use-paginated-data"

export default function PluginPage() {
  const params = useSearchParams()
  const name = params.get("name") ?? ""
  const version = params.get("version") ?? ""
  const pluginData = useData<Plugin>(
    `http://localhost:8080/api/plugins/${encodeURIComponent(name)}?version=${encodeURIComponent(version)}`,
    "plugin"
  )

  const [pluginId, setPluginId] = React.useState<string>('')
  const [orderDirection, setOrderDirection] = React.useState("asc")
  const [orderField, setOrderField] = React.useState<string>("created_at")

  React.useEffect(() => {
    if (pluginData.data && pluginId === ''){
      setPluginId(pluginData.data.id)
    }
  }, [pluginData.data, pluginId])

  const routesData = usePaginatedData<Route>(
    `/api/routes?pids=${pluginId}`,
    "routes",
    10,
    orderField,
    orderDirection as "asc" | "desc",
    { preload: true }
  )

  return (
    <SidebarLayout page_title={"Plugin"}>
      <div className={"flex flex-col p-6"}>
        {pluginData.loading && pluginData.data === null ? (
          <div className={"flex items-center justify-center py-50"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <div className={"flex flex-col"}>
            {pluginData.error ? (
              <div className={"flex items-center justify-center py-50"}>
                <p className={"text-muted-foreground text-xl"}>Something went wrong</p>
              </div>
            ) : (
              <>
                <div className={"flex flex-row items-center justify-end pb-5"}></div>
                <div className={"flex flex-col gap-5 lg:flex-row"}>
                  <div className={"w-full lg:w-1/3"}>
                    <PluginCard plugin={pluginData.data!} className={"w-full"} />
                  </div>
                  <div className={"w-full lg:w-2/3"}>

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