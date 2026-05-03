#[link(wasm_import_module = "env")]
unsafe extern "C" {
    fn host_get_header(k_ptr: *const u8, k_size: u32, b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_method(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_path(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_query_param(k_ptr: *const u8, k_size: u32, b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_get_raw_query(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_send_response(status_code: u32, msg_ptr: *const u8, msg_size: u32);
    fn host_log(level: u32, msg_ptr: *const u8, msg_size: u32);
    fn host_set_header(k_ptr: *const u8, k_size: u32, v_ptr: *const u8, v_size: u32);
    fn host_get_json_config(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_auth_is_authenticated() -> u32;
    fn host_auth_subject(b_ptr: *mut u8, b_size: u32) -> u32;
    fn host_auth_claim(k_ptr: *const u8, k_size: u32, b_ptr: *mut u8, b_size: u32) -> u32;
}

pub mod proxy {
    use super::*;

    #[derive(Clone, Copy)]
    pub enum LogLevel {
        DEBUG = 0,
        INFO = 1,
        WARN = 2,
        ERROR = 3,
    }

    pub const NOT_FOUND: u32 = 0xFFFF_FFFF;

    fn get_string_from_host<F>(f: F) -> Option<String> 
    where F: FnOnce(*mut u8, u32) -> u32 {
        let mut buf = [0u8; 4096];
        let wrote = f(buf.as_mut_ptr(), 4096);
        if wrote == NOT_FOUND {
            return None;
        }
        Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
    }

    pub fn get_header(key: &str) -> Option<String> {
        let mut buf = [0u8; 4096];
        let wrote = unsafe { host_get_header(key.as_ptr(), key.len() as u32, buf.as_mut_ptr(), 4096) };
        if wrote == NOT_FOUND {
            return None;
        }
        Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
    }

    pub fn get_method() -> Option<String> { unsafe { get_string_from_host(|ptr, len| host_get_method(ptr, len)) } }
    pub fn get_path() -> Option<String> { unsafe { get_string_from_host(|ptr, len| host_get_path(ptr, len)) } }

    pub fn get_query_param(key: &str) -> Option<String> {
        let mut buf = [0u8; 4096];
        let wrote = unsafe { host_get_query_param(key.as_ptr(), key.len() as u32, buf.as_mut_ptr(), 4096) };
        if wrote == NOT_FOUND {
            return None;
        }
        Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
    }

    pub fn get_raw_query() -> Option<String> { unsafe { get_string_from_host(|ptr, len| host_get_raw_query(ptr, len)) } }

    pub fn send_response(status_code: u32, msg: &str) {
        unsafe { host_send_response(status_code, msg.as_ptr(), msg.len() as u32); }
    }

    pub fn log(level: LogLevel, msg: &str) {
        unsafe { host_log(level as u32, msg.as_ptr(), msg.len() as u32); }
    }

    pub fn set_header(key: &str, value: &str) {
        unsafe { host_set_header(key.as_ptr(), key.len() as u32, value.as_ptr(), value.len() as u32); }
    }

    pub fn get_json_config() -> Option<String> { unsafe { get_string_from_host(|ptr, len| host_get_json_config(ptr, len)) } }

    pub fn is_authenticated() -> bool { unsafe { host_auth_is_authenticated() != 0 } }

    pub fn get_auth_subject() -> Option<String> { unsafe { get_string_from_host(|ptr, len| host_auth_subject(ptr, len)) } }

    pub fn get_auth_claim(key: &str) -> Option<String> {
        let mut buf = [0u8; 4096];
        let wrote = unsafe { host_auth_claim(key.as_ptr(), key.len() as u32, buf.as_mut_ptr(), 4096) };
        if wrote == NOT_FOUND {
            return None;
        }
        Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
    }
}

#[macro_export]
macro_rules! proxy_handler {
    ($handler: path) => {
        #[no_mangle]
        pub extern "C" fn on_request() {
            $handler();
        }
    };
}
