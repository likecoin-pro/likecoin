package xhash

import "testing"

func BenchmarkGenerateKeyBySecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateKeyByPassword("abc", 256)
	}
}
