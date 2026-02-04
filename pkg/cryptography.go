package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// --------------------------------------------------------- //

const (
	GCM_TAG_SIZE    = 16 // equal to EVP_GCM_TLS_TAG_LEN in OpenSSL
	ARGON2_MIN_SALT = 16
)

type Argon2idParams struct {
	Computation   uint32 // time cost
	Block         uint32 // memory cost
	Parallelism   uint32 // threads
	DerivedLength uint32 // output hash len in bytes
}

// @brief default params for Argon2idParams
//
// @note specification is reflect from my c++ backend
var Argon2idParams_default = Argon2idParams{
	Computation:   2,
	Block:         1 << 20,
	Parallelism:   2,
	DerivedLength: 32,
}

// --------------------------------------------------------- //

func PadPKCS7(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := make([]byte, padding)

	for i := range padtext {
		padtext[i] = byte(padding)
	}

	return append(src, padtext...)
}

func UnpadPKCS7(src []byte) ([]byte, error) {
	length := len(src)

	if length == 0 {
		return nil, errors.New("UnpadPKCS7 src is empty")
	}

	padding := int(src[length-1])

	if padding > length || padding > aes.BlockSize {
		return nil, errors.New("UnpadPKCS7 wrong padding")
	}

	for i := 0; i < padding; i++ {
		if src[length-1-i] != byte(padding) {
			return nil, errors.New("UnpadPKCS7 wrong padding")
		}
	}

	return src[:length-padding], nil
}

func GenerateSalt(n uint32) ([]byte, error) {
	salt := make([]byte, n)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// --------------------------------------------------------- //

func AES_CBC_Encrypt(plaintext, iv, ik []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("iv must be 16 bytes")
	}
	if len(ik) != 32 {
		return nil, errors.New("ik must be 32 bytes")
	}

	block, err := aes.NewCipher(ik)
	if err != nil {
		return nil, err
	}

	plaintext = PadPKCS7(plaintext)
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

func AES_CBC_Decrypt(ciphertext, iv, ik []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("iv must be 16 bytes")
	}
	if len(ik) != 32 {
		return nil, errors.New("ik must be 32 bytes")
	}

	block, err := aes.NewCipher(ik)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext length must be multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return UnpadPKCS7(plaintext)
}

func AES_GCM_Encrypt(plaintext, iv, ik []byte) ([]byte, error) {
	if len(ik) != 32 {
		return nil, errors.New("ik must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(ik)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, err
	}

	fullOutput := aesGCM.Seal(nil, iv, plaintext, nil)

	if len(fullOutput) < GCM_TAG_SIZE {
		return nil, errors.New("unexpected output length")
	}

	return fullOutput, nil
}

func AES_GCM_Decrypt(fullInput, iv, ik []byte) ([]byte, error) {
	if len(ik) != 32 {
		return nil, errors.New("ik must be 32 bytes for AES-256")
	}
	if len(fullInput) < GCM_TAG_SIZE {
		return nil, errors.New("input too short (missing tag)")
	}

	block, err := aes.NewCipher(ik)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, iv, fullInput, nil)
	if err != nil {
		return nil, errors.New("decryption failed: authentication tag mismatch or invalid data")
	}

	return plaintext, nil
}

// --------------------------------------------------------- //

func Argon2id(input string, salt []byte, params Argon2idParams) (string, error) {
	if len(input) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters")
	}
	if len(salt) < 16 {
		return "", fmt.Errorf("salt must be at least 16 bytes")
	}
	if params.DerivedLength == 0 {
		params.DerivedLength = 32
	}

	hash := argon2.IDKey(
		[]byte(input),
		salt,
		params.Computation,
		params.Block,
		uint8(params.Parallelism),
		params.DerivedLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Block,
		params.Computation,
		params.Parallelism,
		b64Salt,
		b64Hash)

	return encoded, nil
}

func Argon2idVerify(input, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid argon2 encoded hash format")
	}

	var mem, time, threads uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &time, &threads); err != nil {
		return false, fmt.Errorf("failed to parse argon2 parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %w", err)
	}

	actualHash := argon2.IDKey(
		[]byte(input),
		salt,
		time,
		mem,
		uint8(threads),
		uint32(len(expectedHash)),
	)

	return string(actualHash) == string(expectedHash), nil
}
