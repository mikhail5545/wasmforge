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

import { Separator } from "@workspace/ui/components/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@workspace/ui/components/sidebar"
import { AppSidebar } from "@/components/navigation/app-sidebar"
import { useTheme } from "next-themes"
import { Button } from "@workspace/ui/components/button"
import { SunMoon } from "lucide-react"
import React, { Suspense } from "react"
import { Spinner } from "@workspace/ui/components/spinner"

export const SidebarLayout = ({
  page_title,
  children,
}: {
  page_title: string
  children: React.ReactNode
}) => {
  const { resolvedTheme, setTheme } = useTheme()

  React.useEffect(() => {
    document.title = `${page_title} | WasmForge Dashboard`
  }, [page_title])
  
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className={"flex h-16 shrink-0 items-center gap-2 border-b"}>
          <div
            className={"flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6"}
          >
            <SidebarTrigger className={"-ml-1"} />
            <Separator
              orientation={"vertical"}
              className={"mx-2 data-[orientation=vertical]:h-8"}
            />
            <h1 className={"ml-1 text-base font-medium"}>{page_title}</h1>
            <div className={"ml-auto flex items-center gap-2"}>
              <Button
                variant={"ghost"}
                size={"icon"}
                onClick={() =>
                  setTheme(resolvedTheme === "light" ? "dark" : "light")
                }
              >
                <SunMoon />
              </Button>
              <Button variant={"ghost"} size={"sm"} asChild>
                <a
                  href={"https://github.com/mikhail5545/wasmforge"}
                  rel={"noopener noreferrer"}
                  target={"_blank"}
                >
                  GitHub
                </a>
              </Button>
            </div>
          </div>
        </header>
        <Suspense fallback={
          <div className={"flex items-center justify-center py-50"}>
            <Spinner />
          </div>
        }>{children}</Suspense>
      </SidebarInset>
    </SidebarProvider>
  )
}
