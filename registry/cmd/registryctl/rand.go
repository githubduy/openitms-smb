package main

import "crypto/rand"

// readRand — tách ra để CLI dùng crypto/rand (an toàn cho keygen thật).
func readRand(b []byte) (int, error) { return rand.Read(b) }
