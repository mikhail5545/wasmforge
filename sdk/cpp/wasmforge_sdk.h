// Copyright (c) 2026. Mikhail Kulik.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef WASMFORGE_SDK_H
#define WASMFORGE_SDK_H
#include <cstdint>
#include <string>
#include <stdexcept>
#include <vector>

extern "C"{
    // Pointers in WASM are 32-bit integers (uint32_t)
    uint32_t host_get_header(uint32_t k_ptr, uint32_t k_len, uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_method(uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_path(uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_query_param(uint32_t k_ptr, uint32_t k_len, uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_raw_query(uint32_t b_ptr, uint32_t b_len);
    void host_send_response(uint32_t status_code, uint32_t b_ptr, uint32_t b_len);
    void host_log(uint32_t level, uint32_t msg_ptr, uint32_t msg_len);
    void host_set_header(uint32_t k_ptr, uint32_t k_len, uint32_t v_ptr, uint32_t v_len);
}

enum LogLevel {
    DEBUG, INFO, WARN, ERROR
};

class Proxy {
public:
    static std::string get_header(const std::string& key) {
        std::vector<char> buffer(1024);

        const uint32_t wrote = host_get_header(
            reinterpret_cast<uintptr_t>(key.c_str()),
            key.size(),
            reinterpret_cast<uintptr_t>(buffer.data()),
            buffer.size()
        );
        if (wrote == 0xFFFFFF) {
            throw std::runtime_error("host_get_header call failed");
        }
        return {buffer.data(), wrote};
    }

    static std::string get_method() {
        std::vector<char> buffer(1024);

        const uint32_t wrote = host_get_method(reinterpret_cast<uintptr_t>(buffer.data()), 1024);
        if (wrote == 0xFFFFFF) {
            throw std::runtime_error("host_get_method call failed");
        }
        return {buffer.data(), wrote};
    }

    static std::string get_path() {
        std::vector<char> buffer(1024);

        const uint32_t wrote = host_get_path(reinterpret_cast<uintptr_t>(buffer.data()), buffer.size());
        if (wrote == 0xFFFFFF) {
            throw std::runtime_error("host_get_path call failed");
        }
        return {buffer.data(), wrote};
    }

    static std::string get_query_param(const std::string& key) {
        std::vector<char> buffer(1024);

        const uint32_t wrote = host_get_query_param(
            reinterpret_cast<uintptr_t>(key.c_str()),
            key.size(),
            reinterpret_cast<uintptr_t>(buffer.data()),
            buffer.size()
        );
        if (wrote == 0xFFFFFF) {
            throw std::runtime_error("host_get_query_param call failed");
        }
        return {buffer.data(), wrote};
    }

    static std::string get_raw_query() {
        std::vector<char> buffer(1024);

        const uint32_t wrote = host_get_raw_query(reinterpret_cast<uintptr_t>(buffer.data()), buffer.size());
        if (wrote == 0xFFFFFF) {
            throw std::runtime_error("host_get_raw_query call failed");
        }
        return {buffer.data(), wrote};
    }

    static void send_response(const uint32_t status_code, const std::string& response) {
        host_send_response(status_code, reinterpret_cast<uintptr_t>(response.data()), response.size());
    }

    static void log(const LogLevel log_level, const std::string& msg) {
        host_log(log_level, reinterpret_cast<uintptr_t>(msg.data()), msg.size());
    }

    static void set_header(const std::string& key, const std::string& value) {
        host_set_header(
            reinterpret_cast<uintptr_t>(key.data()),
            key.size(),
            reinterpret_cast<uintptr_t>(value.data()),
            value.size()
            );
    }
};

// The Entry Point Macro
#define PROXY_PLUGIN(HandlerFunc)\
    extern "C" __attribute__((visibility("default"))) void _start() {\
        HandlerFunc();\
    }

#endif //WASMFORGE_SDK_H