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
import { useRouter, useSearchParams } from "next/navigation"
import {
  ArrowLeft, ChevronDownIcon, ChevronLeft, ChevronRight,
  CircleAlert,
  CircleCheck,
  KeyRound,
  ShieldCheck, ToyBrick,
  Trash2,
} from "lucide-react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"
import { AlertModal } from "@/components/dialog/alert-modal"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { useData } from "@/hooks/use-data"
import { useMutation } from "@/hooks/use-mutation"
import {
  AuthConfig,
  AuthConfigPayload,
  AuthKey,
  KeyBackendType,
  ValidatedTokenResponse,
} from "@/types/auth"
import { ErrorResponse } from "@/types/ErrorResponse"
import { Route } from "@/types/route"
import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Checkbox } from "@workspace/ui/components/checkbox"
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { Separator } from "@workspace/ui/components/separator"
import { Spinner } from "@workspace/ui/components/spinner"
import { Switch } from "@workspace/ui/components/switch"
import { Textarea } from "@workspace/ui/components/textarea"
import { usePaginatedData } from "@/hooks/use-paginated-data"
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@workspace/ui/components/empty"
import { DropdownMenu, DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuTrigger } from "@workspace/ui/components/dropdown-menu"
import { DropdownMenuContent } from "@radix-ui/react-dropdown-menu"

type AuthForm = {
  validate_tokens: boolean
  issue_tokens: boolean
  key_backend_type: KeyBackendType
  jwks_url: string
  jwks_cache_ttl_seconds: string
  token_ttl_seconds: string
  required_claims: string
  allowed_algorithms: string
  issuer: string
  audience: string
  claims_mapping: string
  metadata: string
}

const emptyAuthForm: AuthForm = {
  validate_tokens: true,
  issue_tokens: false,
  key_backend_type: "database",
  jwks_url: "",
  jwks_cache_ttl_seconds: "300",
  token_ttl_seconds: "3600",
  required_claims: "",
  allowed_algorithms: "RS256",
  issuer: "",
  audience: "",
  claims_mapping: "",
  metadata: "",
}

function parseJSON<T>(value: string | undefined, fallback: T): T {
  if (!value?.trim()) {
    return fallback
  }
  try {
    return JSON.parse(value) as T
  } catch {
    return fallback
  }
}

function formatJSON(value: string | undefined): string {
  const parsed = parseJSON<unknown>(value, null)
  return parsed === null ? "" : JSON.stringify(parsed, null, 2)
}

function csv(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
}

