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

import React from "react"

import { LayoutDashboard, Route, ToyBrick, Webhook, BookText, Router } from "lucide-react"

import {NavMain} from "@/components/navigation/nav-main"
import {NavSecondary} from "@/components/navigation/nav-secondary"

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@workspace/ui/components/sidebar"

const data = {
  navMain: [
    {
      title: "Dashboard",
      link: "/dashboard",
      icon: LayoutDashboard,
      isActive: false,
    },
    {
      title: "Routes",
      link: "/routes",
      icon: Route,
      isActive: false,
    },
    {
      title: "Plugins",
      link: "/plugins",
      icon: ToyBrick,
      isActive: false,
    }
  ],
  navSecondary: {
    items: [
      {
        title: "API Reference",
        link: "/api-reference",
        icon: Webhook,
        isActive: false,
      },
      {
        title: "Documentation",
        link: "/documentation",
        icon: BookText,
        isActive: false,
      }
    ],
    label: "API Reference",
  }
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar variant={"inset"} {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size={"lg"} asChild>
              <a href={"/"}>
                <div
                  className={
                    "flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"
                  }
                >
                  <Router className={"size-4"} />
                </div>
                <div className={"grid flex-1 text-left text-sm leading-tight"}>
                  <span className={"truncate font-medium"}>WasmForge</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary
          items={data.navSecondary.items}
          label={data.navSecondary.label}
          className={"mt-auto"}
        />
      </SidebarContent>
    </Sidebar>
  )
}