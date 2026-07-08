package winrsexec

// noProxyCertTransport — WinRM client-cert transport ÉP KHÔNG đi qua HTTP proxy.
// WinRM là kết nối LAN trực tiếp tới host đích (5986) — KHÔNG bao giờ nên qua proxy công ty.
// masterzen/winrm.ClientAuthRequest hardcode Proxy: http.ProxyFromEnvironment (field unexported,
// không override được) → khi HTTPS_PROXY set, request WinRM bị route qua proxy → "cannot connect".
// Đây là bản thay thế: y hệt ClientAuthRequest nhưng Proxy: nil. URL dựng từ Endpoint (exported).

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/masterzen/winrm"
	"github.com/masterzen/winrm/soap"
)

type noProxyCertTransport struct {
	url       string
	transport http.RoundTripper
}

func (c *noProxyCertTransport) Transport(endpoint *winrm.Endpoint) error {
	cert, err := tls.X509KeyPair(endpoint.Cert, endpoint.Key)
	if err != nil {
		return err
	}
	scheme := "http"
	if endpoint.HTTPS {
		scheme = "https"
	}
	c.url = fmt.Sprintf("%s://%s:%d/wsman", scheme, endpoint.Host, endpoint.Port)

	//nolint:gosec
	c.transport = &http.Transport{
		Proxy: nil, // WinRM = LAN trực tiếp, KHÔNG qua proxy
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateOnceAsClient,
			InsecureSkipVerify: endpoint.Insecure,
			Certificates:       []tls.Certificate{cert},
			ServerName:         endpoint.TLSServerName,
			// WinRM certificate-auth KHÔNG hỗ trợ TLS 1.3 (client-cert đổi cách xử lý → HTTP.sys
			// reset connection "forcibly closed"). Ép TLS 1.2 để client-cert trao đổi trong handshake.
			MaxVersion: tls.VersionTLS12,
		},
		Dial:                  (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).Dial,
		ResponseHeaderTimeout: endpoint.Timeout,
	}
	return nil
}

func (c *noProxyCertTransport) Post(_ *winrm.Client, request *soap.SoapMessage) (string, error) {
	httpClient := &http.Client{Transport: c.transport}

	req, err := http.NewRequest("POST", c.url, strings.NewReader(request.String()))
	if err != nil {
		return "", fmt.Errorf("tạo http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/soap+xml;charset=UTF-8")
	req.Header.Set("Authorization", "http://schemas.dmtf.org/wbem/wsman/1/wsman/secprofile/https/mutual")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unknown error %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		return "", fmt.Errorf("đọc response WinRM: %w", rerr)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}
