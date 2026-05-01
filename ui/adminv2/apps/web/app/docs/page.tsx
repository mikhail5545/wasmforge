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

import { CodeBlock } from "@/components/code-block"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Separator } from "@workspace/ui/components/separator"

const toc = [
  ["getting-started", "Getting Started"],
  ["auth-overview", "Auth Overview"],
  ["auth-config", "Route Auth Config"],
  ["key-backends", "Key Backends"],
  ["key-encryption", "Private Key Encryption"],
  ["admin-api", "Admin API"],
  ["plugin-auth-context", "WASM Plugin Auth Context"],
  ["troubleshooting", "Troubleshooting"],
]

const authConfigExample = `{
  "validate_tokens": true,
  "issue_tokens": true,
  "key_backend_type": "database",
  "token_ttl_seconds": 3600,
  "issuer": "wasmforge",
  "audience": "upstream-api",
  "allowed_algorithms": ["RS256"],
  "required_claims": ["sub"],
  "metadata": {
    "upstream_auth_header": "Authorization"
  }
}`

const databaseKeyExample = `{
  "route_id": "018f0000-0000-7000-8000-000000000001",
  "key_id": "route-signing-key-v1",
  "private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\\n...\\n-----END RSA PRIVATE KEY-----",
  "public_key_pem": "-----BEGIN PUBLIC KEY-----\\n...\\n-----END PUBLIC KEY-----",
  "metadata": {
    "primary": true
  }
}`

const envBackendExample = `{
  "validate_tokens": true,
  "issue_tokens": true,
  "key_backend_type": "env",
  "token_ttl_seconds": 3600,
  "allowed_algorithms": ["RS256"],
  "metadata": {
    "env_public_key_var": "WASMFORGE_ROUTE_PUBLIC_KEY",
    "env_private_key_var": "WASMFORGE_ROUTE_PRIVATE_KEY",
    "env_key_id": "env-key-v1"
  }
}`

const jwksBackendExample = `{
  "validate_tokens": true,
  "issue_tokens": false,
  "key_backend_type": "jwks",
  "jwks_url": "https://issuer.example.com/.well-known/jwks.json",
  "jwks_cache_ttl_seconds": 300,
  "token_ttl_seconds": 3600,
  "allowed_algorithms": ["RS256"],
  "issuer": "https://issuer.example.com/",
  "audience": "wasmforge"
}`

