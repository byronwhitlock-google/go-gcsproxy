package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/tink"
	log "github.com/sirupsen/logrus"
)
var kh *keyset.Handle
var err error
var a tink.AEAD
var aad = []byte("")




func base64_md5hash(bytestream []byte) string {
	hash := md5.New()
	_, err := io.WriteString(hash, string(bytestream))
	if err != nil {
		fmt.Println("Error computing MD5 hash:", err)
		log.Fatal(err)
	}
	md5Hash := hash.Sum(nil) // Get the computed MD5 hash as a byte array

	// Step 2: Encode the MD5 hash as Base64
	base64MD5Hash := base64.StdEncoding.EncodeToString(md5Hash)

	// Print the result
	fmt.Println("Base64-encoded MD5 hash:", base64MD5Hash)
	return base64MD5Hash
}

func encrypt_tink(plaintext []byte) ([]byte, error) {
	// Generate a new keyset handle using the AES256-GCM template
	kh, err = keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		fmt.Printf("Error generating keyset: %v\n", err)
		log.Fatal(err)
	}

	// Get an AEAD primitive from the keyset handle
	a, err = aead.New(kh)
	if err != nil {
		fmt.Printf("Error creating AEAD primitive: %v\n", err)
		log.Fatal(err)
	}

	
	// Encrypt the string
	ciphertext, err := a.Encrypt(plaintext, aad)
	if err != nil {
		fmt.Printf("Error encrypting data: %v\n", err)
		log.Fatal(err)
	}

	return ciphertext, nil

}

func decrypt_tink(ciphertext []byte) ([]byte,error) {
	
	// Decrypt the ciphertext back to the original plaintext
    decrypted, err := a.Decrypt(ciphertext, aad)
    if err != nil {
        fmt.Printf("Error decrypting data: %v\n", err)
        log.Fatal(err)
    }

    // Print the decrypted text
    fmt.Printf("Decrypted text: %s\n", string(decrypted))
	return decrypted,nil
}
