package crypt

import (
	"encoding/base64"
	"encoding/json"
)

type MessageWrapper struct {
	Message   []byte `json:"message"`
	Signature []byte `json:"signature"`
	Pubkey    []byte `json:"pubkey"`
}

func PackMessage(message []byte, signature []byte, pubkey []byte) (string, error) {
	wrapper := &MessageWrapper{
		Message:   message,
		Signature: signature,
		Pubkey:    pubkey,
	}

	jsonBytes, err := json.Marshal(wrapper)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

func UnpackMessage(message string) ([]byte, []byte, []byte, error) {
	jsonBytes, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return nil, nil, nil, err
	}

	wrapper := &MessageWrapper{}
	err = json.Unmarshal(jsonBytes, wrapper)
	if err != nil {
		return nil, nil, nil, err
	}

	return wrapper.Message, wrapper.Signature, wrapper.Pubkey, nil
}
