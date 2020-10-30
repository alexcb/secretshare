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
	"io/ioutil"
	"os"
	"path/filepath"
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func genOrLoadKeys() (string, string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	privKeyPath := filepath.Join(homeDir, ".secretshare")
	pubKeyPath := filepath.Join(homeDir, ".secretshare.pub")

	if !fileExists(privKeyPath) {
		pub, priv := generateKey()
		err := ioutil.WriteFile(privKeyPath, []byte(priv), 0600)
		if err != nil {
			return "", "", err
		}
		err = ioutil.WriteFile(pubKeyPath, []byte(pub), 0644)
		if err != nil {
			return "", "", err
		}
	}

	priv, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return "", "", err
	}
	pub, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return "", "", err
	}

	return string(pub), string(priv), nil
}

func main() {
	test()
	pub, priv, err := genOrLoadKeys()
	if err != nil {
		panic(err)
	}
	pub = strings.TrimPrefix(pub, "ssh-rsa ")

	if len(os.Args) <= 1 {
		appName := "secret-share"
		if len(os.Args) > 0 {
			appName = os.Args[0]
		}
		fmt.Printf("To decrypt data, run: %s decrypt < file_to_decrypt\n", appName)
		fmt.Printf("To encrypt data, run: %s <encryption_key> < data_to_encrypt\n", appName)
		fmt.Printf("\n")
		fmt.Printf("For example if someone wanted to send you data, they would run:\n%s %s < data_to_encrypt\n", appName, pub)
		return
	}

	arg := os.Args[1]

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if arg == "decrypt" {
		data2, err := decrypt(string(data), priv)
		if err != nil {
			panic(err)
		}
		fmt.Println(data2)
		return
	}

	encrypted, err := encrypt(string(data), "ssh-rsa "+arg)
	if err != nil {
		panic(err)
	}
	fmt.Println(encrypted)
}
