package kernel_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alireza-constantin/case-platform/server/internal/kernel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The approved system seam: HTTP against real PostgreSQL, no mocked authority.
func TestGuestBootstrapAndCatalogue(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated local PostgreSQL database")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(kernel.NewHandler(pool, "http://localhost:5173", false, nil))
	defer server.Close()

	request := func(cookie *http.Cookie) *http.Response {
		t.Helper()
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/guest", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://localhost:5173")
		if cookie != nil {
			req.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { response.Body.Close() })
		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("bootstrap status %d", response.StatusCode)
		}
		return response
	}
	first := request(nil)
	cookies := first.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected a guest cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge <= 0 || cookie.Path != "/" {
		t.Fatal("guest cookie must be persistent, HttpOnly, SameSite=Lax and usable on local HTTP")
	}
	if request(cookie).Cookies()[0].Value != cookie.Value {
		t.Fatal("returning browser must reuse its stored guest")
	}
	// A new API instance must recognize the same database-backed credential.
	server.Close()
	server = httptest.NewServer(kernel.NewHandler(pool, "http://localhost:5173", false, nil))
	defer server.Close()
	if request(cookie).Cookies()[0].Value != cookie.Value {
		t.Fatal("API restart must retain the guest identity")
	}
	forged := *cookie
	forged.Value = strings.Repeat("a", 64)
	if request(&forged).Cookies()[0].Value == forged.Value {
		t.Fatal("an unknown credential must be replaced with a stored guest")
	}

	response, err := http.Get(server.URL + "/api/cases")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var catalogue struct {
		Cases []json.RawMessage `json:"cases"`
	}
	if err := json.NewDecoder(response.Body).Decode(&catalogue); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || catalogue.Cases == nil || len(catalogue.Cases) != 0 {
		t.Fatal("bootstrap catalogue must be an empty array, not null or speculative demos")
	}
	for _, input := range []struct{ name, body, contentType, origin string }{
		{"cross-origin", "{}", "application/json", "https://other.example"},
		{"opaque origin", "{}", "application/json", "null"},
		{"wrong media type", "{}", "text/plain", ""},
		{"nonempty object", `{"state":true}`, "application/json", ""},
		{"null", "null", "application/json", ""},
		{"trailing JSON", "{} {}", "application/json", ""},
		{"oversized", "{}" + strings.Repeat(" ", 1024), "application/json", ""},
	} {
		t.Run(input.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/guest", strings.NewReader(input.body))
			req.Header.Set("Content-Type", input.contentType)
			req.Header.Set("Origin", input.origin)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != 400 || envelope.Error.Code != "invalid_request" || len(res.Cookies()) != 0 {
				t.Fatal("invalid bootstrap must return a neutral error without establishing a guest")
			}
		})
	}
}
