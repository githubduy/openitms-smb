module quickwin.dev/plugins/device-inventory

go 1.24.6

require (
	github.com/go-sql-driver/mysql v1.8.1
	quickwin.dev/proto v0.0.0
	quickwin.dev/sdk v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace (
	quickwin.dev/pluginmanager => ../../plugin-manager
	quickwin.dev/proto => ../../proto/gen/go
	quickwin.dev/sdk => ../../sdk/go
	quickwin.dev/winrsexec => ../../winrs-exec
)
