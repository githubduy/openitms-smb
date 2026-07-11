package giteamanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test client bằng Gitea API giả (httptest) — không cần binary Gitea thật.

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Host Management": "host-management",
		"  Dự Án #1  ":    "d-n-1",
		"web/app":         "web-app",
		"":                "project",
		"ALL-CAPS_Test":   "all-caps-test",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q)=%q, muốn %q", in, got, want)
		}
	}
}

func TestEnsureOrg_IdempotentAndAuth(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := json.Marshal(nil)
		_ = b
		if r.URL.Path == "/api/v1/orgs" && r.Method == "POST" {
			var m map[string]any
			json.NewDecoder(r.Body).Decode(&m)
			gotBody, _ = m["username"].(string)
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "tok123")
	if err := c.EnsureOrg(context.Background(), "openitms"); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "token tok123" {
		t.Fatalf("auth header sai: %q", gotAuth)
	}
	if gotBody != "openitms" {
		t.Fatalf("org name sai: %q", gotBody)
	}
}

func TestEnsureOrg_AlreadyExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity) // Gitea trả khi org đã tồn tại
	}))
	defer srv.Close()
	if err := NewClient(srv.URL, "t").EnsureOrg(context.Background(), "openitms"); err != nil {
		t.Fatalf("org đã tồn tại phải OK (idempotent), got: %v", err)
	}
}

func TestCreateRepo_NewAndIdempotent(t *testing.T) {
	created := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/api/v1/orgs/openitms/repos":
			if created {
				w.WriteHeader(http.StatusConflict) // lần 2: đã tồn tại
				return
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Repo{Name: "host", FullName: "openitms/host",
				CloneURL: "http://127.0.0.1:3080/openitms/host.git"})
		case r.Method == "GET" && r.URL.Path == "/api/v1/repos/openitms/host":
			json.NewEncoder(w).Encode(Repo{Name: "host", FullName: "openitms/host",
				CloneURL: "http://127.0.0.1:3080/openitms/host.git"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "t")
	r1, err := c.CreateRepo(context.Background(), "openitms", "host", true)
	if err != nil || r1.FullName != "openitms/host" || r1.CloneURL == "" {
		t.Fatalf("tạo repo mới fail: %+v err=%v", r1, err)
	}
	// lần 2 (đã tồn tại) → conflict → fallback GetRepo, không lỗi
	r2, err := c.CreateRepo(context.Background(), "openitms", "host", true)
	if err != nil || r2.FullName != "openitms/host" {
		t.Fatalf("idempotent create fail: %+v err=%v", r2, err)
	}
}

func TestGetFile_NotFoundAndFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/repos/openitms/host/contents/site.yml" {
			json.NewEncoder(w).Encode(map[string]any{
				"content": base64.StdEncoding.EncodeToString([]byte("hello: world\n")),
				"sha":     "abc123",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "t")

	content, sha, exists, err := c.GetFile(context.Background(), "openitms", "host", "main", "site.yml")
	if err != nil || !exists || content != "hello: world\n" || sha != "abc123" {
		t.Fatalf("GetFile found sai: content=%q sha=%q exists=%v err=%v", content, sha, exists, err)
	}
	_, _, exists, err = c.GetFile(context.Background(), "openitms", "host", "main", "missing.sh")
	if err != nil || exists {
		t.Fatalf("GetFile missing phải exists=false không lỗi: exists=%v err=%v", exists, err)
	}
}

func TestPutFile_CreateThenUpdate(t *testing.T) {
	fileExists := false
	var lastMethod string
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := "/api/v1/repos/openitms/host/contents/deploy/site.yml"
		if r.URL.Path != p {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			if !fileExists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"content": base64.StdEncoding.EncodeToString([]byte("old")),
				"sha":     "sha-old",
			})
		case http.MethodPost, http.MethodPut:
			lastMethod = r.Method
			json.NewDecoder(r.Body).Decode(&lastBody)
			fileExists = true
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"commit": map[string]any{"html_url": "http://gitea/commit/1"},
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "t")
	ctx := context.Background()

	// lần 1: file chưa có → POST (create), không kèm sha
	url1, err := c.PutFile(ctx, "openitms", "host", "main", "deploy/site.yml", "content-1", "msg1")
	if err != nil || lastMethod != http.MethodPost || url1 != "http://gitea/commit/1" {
		t.Fatalf("create fail: method=%s url=%q err=%v", lastMethod, url1, err)
	}
	if _, hasSHA := lastBody["sha"]; hasSHA {
		t.Fatalf("create không được kèm sha, body=%v", lastBody)
	}
	if got, _ := lastBody["content"].(string); got != base64.StdEncoding.EncodeToString([]byte("content-1")) {
		t.Fatalf("content base64 sai: %q", got)
	}

	// lần 2: file đã có → PUT (update) kèm sha
	if _, err := c.PutFile(ctx, "openitms", "host", "main", "deploy/site.yml", "content-2", "msg2"); err != nil {
		t.Fatalf("update fail: %v", err)
	}
	if lastMethod != http.MethodPut {
		t.Fatalf("update phải dùng PUT, got %s", lastMethod)
	}
	if lastBody["sha"] != "sha-old" {
		t.Fatalf("update phải kèm sha cũ, got %v", lastBody["sha"])
	}
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	if err := NewClient(srv.URL, "").Healthz(context.Background()); err != nil {
		t.Fatalf("healthz phải OK: %v", err)
	}
}
