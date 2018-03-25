package crypto

import "testing"

func BenchmarkGenerateKeyBySecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewPrivateKeyBySecret("secret-password")
	}
}

func BenchmarkGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewPrivateKey()
	}
}

func BenchmarkSign(b *testing.B) {
	prv := NewPrivateKey()
	data := hash256([]byte("Abc Ёпрст"))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		prv.Sign(data)
	}
}

func BenchmarkVerify(b *testing.B) {
	prv := NewPrivateKey()
	pub := prv.PublicKey
	data := hash256([]byte("Abc Ёпрст"))
	sign := prv.Sign(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !pub.Verify(data, sign) {
			b.Fatal("Verify fail")
		}
	}
}
