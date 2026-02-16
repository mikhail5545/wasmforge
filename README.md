# WasmForge: Customizable WebAssembly API Gateway 

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golanng.org)
![License](https://img.shields.io/badge/license-Apache_2.0-green)
![WASM](https://img.shields.io/badge/WASM-Powered-654FF0?style=flat&logo=webassembly)
![React](https://img.shields.io/badge/Frontend-Next.js-black?style=flat&logo=next.js)

A high-performance, extensible API Gateway written in **Go**. It routes traffic to your microservices while applying dynamic policies (Rate Limiting, Auth, Transformation) using **WebAssembly (WASM)** modules.

**The Killer Feature:** You can write plugins in **C++, Rust, Python, or Go**, compile them to WASM, and hot-swap them into the running gateway without restarting the server.

## 🚀 Features

*   **⚡️ High Performance Data Plane:** Built on Go's standard `net/http` with atomic router swapping for zero-downtime reconfiguration.
*   **🔌 Polyglot Plugins:** Embeds the [wazero](https://wazero.io/) runtime to execute untrusted code safely.
*   **⚛️ Embedded Dashboard:** A full React/Next.js Admin UI baked into a single binary.
*   **💾 Persistent Config:** SQLite + GORM stores routes and plugins, surviving restarts.
*   **🛡️ Secure & Sandboxed:** Plugins run in a memory-safe sandbox and cannot crash the main gateway process.

## 🏗 Architecture

The system is divided into two distinct planes:

1.  **Data Plane (The Proxy):**
    *   Handles high-throughput HTTP traffic.
    *   Executes WASM middleware chains using a zero-dependency runtime.
    *   Uses an **Atomic Swapping** strategy to update routes without locking.

2.  **Control Plane (The UI):**
    *   Exposes a REST API for management.
    *   Serves a Next.js Single Page App (SPA) via Go's `embed` package.
    *   Manages the SQLite database for configuration persistence.

## 🛠 Tech Stack

*   **Core:** Go (Golang) 1.22+
*   **WASM Runtime:** [wazero](https://github.com/tetratelabs/wazero)
*   **Database:** SQLite (Pure Go via `glebarez/go-sqlite`) + GORM
*   **Frontend:** Next.js (TypeScript) + TailwindCSS
*   **API Framework:** Echo (for the Admin API)

## 📦 Quick Start

### Prerequisites
*   Go 1.22+
*   Node.js 18+ (for building the UI)
*   Make

### Installation

1.  **Clone the repository**
    ```bash
    git clone https://github.com/yourusername/go-wasm-gateway.git
    cd go-wasm-gateway
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
    *   **Proxy:** Listening on `:8080`
    *   **Admin Dashboard:** Listening on `:9090`

4.  **Open the Dashboard**
    Visit `http://localhost:9090` to configure routes and upload plugins.

## 🔌 Writing a Plugin (C++ Example)

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