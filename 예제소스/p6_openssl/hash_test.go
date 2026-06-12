package main

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// 표준 라이브러리 crypto/sha256과 교차 검증
func TestSHA256MatchesStdlib(t *testing.T) {
	inputs := []string{"", "Hello, CGo!", "한글도 테스트"}
	for _, in := range inputs {
		want := sha256.Sum256([]byte(in))
		got := sha256OpenSSL([]byte(in))
		if got != hex.EncodeToString(want[:]) {
			t.Errorf("입력 %q: got %s, want %x", in, got, want)
		}
	}
}
