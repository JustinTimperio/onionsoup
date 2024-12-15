package crypt

import (
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
)

func PGPEncryptAndSignMessage(privateKey *crypto.Key, publicKey *crypto.Key, message []byte) ([]byte, error) {
	enc, err := crypto.PGP().Encryption().Recipient(publicKey).SigningKey(privateKey).New()
	if err != nil {
		return nil, err
	}

	encryptedMessage, err := enc.Encrypt(message)
	if err != nil {
		return nil, err
	}

	armoredBytes, err := encryptedMessage.ArmorBytes()
	if err != nil {
		return nil, err
	}

	return armoredBytes, nil
}

func PGPDecryptAndVerifyMessage(privateKey *crypto.Key, publicKey *crypto.Key, message []byte) ([]byte, error) {
	enc, err := crypto.PGP().Decryption().DecryptionKey(privateKey).VerificationKey(publicKey).New()
	if err != nil {
		return nil, err
	}

	decryptedMessage, err := enc.Decrypt(message, crypto.Armor)
	if err != nil {
		return nil, err
	}

	return decryptedMessage.Bytes(), nil
}

func PGPGenerateKeyPair(accountName string) (*crypto.Key, error) {
	if accountName == "" {
		accountName = "anon"
	}

	return crypto.PGPWithProfile(profile.Default()).
		KeyGeneration().
		AddUserId(accountName, accountName).
		New().
		GenerateKey()
}

func PGPPublicKeyToBytes(key *crypto.Key) ([]byte, error) {
	b, err := key.GetArmoredPublicKey()
	return []byte(b), err
}

func PGPPublicKeyToMem(publicBytes []byte) (*crypto.Key, error) {
	return crypto.NewKeyFromArmored(string(publicBytes))
}

func PGPPrivateKeyToBytes(key *crypto.Key, password string) ([]byte, error) {
	lockedKey, err := crypto.PGP().LockKey(key, []byte(password))
	if err != nil {
		return nil, err
	}

	b, err := lockedKey.Armor()
	if err != nil {
		return nil, err
	}

	return []byte(b), nil
}

func PGPPrivateKeyToMem(privateBytes []byte, password string) (*crypto.Key, error) {
	return crypto.NewPrivateKeyFromArmored(string(privateBytes), []byte(password))
}
