// Package giteamanager — quản lý Gitea bundle (git server local mặc định của OpenITMS-SMB).
// ADR-0005 / gitea-integration.md. Package này chạy NGOÀI cây upstream; core-patch 0007 chỉ
// gọi hook mỏng khi tạo project → tự tạo repo local.
package giteamanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client gọi Gitea API bằng token (admin token sinh lúc provision).
type Client struct {
	baseURL string // vd http://127.0.0.1:3080
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Repo — thông tin repo trả về từ Gitea.
type Repo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
	HTMLURL  string `json:"html_url"`
	Empty    bool   `json:"empty"`
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.http.Do(req)
}

// Healthz kiểm tra Gitea sống.
func (c *Client) Healthz(ctx context.Context) error {
	resp, err := c.http.Do(mustReq(ctx, c.baseURL+"/api/healthz"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gitea healthz HTTP %d", resp.StatusCode)
	}
	return nil
}

// EnsureOrg tạo org nếu chưa có (idempotent). 201 tạo mới / 422 đã tồn tại đều OK.
func (c *Client) EnsureOrg(ctx context.Context, org string) error {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/orgs", map[string]any{
		"username":   org,
		"visibility": "private",
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusUnprocessableEntity, http.StatusConflict:
		return nil // tạo mới hoặc đã tồn tại
	default:
		return apiErr("EnsureOrg", resp)
	}
}

// CreateRepo tạo repo dưới org (idempotent — đã có → lấy repo hiện tại). autoInit=true → có commit đầu.
func (c *Client) CreateRepo(ctx context.Context, org, name string, autoInit bool) (*Repo, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/v1/orgs/"+org+"/repos", map[string]any{
		"name":      name,
		"private":   true,
		"auto_init": autoInit,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusCreated:
		return decodeRepo(resp)
	case http.StatusConflict, http.StatusUnprocessableEntity:
		return c.GetRepo(ctx, org, name) // đã tồn tại → trả repo hiện có
	default:
		return nil, apiErr("CreateRepo", resp)
	}
}

// GetRepo lấy repo org/name.
func (c *Client) GetRepo(ctx context.Context, org, name string) (*Repo, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/repos/"+org+"/"+name, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apiErr("GetRepo", resp)
	}
	return decodeRepo(resp)
}

func decodeRepo(resp *http.Response) (*Repo, error) {
	var r Repo
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func apiErr(op string, resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("gitea %s: HTTP %d: %s", op, resp.StatusCode, strings.TrimSpace(string(b)))
}

func mustReq(ctx context.Context, url string) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	return req
}

// Slug chuẩn hóa tên project → tên repo hợp lệ (lowercase, kebab).
func Slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		s = "project"
	}
	return s
}
