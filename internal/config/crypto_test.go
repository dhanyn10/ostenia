package config

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"Normal text", "hello world"},
		{"Empty text", ""},
		{"Special characters", "!@#$%^&*()_+"},
		{"Long text", "this is a much longer text to ensure that the encryption and decryption work correctly with larger data sets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.text)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if tt.text != "" && encrypted == tt.text {
				t.Errorf("Encrypt() returned plaintext")
			}

			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.text {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.text)
			}
		})
	}
}

func TestDecryptInvalid(t *testing.T) {
	_, err := Decrypt("invalid-base64-!")
	if err == nil {
		t.Error("Decrypt() expected error for invalid base64")
	}

	_, err = Decrypt("YQ==") // "a" in base64, too short for IV
	if err == nil {
		t.Error("Decrypt() expected error for too short ciphertext")
	}
}
