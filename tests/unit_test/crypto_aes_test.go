package test_unittest

import (
	"testing"

	"showcase-backend-go/pkg"
)

func Test_AES_CBC(t *testing.T) {
	messageS := "this is your message"
	ivS := "abcdefghijklmnop"
	ikS := "abcdefghijklmnopqrstuvwxyz012345"

	messageB := []byte(messageS)
	ivB := []byte(ivS)
	ikB := []byte(ikS)

	encrypted, err := pkg.AES_CBC_Encrypt(messageB, ivB, ikB)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	decrypted, err := pkg.AES_CBC_Decrypt(encrypted, ivB, ikB)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	if string(decrypted) != messageS {
		t.Errorf("ERROR: original message is not match with decrypted value\n")
	}
}

func Test_AES_GCM(t *testing.T) {
	messageS := "this is your message"
	ivS := "abcdefghijklmnop"
	ikS := "abcdefghijklmnopqrstuvwxyz012345"

	messageB := []byte(messageS)
	ivB := []byte(ivS)
	ikB := []byte(ikS)

	encrypted, err := pkg.AES_GCM_Encrypt(messageB, ivB, ikB)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	decrypted, err := pkg.AES_GCM_Decrypt(encrypted, ivB, ikB)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	if string(decrypted) != messageS {
		t.Errorf("ERROR: original message is not match with decrypted value\n")
	}
}
