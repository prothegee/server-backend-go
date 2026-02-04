package test_unittest

import (
	"testing"

	"showcase-backend-go/pkg"
)

func Test_Argon2id(t *testing.T) {
	var err error
	var password, salt, result string = "strong123!", "abcdefghijklmnop", ""

	argon2Params := pkg.Argon2idParams_default

	result, err = pkg.Argon2id(password, []byte(salt), argon2Params)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	_, err = pkg.Argon2idVerify(password, result)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}
}
