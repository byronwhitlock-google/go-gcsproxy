package main

import (
	"fmt"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	log "github.com/sirupsen/logrus"
)

func encrypt_tink(plaintext []byte) ([]byte,error){
	// Generate a new keyset handle using the AES256-GCM template
    kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
    if err != nil {
        fmt.Printf("Error generating keyset: %v\n", err)
        log.Fatal(err)
    }

    // Get an AEAD primitive from the keyset handle
    a, err := aead.New(kh)
    if err != nil {
        fmt.Printf("Error creating AEAD primitive: %v\n", err)
        log.Fatal(err)
    }

	aad := []byte("")
	// Encrypt the string
    ciphertext, err := a.Encrypt(plaintext,aad)
    if err != nil {
        fmt.Printf("Error encrypting data: %v\n", err)
        log.Fatal(err)
    }

    // Print the ciphertext
    //fmt.Printf("Encrypted text: %x\n", ciphertext)
	return ciphertext,nil

}

// func decrypt_tink(ciphertext string) (string,error) {
// 	// Generate a new keyset handle using the AES256-GCM template
//     kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
//     if err != nil {
//         fmt.Printf("Error generating keyset: %v\n", err)
//         return
//     }

//     // Get an AEAD primitive from the keyset handle
//     a, err := aead.New(kh)
//     if err != nil {
//         fmt.Printf("Error creating AEAD primitive: %v\n", err)
//         return
//     }

// 	// Decrypt the ciphertext back to the original plaintext
//     decrypted, err := a.Decrypt(ciphertext, aad)
//     if err != nil {
//         fmt.Printf("Error decrypting data: %v\n", err)
//         return
//     }

//     // Print the decrypted text
//     fmt.Printf("Decrypted text: %s\n", string(decrypted))
// 	return string(decrypted),nil
// }

// // func main() {
// //     // Generate a new keyset handle using the AES256-GCM template
// //     kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
// //     if err != nil {
// //         fmt.Printf("Error generating keyset: %v\n", err)
// //         return
// //     }

// //     // Get an AEAD primitive from the keyset handle
// //     a, err := aead.New(kh)
// //     if err != nil {
// //         fmt.Printf("Error creating AEAD primitive: %v\n", err)
// //         return
// //     }

// //     // String to be encrypted
// //     plaintext := "this is a secret message"

// //     // Additional data to be authenticated but not encrypted (can be nil)
// //     aad := []byte("")

// //     // Encrypt the string
// //     ciphertext, err := a.Encrypt([]byte(plaintext), aad)
// //     if err != nil {
// //         fmt.Printf("Error encrypting data: %v\n", err)
// //         return
// //     }

// //     // Print the ciphertext
// //     fmt.Printf("Encrypted text: %x\n", ciphertext)

// //     // Decrypt the ciphertext back to the original plaintext
// //     decrypted, err := a.Decrypt(ciphertext, aad)
// //     if err != nil {
// //         fmt.Printf("Error decrypting data: %v\n", err)
// //         return
// //     }

// //     // Print the decrypted text
// //     fmt.Printf("Decrypted text: %s\n", string(decrypted))
// // }
