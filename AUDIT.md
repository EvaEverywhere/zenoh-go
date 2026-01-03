# CGO Bindings Audit

## Summary

Audit of zenoh-go CGO bindings against zenoh-c API.

**Last Updated:** 2026-01-02

## Issues Found & Status

### 1. ✅ Config Endpoint Setup - FIXED

**Status:** Fixed with `z_config_from_str` and fallback to `zc_config_insert_json5`

The config now properly creates JSON5 configuration and applies it using:
- Primary: `z_config_from_str(&zconfig, jsonString)`
- Fallback: `zc_config_insert_json5(config, "connect", connectJSON)`

### 2. ✅ z_view_string_t Field Access - FIXED

**Status:** Fixed with C helper functions

Added `view_string_data()` and `view_string_len()` helpers that use proper zenoh-c API:
```c
static const char* view_string_data(const z_view_string_t* s) {
    return z_string_data(z_view_string_loan(s));
}
```

### 3. ✅ z_bytes_reader API - FIXED

**Status:** Fixed with helper function

Added `read_bytes_to_buffer()` helper that encapsulates the reader API:
```c
static size_t read_bytes_to_buffer(const z_loaned_bytes_t* bytes, uint8_t* buffer, size_t len) {
    z_bytes_reader_t reader;
    z_bytes_reader_init(&reader, bytes);
    return z_bytes_reader_read(&reader, buffer, len);
}
```

### 4. ✅ Callback Signature - OK

**Status:** Verified correct

The C callback receives `z_loaned_sample_t*` and passes it to Go - this is correct.

### 5. ⚠️ z_declare_publisher Argument Order - NEEDS VERIFICATION

**Status:** Needs testing on robot

```go
C.z_declare_publisher(
    C.z_session_loan(&s.session),  // session
    &pub,                           // out publisher
    C.z_view_keyexpr_loan(&ke),    // key expr
    nil,                            // options
)
```

This matches zenoh-c 1.0 API pattern but needs hardware verification.

### 6. ⚠️ z_declare_subscriber Argument Order - NEEDS VERIFICATION

**Status:** Needs testing on robot

### 7. ✅ Mode Configuration - FIXED

**Status:** Mode is now included in JSON5 config

```json
{"mode":"client","connect":{"endpoints":["tcp/..."]}}
```

### 8. ℹ️ Error Code Handling - ACCEPTABLE

**Status:** Basic error codes returned

Could be improved in future to extract more detailed error messages.

## C Helper Functions Added

```c
// Config creation
int config_from_json5(z_owned_config_t* config, const char* json5);
int config_insert_json5(z_loaned_config_t* config, const char* key, const char* json5);

// String access (version-agnostic)
const char* view_string_data(const z_view_string_t* s);
size_t view_string_len(const z_view_string_t* s);

// Bytes reading
size_t read_bytes_to_buffer(const z_loaned_bytes_t* bytes, uint8_t* buffer, size_t len);

// Closure creation
z_owned_closure_sample_t make_sample_closure(void* context);
```

## Testing Checklist

- [x] Mock mode builds and tests pass
- [ ] Test with zenoh-c installed locally on macOS (if available)
- [ ] Test with zenoh-c on ARM64 Linux (robot)
- [ ] Verify config JSON is properly applied
- [ ] Verify subscriber callbacks work
- [ ] Verify publisher put works
- [ ] Test with actual Reachy Mini robot

## Next Steps

1. **Install zenoh-c on robot** and test CGO build
2. **Verify function signatures** against installed zenoh.h
3. **Test with real Zenoh router** (Pollen daemon)
4. **Integrate into go-reachy**

## Version Compatibility

Target: zenoh-c 1.0.x

Key API patterns used:
- `z_owned_*` - Owned types (caller manages lifecycle)
- `z_loaned_*` - Borrowed references
- `z_*_move()` - Transfer ownership
- `z_*_loan()` - Borrow reference
- `z_*_drop()` - Release owned resource

