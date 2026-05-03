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
import { DocSection, InfoCard, Endpoint } from "@/components/doc"
import {
  Table,
  TableBody,
  TableHeader,
  TableRow,
  TableHead,
  TableCell,
} from "@workspace/ui/components/table"

const toc = [
  ["overview", "Authentication Overview"],
  ["auth-config", "Authentication Config"],
  ["configuration-policies", "Configuration Policies"],
  ["key-backends", "Key Backends"],
  ["key-encryption", "Private Key Encryption"],
  ["plugin-auth-context", "Plugin Auth Context"],
  ["considerations", "Security Considerations"],
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

export default function AuthDocsPage() {
  return (
    <SidebarLayout page_title={"Auth Integration"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className="w-fit">
            WasmForge Gateway
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>
            Auth Integration
          </h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            WasmForge provides a robust, route-scoped authentication system that
            integrates seamlessly with both native middleware and WASM plugins.
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
                  className={
                    "rounded-lg border px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
                  }
                >
                  {label}
                </a>
              ))}
            </nav>
          </CardContent>
        </Card>

        <DocSection
          id={"overview"}
          title={"Auth Overview"}
          description={
            "How current authentication flow is structured and what are the key steps of the auth integration."
          }
        >
          <p>
            WasmForge provides a robust, route-scoped authentication system that
            integrates seamlessly with both native middleware and WASM plugins.
            The system is designed to handle JWT validation, deterministic key
            resolution across multiple backends (Database, JWKS, Environment
            Variables), and secure token issuance for upstream services.
          </p>
          <p className={"text-xl font-semibold"}>Authentication Flow</p>
          <ol className={"mx-5 list-decimal space-y-2"}>
            <li className={"ml-3"}>
              <strong>Middleware Entry: </strong> The{" "}
              <code>authMiddleware</code> (native Go middleware) intercepts the
              request.
            </li>
            <li className={"ml-3"}>
              <strong>Config Resolution: </strong> It retrieves the{" "}
              <code>AuthConfig</code> associated with the current route. If
              authentication is disabled or no configuration is found, the
              request proceeds.
            </li>
            <li className={"ml-3"}>
              <strong>Token Extraction: </strong> The middleware extracts the
              Bearer token from the <code>Authorization</code> header.
            </li>
            <li className={"ml-3"}>
              <strong>Validation (Optional): </strong> If{" "}
              <code>ValidateTokens</code> is enabled:
              <ol className={"mx-5 list-disc space-y-1"}>
                <li className={"ml-3"}>
                  The <code>TokenValidator</code> parses the JWT.
                </li>
                <li className={"ml-3"}>
                  The <code>KeyManager</code> resolves the appropriate public
                  key based on the token's <code>kid</code>
                  header or a primary key fallback.
                </li>
                <li className={"ml-3"}>
                  Standard claims (iss, aud, exp, nbf) and custom{" "}
                  <code>RequiredClaims</code> are verified.
                </li>
                <li className={"ml-3"}>
                  On failure, a <code>401 Unauthorized</code> or{" "}
                  <code>403 Forbidden</code> response is returned.
                </li>
                <li className={"ml-3"}>
                  Standard claims (iss, aud, exp, nbf) and custom{" "}
                  <code>RequiredClaims</code> are verified.
                </li>
              </ol>
            </li>
            <li className={"ml-3"}>
              <strong>Context Propagation: </strong> Structured auth context is
              injected into the request state, it contains the following claims:
              <ol className={"list-disc space-y-1"}>
                <li className={"ml-3"}>
                  <code>IsAuthenticated</code> (boolean)
                </li>
                <li className={"ml-3"}>
                  <code>Subject</code> (sub claim)
                </li>
                <li className={"ml-3"}>
                  <code>Claims</code> (full claim set)
                </li>
                <li className={"ml-3"}>
                  <code>AuthConfig</code> (the active configuration)
                </li>
              </ol>
            </li>
            <li className={"ml-3"}>
              <strong>Token Issuance (Optional): </strong> If{" "}
              <code>IssueTokens</code> is enabled:
              <ol className={"list-disc space-y-1"}>
                <li className={"ml-3"}>
                  The <code>TokenIssuer</code> generates a new JWT signed with a
                  private key managed by the system.
                </li>
                <li className={"ml-3"}>
                  The new token is injected into an upstream header (default:{" "}
                  <code>Authorization</code> or configured via metadata)
                </li>
              </ol>
            </li>
            <li className={"ml-3"}>
              <strong>Plugin Execution: </strong> Subsequent WASM plugins can
              access the verified <code>AuthContext</code>
              through{" "}
              <a
                className={"underline"}
                href={"/docs/api-reference/wasm-exports"}
              >
                host exports
              </a>
              .
            </li>
          </ol>
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Validate inbound JWTs"}>
              Enable{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                validate_tokens
              </code>{" "}
              to require a bearer token before the request reaches plugins or
              the upstream service.
            </InfoCard>
            <InfoCard title={"Issue upstream JWTs"}>
              Enable{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                issue_tokens
              </code>{" "}
              to mint a gateway-owned token and inject it into the upstream
              request header.
            </InfoCard>
            <InfoCard title={"Expose safe plugin context"}>
              Plugins can read the authenticated flag, subject, and claims. They
              cannot read private keys or call the signer.
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
            <code className={"rounded bg-muted px-1 font-mono"}>
              validate_tokens
            </code>{" "}
            or{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              issue_tokens
            </code>{" "}
            must be enabled. RS256 is currently the supported signing algorithm.
          </p>
          <CodeBlock language={"json"} code={authConfigExample} />
          <p>
            By default issued upstream tokens are injected as{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              Authorization: Bearer &lt;token&gt;
            </code>
            . Set{" "}
            <code className={"rounded bg-muted px-1 font-mono"}>
              metadata.upstream_auth_header
            </code>{" "}
            to use a custom header.
          </p>
        </DocSection>

        <DocSection
          id={"configuration-policies"}
          title={"Configuration Policies"}
          description={
            "Description of authentication configuration policies and components."
          }
        >
          <p>
            As already told in the previous section, authentication is
            configured via <code>AuthConfig</code> model:
          </p>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Field</TableHead>
                <TableHead>Description</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              <TableRow>
                <TableCell>
                  <code>Enabled</code>
                </TableCell>
                <TableCell>Global toggle for the route.</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>ValidateTokens</code>
                </TableCell>
                <TableCell>Enables incoming JWT verification.</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>IssueTokens</code>
                </TableCell>
                <TableCell>
                  Enables outbound JWT signing for upstream.
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>KayBackendType</code>
                </TableCell>
                <TableCell>
                  Source of keys (<code>database</code>, <code>jwks</code>,{" "}
                  <code>env</code>).
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>JWKSUrl</code>
                </TableCell>
                <TableCell>Endpoint for fetching remote public keys.</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>TokenAudience</code>
                </TableCell>
                <TableCell>
                  Required <code>aud</code> claim value.
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>TokenIssuer</code>
                </TableCell>
                <TableCell>
                  Required <code>iss</code> claim value.
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>RequiredClaims</code>
                </TableCell>
                <TableCell>
                  JSON list of mandatory claims in the incoming token.
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>AllowedAlgorithms</code>
                </TableCell>
                <TableCell>
                  Permitted signing algorithms (e.g., <code>["RS256"]</code>).
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>
                  <code>Metadata</code>
                </TableCell>
                <TableCell>
                  Flexible JSON storage for backend-specific settings (e.g., Env
                  var names).
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </DocSection>

        <DocSection
          id={"key-backends"}
          title={"Key Backends"}
          description={"Choose where validation and signing keys come from."}
        >
          <div className={"grid gap-4 lg:grid-cols-3"}>
            <InfoCard title={"database"}>
              Stores route key material in{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                key_materials
              </code>
              . Supports validation and issuance. Multiple active keys require
              one{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                metadata.primary
              </code>{" "}
              key.
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
              {
                label: "database key",
                language: "json",
                code: databaseKeyExample,
              },
              {
                label: "env config",
                language: "json",
                code: envBackendExample,
              },
              {
                label: "jwks config",
                language: "json",
                code: jwksBackendExample,
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"key-encryption"}
          title={"Private Key Encryption"}
          description={
            "Database-backed private keys are encrypted before persistence."
          }
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
              <code className={"rounded bg-muted px-1 font-mono"}>
                WASMFORGE_AUTH_MASTER_KEY
              </code>
              . This is the default provider.
            </InfoCard>
            <InfoCard title={"1password"}>
              Resolves the master key from a 1Password secret reference using{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                OP_SERVICE_ACCOUNT_TOKEN
              </code>
              .
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
                code: '$env:WASMFORGE_AUTH_MASTER_KEY = "<base64-32-byte-key>"\n./bin/wasmforge.exe --auth-encryption-provider local',
              },
              {
                label: "1password",
                code: '$env:OP_SERVICE_ACCOUNT_TOKEN = "<token>"\n./bin/wasmforge.exe --auth-encryption-provider 1password \\ \n--auth-encryption-1password-reference "op://Vault/WasmForge/master-key"',
              },
              {
                label: "aws-kms",
                code: "./bin/wasmforge.exe --auth-encryption-provider aws-kms \\ \n--auth-encryption-aws-kms-region us-east-1 --auth-encryption-aws-kms-key-id alias/wasmforge-auth",
              },
            ]}
          />
        </DocSection>

        <DocSection
          id={"plugin-auth-context"}
          title={"WASM Plugin Auth Context"}
          description={
            "Plugins can inspect verified auth state after native auth succeeds."
          }
        >
          <p>
            WASM plugins run after native auth. Use host exports to read
            verified context instead of reparsing JWTs inside plugins.
          </p>
          <div className={"grid gap-3"}>
            <Endpoint
              method={"WASM"}
              path={"host_auth_is_authenticated() -> u32"}
            />
            <Endpoint
              method={"WASM"}
              path={"host_auth_subject(buf_ptr, buf_len) -> u32"}
            />
            <Endpoint
              method={"WASM"}
              path={
                "host_auth_claim(key_ptr, key_len, buf_ptr, buf_len) -> u32"
              }
            />
          </div>
          <p>
            These exports are read-only. Plugins can still set ordinary request
            headers through existing header APIs, but they never receive raw
            signer access, private PEMs, or encryption keys.
          </p>
        </DocSection>

        <DocSection
          id={"considerations"}
          title={"Security Considerations"}
          description={
            "Security considerations in the authentication workflow."
          }
        >
          <ol className={"mx-5 list-disc space-y-2"}>
            <li className={"ml-3"}>
              <strong>Key Encryption: </strong> Private keys stored in the
              database are encrypted at rest using an envelope encryption
              strategy (Local, 1Password, or AWS KMS).
            </li>
            <li className={"ml-3"}>
              <strong>Deterministic Resolution: </strong> The{" "}
              <code>KeyManager</code> follows a strict priority:{" "}
              <code>kid</code> match {"->"} Primary key {"->"} First active key.
            </li>
            <li className={"ml-3"}>
              <strong>Auditability: </strong> Every validation failure is logged
              to prevent silent bypass attempts.
            </li>
          </ol>
        </DocSection>

        <DocSection
          id={"troubleshooting"}
          title={"Troubleshooting"}
          description={
            "Common auth setup failures and the first place to look."
          }
        >
          <div className={"grid gap-4 md:grid-cols-2"}>
            <InfoCard title={"Startup fails with missing master key"}>
              Set the required provider env var. Local provider needs{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                WASMFORGE_AUTH_MASTER_KEY
              </code>
              ; 1Password needs the service account token; AWS KMS needs usable
              AWS credentials and a key ID.
            </InfoCard>
            <InfoCard title={"JWKS issuance is rejected"}>
              JWKS is validation-only because it provides public keys. Use the
              database or env backend when WasmForge must issue upstream tokens.
            </InfoCard>
            <InfoCard title={"Multiple database keys fail validation"}>
              Mark exactly one active key with{" "}
              <code className={"rounded bg-muted px-1 font-mono"}>
                {'{"primary":true}'}
              </code>
              . WasmForge does not choose an arbitrary active key.
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
