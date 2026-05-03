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

extern "C" {
    uint32_t host_get_header(uint32_t k_ptr, uint32_t k_len, uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_method(uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_path(uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_query_param(uint32_t k_ptr, uint32_t k_len, uint32_t b_ptr, uint32_t b_len);
    uint32_t host_get_raw_query(uint32_t b_ptr, uint32_t b_len);
    void host_send_response(uint32_t status_code, uint32_t b_ptr, uint32_t b_len);
    void host_log(uint32_t level, uint32_t msg_ptr, uint32_t msg_len);
    void host_set_header(uint32_t k_ptr, uint32_t k_len, uint32_t v_ptr, uint32_t v_len);
    uint32_t host_get_json_config(uint32_t b_ptr, uint32_t b_len);
    uint32_t host_auth_is_authenticated();
    uint32_t host_auth_subject(uint32_t b_ptr, uint32_t b_len);
    uint32_t host_auth_claim(uint32_t k_ptr, uint32_t k_len, uint32_t b_ptr, uint32_t b_len);
}

namespace Proxy {
    enum LogLevel {
        DEBUG = 0, INFO = 1, WARN = 2, ERROR = 3
    };

    const uint32_t NOT_FOUND = 0xFFFFFFFF;

    inline std::string get_string_from_host(uint32_t (*func)(uint32_t, uint32_t)) {
        std::vector<char> buffer(4096);
        uint32_t wrote = func(reinterpret_cast<uintptr_t>(buffer.data()), buffer.size());
        if (wrote == NOT_FOUND) return "";
        return {buffer.data(), wrote};
    }

    inline std::string get_header(const std::string& key) {
        std::vector<char> buffer(4096);
        uint32_t wrote = host_get_header(
            reinterpret_cast<uintptr_t>(key.c_str()), key.size(),
            reinterpret_cast<uintptr_t>(buffer.data()), buffer.size()
        );
        if (wrote == NOT_FOUND) return "";
        return {buffer.data(), wrote};
    }

    inline std::string get_method() { return get_string_from_host(host_get_method); }
    inline std::string get_path() { return get_string_from_host(host_get_path); }

    inline std::string get_query_param(const std::string& key) {
        std::vector<char> buffer(4096);
        uint32_t wrote = host_get_query_param(
            reinterpret_cast<uintptr_t>(key.c_str()), key.size(),
            reinterpret_cast<uintptr_t>(buffer.data()), buffer.size()
        );
        if (wrote == NOT_FOUND) return "";
        return {buffer.data(), wrote};
    }

    inline std::string get_raw_query() { return get_string_from_host(host_get_raw_query); }

    inline void send_response(uint32_t status_code, const std::string& response) {
        host_send_response(status_code, reinterpret_cast<uintptr_t>(response.data()), response.size());
    }

    inline void log(LogLevel level, const std::string& msg) {
        host_log(level, reinterpret_cast<uintptr_t>(msg.data()), msg.size());
    }

    inline void set_header(const std::string& key, const std::string& value) {
        host_set_header(reinterpret_cast<uintptr_t>(key.data()), key.size(), reinterpret_cast<uintptr_t>(value.data()), value.size());
    }

    inline std::string get_json_config() { return get_string_from_host(host_get_json_config); }

    inline bool is_authenticated() { return host_auth_is_authenticated() != 0; }

    inline std::string get_auth_subject() { return get_string_from_host(host_auth_subject); }

    inline std::string get_auth_claim(const std::string& key) {
        std::vector<char> buffer(4096);
        uint32_t wrote = host_auth_claim(
            reinterpret_cast<uintptr_t>(key.c_str()), key.size(),
            reinterpret_cast<uintptr_t>(buffer.data()), buffer.size()
        );
        if (wrote == NOT_FOUND) return "";
        return {buffer.data(), wrote};
    }
}

#define PROXY_PLUGIN(HandlerFunc)\
    extern "C" __attribute__((visibility("default"))) void on_request() {\
        HandlerFunc();\
    }

#endif
