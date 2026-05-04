/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import type { ReactNode } from "react"

import { CodeBlock } from "@/components/code-block"
import { Endpoint } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Separator } from "@workspace/ui/components/separator"

const authConfigPayload = `{
  "validate_tokens": true,
  "issue_tokens": true,
  "key_backend_type": "database",
  "token_ttl_seconds": 3600,
  "issuer": "wasmforge",
  "audience": "orders-api",
  "allowed_algorithms": ["RS256"],
  "required_claims": ["sub"],
  "claims_mapping": {},
  "metadata": {
    "upstream_auth_header": "Authorization"
  }
}`

export default function AuthApiReferencePage() {
  return (
    <SidebarLayout page_title={"Auth API Reference"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className="w-fit">
            Reference
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Auth API</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Admin API endpoints for configuring route-scoped JWT validation,
            managing token issuance, and handling cryptographic key material.
          </p>
        </header>

        <ReferenceSection
          title={"Auth Config"}
          description={"Route-scoped JWT validation and issuance settings."}
        >
          <EndpointGrid>
            <Endpoint method={"GET"} path={"/api/auth/routes/:route_id/config"}>
              Read the route auth config.
            </Endpoint>
            <Endpoint method={"PUT"} path={"/api/auth/routes/:route_id/config"}>
              Create or replace the route auth config.
            </Endpoint>
            <Endpoint
              method={"DELETE"}
              path={"/api/auth/routes/:route_id/config"}
            >
              Disable and delete route auth config.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/auth/validate"}>
              Validate a token with{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                {'{ "route_id", "token" }'}
              </code>
              .
            </Endpoint>
          </EndpointGrid>
          <CodeBlock language={"json"} code={authConfigPayload} />
          <FieldList
            fields={[
              ["key_backend_type", "database, env, or jwks."],
              ["validate_tokens", "Require and validate inbound Bearer JWTs."],
              ["issue_tokens", "Mint an upstream JWT. Not supported by jwks."],
              [
                "issuer / audience",
                "Issuer and audience checks for inbound validation and values for issued tokens.",
              ],
              [
                "required_claims",
                "Claims that must be present in validated tokens.",
              ],
              [
                "claims_mapping",
                "Reserved mapping data for translating claims into gateway/upstream context.",
              ],
              [
                "metadata.upstream_auth_header",
                "Header used for issued upstream tokens. Defaults to Authorization.",
              ],
            ]}
          />
        </ReferenceSection>

        <ReferenceSection
          title={"Auth Keys"}
          description={
            "Database-backed key import, generation, listing, and deletion."
          }
        >
          <EndpointGrid>
            <Endpoint
              method={"GET"}
              path={
                "/api/auth/keys?ids=&r_ids=&auth_config_ids=&types=&alg=&is_active=&of=&od=&ps=&pt="
              }
            >
              List key material records.
            </Endpoint>
            <Endpoint method={"GET"} path={"/api/auth/keys/:kid"}>
              Get key material by key ID.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/auth/keys"}>
              Import PEM public/private key material for a route.
            </Endpoint>
            <Endpoint method={"POST"} path={"/api/auth/keys/generate"}>
              Generate a new route RSA key pair.
            </Endpoint>
            <Endpoint method={"DELETE"} path={"/api/auth/keys/:kid"}>
              Deactivate/delete key material by key ID.
            </Endpoint>
          </EndpointGrid>
          <CodeBlock
            tabs={[
              {
                label: "import",
                language: "json",
                code: `{
  "route_id": "<route-id>",
  "key_id": "orders-signing-key-v1",
  "private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\\n...\\n-----END RSA PRIVATE KEY-----",
  "public_key_pem": "-----BEGIN PUBLIC KEY-----\\n...\\n-----END PUBLIC KEY-----",
  "expires_at": "2026-08-01T00:00:00Z",
  "metadata": { "primary": true }
}`,
              },
              {
                label: "generate",
                language: "json",
                code: `{
  "route_id": "<route-id>",
  "key_id": "orders-signing-key-v2",
  "expires_in_days": 90,
  "metadata": { "primary": true }
}`,
              },
            ]}
          />
          <p className={"text-sm leading-7 text-muted-foreground"}>
            Generated responses include{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              private_key_pem
            </code>{" "}
            once. Persist it externally if operators need a copy. Database
            private keys are encrypted at rest by the configured auth encryption
            provider.
          </p>
          <CodeBlock
            tabs={[
              {
                label: "set config",
                code: 'curl -X PUT http://localhost:8080/api/auth/routes/<route-id>/config \\\n  -H "Content-Type: application/json" \\\n  -d @auth-config.json',
              },
              {
                label: "generate key",
                code: 'curl -X POST http://localhost:8080/api/auth/keys/generate \\\n  -H "Content-Type: application/json" \\\n  -d \'{"route_id":"<route-id>","key_id":"route-signing-key-v1","expires_in_days":90,"metadata":{"primary":true}}\'',
              },
              {
                label: "validate token",
                code: 'curl -X POST http://localhost:8080/api/auth/validate \\\n  -H "Content-Type: application/json" \\\n  -d \'{"route_id":"<route-id>","token":"<jwt>"}\'',
              },
            ]}
          />
        </ReferenceSection>
      </main>
    </SidebarLayout>
  )
}

function ReferenceSection({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <section className={"scroll-mt-20"}>
      <Separator className={"mb-8"} />
      <div className={"flex flex-col gap-5"}>
        <div className={"flex flex-col gap-2"}>
          <h2 className={"text-2xl font-bold tracking-tight"}>{title}</h2>
          <p className={"text-muted-foreground"}>{description}</p>
        </div>
        <div className={"flex flex-col gap-5"}>{children}</div>
      </div>
    </section>
  )
}

function EndpointGrid({ children }: { children: ReactNode }) {
  return <div className={"grid gap-3"}>{children}</div>
}

function FieldList({ fields }: { fields: [string, string][] }) {
  return (
    <div className={"grid gap-2"}>
      {fields.map(([name, description]) => (
        <div
          key={name}
          className={"rounded-lg border px-3 py-2 text-sm leading-6"}
        >
          <code className={"font-mono"}>{name}</code>
          <span className={"text-muted-foreground"}> - {description}</span>
        </div>
      ))}
    </div>
  )
}
