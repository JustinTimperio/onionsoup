package crypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// Encrypts and signs the message using the sender's private key and recipient's public key.
// Returns the encrypted message on success and an error otherwise.
func RSAEncryptAndSignMessage(senderPrivateKey *rsa.PrivateKey, recipientPublicKey *rsa.PublicKey, message []byte) ([]byte, []byte, error) {
	// Encrypt the message using the recipient's public key
	encryptedMessage, err := rsa.EncryptPKCS1v15(rand.Reader, recipientPublicKey, message)
	if err != nil {
		return nil, nil, fmt.Errorf("Error encrypting message: %v", err)
	}

	// Sign the message using the sender's private key
	hash := sha256.New()
	hash.Write(message)
	hashedMessage := hash.Sum(nil)

	// Sign the message using the sender's private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, senderPrivateKey, crypto.SHA256, hashedMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("Error signing message: %v", err)
	}

	return encryptedMessage, signature, nil
}

// Decrypts and verifies the message using the recipient's private key and sender's public key.
// Returns the decrypted message on success and an error otherwise.
func RSADecryptAndVerifyMessage(receiverPrivateKey *rsa.PrivateKey, senderPublicKey *rsa.PublicKey, message []byte, signature []byte) ([]byte, error) {
	// Safety check that the message and signature are not empty
	if message == nil || signature == nil {
		return nil, fmt.Errorf("Message and Signature cannot be empty")
	}

	// Decrypt the message using the recipient's private key
	decryptedMessage, err := rsa.DecryptPKCS1v15(rand.Reader, receiverPrivateKey, message)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting message: %v", err)
	}

	// Verify the signature using the sender's public key
	hash := sha256.New()
	hash.Write(decryptedMessage)
	hashedMessage := hash.Sum(nil)

	// Verify the signature using the sender's public key
	err = rsa.VerifyPKCS1v15(senderPublicKey, crypto.SHA256, hashedMessage, signature)
	if err != nil {
		return nil, fmt.Errorf("Error verifying message signature: %v", err)
	}

	return decryptedMessage, nil
}

// Generates a new RSA Key Pair and returns them as pointers.
func RSAGenerateKeyPair(size int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if size < 4096 {
		return nil, nil, fmt.Errorf("RSA Key Length must be at least 4096 bits")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// Converts a Private Key to a PEM encoded byte array.
func RSAPrivateKeyToBytes(privateKey *rsa.PrivateKey, password string) ([]byte, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encrypt the Private Key using a password if one is provided
	if len(password) > 0 {
		var err error
		privateKeyPEM, err = AESEncrypt([]byte(password), privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("Failed to encrypt private key: %v", err)
		}
	}

	return privateKeyPEM, nil
}

// Converts a PEM encoded byte array to a Private Key Pointer.
func RSAPrivateKeyToMem(privateKeyBytes []byte, password string) (*rsa.PrivateKey, error) {

	// Decrypt the Private Key using a password if one is provided
	if len(password) > 0 {
		var err error
		privateKeyBytes, err = AESDecrypt([]byte(password), privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to decrypt private key: %v", err)
		}
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("Failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// Converts a Public Key to a PEM encoded byte array.
func RSAPublicKeyToBytes(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err

	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return publicKeyPEM, nil
}

// Converts a PEM encoded byte array to a Public Key Pointer.
func RSAPublicKeyToMem(publicKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("Failed to decode PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Failed to parse RSA public key")
	}

	return publicKey, nil
}
