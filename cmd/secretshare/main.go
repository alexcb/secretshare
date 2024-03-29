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
	"path"
	"path/filepath"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh"
)

var Version = "development"

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

	sum := sha256.Sum256(plaintext)
	plaintext = append(sum[:], plaintext...)

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

	expectedSum := ciphertext[:32]
	actualSum := sha256.Sum256(ciphertext[32:])
	if !bytes.Equal(expectedSum, actualSum[:]) {
		return nil, fmt.Errorf("sha256 mismatch %v vs %v", expectedSum, actualSum)
	}

	buf := bytes.NewReader(ciphertext[32:])
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
	logger("parsing public key")
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
		logger("encrypting with OAEP only")
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

	logger("encrypting with OAEP+AES256")
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
		logger("decrypted OAEP only")
		return payload, nil
	}

	decryptedAESKey := payload
	decrypted, err := decryptAES(decryptedAESKey, aesData)
	if err != nil {
		return nil, err
	}

	logger("decrypted OAEP+AES256")
	return decrypted, nil
}

func test() {
	pub, priv, err := generateKey()
	if err != nil {
		panic(err)
	}
	pub = strings.TrimPrefix(pub, "ssh-rsa ")

	for _, data := range [][]byte{
		[]byte("hello test"),
		[]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
	} {
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

type opts struct {
	Verbose bool `long:"verbose" short:"v" description:"Enable verbose logging"`
	Version bool `long:"version" short:"V" description:"Print version and exit"`
}

var loggerEnabled bool

func logger(msg string, args ...interface{}) {
	if loggerEnabled {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
}

func main() {
	test()

	programName := "secretshare"
	if len(os.Args) > 0 {
		programName = path.Base(os.Args[0])
	}

	progOpts := opts{}
	p := flags.NewNamedParser("", flags.PrintErrors|flags.PassDoubleDash|flags.PassAfterNonOption|flags.HelpFlag)
	_, err := p.AddGroup(fmt.Sprintf("%s [options] args", programName), "", &progOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	args, err := p.ParseArgs(os.Args[1:])
	if err != nil {
		p.WriteHelp(os.Stderr)
		os.Exit(1)
	}
	loggerEnabled = progOpts.Verbose

	if progOpts.Version {
		fmt.Printf("secretshare version: %s\n", Version)
		os.Exit(0)
	}

	pub, priv, err := genOrLoadKeys()
	if err != nil {
		panic(err)
	}
	pub = strings.TrimPrefix(pub, "ssh-rsa ")

	var data []byte
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed while reading from stdin: %s\n", err.Error())
			os.Exit(1)
		}
	}

	mode := "help"
	var pubKey string
	if len(data) > 0 {
		if len(args) == 0 {
			mode = "decrypt"
		} else if len(args) == 1 {
			mode = "encrypt"
			pubKey = args[0]
		}
	}

	if mode == "help" {
		fmt.Printf("To decrypt data, run: %s < file_to_decrypt\n", programName)
		fmt.Printf("To encrypt data, run: %s <encryption_key> < data_to_encrypt\n", programName)
		fmt.Printf("\n")
		fmt.Printf("For example if someone wanted to send you data, they would run:\n%s %s < data_to_encrypt\n", programName, pub)
		return
	}

	if mode == "decrypt" {
		logger("decrypting %d bytes", len(data))
		data2, err := decrypt(string(data), priv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed while decrypting: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s", data2)
		return
	}

	if mode != "encrypt" {
		panic("shouldnt happen")
	}

	logger("encrypting %d bytes", len(data))
	encrypted, err := encrypt(data, []byte("ssh-rsa "+pubKey))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed while encrypting: %s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%s\n", encrypted)
}