function optionalNumber(value: string): number | undefined {
  if (!value.trim()) {
    return undefined
  }
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function authConfigToForm(config: AuthConfig | null): AuthForm {
  if (!config) {
    return emptyAuthForm
  }
  return {
    validate_tokens: config.validate_tokens,
    issue_tokens: config.issue_tokens,
    key_backend_type: config.key_backend_type,
    jwks_url: config.jwks_url ?? "",
    jwks_cache_ttl_seconds:
      config.jwks_cache_ttl_seconds?.toString() ??
      emptyAuthForm.jwks_cache_ttl_seconds,
    token_ttl_seconds: config.token_ttl_seconds?.toString() ?? "3600",
    required_claims: parseJSON<string[]>(config.required_claims, []).join(", "),
    allowed_algorithms: parseJSON<string[]>(
      config.allowed_algorithms,
      []
    ).join(", "),
    issuer: config.token_issuer ?? "",
    audience: config.token_audience ?? "",
    claims_mapping: formatJSON(config.claims_mapping),
    metadata: formatJSON(config.metadata),
  }
}

function buildPayload(form: AuthForm): AuthConfigPayload {
  const metadata = parseJSON<Record<string, unknown>>(form.metadata, {})
  const claimsMapping = parseJSON<Record<string, string>>(form.claims_mapping, {})
  return {
    validate_tokens: form.validate_tokens,
    issue_tokens: form.issue_tokens,
    key_backend_type: form.key_backend_type,
    jwks_url: form.key_backend_type === "jwks" ? form.jwks_url : undefined,
    jwks_cache_ttl_seconds:
      form.key_backend_type === "jwks"
        ? optionalNumber(form.jwks_cache_ttl_seconds)
        : undefined,
    token_ttl_seconds: optionalNumber(form.token_ttl_seconds) ?? 3600,
    required_claims: csv(form.required_claims),
    allowed_algorithms: csv(form.allowed_algorithms),
    issuer: form.issuer,
    audience: form.audience,
    claims_mapping:
      Object.keys(claimsMapping).length > 0 ? claimsMapping : undefined,
    metadata: Object.keys(metadata).length > 0 ? metadata : undefined,
  }
}

async function fetchAuthConfig(routeID: string): Promise<AuthConfig | null> {
  const response = await fetch(
    `/api/auth/routes/${routeID}/config`
  )
  if (response.status === 404) {
    return null
  }
  if (!response.ok) {
    throw new Error(`Auth config request failed with ${response.status}`)
  }
  const body = (await response.json()) as { auth_config: AuthConfig }
  return body.auth_config
}

function RouteAuthPageContent() {
  const router = useRouter()
  const params = useSearchParams()
  const path = params.get("path") ?? ""
  const routeData = useData<Route>(
    path
      ? `/api/routes/${encodeURIComponent(path)}`
      : null,
    "route"
  )
  const [orderField, setOrderField] = React.useState('created_at')
  const [orderDirection, setOrderDirection] = React.useState("asc")
  const [perPage, setPerPage] = React.useState('10')
  const keysData = usePaginatedData<AuthKey>(
    `/api/auth/keys?r_ids=${routeData.data?.id ?? null}&is_active=true`,
    "keys",
    parseInt(perPage),
    orderField,
    orderDirection as 'asc' | 'desc',
    { preload: true }
  )
  
  const { loading, error, mutate, reset } = useMutation()
  const [authConfig, setAuthConfig] = React.useState<AuthConfig | null>(null)
  const [configLoading, setConfigLoading] = React.useState(false)
  const [form, setForm] = React.useState<AuthForm>(emptyAuthForm)
  const [validationErrors, setValidationErrors] = React.useState<Record<
    string,
    string
  > | null>(null)
  const [showErrorAlert, setShowErrorAlert] = React.useState(false)
  const [showSuccess, setShowSuccess] = React.useState(false)
  const [successMessage, setSuccessMessage] = React.useState("")
  const [generatedKey, setGeneratedKey] = React.useState<AuthKey | null>(null)
  const [keyID, setKeyID] = React.useState("")
  const [expiresInDays, setExpiresInDays] = React.useState("90")
  const [primaryKey, setPrimaryKey] = React.useState(true)
  const [privateKeyPEM, setPrivateKeyPEM] = React.useState("")
  const [publicKeyPEM, setPublicKeyPEM] = React.useState("")
  const [tokenToValidate, setTokenToValidate] = React.useState("")
  const [validationResult, setValidationResult] =
    React.useState<ValidatedTokenResponse | null>(null)

  const loadAuthConfig = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }
    setConfigLoading(true)
    try {
      const config = await fetchAuthConfig(routeData.data.id)
      setAuthConfig(config)
      setForm(authConfigToForm(config))
    } catch (err) {
      setShowErrorAlert(true)
    } finally {
      setConfigLoading(false)
    }
  }, [routeData.data])

  React.useEffect(() => {
    void loadAuthConfig()
  }, [loadAuthConfig])

  const resolveError = React.useCallback((mutationError: ErrorResponse) => {
    if (mutationError.code === "VALIDATION_FAILED") {
      setValidationErrors(mutationError.validationErrors ?? null)
      return
    }
    setValidationErrors(null)
    setShowErrorAlert(true)
  }, [])

  const updateForm = React.useCallback((patch: Partial<AuthForm>) => {
    setForm((prev) => ({ ...prev, ...patch }))
  }, [])

  const validateJSONFields = React.useCallback(() => {
    for (const [field, value] of [
      ["metadata", form.metadata],
      ["claims_mapping", form.claims_mapping],
    ] as const) {
      if (!value.trim()) {
        continue
      }
      try {
        JSON.parse(value)
      } catch {
        setValidationErrors({ [field]: `${field} must be valid JSON` })
        return false
      }
    }
    return true
  }, [form])

  const saveConfig = React.useCallback(async () => {
    if (!routeData.data || !validateJSONFields()) {
      return
    }
    const result = await mutate(
      `/api/auth/routes/${routeData.data.id}/config`,
      "PUT",
      JSON.stringify(buildPayload(form)),
      { "Content-Type": "application/json" }
    )
    if (!result.success) {
      resolveError(result.error)
      return
    }
    setValidationErrors(null)
    setSuccessMessage("Auth config saved for this route.")
    setShowSuccess(true)
    await loadAuthConfig()
  }, [form, loadAuthConfig, mutate, resolveError, routeData.data, validateJSONFields])

  const deleteConfig = React.useCallback(async () => {
    if (!routeData.data || !authConfig) {
      return
    }
    const result = await mutate(
      `/api/auth/routes/${routeData.data.id}/config`,
      "DELETE"
    )
    if (!result.success) {
      resolveError(result.error)
      return
    }
    setAuthConfig(null)
    setForm(emptyAuthForm)
    setSuccessMessage("Auth config deleted for this route.")
    setShowSuccess(true)
  }, [authConfig, mutate, resolveError, routeData.data])

  const generateKey = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }
    const result = await mutate(
      "/api/auth/keys/generate",
      "POST",
      {
        route_id: routeData.data.id,
        key_id: keyID,
        expires_in_days: optionalNumber(expiresInDays) ?? 90,
        metadata: primaryKey ? { primary: true } : undefined,
      }
    )
    if (!result.success) {
      resolveError(result.error)
      return
    }
    const body = (await result.response.json()) as { key: AuthKey }
    setGeneratedKey(body.key)
    setKeyID("")
    setSuccessMessage("Signing key generated. Copy the private key now if you need it; it is returned only once.")
    setShowSuccess(true)
    await keysData.refetch()
  }, [expiresInDays, keyID, keysData, mutate, primaryKey, resolveError, routeData.data])

  const importKey = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }
    const result = await mutate("/api/auth/keys", "POST", {
      route_id: routeData.data.id,
      key_id: keyID,
      private_key_pem: privateKeyPEM,
      public_key_pem: publicKeyPEM,
      metadata: primaryKey ? JSON.stringify({ primary: true }) : undefined,
    })
    if (!result.success) {
      resolveError(result.error)
      return
    }
    setKeyID("")
    setPrivateKeyPEM("")
    setPublicKeyPEM("")
    setSuccessMessage("Signing key imported.")
    setShowSuccess(true)
    await keysData.refetch()
  }, [keyID, keysData, mutate, primaryKey, privateKeyPEM, publicKeyPEM, resolveError, routeData.data])

  const deleteKey = React.useCallback(
    async (key: AuthKey) => {
      const result = await mutate(
        `/api/auth/keys/${encodeURIComponent(key.key_id)}`,
        "DELETE"
      )
      if (!result.success) {
        resolveError(result.error)
        return
      }
      await keysData.refetch()
    },
    [keysData, mutate, resolveError]
  )

  const validateToken = React.useCallback(async () => {
    if (!routeData.data) {
      return
    }
    const result = await mutate(
      "/api/auth/validate",
      "POST",
      { route_id: routeData.data.id, token: tokenToValidate }
    )
    if (!result.success) {
      resolveError(result.error)
      return
    }
    const body = (await result.response.json()) as {
      validation_result: ValidatedTokenResponse
    }
    setValidationResult(body.validation_result)
  }, [mutate, resolveError, routeData.data, tokenToValidate])

  const pageLoading = routeData.loading || configLoading

  return (
    <SidebarLayout page_title={"Route Auth"}>
      <AlertModal
        variant={"alert"}
        size={"sm"}
        visible={showErrorAlert || !!routeData.error || !!keysData.error}
        title={"Auth configuration error"}
        description={
          routeData.error?.message ||
          keysData.error?.message ||
          error?.message ||
          "No additional details available."
        }
        icon={<CircleAlert size={15} />}
        onClose={() => {
          setShowErrorAlert(false)
          reset()
        }}
      />
      <AlertModal
        variant={"default"}
        size={"sm"}
        visible={showSuccess}
        title={"Success"}
        description={successMessage}
        icon={<CircleCheck size={15} />}
        onClose={() => setShowSuccess(false)}
      />
      <div className={"flex flex-col gap-5 p-6"}>
        {pageLoading ? (
          <div className={"flex items-center justify-center py-40"}>
            <Spinner className={"size-10"} />
          </div>
        ) : (
          <>
            <Card>
              <CardHeader
                className={
                  "flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
                }
              >
                <div>
                  <CardTitle className={"flex items-center gap-2"}>
                    <ShieldCheck size={18} />
                    Native Auth Config
                  </CardTitle>
                  <CardDescription>
                    Route-scoped JWT validation and token issuing for{" "}
                    <span className={"font-mono"}>{routeData.data?.path}</span>.
                  </CardDescription>
                </div>
                <Badge variant={authConfig?.enabled ? "default" : "outline"}>
                  {authConfig?.enabled ? "Configured" : "Not configured"}
                </Badge>
              </CardHeader>
              <CardContent>
                <FieldSet>
                  <FieldGroup className={"grid gap-5 lg:grid-cols-3"}>
                    <Field orientation={"horizontal"}>
                      <Switch
                        checked={form.validate_tokens}
                        onCheckedChange={(checked) =>
                          updateForm({ validate_tokens: checked })
                        }
                      />
                      <FieldContent>
                        <FieldLabel>Validate tokens</FieldLabel>
                        <FieldDescription>
                          Reject requests with invalid JWTs.
                        </FieldDescription>
                      </FieldContent>
                    </Field>
                    <Field orientation={"horizontal"}>
                      <Switch
                        checked={form.issue_tokens}
                        onCheckedChange={(checked) =>
                          updateForm({ issue_tokens: checked })
                        }
                      />
                      <FieldContent>
                        <FieldLabel>Issue tokens</FieldLabel>
                        <FieldDescription>
                          Allow this route config to sign tokens.
                        </FieldDescription>
                      </FieldContent>
                    </Field>
                    <Field>
                      <FieldLabel>Key backend</FieldLabel>
                      <Select
                        value={form.key_backend_type}
                        onValueChange={(value) =>
                          updateForm({
                            key_backend_type: value as KeyBackendType,
                          })
                        }
                      >
                        <SelectTrigger className={"w-full"}>
                          <SelectValue placeholder={"Key backend"} />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectLabel>Backend</SelectLabel>
                            <SelectItem value={"database"}>Database</SelectItem>
                            <SelectItem value={"jwks"}>JWKS</SelectItem>
                            <SelectItem value={"env"}>Environment</SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </Field>
                    <Field>
                      <FieldLabel>Issuer</FieldLabel>
                      <Input
                        value={form.issuer}
                        placeholder={"wasmforge"}
                        onChange={(event) =>
                          updateForm({ issuer: event.target.value })
                        }
                      />
                    </Field>
                    <Field>
                      <FieldLabel>Audience</FieldLabel>
                      <Input
                        value={form.audience}
                        placeholder={"api://gateway"}
                        onChange={(event) =>
                          updateForm({ audience: event.target.value })
                        }
                      />
                    </Field>
                    <Field>
                      <FieldLabel>Token TTL seconds</FieldLabel>
                      <Input
                        type={"number"}
                        min={1}
                        value={form.token_ttl_seconds}
                        onChange={(event) =>
                          updateForm({ token_ttl_seconds: event.target.value })
                        }
                      />
                    </Field>
                    {form.key_backend_type === "jwks" && (
                      <>
                        <Field>
                          <FieldLabel>JWKS URL</FieldLabel>
                          <Input
                            value={form.jwks_url}
                            placeholder={"https://issuer/.well-known/jwks.json"}
                            onChange={(event) =>
                              updateForm({ jwks_url: event.target.value })
                            }
                          />
                        </Field>
                        <Field>
                          <FieldLabel>JWKS cache TTL seconds</FieldLabel>
                          <Input
                            type={"number"}
                            min={0}
                            value={form.jwks_cache_ttl_seconds}
                            onChange={(event) =>
                              updateForm({
                                jwks_cache_ttl_seconds: event.target.value,
                              })
                            }
                          />
                        </Field>
                      </>
                    )}
                    <Field>
                      <FieldLabel>Required claims</FieldLabel>
                      <Input
                        value={form.required_claims}
                        placeholder={"sub, scope"}
                        onChange={(event) =>
                          updateForm({ required_claims: event.target.value })
                        }
                      />
                      <FieldDescription>
                        Comma-separated claim names.
                      </FieldDescription>
                    </Field>
                    <Field>
                      <FieldLabel>Allowed algorithms</FieldLabel>
                      <Input
                        value={form.allowed_algorithms}
                        placeholder={"RS256"}
                        onChange={(event) =>
                          updateForm({ allowed_algorithms: event.target.value })
                        }
                      />
                    </Field>
                  </FieldGroup>
                  <Separator className={"my-5"} />
                  <FieldGroup className={"grid gap-5 lg:grid-cols-2"}>
                    <Field>
                      <FieldLabel>Claims mapping JSON</FieldLabel>
                      <Textarea
                        className={"min-h-28 font-mono text-xs"}
                        value={form.claims_mapping}
                        placeholder={'{"email":"preferred_username"}'}
                        onChange={(event) =>
                          updateForm({ claims_mapping: event.target.value })
                        }
                      />
                      <FieldError>
                        {validationErrors?.claims_mapping}
                      </FieldError>
                    </Field>
                    <Field>
                      <FieldLabel>Metadata JSON</FieldLabel>
                      <Textarea
                        className={"min-h-28 font-mono text-xs"}
                        value={form.metadata}
                        placeholder={
                          form.key_backend_type === "env"
                            ? '{"env_public_key_var":"JWT_PUBLIC_KEY","env_private_key_var":"JWT_PRIVATE_KEY","env_key_id":"env-v1"}'
                            : '{"upstream_auth_header":"Authorization"}'
                        }
                        onChange={(event) =>
                          updateForm({ metadata: event.target.value })
                        }
                      />
                      <FieldDescription>
                        Use metadata for upstream auth header behavior, env key
                        variables, or provider-specific settings.
                      </FieldDescription>
                      <FieldError>{validationErrors?.metadata}</FieldError>
                    </Field>
                  </FieldGroup>
                </FieldSet>
                <div className={"mt-6 flex justify-between gap-3"}>
                  <Button variant={"outline"} onClick={() => router.back()}>
                    <ArrowLeft size={15} />
                    Back
                  </Button>
                  <div className={"flex gap-2"}>
                    <Button
                      variant={"outline"}
                      onClick={deleteConfig}
                      disabled={loading || !authConfig}
                    >
                      <Trash2 size={15} />
                      Delete config
                    </Button>
                    <Button onClick={saveConfig} disabled={loading}>
                      {loading && <Spinner />}
                      Save Auth Config
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
            <div className={"grid gap-5 xl:grid-cols-2"}>
              <Card>
                <CardHeader>
                  <CardTitle className={"flex items-center gap-2"}>
                    <KeyRound size={18} />
                    Database Keys
                  </CardTitle>
                  <CardDescription>
                    Generate or import RS256 signing keys when the backend is
                    set to database.
                  </CardDescription>
                </CardHeader>
                <CardContent className={"flex flex-col gap-5"}>
                  <FieldGroup className={"grid gap-4 lg:grid-cols-3"}>
                    <Field>
                      <FieldLabel>Key ID</FieldLabel>
                      <Input
                        value={keyID}
                        placeholder={"route-key-v1"}
                        onChange={(event) => setKeyID(event.target.value)}
                      />
                    </Field>
                    <Field>
                      <FieldLabel>Expires in days</FieldLabel>
                      <Input
                        type={"number"}
                        value={expiresInDays}
                        onChange={(event) =>
                          setExpiresInDays(event.target.value)
                        }
                      />
                    </Field>
                    <Field orientation={"horizontal"} className={"pt-6"}>
                      <Checkbox
                        checked={primaryKey}
                        onCheckedChange={(checked) =>
                          setPrimaryKey(checked === true)
                        }
                      />
                      <FieldContent>
                        <FieldLabel>Primary key</FieldLabel>
                      </FieldContent>
                    </Field>
                  </FieldGroup>
                  <div className={"flex flex-wrap gap-2"}>
                    <Button
                      onClick={generateKey}
                      disabled={loading || !keyID || !routeData.data}
                    >
                      {loading && <Spinner />}
                      Generate key
                    </Button>
                    <Button
                      variant={"outline"}
                      onClick={importKey}
                      disabled={
                        loading || !keyID || !privateKeyPEM || !publicKeyPEM
                      }
                    >
                      Import PEM key
                    </Button>
                  </div>
                  <FieldGroup className={"grid gap-4 lg:grid-cols-2"}>
                    <Field>
                      <FieldLabel>Private key PEM</FieldLabel>
                      <Textarea
                        className={"min-h-32 font-mono text-xs"}
                        value={privateKeyPEM}
                        onChange={(event) =>
                          setPrivateKeyPEM(event.target.value)
                        }
                      />
                    </Field>
                    <Field>
                      <FieldLabel>Public key PEM</FieldLabel>
                      <Textarea
                        className={"min-h-32 font-mono text-xs"}
                        value={publicKeyPEM}
                        onChange={(event) =>
                          setPublicKeyPEM(event.target.value)
                        }
                      />
                    </Field>
                  </FieldGroup>
                  {generatedKey?.private_key_pem && (
                    <Field>
                      <FieldLabel>Generated private key</FieldLabel>
                      <Textarea
                        readOnly
                        className={"min-h-32 font-mono text-xs"}
                        value={generatedKey.private_key_pem}
                      />
                      <FieldDescription>
                        The API returns generated private key material only
                        once.
                      </FieldDescription>
                    </Field>
                  )}
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Active Key Material</CardTitle>
                  <CardDescription>
                    Keys are encrypted at rest by the configured gateway
                    provider.
                  </CardDescription>
                </CardHeader>
                <CardContent className={"flex flex-col gap-3"}>
                  {keysData.loading ? (
                    <div className={"flex justify-center py-10"}>
                      <Spinner />
                    </div>
                  ) : (keysData.data ?? []).length === 0 ? (
                    <p className={"text-sm text-muted-foreground"}>
                      No keys found for this route.
                    </p>
                  ) : (
                    keysData.data?.map((key) => (
                      <div
                        key={key.id}
                        className={
                          "flex items-center justify-between rounded-lg border p-3"
                        }
                      >
                        <div className={"flex flex-col gap-1"}>
                          <div className={"flex items-center gap-2"}>
                            <span className={"font-mono text-sm"}>
                              {key.key_id}
                            </span>
                            <Badge
                              variant={key.is_active ? "default" : "outline"}
                            >
                              {key.is_active ? "active" : "inactive"}
                            </Badge>
                            {key.metadata?.primary === true && (
                              <Badge variant={"outline"}>primary</Badge>
                            )}
                          </div>
                          <span className={"text-xs text-muted-foreground"}>
                            {key.algorithm} · expires{" "}
                            {key.expires_at
                              ? new Date(key.expires_at).toLocaleString()
                              : "never"}
                          </span>
                        </div>
                        <Button
                          variant={"ghost"}
                          size={"icon"}
                          onClick={() => deleteKey(key)}
                          disabled={loading}
                        >
                          <Trash2 size={15} />
                        </Button>
                      </div>
                    ))
                  )}
                  <div className={"mt-5 flex flex-row justify-end gap-5"}>
                    <div
                      className={
                        "flex flex-row items-center justify-center gap-2"
                      }
                    >
                      <p className={"text-sm font-semibold"}>Rows per page</p>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant={"outline"}
                            disabled={keysData.loading}
                          >
                            {perPage}
                            <ChevronDownIcon />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent>
                          <DropdownMenuRadioGroup
                            value={perPage}
                            onValueChange={setPerPage}
                          >
                            <DropdownMenuRadioItem value={"5"}>
                              5
                            </DropdownMenuRadioItem>
                            <DropdownMenuRadioItem value={"10"}>
                              10
                            </DropdownMenuRadioItem>
                            <DropdownMenuRadioItem value={"20"}>
                              20
                            </DropdownMenuRadioItem>
                            <DropdownMenuRadioItem value={"30"}>
                              30
                            </DropdownMenuRadioItem>
                            <DropdownMenuRadioItem value={"40"}>
                              40
                            </DropdownMenuRadioItem>
                          </DropdownMenuRadioGroup>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                    <div
                      className={
                        "flex flex-row items-center justify-center gap-2"
                      }
                    >
                      <Button
                        variant={"outline"}
                        size={"icon"}
                        disabled={
                          keysData.loading || keysData.previousPageToken === ""
                        }
                        onClick={() => keysData.previousPage()}
                      >
                        <ChevronLeft />
                      </Button>
                      <Button
                        variant={"outline"}
                        size={"icon"}
                        disabled={
                          keysData.loading || keysData.nextPageToken === ""
                        }
                        onClick={() => keysData.nextPage()}
                      >
                        <ChevronRight />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
            <Card>
              <CardHeader>
                <CardTitle>Validate Token</CardTitle>
                <CardDescription>
                  Verify a JWT against this route auth config and inspect the
                  claims that plugins can read after native auth succeeds.
                </CardDescription>
              </CardHeader>
              <CardContent className={"flex flex-col gap-4"}>
                <Textarea
                  className={"min-h-28 font-mono text-xs"}
                  value={tokenToValidate}
                  placeholder={"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."}
                  onChange={(event) => setTokenToValidate(event.target.value)}
                />
                <div>
                  <Button
                    onClick={validateToken}
                    disabled={loading || !tokenToValidate}
                  >
                    Validate token
                  </Button>
                </div>
                {validationResult && (
                  <pre
                    className={"overflow-auto rounded-lg bg-muted p-4 text-xs"}
                  >
                    {JSON.stringify(validationResult, null, 2)}
                  </pre>
                )}
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </SidebarLayout>
  )
}

export default function RouteAuthPage() {
  return (
    <React.Suspense
      fallback={
        <SidebarLayout page_title={"Route Auth"}>
          <div className={"flex items-center justify-center p-6"}>
            <Spinner className={"h-10 w-10"} />
          </div>
        </SidebarLayout>
      }
    >
      <RouteAuthPageContent />
    </React.Suspense>
  )
}
