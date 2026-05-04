use wasmforge_sdk::{proxy, proxy_handler};

proxy_handler!(on_request);

const HEADER_NAME: &str = "x-api-key";
const EXPECTED: &str = "bench-secret";

pub fn on_request() {
    let header = proxy::get_header(HEADER_NAME);
    
    match header {
        Some(val) if val == EXPECTED => {
            // Authorized
        }
        _ => {
            proxy::send_response(401, "unauthorized");
        }
    }
}
