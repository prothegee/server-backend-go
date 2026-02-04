package pkg

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// --------------------------------------------------------- //

var (
	regexUuidV1 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-1[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	regexUuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	regexUuidV7 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

type Uuid_e int

const (
	UUID_UNDEFINED Uuid_e = iota
	UUID_V1
	UUID_V4
	UUID_V7
)

const (
	ALPHANUMERIC = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// --------------------------------------------------------- //

func GenerateUUID(t Uuid_e) (uuid.UUID, error) {
	switch t {
	case UUID_V1:
		{
			return uuid.NewUUID()
		}
	case UUID_V4:
		{
			return uuid.NewRandom()
		}
	case UUID_V7:
		{
			return uuid.NewV7()
		}
	default:
		{
			return uuid.Nil, errors.New("undefined uuid")
		}
	}
}

// copy origin dir to target dir
//
// origin & target are relative from executeable
//
// params:
// 	o string - origin dir
// 	t string - target dir
// 	f bool - force overwrite if true
//
// return: error
func CopyDir(o, t string, f bool) error {
	return filepath.WalkDir(o, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(o, p)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(t, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if _, err := os.Stat(dstPath); err == nil {
			if !f {
				return nil
			}
			// overwrite otherwise
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0644)
	})
}

// check if id is valid
//
// apply for unsigned integer
//
// params:
// 	id string
//
// return: error
func IsValidId(id string) error {
	if id == "" {
		return errors.New("empty value")
	}

	_, err := strconv.ParseUint(id, 10, 64)
	if err == nil {
		return nil
	}

	return errors.New("unknown uint")
}

// check if uuid is valid
//
// apply for uuid, check your trim parse
//
// params:
// 	input string
//
// return: (Uuid_e, error)
func IsValidUuid(input string) (Uuid_e, error) {
	if input == "" {
		return UUID_UNDEFINED, errors.New("empty value")
	}

	uuid := strings.TrimSpace(input)

	if regexUuidV1.MatchString(strings.ToLower(uuid)) {
		return UUID_V1, nil
	}

	if regexUuidV4.MatchString(strings.ToLower(uuid)) {
		return UUID_V4, nil
	}

	if regexUuidV7.MatchString(strings.ToLower(uuid)) {
		return UUID_V7, nil
	}

	return UUID_UNDEFINED, errors.New("unknown uuid")
}

// parse param and check what kind of authorization header
//
// params:
//	v string - "Authorization" header value
//
// return: (scheme, credential string, err error)
func ParseAuthorizationHeader(v string) (scheme, credential string, err error) {
	if v == "" {
		return "", "", errors.New("expecting authorization value")
	}

	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 {
		return "", "", errors.New("wrong authorization format")
	}

	scheme = strings.TrimSpace(parts[0])
	credential = strings.TrimSpace(parts[1])

	return scheme, credential, nil
}

// generate random alphanumeric
//
// params:
//	length int
//
// return: (string, error) - (actual string, nil if ok)
func GenRandomAlphanumeric(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length can't be empty")
	}

	result := make([]byte, length)

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		result[i] = ALPHANUMERIC[randomBytes[i]%byte(len(ALPHANUMERIC))]
	}

	return string(result), nil
}

// generate random number
//
// params:
// 	min int - min range
// 	max int - max range
//
// return: (int, error)
func GenRandomNumber(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min (%d) cannot be greater than max (%d)", min, max)
	}
	if min == max {
		return min, nil
	}

	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return 0, err
	}

	// convert to uint64 then int (sys 64)
	r := binary.LittleEndian.Uint64(buf[:])
	rangeSize := uint64(max - min + 1)
	result := min + int(r%rangeSize)
	return result, nil
}

// timestamp now
//
// return: string
func TimestampNow() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05.000000000")
}
