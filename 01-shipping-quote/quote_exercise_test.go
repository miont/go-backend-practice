//go:build exercise

package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShippingQuote(t *testing.T) {
	const jsonType = "application/json"
	tests := []struct {
		name, method, contentType, body string
		status, price                   int
		errorCode, allow                string
	}{
		{"exact kg", "POST", jsonType, `{"zone":"local","weight_grams":1000}`, 200, 18000, "", ""},
		{"round up", "POST", jsonType, `{"zone":"local","weight_grams":1001}`, 200, 21000, "", ""},
		{"max weight", "POST", jsonType, `{"zone":"remote","weight_grams":20000}`, 200, 150000, "", ""},
		{"zero weight", "POST", jsonType, `{"zone":"local","weight_grams":0}`, 400, 0, "invalid_request", ""},
		{"above max", "POST", jsonType, `{"zone":"local","weight_grams":20001}`, 400, 0, "invalid_request", ""},
		{"fractional weight", "POST", jsonType, `{"zone":"local","weight_grams":1.5}`, 400, 0, "invalid_request", ""},
		{"missing weight", "POST", jsonType, `{"zone":"local"}`, 400, 0, "invalid_request", ""},
		{"unknown zone", "POST", jsonType, `{"zone":"unknown","weight_grams":1}`, 400, 0, "invalid_request", ""},
		{"extra field", "POST", jsonType, `{"zone":"local","weight_grams":1,"currency":"RUB"}`, 400, 0, "invalid_request", ""},
		{"second value", "POST", jsonType, `{"zone":"local","weight_grams":1} {}`, 400, 0, "invalid_request", ""},
		{"malformed JSON", "POST", jsonType, `{"zone":`, 400, 0, "invalid_request", ""},
		{"charset", "POST", jsonType + "; charset=utf-8", `{"zone":"local","weight_grams":1}`, 200, 18000, "", ""},
		{"plain text", "POST", "text/plain", `{"zone":"local","weight_grams":1}`, 415, 0, "unsupported_media_type", ""},
		{"missing media type", "POST", "", `{"zone":"local","weight_grams":1}`, 415, 0, "unsupported_media_type", ""},
		{"wrong method", "GET", "", "", 405, 0, "method_not_allowed", "POST"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/v1/shipping/quote", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rec := httptest.NewRecorder()
			routes().ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tc.status, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != jsonType {
				t.Errorf("Content-Type = %q, want %q", got, jsonType)
			}
			if got := rec.Header().Get("Allow"); got != tc.allow {
				t.Errorf("Allow = %q, want %q", got, tc.allow)
			}

			if tc.errorCode != "" {
				var got errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.Error != tc.errorCode {
					t.Errorf("error = %q, want %q", got.Error, tc.errorCode)
				}
				return
			}

			var got quoteResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.PriceKopecks != tc.price {
				t.Errorf("price = %d, want %d", got.PriceKopecks, tc.price)
			}
		})
	}
}
