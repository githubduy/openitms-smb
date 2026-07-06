// registryctl — CLI cho maintainer: sinh khoá, build + ký registry từ artifact.
// Private key KHÔNG commit (publishing-policy). Public key công khai, nhúng vào core.
//
//	registryctl keygen                                  # in pub + priv (base64) ra stdout
//	registryctl build <registry-name> <out-dir> <spec.json>   # build + ký (priv qua env REGISTRY_PRIVATE_KEY)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"quickwin.dev/registry"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "keygen":
		keygen()
	case "build":
		build(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: registryctl keygen | build <name> <out-dir> <spec.json>")
	os.Exit(2)
}

func keygen() {
	// seed ngẫu nhiên: đọc từ crypto/rand qua registry (dùng thời gian + os pid tránh Date cấm ở workflow;
	// ở CLI thường/không phải workflow nên dùng rand trực tiếp).
	seed := make([]byte, 32)
	if _, err := readRand(seed); err != nil {
		fatal(err)
	}
	pub, priv := registry.GenerateKeyFromSeed(seed)
	fmt.Println("PUBLIC_KEY=" + registry.EncodePublicKey(pub))
	fmt.Fprintln(os.Stderr, "PRIVATE_KEY (giữ bí mật, KHÔNG commit — set env REGISTRY_PRIVATE_KEY):")
	fmt.Println("PRIVATE_KEY=" + registry.EncodePrivateKey(priv))
}

// spec.json: [{"type","name","version","license","dir","description","min_core_version"}]
type specEntry struct {
	Type           registry.ArtifactType `json:"type"`
	Name           string                `json:"name"`
	Version        string                `json:"version"`
	License        string                `json:"license"`
	Dir            string                `json:"dir"`
	Description    string                `json:"description"`
	MinCoreVersion string                `json:"min_core_version"`
}

func build(args []string) {
	if len(args) != 3 {
		usage()
	}
	name, outDir, specPath := args[0], args[1], args[2]

	raw, err := os.ReadFile(specPath)
	if err != nil {
		fatal(err)
	}
	var entries []specEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		fatal(fmt.Errorf("spec.json: %w", err))
	}
	arts := make([]registry.SourceArtifact, 0, len(entries))
	for _, e := range entries {
		arts = append(arts, registry.SourceArtifact{
			Type: e.Type, Name: e.Name, Version: e.Version, License: e.License, Dir: e.Dir,
			Meta: map[string]string{"description": e.Description, "min_core_version": e.MinCoreVersion},
		})
	}

	var priv []byte
	if k := os.Getenv("REGISTRY_PRIVATE_KEY"); k != "" {
		p, err := registry.DecodePrivateKey(k)
		if err != nil {
			fatal(fmt.Errorf("REGISTRY_PRIVATE_KEY: %w", err))
		}
		priv = p
	} else {
		fmt.Fprintln(os.Stderr, "CẢNH BÁO: REGISTRY_PRIVATE_KEY rỗng — build KHÔNG ký (chỉ dev local)")
	}

	ix, built, sig, err := registry.BuildIndex(name, arts, time.Now(), priv)
	if err != nil {
		fatal(err)
	}
	if err := registry.WriteRegistry(outDir, ix, built, sig); err != nil {
		fatal(err)
	}
	fmt.Printf("OK: registry %q → %s (%d artifact, signed=%v)\n", name, outDir, len(built), sig != "")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ERROR:", err)
	os.Exit(1)
}
