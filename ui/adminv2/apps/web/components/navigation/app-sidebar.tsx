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

import React from "react"

import {
  BookText,
  Info,
  LayoutDashboard,
  Route,
  ToyBrick,
  Webhook,
} from "lucide-react"

import { NavMain } from "@/components/navigation/nav-main"
import { NavSecondary } from "@/components/navigation/nav-secondary"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
} from "@workspace/ui/components/sidebar"

const WasmForgeLogo = ({ className }: { className?: string }) => (
  <svg
    viewBox="0 0 814 950"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <path
      d="M561.001 598.58L342.501 912.08L249.001 857.08V709.08L329.001 755.58L409.501 638.58V508.58L561.001 598.58Z"
      fill="currentColor"
    />
    <path
      d="M329.001 589.08L173.001 812.08L1.50146 709.08V479.08L163.001 367.58L245.501 415.58V537.58L163.001 487.58L100.001 530.58V650.58L150.501 678.58L249.001 539.08L329.001 589.08Z"
      fill="currentColor"
    />
    <path
      d="M120.501 359.58L1.50146 443.58L0.501465 240.08L407.001 0.580444L813.001 240.08V712.58L407.501 949.08L367.501 924.58L478.001 767.58L693.501 640.58V311.08L407.001 142.58L120.501 311.08V359.58Z"
      fill="currentColor"
    />
    <path
      d="M561.001 598.58L342.501 912.08L249.001 857.08V709.08L329.001 755.58L409.501 638.58V508.58L561.001 598.58Z"
      stroke="currentColor"
    />
    <path
      d="M329.001 589.08L173.001 812.08L1.50146 709.08V479.08L163.001 367.58L245.501 415.58V537.58L163.001 487.58L100.001 530.58V650.58L150.501 678.58L249.001 539.08L329.001 589.08Z"
      stroke="currentColor"
    />
    <path
      d="M120.501 359.58L1.50146 443.58L0.501465 240.08L407.001 0.580444L813.001 240.08V712.58L407.501 949.08L367.501 924.58L478.001 767.58L693.501 640.58V311.08L407.001 142.58L120.501 311.08V359.58Z"
      stroke="currentColor"
    />
  </svg>
)

const data = {
  navMain: [
    {
      title: "Dashboard",
      link: "/",
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
    },
  ],
  navSecondary: {
    items: [
      {
        title: "API Reference",
        link: "/docs/api-reference",
        icon: Webhook,
        isActive: false,
      },
      {
        title: "Documentation",
        link: "/docs",
        icon: BookText,
        isActive: false,
      },
      {
        title: "About",
        link: "/about",
        icon: Info,
        isActive: false,
      },
    ],
    label: "Resources",
  },
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar variant={"inset"} {...props}>
      <SidebarHeader>
        <div className={"flex flex-row items-center justify-start gap-3"}>
          <WasmForgeLogo className={"h-10 w-10 p-1.5"} />
          <span className={"truncate font-medium"}>
            Wasm<strong>Forge</strong>
          </span>
        </div>
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
