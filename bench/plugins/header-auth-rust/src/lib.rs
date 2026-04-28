#[link(wasm_import_module = "env")]
extern "C" {
    fn host_get_header(
        key_ptr: *const u8,
        key_len: u32,
        buf_ptr: *mut u8,
        buf_max_len: u32,
    ) -> u32;
    fn host_send_response(status_code: u32, body_ptr: *const u8, body_len: u32);
}

const NOT_FOUND: u32 = 0xFFFF_FFFF;
const HEADER_NAME: &[u8] = b"x-api-key";
const EXPECTED: &[u8] = b"bench-secret";
const RESP_UNAUTHORIZED: &[u8] = b"unauthorized";

#[no_mangle]
pub extern "C" fn on_request() {
    let mut buf = [0u8; 64];
    let read = unsafe {
        host_get_header(
            HEADER_NAME.as_ptr(),
            HEADER_NAME.len() as u32,
            buf.as_mut_ptr(),
            buf.len() as u32,
        )
    };

    if read == NOT_FOUND {
        unauthorized();
        return;
    }

    let read_len = read as usize;
    if read_len != EXPECTED.len() || &buf[..read_len] != EXPECTED {
        unauthorized();
    }
}

fn unauthorized() {
    unsafe {
        host_send_response(
            401,
            RESP_UNAUTHORIZED.as_ptr(),
            RESP_UNAUTHORIZED.len() as u32,
        );
    }
}
