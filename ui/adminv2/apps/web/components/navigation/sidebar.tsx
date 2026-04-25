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

import {
  Sidebar,
  SidebarHeader,
  SidebarContent,
  SidebarGroup,
  SidebarFooter,
  SidebarRail,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuAction,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarGroupAction,
  SidebarSeparator,
  SidebarProvider,
  SidebarMenuSub,
} from "@workspace/ui/components/sidebar"
import { Button } from "@workspace/ui/components/button"
import {useRouter} from "next/navigation"
import {
  Blocks,
  ChartLine,
  LayoutDashboard,
  PanelRight,
  PlusIcon,
  Route,
  SunMoon,
  ToyBrick,
} from "lucide-react"
import React from "react"
import { cn } from "@workspace/ui/lib/utils"
import { useTheme } from "next-themes"
import Image from "next/image"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  DropdownMenuItem,
} from "@workspace/ui/components/dropdown-menu"

interface SidebarLayoutProps {
  children: React.ReactNode
  pageTitle?: string
  className?: string
}

const SidebarLayout = ({
  pageTitle,
  children,
  className,
}: SidebarLayoutProps) => {
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const { resolvedTheme, setTheme } = useTheme()

  return (
    <SidebarProvider open={open} onOpenChange={setOpen}>
      <div className={"min-w-screen bg-sidebar"}>
        <Sidebar className={cn("p-3", className)}>
          <SidebarHeader>
            <div className={"flex flex-row items-center justify-start"}>
              <p className={"text-2xl font-bold"}>WasmForge</p>
            </div>
          </SidebarHeader>
          <SidebarContent className={"mt-5"}>
            <SidebarMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant={"default"}
                    size={"lg"}
                    className={
                      "w-full justify-start bg-sidebar text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                    }
                  >
                    <PlusIcon />
                    Quick Create
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem onClick={() => router.push("/routes/new")}>
                    Route
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => router.push("/plugins/new")}>
                    Plugin
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => router.push("/routes/plugins/new")}
                  >
                    Attach plugins
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <SidebarGroup>
                <SidebarGroupLabel>Overview and stats</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <Button
                        onClick={() => router.push("/routes")}
                        variant={"default"}
                        size={"lg"}
                        className={
                          "w-full justify-start bg-sidebar text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                        }
                      >
                        <LayoutDashboard />
                        Dashboard
                      </Button>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <Button
                        onClick={() => router.push("/routes")}
                        variant={"default"}
                        size={"lg"}
                        className={
                          "w-full justify-start bg-sidebar text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                        }
                      >
                        <ChartLine />
                        Statistics
                      </Button>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
              <SidebarGroup>
                <SidebarGroupLabel>Items and content</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    <SidebarMenuItem>
                      <Button
                        onClick={() => router.push("/routes")}
                        variant={"default"}
                        size={"lg"}
                        className={
                          "w-full justify-start bg-sidebar text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                        }
                      >
                        <LayoutDashboard />
                        Routes
                      </Button>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <Button
                        onClick={() => router.push("/plugins")}
                        variant={"default"}
                        size={"lg"}
                        className={
                          "w-full justify-start bg-sidebar text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                        }
                      >
                        <ToyBrick />
                        Plugins
                      </Button>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                      <Button
                        onClick={() => router.push("/routes")}
                        variant={"default"}
                        size={"lg"}
                        className={
                          "w-full justify-start bg-sidebar text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                        }
                      >
                        <Blocks />
                        Plugins on routes
                      </Button>
                    </SidebarMenuItem>
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarMenu>
          </SidebarContent>
          <SidebarFooter>
            <a
              href={"https://github.com/mikhail5545/wasmforge"}
              className={
                "flex h-10 flex-row items-center justify-start gap-4 p-3 hover:underline"
              }
            >
              {resolvedTheme === "dark" ? (
                <Image
                  src={"/GitHub_Invertocat_White_Clearspace.svg"}
                  alt={"GitHub"}
                  width={32}
                  height={32}
                />
              ) : (
                <Image
                  src={"/GitHub_Invertocat_Black_Clearspace.svg"}
                  alt={"GitHub"}
                  width={32}
                  height={32}
                />
              )}
              <p className={"text-sm"}>View on GitHub</p>
            </a>
          </SidebarFooter>
          <SidebarRail />
        </Sidebar>
        <div
          className={cn(
            "mx-auto bg-sidebar p-5 h-full transition-all duration-300",
            open ? "ml-[16rem]" : "ml-0"
          )}
        >
          <div
            className={cn(
              "rounded-xl bg-background h-full transition-all duration-300",
              open ? "px-0 md:px-5 lg:px-10" : "px-10 md:px-15 lg:px-30"
            )}
          >
            <div
              className={
                "flex flex-row items-center justify-between border-b border-b-border p-2"
              }
            >
              <div className={"flex flex-row"}>
                <div
                  className={
                    "flex items-center justify-center border-r border-r-border pr-2"
                  }
                >
                  <Button
                    variant={"secondary"}
                    size={"icon"}
                    onClick={() => setOpen(!open)}
                  >
                    <PanelRight />
                  </Button>
                </div>
                <p className={"px-3 text-lg font-bold"}>{pageTitle}</p>
              </div>
              <Button
                variant={"ghost"}
                size={"icon"}
                onClick={() =>
                  setTheme(resolvedTheme === "light" ? "dark" : "light")
                }
              >
                <SunMoon />
              </Button>
            </div>
            <div className={"flex w-full flex-col p-5"}>{children}</div>
          </div>
        </div>
      </div>
    </SidebarProvider>
  )
}

export {SidebarLayout}