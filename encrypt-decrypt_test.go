package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"testing"
)

func TestBase64_md5hash(t *testing.T) {
	testCases := []struct {
		input    []byte
		expected string
	}{
		{[]byte("hello"), "XUFAKrxLKna5cZ2REBfFkg=="},
		{[]byte(""), "1B2M2Y8AsgTpgAmY7PhCfg=="}, // MD5 hash of an empty string
		{[]byte("a longer string with spaces"), "h/BochuK2O56DxYHJbWijA=="},
	}

	for _, tc := range testCases {
		actual := Base64MD5Hash(tc.input)
		if actual != tc.expected {
			t.Errorf("For input %q, expected hash %q but got %q", tc.input, tc.expected, actual)
		}
	}
}


//Since it uses Google Cloud KMS, integration tests needed. But implemented test using static data and context.
func TestEncryptBytes(t *testing.T) {
	ctx := context.Background()
	resourceName := "projects/cmetestproj/locations/global/keyRings/gcsproxytest/cryptoKeys/gcsproxy"

	plaintext := []byte("Test data to encrypt")

	ciphertext, err := EncryptBytes(ctx, resourceName, plaintext)
	if err == nil {
		fmt.Printf("Ciphertext: %x\n", ciphertext)
		_, err := DecryptBytes(ctx, resourceName, ciphertext)
		if err != nil {
			t.Errorf("decryption error %v", err)
		}

	} else {
		t.Errorf("encryption error %v", err)
	}

}

//Since it uses Google Cloud KMS, integration tests needed. But implemented test using static data and context.
func TestDecryptBytes(t *testing.T) {

	ctx := context.Background()
	resourceName := "gcp-kms://projects/gcp-project-id/locations/global/keyRings/test-keyring/cryptoKeys/test-key/cryptoKeyVersions/1"
	// Following encrypted value is dummy, change it when testing. It does not gurantee correct value when running with actual integration tests.
	encryptedText := []byte{0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0}

	decrypted, err := DecryptBytes(ctx, resourceName, encryptedText)
	if err == nil {
		//fmt.Printf("Decrypted text: %s\n", decrypted)
		_, err := EncryptBytes(ctx, resourceName, decrypted)
		if err != nil {
			t.Errorf("encryption error %v", err)
		}
	} else {
		t.Errorf("decryption error %v", err)
	}
}


func TestBase64Md5HashEmpty(t *testing.T) {
	emptyHash := Base64MD5Hash([]byte(""))

	// Compare with the known MD5 hash of an empty string
	expectedHash := "1B2M2Y8AsgTpgAmY7PhCfg=="
	if emptyHash != expectedHash {
		t.Errorf("MD5 hash of empty string mismatch.\nExpected: %s\nGot: %s", expectedHash, emptyHash)
	}
}

// benchmark test for Base64 function
func BenchmarkBase64_md5hash(b *testing.B) {
	data := []byte("test data")
	for i := 0; i < b.N; i++ {
		Base64MD5Hash(data)
	}
}

func BenchmarkMd5Hash(b *testing.B) {
	data := []byte("test data")
	h := md5.New()
	for i := 0; i < b.N; i++ {
		h.Write(data)
		h.Sum(nil) // Reset the hash for the next iteration
	}

}


