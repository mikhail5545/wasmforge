# WasmForge: Customizable WASM-powered API Gateway 

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golanng.org)
[![License](https://img.shields.io/badge/license-Apache_2.0-green)](https://www.apache.org/licenses/LICENSE-2.0)
[![WASM](https://img.shields.io/badge/WASM-Powered-654FF0?style=flat&logo=webassembly)](https://webassembly.org/)
[![Next.js](https://img.shields.io/badge/Frontend-Next.js-black?style=flat&logo=next.js)](https://nextjs.org/)

A high-performance, extensible API Gateway written in **Go**. It routes traffic to your microservices while applying dynamic policies (Rate Limiting, Auth, Transformation) using **WebAssembly (WASM)** modules.

**Customizable:** You can write plugins in **C++, Rust, Python, or Go**, compile them to WASM, and hot-swap them into the running gateway without restarting the server.

## Features
*   **High Performance Routing:** Build on Go's standard `net/http` with an atomic router swapping mechanism for zero-downtime updates.
*   **WASM Middleware:** Execute custom logic in a secure, sandboxed environment. Perfect for rate limiting, authentication, request/response transformation, and more.
*   **Dynamic Configuration:** Manage routes and plugins via a user-friendly Next.js dashboard or REST API.
*   **GUI:** Clean Next.js/React dashboard for management baked into the Go binary.
*   **Persistent Config:** SQLite + GORM stores routes, files, plugins, and configurations surviving restarts.
*   **Secure & Sandboxed:** Plugins run in a memory-safe sandbox and cannot crash the main gateway process.

## Architecture

The system is divided into two distinct parts:

1.  **Data Plane (The Proxy):**
    *   Handles high-throughput HTTP traffic.
    *   Executes WASM middleware chains using a zero-dependency runtime.
    *   Uses an **Atomic Swapping** strategy to update routes without locking.

2.  **Control Plane (The UI):**
    *   Exposes a REST API for management.
    *   Serves a Next.js Single Page App (SPA) via Go's `embed` package.
    *   Manages the SQLite database for configuration persistence.

## Tech Stack

*   **Core:** Go (Golang) 1.25
*   **WASM Runtime:** [wazero](https://github.com/tetratelabs/wazero)
*   **Database:** SQLite + GORM
*   **Frontend:** Next.js (TypeScript) + TailwindCSS
*   **API Framework:** Echo (for the Admin API)

## Quick Start

### Prerequisites
*   Go 1.25
*   Node.js 18+ (for building the UI)
*   Make

### Installation

1.  **Clone the repository**
    ```bash
    git clone https://github.com/mikhail5545/wasmforge.git
    cd wasmforge
    ```

2.  **Build the Project**
    This command builds the Next.js UI, embeds it into the Go binary, and compiles the gateway.
    ```bash
    make build
    ```

3.  **Run the Gateway**
    ```bash
    ./bin/gateway
    ```
    *   **Proxy:** Listening on `:8000`
    *   **Admin Dashboard:** Listening on `:8080`

4.  **Open the Dashboard**
    Visit `http://localhost:8080` to configure routes and upload plugins.

### Building options

*   `make build` - Builds the entire project (UI + Go binary).
*   `make build-ui` - Builds the Next.js UI as static files.
*   `make build-go` - Compiles the Go gateway without rebuilding the UI.
*   `make clean` - Cleans the build artifacts.
*   `make npm-run` - Runs the npm build for the UI without affecting the Go build.
*   `run-separate` - Runs the UI build and Go build in parallel, ensuring the UI is ready before the Go build starts. Default ports are :3000 for the UI and :8080 for the Go server.

## Writing a Plugin (C++ Example)

You can write plugins in any language that targets WASI. Here is a simple C++ example that blocks requests missing a header.

```cpp
#include "proxy_sdk.h"

void on_request() {
    // 1. Get Configuration from Gateway
    auto config = Proxy::getConfig(); 
    
    // 2. Check for Header
    std::string secret = Proxy::getHeader("X-Secret-Token");
    
    if (secret != "super-secret") {
        Proxy::log(WARN, "Unauthorized access attempt");
        Proxy::sendResponse(403, "Access Denied");
        return;
    }
    
    // 3. Allow request to proceed
    Proxy::log(INFO, "Access granted");
}