package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

func marshalRSAPrivate(priv *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))
}

func generateKey() (string, string) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(err)
	}

	pub, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		panic(err)
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(pub))
	privKeyStr := marshalRSAPrivate(key)

	return pubKeyStr, privKeyStr
}

func encrypt(msg, publicKey string) (string, error) {
	parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", err
	}
	// To get back to an *rsa.PublicKey, we need to first upgrade to the
	// ssh.CryptoPublicKey interface
	parsedCryptoKey := parsed.(ssh.CryptoPublicKey)

	// Then, we can call CryptoPublicKey() to get the actual crypto.PublicKey
	pubCrypto := parsedCryptoKey.CryptoPublicKey()

	// Finally, we can convert back to an *rsa.PublicKey
	pub := pubCrypto.(*rsa.PublicKey)

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		[]byte(msg),
		nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

func promptLine(msg string) string {
	fmt.Println(msg)
	reader := bufio.NewReader(os.Stdin)
	s, _ := reader.ReadString('\n')
	return s
}

func decrypt(data, priv string) (string, error) {
	data2, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode([]byte(priv))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, data2, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func encryptMain(pub string) {
	data := promptLine("enter data to encrypt")
	encrypted, err := encrypt(data, "ssh-rsa "+pub)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--- Below is the encrypted data ---\n")
	fmt.Println(encrypted)
}

func test() {
	pub, priv := generateKey()
	pub = strings.TrimPrefix(pub, "ssh-rsa ")

	data := "hello test"
	encrypted, err := encrypt(data, "ssh-rsa "+pub)
	if err != nil {
		panic(err)
	}

	data2, err := decrypt(encrypted, priv)
	if err != nil {
		panic(err)
	}

	if data != data2 {
		panic("missmatch")
	}
}

func main() {
	test()
	if len(os.Args) > 1 {
		encryptMain(os.Args[1])
		return
	}

	pub, priv := generateKey()
	pub = strings.TrimPrefix(pub, "ssh-rsa ")

	fmt.Printf("run this command to encrypt data: %s %s\n", os.Args[0], pub)

	data := promptLine("enter encrypted data")

	data2, err := decrypt(data, priv)
	if err != nil {
		panic(err)
	}
	fmt.Println(data2)
}
