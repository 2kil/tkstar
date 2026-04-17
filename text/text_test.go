package text

import (
	"bytes"
	"testing"
)

func TestTextEncryptDecryptPreservesLeadingZeroBytes(t *testing.T) {
	keyPair, err := TextGetKeyPair(1024)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	plaintext := []byte{0, 0, 1, 2, 3, 4}
	ciphertext, err := TextEncrypt(keyPair.PublicKey, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := TextDecrypt(keyPair.PrivateKey, ciphertext)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted = %v, want %v", decrypted, plaintext)
	}
}
