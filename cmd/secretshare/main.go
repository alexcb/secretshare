package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
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

func generateKey() (string, string, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return "", "", err
	}

	pub, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return "", "", err
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(pub))
	privKeyStr := marshalRSAPrivate(key)

	return pubKeyStr, privKeyStr, nil
}

// encryptAES256 returns a random passphrase and corresponding bytes encrypted with it
func encryptAES256(data []byte) ([]byte, []byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, nil, err
	}

	n := len(data)
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(n)); err != nil {
		return nil, nil, err
	}
	if _, err := buf.Write(data); err != nil {
		return nil, nil, err
	}

	paddingN := aes.BlockSize - (buf.Len() % aes.BlockSize)
	if paddingN > 0 {
		padding := make([]byte, paddingN)
		if _, err := rand.Read(padding); err != nil {
			return nil, nil, err
		}
		if _, err := buf.Write(padding); err != nil {
			return nil, nil, err
		}
	}
	plaintext := buf.Bytes()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return key, ciphertext, nil
}

func decryptAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// works inplace when both args are the same
	mode.CryptBlocks(ciphertext, ciphertext)

	buf := bytes.NewReader(ciphertext)
	var n uint64
	if err = binary.Read(buf, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	payload := make([]byte, n)
	if _, err = buf.Read(payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func encrypt(msg, publicKey []byte) (string, error) {
	parsed, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
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

	if len(msg) <= 256 {
		// msg is small enough to only use OAEP encryption; this will result in less bytes to transfer.
		encryptedBytes, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			pub,
			msg,
			nil)
		if err != nil {
			return "", err
		}
		if len(encryptedBytes) != 256 {
			panic(len(encryptedBytes))
		}
		return base64.StdEncoding.EncodeToString(encryptedBytes), nil
	}

	// otherwise, encrypt using AES256

	key, ciphertext, err := encryptAES256(msg)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		key,
		nil)
	if err != nil {
		return "", err
	}
	if len(encryptedBytes) != 256 {
		panic(len(encryptedBytes))
	}
	return base64.StdEncoding.EncodeToString(append(encryptedBytes, ciphertext...)), nil
}

func decrypt(data, priv string) ([]byte, error) {
	data2, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	if len(data2) < 256 {
		return nil, fmt.Errorf("not enough data to decrypt")
	}

	block, _ := pem.Decode([]byte(priv))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	oaepData := data2[:256]
	aesData := data2[256:]
	payload, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, oaepData, nil)
	if err != nil {
		return nil, err
	}

	if len(aesData) == 0 {
		return payload, nil
	}

	decryptedAESKey := payload
	decrypted, err := decryptAES(decryptedAESKey, aesData)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func test() {
	pub, priv, err := generateKey()
	if err != nil {
		panic(err)
	}
	pub = strings.TrimPrefix(pub, "ssh-rsa ")

	data := []byte("hello test")
	encrypted, err := encrypt(data, []byte("ssh-rsa "+pub))
	if err != nil {
		panic(err)
	}

	data2, err := decrypt(encrypted, priv)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(data, data2) {
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
		pub, priv, err := generateKey()
		if err != nil {
			return "", "", err
		}
		err = ioutil.WriteFile(privKeyPath, []byte(priv), 0600)
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
		fmt.Fprintf(os.Stderr, "failed while reading from stdin: %s\n", err.Error())
		os.Exit(1)
	}

	if arg == "decrypt" {
		data2, err := decrypt(string(data), priv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed while decrypting: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s", data2)
		return
	}

	encrypted, err := encrypt(data, []byte("ssh-rsa "+arg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed while encrypting: %s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%s\n", encrypted)
}
