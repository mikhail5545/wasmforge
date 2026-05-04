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
import { DocSection, InfoCard, PayloadTable } from "@/components/doc"
import { SidebarLayout } from "@/components/navigation/sidebar-layout"
import { Badge } from "@workspace/ui/components/badge"

const rustImportsExample = `#[link(wasm_import_module = "env")]
unsafe extern "C" {
    fn host_get_header(k_ptr: *const u8, k_size: u32, b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_set_header(k_ptr: *const u8, k_size: u32, v_ptr: *const u8, v_size: u32);
    fn host_get_method(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_path(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_query_param(k_ptr: *const u8, k_size: u32, b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_raw_query(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_json_config(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_send_response(status_code: u32, body_ptr: *const u8, body_len: u32);
    fn host_log(level: u32, msg_ptr: *const u8, msg_len: u32);
    fn host_auth_is_authenticated() -> u32;
    fn host_auth_subject(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_auth_claim(k_ptr: *const u8, k_size: u32, b_ptr: *mut u8, b_size: u32) -> u32;
}

#[no_mangle]
pub extern "C" fn on_request() {
    // required export called by the gateway
}`

const sdkContractRust = `pub mod proxy {
    pub enum LogLevel { DEBUG, INFO, WARN, ERROR }
    pub const NOT_FOUND: u32 = 0xFFFF_FFFF;

    pub fn get_header(key: &str) -> Option<String>;
    pub fn get_method() -> Option<String>;
    pub fn get_path() -> Option<String>;
    pub fn get_query_param(key: &str) -> Option<String>;
    pub fn get_raw_query() -> Option<String>;
    pub fn send_response(status_code: u32, msg: &str);
    pub fn log(level: LogLevel, msg: &str);
    pub fn set_header(key: &str, value: &str);
    pub fn get_json_config() -> Option<String>;
    pub fn is_authenticated() -> bool;
    pub fn get_auth_subject() -> Option<String>;
    pub fn get_auth_claim(key: &str) -> Option<String>;
}

#[macro_export]
macro_rules! proxy_handler { ($handler: path) => { ... } }`

const sdkContractCpp = `namespace Proxy {
    enum LogLevel { DEBUG, INFO, WARN, ERROR };
    const uint32_t NOT_FOUND = 0xFFFFFFFF;

    std::string get_header(const std::string& key);
    std::string get_method();
    std::string get_path();
    std::string get_query_param(const std::string& key);
    std::string get_raw_query();
    void send_response(uint32_t status_code, const std::string& response);
    void log(LogLevel level, const std::string& msg);
    void set_header(const std::string& key, const std::string& value);
    std::string get_json_config();
    bool is_authenticated();
    std::string get_auth_subject();
    std::string get_auth_claim(const std::string& key);
}

#define PROXY_PLUGIN(HandlerFunc) ...`

const requestExampleRust = `use wasmforge_sdk::{proxy, proxy_handler};

proxy_handler!(on_request);

pub fn on_request() {
    let path = proxy::get_path().unwrap_or_default();
    let method = proxy::get_method().unwrap_or_default();

    proxy::log(proxy::LogLevel::INFO, &format!("{} {}", method, path));

    if path == "/admin" {
        proxy::send_response(403, "Forbidden area");
    }
    
    proxy::set_header("X-Wasm-Filtered", "true");
}`

const requestExampleCpp = `#include <wasmforge_sdk.h>

void on_plugin_load() {
    auto path = Proxy::get_path();
    auto method = Proxy::get_method();

    Proxy::log(Proxy::INFO, method + " " + path);

    if (path == "/admin") {
        Proxy::send_response(403, "Forbidden area");
    }

    Proxy::set_header("X-Wasm-Filtered", "true");
}

PROXY_PLUGIN(on_plugin_load)`

