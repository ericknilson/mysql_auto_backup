package crypto

import (
	"encoding/base64"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	c, err := New("uma-chave-qualquer")
	if err != nil {
		t.Fatal(err)
	}
	cases := []string{"", "x", "host.example.com", "p@ssw0rd!çãoº"}
	for _, plain := range cases {
		enc, err := c.Encrypt(plain)
		if err != nil {
			t.Fatalf("encrypt %q: %v", plain, err)
		}
		dec, err := c.Decrypt(enc)
		if err != nil {
			t.Fatalf("decrypt %q: %v", plain, err)
		}
		if dec != plain {
			t.Fatalf("roundtrip falhou: %q != %q", dec, plain)
		}
	}
}

func TestTamperDetection(t *testing.T) {
	c, _ := New("seg")
	enc, _ := c.Encrypt("dados sensíveis")
	raw, _ := base64.StdEncoding.DecodeString(enc)
	raw[len(raw)-1] ^= 0x01
	tampered := base64.StdEncoding.EncodeToString(raw)
	if _, err := c.Decrypt(tampered); err == nil {
		t.Fatal("decrypt deveria falhar em ciphertext alterado")
	}
}

func TestUniqueCiphertext(t *testing.T) {
	c, _ := New("seg")
	a, _ := c.Encrypt("igual")
	b, _ := c.Encrypt("igual")
	if a == b {
		t.Fatal("nonces devem produzir ciphertexts distintos para o mesmo texto")
	}
}
