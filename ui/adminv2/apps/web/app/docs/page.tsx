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

import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { CodeBlock } from "@/components/code-block"
import { Separator } from "@workspace/ui/components/separator"

export default function DocsPage() {
  return (
    <SidebarLayout page_title={"Documentation"}>
      <div
        className={"flex flex-col gap-5 px-5 py-6 md:px-20 lg:px-35 xl:px-60"}
      >
        <h1 className={"text-3xl font-bold"}>Documentation</h1>
        <div className={"flex flex-col gap-3"}>
          <p className={"text-xl font-semibold"}>Table of Contents</p>
          <div className={"flex flex-col gap-3"}>
            <a href={"#getting-started"} className={"underline"}>
              Getting Started
            </a>
          </div>
        </div>

        <Separator />
        <section id={"getting-started"} className={"flex flex-col gap-10"}>
          <div className={"flex flex-col gap-2"}>
            <h2 className={"text-2xl font-bold"}>Getting Started</h2>
            <p className={"text-md font-semibold"}>
              This section describes the process of installation and building
            </p>
          </div>
          <div className={"flex flex-col gap-5"}>
            <p>
              Firstly, you need to have <strong>Go 1.25+</strong> and{" "}
              <strong>Node 22+</strong>. It is also recommended to have{" "}
              <strong>Make</strong> installed, but project can be built without
              it as well. After you have all the dependencies, you can clone the
              repository with:
            </p>
            <CodeBlock
              code={
                "git clone https://github.com/mikhail5545/wasmforge.git\ncd wasmforge"
              }
            />
          </div>
          <div className={"flex flex-col gap-5"}>
            <p className={"text-lg font-medium"}>Build the application</p>
            <p>
              To build the application, navigate to the root directory of the
              application if you haven&#39;t done this yet. Then, you can run
              one of the following commands to build the application. You need
              just one line of code. This will handle both building ui and
              application binary itself.
            </p>
            <CodeBlock
              tabs={[
                { label: "Make", code: "make build" },
                { label: "Bash", code: "bash ./scripts/build.sh" },
                { label: "PowerShell", code: "powershell ./scripts/build.ps1" },
              ]}
            />
          </div>
          <div className={"flex flex-col gap-5"}>
            <p className={"text-lg font-medium"}>Running Binary</p>
            <p>
              After you have built the application, you can run it with the
              following command. This will start the application and open the UI
              in your default browser.
            </p>
            <CodeBlock
              tabs={[
                { label: "Bash", code: "./bin/wasmforge" },
                { label: "PowerShell", code: "./bin/wasmforge.exe" },
              ]}
            />
          </div>
          <p>
            After this steps, admin panel will be available at{" "}
            <strong>http://localhost::8080</strong>
          </p>
        </section>

        <Separator />
        <section id={"creating-a-route"} className={"flex flex-col gap-10"}>
          <div className={"flex flex-col gap-2"}>
            <h2 className={"text-2xl font-bold"}>Creating a Route</h2>
            <p className={"text-md font-semibold"}>
              This section describes the process of creating a new route
            </p>
          </div>
          <div className={"flex flex-col gap-5"}>
            <p className={"text-lg font-medium"}>Fill the form</p>
            <p>
              To create a new route using UI, you need to navigate to{" "}
              <a href={"/routes/new"} className={"underline"}>
                new route
              </a>{" "}
              page and fill the necessary fields. All fields and their meaning
              are listed below.
            </p>
            <ul className={"my-6 list-disc [&>li]:mt-2 mx-4"}>
              <li>
                <strong>Path</strong> - defines the path of the route. It can be any valid URL path, for example:{" "}
                <code className={"bg-muted px-1 font-mono"}>/my-route</code>. It is required and must be unique for every route.
              </li>
              <li>
                <strong>Target URL</strong> - where requests that are received in this route will be forwarded. It can be any valid URL, for example:{" "}
                <code className={"bg-muted px-1 font-mono"}>
                  http://localhost:3000
                </code>. This field is also required and must be a valid URL.
              </li>
              <li>
                <strong>Idle Connection Timeout</strong> - defines the maximum amount of time a connection can be idle before it is closed. It is an optional field and if it is not set, default value of 30 seconds will be used. The value must be a valid duration string, for example:{" "}
              </li>
            </ul>
          </div>
        </section>
      </div>
    </SidebarLayout>
  )
}