const authExampleRust = `use wasmforge_sdk::{proxy, proxy_handler};

proxy_handler!(on_request);

pub fn on_request() {
    if !proxy::is_authenticated() {
        proxy::send_response(401, "Please log in");
        return;
    }

    if let Some(sub) = proxy::get_auth_subject() {
        proxy::log(proxy::LogLevel::INFO, &format!("User: {}", sub));
    }

    if let Some(role) = proxy::get_auth_claim("role") {
        if role != "admin" {
            proxy::send_response(403, "Admin only");
        }
    }
}`

const authExampleCpp = `#include <wasmforge_sdk.h>

void on_plugin_load() {
    if (!Proxy::is_authenticated()) {
        Proxy::send_response(401, "Please log in");
        return;
    }

    auto sub = Proxy::get_auth_subject();
    if (!sub.empty()) {
        Proxy::log(Proxy::INFO, "User: " + sub);
    }

    auto role = Proxy::get_auth_claim("role");
    if (role != "admin") {
        Proxy::send_response(403, "Admin only");
    }
}

PROXY_PLUGIN(on_plugin_load)`

const configExampleRust = `use wasmforge_sdk::{proxy, proxy_handler};
use serde_json::Value;

proxy_handler!(on_request);

pub fn on_request() {
    let config_raw = proxy::get_json_config().unwrap_or_else(|| "{}".to_string());
    let config: Value = serde_json::from_str(&config_raw).unwrap();

    if let Some(enabled) = config.get("enabled").and_then(|v| v.as_bool()) {
        if !enabled {
            return;
        }
    }
    // ... logic
}`

const configExampleCpp = `#include <wasmforge_sdk.h>
// Use a JSON library like nlohmann/json

void on_plugin_load() {
    std::string config_raw = Proxy::get_json_config();
    // parse config_raw...
}

PROXY_PLUGIN(on_plugin_load)`

