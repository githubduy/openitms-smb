package sdk

import "github.com/hashicorp/go-plugin"

// Serve — điểm vào duy nhất của một plugin process. Gọi trong main():
//
//	func main() { sdk.Serve(&myPlugin{}) }
//
// go-plugin lo handshake (stdout), gRPC server và mTLS tự động.
func Serve(impl Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         map[string]plugin.Plugin{PluginName: &GRPCPlugin{Impl: impl}},
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
