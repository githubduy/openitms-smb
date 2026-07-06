// Package sdk — quickwin-plugin-sdk-go: viết plugin QuickWin bằng Go.
// Plugin chỉ cần implement interface Plugin rồi gọi sdk.Serve(impl) trong main().
package sdk

import "github.com/hashicorp/go-plugin"

// ProtocolVersion = APP-PROTOCOL-VERSION (xem docs/L2-specs/proto-contract.md).
// Bump CHỈ khi breaking change proto — cần ADR được duyệt.
const ProtocolVersion = 1

// PluginName — key duy nhất trong plugin map của go-plugin.
const PluginName = "quickwin"

// Handshake — chung cho core và mọi plugin. Magic cookie chỉ để nhận diện
// "binary này là plugin QuickWin", KHÔNG phải cơ chế bảo mật (bảo mật = mTLS
// tự động của go-plugin + checksum manifest).
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   "QUICKWIN_PLUGIN",
	MagicCookieValue: "quickwin-plugin-v1-2f8a4e",
}