export default function WasmExportsApiReferencePage() {
  return (
    <SidebarLayout page_title={"WASM Host API Reference"}>
      <main
        className={
          "mx-auto flex w-full max-w-5xl flex-col gap-8 px-5 py-6 md:px-10 lg:px-12"
        }
      >
        <header className={"flex flex-col gap-3"}>
          <Badge variant={"outline"} className={"w-fit"}>
            Reference
          </Badge>
          <h1 className={"text-4xl font-bold tracking-tight"}>
            WASM Host Imports
          </h1>
          <p className={"max-w-3xl text-muted-foreground"}>
            Gateway plugins import functions from the
            <code className={"rounded bg-muted px-1 font-mono"}> env</code>
            module and must export
            <code className={"rounded bg-muted px-1 font-mono"}>
              {" "}
              on_request
            </code>
            . These imports are the only supported boundary between the WASM
            guest and the WasmForge request context.
          </p>
        </header>

        <DocSection
          id={"contract"}
          title={"Contract"}
          description={
            "Host functions exchange UTF-8 data through guest memory pointers and lengths."
          }
        >
          <div className={"grid gap-4 md:grid-cols-3"}>
            <InfoCard title={"Pointers"}>
              <code>*_ptr</code> and <code>*_len</code> identify guest memory
              ranges. Invalid memory reads or writes return an error sentinel or
              no-op depending on the function.
            </InfoCard>
            <InfoCard title={"Return Values"}>
              Buffer-returning functions return bytes written. They return
              <code> 0xFFFFFFFF</code> when context data is missing or memory
              access fails.
            </InfoCard>
            <InfoCard title={"Truncation"}>
              If the output is larger than <code>buf_len</code>, the host writes
              as much as fits and returns the truncated byte count.
            </InfoCard>
          </div>
          <CodeBlock language={"rust"} code={rustImportsExample} />
        </DocSection>

        <DocSection
          id={"sdk"}
          title={"SDK Overview"}
          description={
            "WasmForge provides idiomatic wrappers for Rust and C++ to handle memory management and buffer safety automatically."
          }
        >
          <div className={"grid gap-4 md:grid-cols-2 mb-4"}>
            <InfoCard title={"Rust SDK"}>
              Available as a crate. Uses <code>Option&lt;String&gt;</code> for safe retrieval and includes a <code>proxy_handler!</code> macro to export the entry point.
            </InfoCard>
            <InfoCard title={"C++ SDK"}>
              A single-header SDK using <code>namespace Proxy</code>. Provides <code>std::string</code> based wrappers and a <code>PROXY_PLUGIN</code> macro.
            </InfoCard>
          </div>
          <CodeBlock tabs={ [
            { label: "Rust SDK", language: "rust", code: sdkContractRust },
            { label: "C++ SDK", language: "cpp", code: sdkContractCpp }
          ] } />
        </DocSection>

        <DocSection
          id={"request"}
          title={"Request & Metadata"}
          description={
            "Inspect and modify the incoming HTTP request before it reaches the upstream."
          }
        >
          <PayloadTable
            data={[
              {
                property: "get_header(key)",
                type: "Option<String> / std::string",
                description: "Retrieves a request header. Returns empty if missing.",
              },
              {
                property: "set_header(key, val)",
                type: "void",
                description: "Adds or overwrites a request header for the upstream.",
              },
              {
                property: "get_method()",
                type: "String / std::string",
                description: "The HTTP verb (GET, POST, etc).",
              },
              {
                property: "get_path()",
                type: "String / std::string",
                description: "The URL path without query parameters.",
              },
            ]}
          />
          <CodeBlock tabs={ [
            { label: "Rust", language: "rust", code: requestExampleRust },
            { label: "C++", language: "cpp", code: requestExampleCpp }
          ] } />
        </DocSection>

        <DocSection
          id={"auth"}
          title={"Auth Context"}
          description={
            "Access data from the native Authentication middleware without parsing tokens manually."
          }
        >
          <p className={"mb-4 text-muted-foreground"}>
            WasmForge securely propagates authenticated identity and claims to your plugins. This allows you to implement RBAC or personalized logic without exposing signing secrets to WASM.
          </p>
          <PayloadTable
            data={[
              {
                property: "is_authenticated()",
                type: "bool",
                description: "True if the request was successfully authenticated by the gateway.",
              },
              {
                property: "get_auth_subject()",
                type: "Option<String> / std::string",
                description: "The 'sub' claim of the validated token.",
              },
              {
                property: "get_auth_claim(key)",
                type: "Option<String> / std::string",
                description: "Retrieves any claim from the token. Complex types are JSON encoded.",
              },
            ]}
          />
          <CodeBlock tabs={ [
            { label: "Rust", language: "rust", code: authExampleRust },
            { label: "C++", language: "cpp", code: authExampleCpp }
          ] } />
        </DocSection>

        <DocSection
          id={"config"}
          title={"Configuration"}
          description={
            "Plugins can be configured per-route via JSON."
          }
        >
          <p className={"mb-4 text-muted-foreground"}>
            When binding a plugin to a route, you can provide a JSON configuration string. This is accessible within your plugin via <code>get_json_config()</code>.
          </p>
          <CodeBlock tabs={ [
            { label: "Rust", language: "rust", code: configExampleRust },
            { label: "C++", language: "cpp", code: configExampleCpp }
          ] } />
        </DocSection>

        <DocSection
          id={"best-practices"}
          title={"Best Practices"}
          description={"Guidelines for building performant and secure plugins."}
        >
          <div className={"grid gap-6 md:grid-cols-2"}>
            <InfoCard title={"Minimize Allocations"}>
              WASM memory is fast, but frequent allocations can trigger GC or fragmentation. Reuse buffers where possible for high-throughput routes.
            </InfoCard>
            <InfoCard title={"Error Handling"}>
              Always check if a header or claim exists before using it. WasmForge SDKs return <code>None</code> or empty strings rather than panicking.
            </InfoCard>
            <InfoCard title={"Logging Verbosity"}>
              Use <code>LogLevel::DEBUG</code> for development. Excessive logging in production can impact gateway throughput.
            </InfoCard>
            <InfoCard title={"Binary Size"}>
              Strip symbols and use LTO when compiling your Rust or C++ plugins to minimize the cold-start time and memory footprint.
            </InfoCard>
          </div>
        </DocSection>
      </main>
    </SidebarLayout>
  )
}
