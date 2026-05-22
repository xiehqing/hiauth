package security

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5Password(password string) string {
	sum := md5.Sum([]byte(password))
	return hex.EncodeToString(sum[:])
}
