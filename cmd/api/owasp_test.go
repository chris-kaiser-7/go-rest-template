package main

import (
	"net/http"
	"testing"
)

func check_owasp_headers(t *testing.T, header http.Header) {
	check_nostore_header(t, header)
	check_content_security_header(t, header)
	check_content_type_header(t, header)
	check_strict_transport_header(t, header)
	check_nosiff_header(t, header)
	check_deny_xframe_header(t, header)
}

func check_nostore_header(t *testing.T, header http.Header) {
	check_header_value(t, header, CACHE_CONTROL, "no-store")
}

func check_content_security_header(t *testing.T, header http.Header) {
	check_header_value(t, header, CONTENT_SECURITY_POLICY, "frame-ancestors 'none'")
}

func check_content_type_header(t *testing.T, header http.Header) {
	check_header_exists(t, header, CONTENT_TYPE)
}

func check_strict_transport_header(t *testing.T, header http.Header) {
	check_header_exists(t, header, STRICT_TRANSPORT_SECURITY)
}

func check_nosiff_header(t *testing.T, header http.Header) {
	check_header_value(t, header, X_CONTENT_OPTIONS, "nosniff")
}

func check_deny_xframe_header(t *testing.T, header http.Header) {
	check_header_value(t, header, X_FRAME_OPTIONS, "DENY")
}

func check_header_value(t *testing.T, header http.Header, key string, value string) {
	h := header.Get(key)
	if h != value {
		t.Errorf("Expected header %s to have value %s", key, value)
	}
}

func check_header_exists(t *testing.T, header http.Header, key string) {
	h := header.Values(key)
	if len(h) == 0 {
		t.Errorf("Expected header %s to exist", key)
	}
}
