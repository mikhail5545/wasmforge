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

## Not all features are fully implemented yet