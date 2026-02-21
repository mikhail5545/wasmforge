use String;

#[link(wasm_import_module = "env")]
unsafe extern "C" {
    fn host_get_header(k_ptr: *const u8, k_size: i32, b_ptr: *const u8, b_size: i32) -> i32;
    fn host_get_method(b_ptr: *const u8, b_size: i32) -> i32;
    fn host_get_path(b_ptr: *const u8, b_size: i32) -> i32;
    fn host_get_query_param(k_ptr: *const u8, k_size: i32, b_ptr: *const u8, b_size: i32) -> i32;
    fn host_get_raw_query(b_ptr: *const u8, b_size: i32) -> i32;
    fn host_send_response(status_code: i32, msg_ptr: *const u8, msg_size: i32);
    fn host_log(level: i32, msg_ptr: *const u8, msg_size: i32);
    fn host_set_header(k_ptr: *const u8, k_size: i32, v_ptr: *const u8, v_size: i32);
}

pub mod proxy{
    use super::*;

    pub enum LogLevel{
        DEBUG, INFO, WARN, ERROR
    }

    pub fn get_header(key: &str) -> Option<String> {
        let mut buf = [0u8;1024];
        unsafe{
            let wrote = host_get_header(
                key.as_ptr(),
                key.len() as i32,
                buf.as_mut_ptr(),
                1024
            );
            if wrote == 0xFFFFFF{
                return None;
            }
            Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
        }
    }

    pub fn get_method() -> Option<String> {
        let mut buf = [0u8;1024];
        unsafe{
            let wrote = host_get_method(buf.as_mut_ptr(), 1024);
            if wrote == 0xFFFFFF{
                return None;
            }
            Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
        }
    }

    pub fn get_path() -> Option<String> {
        let mut buf= [0u8;1024];
        unsafe{
            let wrote = host_get_path(buf.as_mut_ptr(), 1024);
            if wrote == 0xFFFFFF{
                return None;
            }
            Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
        }
    }

    pub fn get_query_param(key: &str) -> Option<String> {
        let mut buf = [0u8;1024];
        unsafe{
            let wrote = host_get_query_param(
                key.as_ptr(),
                key.len() as i32,
                buf.as_mut_ptr(),
                1024
            );
            if wrote == 0xFFFFFF{
                return None;
            }
            Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
        }
    }

    pub fn get_raw_query() -> Option<String> {
        let mut buf = [0u8;1024];
        unsafe{
            let wrote = host_get_raw_query(buf.as_mut_ptr(), 1024);
            if wrote == 0xFFFFFF{
                return None;
            }
            Some(String::from_utf8_lossy(&buf[..wrote as usize]).to_string())
        }
    }

    pub fn send_response(status_code: i32, msg: &str) {
        unsafe {
            host_send_response(status_code, msg.as_ptr(), msg.len() as i32);
        }
    }

    pub fn log(level: LogLevel, msg: &str) {
        unsafe{
            host_log(level as i32, msg.as_ptr(), msg.len() as i32);
        }
    }

    pub fn set_header(key: &str, value: &str) {
        unsafe {
            host_set_header(key.as_ptr(), key.len() as i32, value.as_ptr(), value.len() as i32);
        }
    }
}

#[macro_export]
macro_rules! proxy_handler {
    ($handler: path) => {
        #[no_mangle]
        pub extern "C" fn _start() {
            $handler();
        }
    };
}