/*
 * Copyright 2020 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cryptogo

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"syscall"
)

const (
	AES_128_KEY_SIZE_BITS  = 128
	AES_128_KEY_SIZE_BYTES = AES_128_KEY_SIZE_BITS / 8

	AES_256_KEY_SIZE_BITS  = 256
	AES_256_KEY_SIZE_BYTES = AES_256_KEY_SIZE_BITS / 8

	MIN_PASSWORD = 12

	privateKeyFileExtension = ".go.private"
)

func RandomString(size uint) (string, error) {
	buffer := make([]byte, size)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func Hash(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

func HashProtobuf(protobuf proto.Message) ([]byte, error) {
	data, err := proto.Marshal(protobuf)
	if err != nil {
		return nil, err
	}
	return Hash(data), nil
}

func RSAPublicKeyToPKCS1Bytes(publicKey *rsa.PublicKey) []byte {
	return x509.MarshalPKCS1PublicKey(publicKey)
}

func RSAPublicKeyToPKIXBytes(publicKey *rsa.PublicKey) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func RSAPublicKeyFromPKCS1Bytes(data []byte) (*rsa.PublicKey, error) {
	pub, err := x509.ParsePKCS1PublicKey(data)
	if err != nil {
		return nil, err
	}
	return PublicKeyToRSAPublicKey(pub)
}

func RSAPublicKeyFromPKIXBytes(data []byte) (*rsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	return PublicKeyToRSAPublicKey(pub)
}

func PublicKeyToRSAPublicKey(key interface{}) (*rsa.PublicKey, error) {
	switch k := key.(type) {
	case *rsa.PublicKey:
		return k, nil
	default:
		return nil, ErrUnsupportedPublicKeyType{Type: fmt.Sprintf("%v", k)}
	}
}

func RSAPublicKeyToPEM(publicKey *rsa.PublicKey) (*pem.Block, error) {
	// Marshal public key into PKIX
	bytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	// Create PEM block
	return &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytes,
	}, nil
}

func RSAPrivateKeyToPKCS1Bytes(privateKey *rsa.PrivateKey) []byte {
	return x509.MarshalPKCS1PrivateKey(privateKey)
}

func RSAPrivateKeyToPKCS8Bytes(privateKey *rsa.PrivateKey) ([]byte, error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func RSAPrivateKeyFromPKCS1Bytes(data []byte) (*rsa.PrivateKey, error) {
	priv, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return nil, err
	}
	return PrivateKeyToRSAPrivateKey(priv)
}

func RSAPrivateKeyFromPKCS8Bytes(data []byte) (*rsa.PrivateKey, error) {
	priv, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		return nil, err
	}
	return PrivateKeyToRSAPrivateKey(priv)
}

func PrivateKeyToRSAPrivateKey(key interface{}) (*rsa.PrivateKey, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return k, nil
	default:
		return nil, ErrUnsupportedPrivateKeyType{Type: fmt.Sprintf("%v", k)}
	}
}

func RSAPrivateKeyToPEM(privateKey *rsa.PrivateKey, password []byte) (*pem.Block, error) {
	// Create encrypted PEM block with private key marshalled into PKCS8
	data, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return x509.EncryptPEMBlock(rand.Reader, "ENCRYPTED PRIVATE KEY", data, password, x509.PEMCipherAES128)
}

func ParseRSAPublicKey(publicKey []byte, format PublicKeyFormat) (*rsa.PublicKey, error) {
	switch format {
	case PublicKeyFormat_PKCS1_PUBLIC:
		return RSAPublicKeyFromPKCS1Bytes(publicKey)
	case PublicKeyFormat_PKIX:
		fallthrough
	case PublicKeyFormat_X509:
		return RSAPublicKeyFromPKIXBytes(publicKey)
	case PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT:
		fallthrough
	default:
		return nil, ErrUnsupportedPublicKeyFormat{Format: format.String()}
	}
}

func ParseRSAPrivateKey(privateKey []byte, format PrivateKeyFormat) (*rsa.PrivateKey, error) {
	switch format {
	case PrivateKeyFormat_PKCS1_PRIVATE:
		return RSAPrivateKeyFromPKCS1Bytes(privateKey)
	case PrivateKeyFormat_PKCS8:
		return RSAPrivateKeyFromPKCS8Bytes(privateKey)
	case PrivateKeyFormat_UNKNOWN_PRIVATE_KEY_FORMAT:
		fallthrough
	default:
		return nil, ErrUnsupportedPrivateKeyFormat{Format: format.String()}
	}
}

func HasRSAPrivateKey(directory, name string) bool {
	_, err := os.Stat(path.Join(directory, name+privateKeyFileExtension))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ListRSAPrivateKeys(directory string) ([]string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, privateKeyFileExtension) {
			keys = append(keys, strings.TrimSuffix(name, privateKeyFileExtension))
		}
	}
	return keys, nil
}

func CreateRSAPrivateKey(directory, name string, password []byte) (*rsa.PrivateKey, error) {
	// Create directory
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return nil, err
	}

	log.Println("Generating RSA-4096bit Public/Private Key Pair")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	if err := WriteRSAPrivateKey(privateKey, directory, name, password); err != nil {
		return nil, err
	}

	return privateKey, nil
}

func WriteRSAPrivateKey(privateKey *rsa.PrivateKey, directory, name string, password []byte) error {
	// Encode Private Key to PEM block
	privateKeyPEM, err := RSAPrivateKeyToPEM(privateKey, password)
	if err != nil {
		return err
	}

	// Write Private Key PEM block to file
	if err := WritePEM(privateKeyPEM, path.Join(directory, name+privateKeyFileExtension)); err != nil {
		return err
	}

	return nil
}

func RSAPrivateKey(directory, name string, password []byte) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := ReadPEM(path.Join(directory, name+privateKeyFileExtension))
	if err != nil {
		return nil, err
	}

	decrypted, err := x509.DecryptPEMBlock(privateKeyPEM, password)
	if err != nil {
		return nil, err
	}

	priv, err := x509.ParsePKCS8PrivateKey(decrypted)
	if err != nil {
		return nil, err
	}
	return PrivateKeyToRSAPrivateKey(priv)
}

func Password() ([]byte, error) {
	pwd, ok := os.LookupEnv("PASSWORD")
	if ok {
		return []byte(pwd), nil
	} else {
		return ReadPassword("Enter password: ")
	}
}

func ReadPassword(prompt string) ([]byte, error) {
	log.Print(prompt)
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	log.Println()
	return password, nil
}

func LoadRSAPrivateKey(directory, name string) (*rsa.PrivateKey, error) {
	log.Println("Alias:", name)
	password, err := Password()
	if err != nil {
		return nil, err
	}
	if HasRSAPrivateKey(directory, name) {
		key, err := RSAPrivateKey(directory, name, password)
		if err != nil {
			return nil, err
		}

		return key, nil
	} else {
		log.Println("Creating key in " + directory + " for " + name)

		confirm, err := ReadPassword("Confirm password: ")
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(password, confirm) {
			return nil, ErrPasswordsDoNotMatch{}
		}

		key, err := CreateRSAPrivateKey(directory, name, password)
		if err != nil {
			return nil, err
		}

		log.Println("Successfully Created Key Pair")
		return key, nil
	}
}

func ExportKeys(host, keystore, name string, password []byte) (string, error) {
	privateKey, err := RSAPrivateKey(keystore, name, password)
	if err != nil {
		return "", err
	}

	// Generate a random access code
	accessCode, err := GenerateRandomKey(AES_128_KEY_SIZE_BYTES)
	if err != nil {
		return "", err
	}

	data, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	encryptedPrivateKeyBytes, err := EncryptAESGCM(accessCode, data)
	if err != nil {
		return "", err
	}
	publicKeyBytes, err := RSAPublicKeyToPKIXBytes(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}
	encryptedPassword, err := EncryptAESGCM(accessCode, password)
	if err != nil {
		return "", err
	}
	response, err := http.PostForm(host+"/keys", url.Values{
		"name":             {name},
		"publicKey":        {base64.RawURLEncoding.EncodeToString(publicKeyBytes)},
		"publicKeyFormat":  {"PKIX"},
		"privateKey":       {base64.RawURLEncoding.EncodeToString(encryptedPrivateKeyBytes)},
		"privateKeyFormat": {"PKCS8"},
		"password":         {base64.RawURLEncoding.EncodeToString(encryptedPassword)},
	})
	if err != nil {
		return "", err
	}
	switch response.StatusCode {
	case http.StatusOK:
		log.Println("Keys exported")
		return base64.RawURLEncoding.EncodeToString(accessCode), nil
	default:
		return "", ErrExportFailed{StatusCode: response.StatusCode, Status: response.Status}
	}
}

func ImportKeys(host, keystore, name, accessCode string) error {
	response, err := http.Get(host + "/keys?name=" + name)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(data) <= 0 {
		return fmt.Errorf("Could not get KeyShare for %s", name)
	}
	keyShare := &KeyShare{}
	if err = proto.Unmarshal(data, keyShare); err != nil {
		return err
	}
	if name != keyShare.Name {
		return fmt.Errorf("Incorrect KeyShare Name: %s vs %s", name, keyShare.Name)
	}
	// Decode Access Code
	decodedAccessCode, err := base64.RawURLEncoding.DecodeString(accessCode)
	if err != nil {
		return err
	}
	// Decrypt Private Key
	decryptedPrivateKey, err := DecryptAESGCM(decodedAccessCode, keyShare.PrivateKey)
	if err != nil {
		return err
	}
	// Parse Private Key
	privateKey, err := ParseRSAPrivateKey(decryptedPrivateKey, keyShare.PrivateFormat)
	if err != nil {
		return err
	}
	// Decrypt Password
	decryptedPassword, err := DecryptAESGCM(decodedAccessCode, keyShare.Password)
	if err != nil {
		return err
	}
	// Write Private Key
	if err := WriteRSAPrivateKey(privateKey, keystore, name, decryptedPassword); err != nil {
		return err
	}
	log.Println("Keys imported")
	return nil
}

func DecryptKey(algorithm EncryptionAlgorithm, secret []byte, key *rsa.PrivateKey) ([]byte, error) {
	switch algorithm {
	case EncryptionAlgorithm_RSA_ECB_OAEPPADDING:
		// Decrypt a shared key
		return rsa.DecryptOAEP(sha512.New(), rand.Reader, key, secret, nil)
	case EncryptionAlgorithm_UNKNOWN_ENCRYPTION:
		return secret, nil
	default:
		return nil, ErrUnsupportedEncryption{Algorithm: algorithm.String()}
	}
}

func DecryptPayload(algorithm EncryptionAlgorithm, key []byte, payload []byte) ([]byte, error) {
	switch algorithm {
	case EncryptionAlgorithm_AES_128_GCM_NOPADDING:
		fallthrough
	case EncryptionAlgorithm_AES_256_GCM_NOPADDING:
		return DecryptAESGCM(key, payload)
	case EncryptionAlgorithm_UNKNOWN_ENCRYPTION:
		return payload, nil
	default:
		return nil, ErrUnsupportedEncryption{Algorithm: algorithm.String()}
	}
}

func GenerateRandomKey(bytes int) ([]byte, error) {
	key := make([]byte, bytes)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func EncryptAESGCM(key, payload []byte) ([]byte, error) {
	// Create cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create galois counter mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt payload
	encrypted := append(nonce, gcm.Seal(nil, nonce, payload, nil)...)

	return encrypted, nil
}

func DecryptAESGCM(key, encrypted []byte) ([]byte, error) {
	// Create cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create galois counter mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// Get nonce
	nonce := encrypted[:gcm.NonceSize()]
	// Get payload
	payload := encrypted[gcm.NonceSize():]

	// Decrypt payload
	return gcm.Open(nil, nonce, payload, nil)
}

func CreateSignature(privateKey *rsa.PrivateKey, data []byte, algorithm SignatureAlgorithm) ([]byte, error) {
	switch algorithm {
	case SignatureAlgorithm_SHA512WITHRSA:
		return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA512, data)
	case SignatureAlgorithm_SHA512WITHRSA_PSS:
		var options rsa.PSSOptions
		options.SaltLength = rsa.PSSSaltLengthAuto
		return rsa.SignPSS(rand.Reader, privateKey, crypto.SHA512, data, &options)
	case SignatureAlgorithm_UNKNOWN_SIGNATURE:
		fallthrough
	default:
		return nil, ErrUnsupportedSignature{Algorithm: algorithm.String()}
	}
}

func VerifySignature(publicKey *rsa.PublicKey, data, signature []byte, algorithm SignatureAlgorithm) error {
	switch algorithm {
	case SignatureAlgorithm_SHA512WITHRSA:
		return rsa.VerifyPKCS1v15(publicKey, crypto.SHA512, data, signature)
	case SignatureAlgorithm_SHA512WITHRSA_PSS:
		var options rsa.PSSOptions
		options.SaltLength = rsa.PSSSaltLengthAuto
		return rsa.VerifyPSS(publicKey, crypto.SHA512, data, signature, &options)
	case SignatureAlgorithm_UNKNOWN_SIGNATURE:
		fallthrough
	default:
		return ErrUnsupportedSignature{Algorithm: algorithm.String()}
	}
}

func ReadPEM(filename string) (*pem.Block, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)

	return block, nil
}

func WritePEM(key *pem.Block, filename string) error {
	return ioutil.WriteFile(filename, pem.EncodeToMemory(key), os.ModePerm)
}
