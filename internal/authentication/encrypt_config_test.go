package authentication

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"hiauth/internal/security"
	"testing"
)

func TestRSAOAEPPasswordEncryptDecryptMe(t *testing.T) {
	publicKeyPEM := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAy2YURbafhRvOxSBb89m68M1bOdFIQiReUu79jvnSUHcQv34CBU2UBWTSQ7VD/6Kyng0mN3lcXV0JDPWI0owdnx1qMZynfVxsnJ2Bsn/W9mrVuSCdnsn7g0jI6mcAa5/ES3KOeG/M/5jI0VeLRV8KGlQ/fxyjwP91ZgA3NkkKHK/fXfPbnfzo9Za9q5OW2K1Gy6dzv5LKi1H7xDBynY9s6PFhhVFaNGwHbxn7Z8FIUUHeVSRhiqln27oO55AORDqgZRQSt0zvUdoICi6YBChJuY5Lp2jRueQju/S0N98tTxctY9sSzcZDb2whuybylHusASOEPFOXZz1I/MFVpWkZQwIDAQAB"
	privateKeyBase64 := "MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDLZhRFtp+FG87FIFvz2brwzVs50UhCJF5S7v2O+dJQdxC/fgIFTZQFZNJDtUP/orKeDSY3eVxdXQkM9YjSjB2fHWoxnKd9XGycnYGyf9b2atW5IJ2eyfuDSMjqZwBrn8RLco54b8z/mMjRV4tFXwoaVD9/HKPA/3VmADc2SQocr99d89ud/Oj1lr2rk5bYrUbLp3O/ksqLUfvEMHKdj2zo8WGFUVo0bAdvGftnwUhRQd5VJGGKqWfbug7nkA5EOqBlFBK3TO9R2ggKLpgEKEm5jkunaNG55CO79LQ33y1PFy1j2xLNxkNvbCG7JvKUe6wBI4Q8U5dnPUj8wVWlaRlDAgMBAAECggEANT0Y3U553ptyucJIj0FUXydBU5bl9PoT/l0z3TKizBD+P0E6Qi0WK3tCVhqiG814N0p10Fthw8ZQUiYKlCG8tyM54paAeJ2yiCPqCNDRcVpxaq2Q1DlGLgzlGgWv5HvDI4RuqjOJUvWoyrLBb3z65f1bSWgzrJaxIeu4h+sCTJN8ruKE7NEuwU1Onhx4S6hk8CM2LLVLAyv6rxxFRvGsZgfRorDPmT+PZVfknYw/kVqTxMZMoGTVVQdeuABOHv8uK7rTMj3rfPQE0ZiLZGCZr3mBjx21WKRShVR5aY7LSHA+n4BDNZ85xb/t2LdSQafvJ3wBa3scravv2e2/VqORkQKBgQDuitKgVoY6qBZz4QHFHwyDc/KHcJ8Cd83bUESidaSaXv263SkTOpuxMmTQDpMM2zIaUjdBaNPOqwGdIZY/Pxl/GcCSdaiZjdLiQBm8UtNMW8BRYArLavDqkmpQNoaM61EiX8UTdhNnUQyDi0s/KM4SLwbtY2wJDe9RyxLV5/BhXQKBgQDaSNRZC/n0nZw5+x1oqSv93XrdkKdvuK0UUY19ble3DdBRG2QWO/9tHdON8TWtbInDuP5bOCsOIks1CyWA1KNWFJntPiIV6M6ey8MuamKZYIC/UjZ4WX26IzK32IMWLbKJc+kREtWkFK13MmC3Mla0EK8gNknmgYeMUCHADiKbHwKBgCdJwgsadR0wFhKb2pjG1l7IOAfKqsXTSZp3i/Zd/fBW+N9QEbXTD1WOAUCrRdj2OThQlj01sLz3OVrR71cXY3Gloiv9KPmxfCw7doGn+pk2+2Prt5ttT6Sy3MO9V0fachCBSYo9BlEb7j20MX6Dj/06tZ9foqmTG/mSwtVsUBEZAoGAfquKxo3jpCceNKtbmpOpWq1/EjpSX8vMbKESuXoh3rFedOKvRxPUGu8XCCS0oIn+vByLRkYm/hG6kPKB9evvSRG1bW4D+7DYzl+ySSolQ5ozvFKqF1bfVff9A6DaGTG1jHw+ANFsNsZlD2mlpEnK9L1F0yyN3/zEuxD5NOk+/cMCgYEAlcQi9zLx/mNrvDO6Fjc4PP2l4kPk88jFnZHbTs93zwMPZPCO0scHh2HAd+/KPf3hEJcES1iTifjcV3/ZntYyHsxiz3eIPwxLCZ/m2Myx+O5oNPDrrULfMFjGDMMCpF0ugv82FhjpLDW51m91kny97r0FcKhlOen2lw8efXitUgQ="
	password := "Admin@123456"
	encryptedPassword, err := rsaEncryptPassword(password, publicKeyPEM)
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}
	t.Log(encryptedPassword)
	parsedPrivateKey, err := security.ParseRSAPrivateKey(privateKeyBase64)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	cipherText, err := security.DecodeBase64(encryptedPassword)
	if err != nil {
		t.Fatalf("decode encrypted password: %v", err)
	}
	plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, parsedPrivateKey, cipherText, nil)
	if err != nil {
		t.Fatalf("decrypt password: %v", err)
	}
	t.Log(string(plainText))
	t.Log(password)
	if string(plainText) != password {
		t.Fatalf("password mismatch, got %q want %q", string(plainText), password)
	}
	cipherText1, err := security.DecodeBase64("CaSdA3cGag8UWtmKDRG/m1GJFw919NrOmmhvJol33ChdP/wBzlM8FXMAgRuO27NxNfGQleMbcYO1YioifWTxXXE7vV8e6tig/4rAKqbavTaMoAJpcNk2aP6SCFlDDkg9JShAfPbyzSd3EgBl5k0T3iQ0G2POgHwFdQxOho4W3F9v+CxocOI7afRPweXYAayCh/CJ/AYN14F/D05uuMJhVlg6Y5Kn+y+LteLSRpP2qEDNmZhich4sgsBD31fag9kAYrSKcOwVb2hhe0Xc7wR7kf1oVVTZYcVjf5eUzhDof2Ro/oyQ9BAn/vubomB+HHR/AjbhxgwnbtlUOBgdrRM7YQ==")
	if err != nil {
		t.Fatalf("decode encrypted password: %v", err)
	}
	plainText1, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, parsedPrivateKey, cipherText1, nil)
	if err != nil {
		t.Fatalf("decrypt password: %v", err)
	}
	t.Log(plainText1)
}

func TestRSAOAEPPasswordEncryptDecrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyDER)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicKeyPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}))

	password := "Admin@123456"
	encryptedPassword, err := rsaEncryptPassword(password, publicKeyPEM)
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}

	parsedPrivateKey, err := security.ParseRSAPrivateKey(privateKeyBase64)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	cipherText, err := security.DecodeBase64(encryptedPassword)
	if err != nil {
		t.Fatalf("decode encrypted password: %v", err)
	}
	plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, parsedPrivateKey, cipherText, nil)
	if err != nil {
		t.Fatalf("decrypt password: %v", err)
	}
	if string(plainText) != password {
		t.Fatalf("password mismatch, got %q want %q", string(plainText), password)
	}
}

func rsaEncryptPassword(password string, publicKey string) (string, error) {
	block, _ := pem.Decode([]byte(security.NormalizePublicKeyPEM(publicKey)))
	publicKeyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPublicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		return "", ErrInvalidRSAKey
	}
	cipherText, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPublicKey, []byte(password), nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}