export default function DocsPage() {
  return (
    <SidebarLayout page_title={"Documentation"}>
      <main className={"mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"}>
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"}>WasmForge Gateway</Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>Documentation</h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Configure routes, attach WASM plugins, and add native JWT auth
            without giving plugins access to signing keys or private material.
          </p>
        </header>

        <Card>
          <CardHeader>
            <CardTitle>Table of Contents</CardTitle>
            <CardDescription>
              Operational reference for the embedded admin panel and API.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <nav className={"grid gap-2 sm:grid-cols-2 lg:grid-cols-4"}>
              {toc.map(([id, label]) => (
                <a
                  key={id}
                  href={`#${id}`}
                  className={"rounded-lg border px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"}
                >
                  {label}
                </a>
              ))}
            </nav>
          </CardContent>
        </Card>

        <DocSection
          id={"getting-started"}
          title={"Getting Started"}
          description={"Build the gateway, embedded adminv2 UI, and Go binary."}
        >
          <p>
            Install <strong>Go 1.25+</strong> and <strong>Node 22+</strong>.
            From the repository root, run one build command. The active admin UI
            is <code className={"rounded bg-muted px-1 font-mono"}>ui/adminv2</code>;
            the old <code className={"rounded bg-muted px-1 font-mono"}>ui/admin-ui</code>{" "}
            directory is deprecated.
          </p>
          <CodeBlock
            tabs={[
              { label: "Make", code: "make build" },
              { label: "Bash", code: "bash ./scripts/build.sh" },
              { label: "PowerShell", code: "powershell ./scripts/build.ps1" },
            ]}
          />
          <CodeBlock
            tabs={[
              { label: "Bash", code: "./bin/wasmforge" },
              { label: "PowerShell", code: "./bin/wasmforge.exe" },
            ]}
          />
          <p>
            The admin panel and admin API are served from{" "}
            <strong>http://localhost:8080</strong>. Admin API routes are under{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>/api</code>.
          </p>
        </DocSection>

        <DocSection
          id={"auth-overview"}
          title={"Auth Overview"}
          description={"Native auth runs before WASM plugins in the route chain."}
        >
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Validate inbound JWTs"}>
              Enable <code className={"rounded bg-muted px-1 font-mono"}>validate_tokens</code>{" "}
              to require a bearer token before the request reaches plugins or
              the upstream service.
            </InfoCard>
            <InfoCard title={"Issue upstream JWTs"}>
              Enable <code className={"rounded bg-muted px-1 font-mono"}>issue_tokens</code>{" "}
              to mint a gateway-owned token and inject it into the upstream
              request header.
            </InfoCard>
            <InfoCard title={"Expose safe plugin context"}>
              Plugins can read the authenticated flag, subject, and claims.
              They cannot read private keys or call the signer.
            </InfoCard>
          </div>
          <p>
            Request order is observer, route state, native auth, WASM plugins,
            method validation, and reverse proxy. This keeps plugin behavior
            intact while giving plugins a verified trust boundary.
          </p>
        </DocSection>

        <DocSection
          id={"auth-config"}
          title={"Route Auth Config"}
          description={"Auth is route-scoped and stored in auth_configs."}
        >
          <p>
            Every route can have one auth config. At least one of{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>validate_tokens</code>{" "}
            or <code className={"rounded bg-muted px-1 font-mono"}>issue_tokens</code>{" "}
            must be enabled. RS256 is currently the supported signing algorithm.
          </p>
          <CodeBlock language={"json"} code={authConfigExample} />
          <p>
            By default issued upstream tokens are injected as{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>Authorization: Bearer &lt;token&gt;</code>.
            Set <code className={"rounded bg-muted px-1 font-mono"}>metadata.upstream_auth_header</code>{" "}
            to use a custom header.
          </p>
        </DocSection>

        <DocSection
          id={"key-backends"}
          title={"Key Backends"}
          description={"Choose where validation and signing keys come from."}
        >
          <div className={"grid gap-4 lg:grid-cols-3"}>
            <InfoCard title={"database"}>
              Stores route key material in <code className={"rounded bg-muted px-1 font-mono"}>key_materials</code>.
              Supports validation and issuance. Multiple active keys require one{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>metadata.primary</code> key.
            </InfoCard>
            <InfoCard title={"env"}>
              Reads PEM keys from environment variables named in config
              metadata. Supports validation and issuance when a private key var
              is configured.
            </InfoCard>
            <InfoCard title={"jwks"}>
              Fetches public keys from a JWKS URL. Supports validation only.
              Issuance must be disabled for this backend.
            </InfoCard>
          </div>
          <CodeBlock
            tabs={[
              { label: "database key", language: "json", code: databaseKeyExample },
              { label: "env config", language: "json", code: envBackendExample },
              { label: "jwks config", language: "json", code: jwksBackendExample },
            ]}
          />
        </DocSection>

        <DocSection
          id={"key-encryption"}
          title={"Private Key Encryption"}
          description={"Database-backed private keys are encrypted before persistence."}
        >
          <p>
            New DB-backed private keys are envelope-encrypted. WasmForge
            generates a per-key data encryption key, encrypts the PEM with
            AES-256-GCM, wraps the data key through the configured provider, and
            stores only ciphertext and metadata.
          </p>
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"local"}>
              Uses a 32-byte master key from{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>WASMFORGE_AUTH_MASTER_KEY</code>.
              This is the default provider.
            </InfoCard>
            <InfoCard title={"1password"}>
              Resolves the master key from a 1Password secret reference using{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>OP_SERVICE_ACCOUNT_TOKEN</code>.
            </InfoCard>
            <InfoCard title={"aws-kms"}>
              Uses AWS KMS to wrap and unwrap data encryption keys. Configure
              the KMS key ID or ARN at startup.
            </InfoCard>
          </div>
          <CodeBlock
            tabs={[
              {
                label: "local",
                code: "$env:WASMFORGE_AUTH_MASTER_KEY = \"<base64-32-byte-key>\"\n./bin/wasmforge.exe --auth-encryption-provider local",
              },
              {
                label: "1password",
                code: "$env:OP_SERVICE_ACCOUNT_TOKEN = \"<token>\"\n./bin/wasmforge.exe --auth-encryption-provider 1password --auth-encryption-1password-reference \"op://Vault/WasmForge/master-key\"",
              },
              {
                label: "aws-kms",
                code: "./bin/wasmforge.exe --auth-encryption-provider aws-kms --auth-encryption-aws-kms-region us-east-1 --auth-encryption-aws-kms-key-id alias/wasmforge-auth",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"admin-api"}
          title={"Admin API"}
          description={"Use these endpoints from adminv2 or automation."}
        >
          <div className={"grid gap-3"}>
            <Endpoint method={"GET"} path={"/api/auth/routes/:route_id/config"} />
            <Endpoint method={"PUT"} path={"/api/auth/routes/:route_id/config"} />
            <Endpoint method={"DELETE"} path={"/api/auth/routes/:route_id/config"} />
            <Endpoint method={"POST"} path={"/api/auth/validate"} />
            <Endpoint method={"GET"} path={"/api/auth/keys"} />
            <Endpoint method={"POST"} path={"/api/auth/keys"} />
            <Endpoint method={"POST"} path={"/api/auth/keys/generate"} />
            <Endpoint method={"GET"} path={"/api/auth/keys/:kid"} />
            <Endpoint method={"DELETE"} path={"/api/auth/keys/:kid"} />
          </div>
          <CodeBlock
            tabs={[
              {
                label: "set config",
                code: "curl -X PUT http://localhost:8080/api/auth/routes/<route-id>/config \\\n  -H \"Content-Type: application/json\" \\\n  -d @auth-config.json",
              },
              {
                label: "generate key",
                code: "curl -X POST http://localhost:8080/api/auth/keys/generate \\\n  -H \"Content-Type: application/json\" \\\n  -d '{\"route_id\":\"<route-id>\",\"key_id\":\"route-signing-key-v1\",\"expires_in_days\":90,\"metadata\":{\"primary\":true}}'",
              },
              {
                label: "validate token",
                code: "curl -X POST http://localhost:8080/api/auth/validate \\\n  -H \"Content-Type: application/json\" \\\n  -d '{\"route_id\":\"<route-id>\",\"token\":\"<jwt>\"}'",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"plugin-auth-context"}
          title={"WASM Plugin Auth Context"}
          description={"Plugins can inspect verified auth state after native auth succeeds."}
        >
          <p>
            WASM plugins run after native auth. Use host exports to read
            verified context instead of reparsing JWTs inside plugins.
          </p>
          <div className={"grid gap-3"}>
            <Endpoint method={"WASM"} path={"host_auth_is_authenticated() -> u32"} />
            <Endpoint method={"WASM"} path={"host_auth_subject(buf_ptr, buf_len) -> u32"} />
            <Endpoint method={"WASM"} path={"host_auth_claim(key_ptr, key_len, buf_ptr, buf_len) -> u32"} />
          </div>
          <p>
            These exports are read-only. Plugins can still set ordinary request
            headers through existing header APIs, but they never receive raw
            signer access, private PEMs, or encryption keys.
          </p>
        </DocSection>

        <DocSection
          id={"troubleshooting"}
          title={"Troubleshooting"}
          description={"Common auth setup failures and the first place to look."}
        >
          <div className={"grid gap-4 md:grid-cols-2"}>
            <InfoCard title={"Startup fails with missing master key"}>
              Set the required provider env var. Local provider needs{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>WASMFORGE_AUTH_MASTER_KEY</code>;
              1Password needs the service account token; AWS KMS needs usable
              AWS credentials and a key ID.
            </InfoCard>
            <InfoCard title={"JWKS issuance is rejected"}>
              JWKS is validation-only because it provides public keys. Use the
              database or env backend when WasmForge must issue upstream tokens.
            </InfoCard>
            <InfoCard title={"Multiple database keys fail validation"}>
              Mark exactly one active key with{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>{"{\"primary\":true}"}</code>.
              WasmForge does not choose an arbitrary active key.
            </InfoCard>
            <InfoCard title={"Plugin cannot see auth claims"}>
              Confirm the route has auth enabled and token validation succeeds.
              Plugins only receive populated auth context after native auth
              validates the inbound token.
            </InfoCard>
          </div>
        </DocSection>
      </main>
    </SidebarLayout>
  )
}

function DocSection({
  id,
  title,
  description,
  children,
}: {
  id: string
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <section id={id} className={"scroll-mt-20"}>
      <Separator className={"mb-8"} />
      <div className={"flex flex-col gap-5"}>
        <div className={"flex flex-col gap-2"}>
          <h2 className={"text-2xl font-bold tracking-tight"}>{title}</h2>
          <p className={"text-muted-foreground"}>{description}</p>
        </div>
        <div className={"flex flex-col gap-5 text-sm leading-7"}>{children}</div>
      </div>
    </section>
  )
}

function InfoCard({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <Card size={"sm"}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className={"text-muted-foreground"}>{children}</CardContent>
    </Card>
  )
}

function Endpoint({ method, path }: { method: string; path: string }) {
  return (
    <div className={"flex flex-wrap items-center gap-2 rounded-lg border bg-card px-3 py-2"}>
      <Badge variant={"secondary"}>{method}</Badge>
      <code className={"font-mono text-sm"}>{path}</code>
    </div>
  )
}
