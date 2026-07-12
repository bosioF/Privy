package crypto

import (
	"crypto/rand"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/sha512"
	"net"
	"errors"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"bufio"
	"strings"
)

func GenSessionKey() ([]byte, error){
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func EncryptMsg(plaintext []byte, key []byte) ([]byte, error){
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	GCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, GCM.NonceSize())
	
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	EncMsg := GCM.Seal(nonce, nonce, plaintext, nil)
	return EncMsg, nil
}

func DecryptMsg(ciphertext []byte, key []byte) ([]byte, error){
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	GCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:GCM.NonceSize()]
	EncMsg := ciphertext[GCM.NonceSize():]

	DecMsg, err := GCM.Open(nil, nonce, EncMsg, nil)
	if err != nil {
		return nil, err
	}

	return DecMsg, nil

}

func GenECDHKeys() (*ecdh.PrivateKey, []byte, error){
	privKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pubBytes := privKey.PublicKey().Bytes()

	return privKey, pubBytes, nil
}

func DeriveSharedSecret(privKey *ecdh.PrivateKey, pubBytes[]byte) ([]byte, error){
	pubKey, err := ecdh.X25519().NewPublicKey(pubBytes)
	if err != nil {
		return nil, err 
	}

	sharedKey, err := privKey.ECDH(pubKey)
	if err != nil {
		return nil, err
	}

	return sharedKey, nil

}

func Handshake(conn net.Conn, scanner *bufio.Scanner) ([]byte, error){
	privKey, pubBytes, err := GenECDHKeys()

	b64pubBytes := base64.StdEncoding.EncodeToString(pubBytes)
	fmt.Fprintln(conn, b64pubBytes)

	if !scanner.Scan() {
		return nil, errors.New("error while receiving pub key")
	}

	b64remotePub := scanner.Text()
	remotePubBytes, err := base64.StdEncoding.DecodeString(b64remotePub)
	if err != nil {
		return nil, err
	}

	sharedKey, err := DeriveSharedSecret(privKey, remotePubBytes)
	if err != nil {
		return nil, err
	}

	fmt.Println("Handshake successful! Connection Secure!")
	fmt.Println("If you suspect someone is atttemping a MitM attack, verify that this code is the same as the other person, over another secure channel.")
	
	sharedKeyHash := sha512.New()
	sharedKeyHash.Write(sharedKey)
	SAScode := sharedKeyHash.Sum(nil)
	hexSAScode := hex.EncodeToString(SAScode)

	var chunks []string
	for i := 0; i < len(hexSAScode)/4; i += 4 {
		end := i + 4
		if end > len(hexSAScode)/4 {
			end = len(hexSAScode)/4
		}
		chunks = append(chunks, hexSAScode[i:end])
	}


	finalSAScode := strings.Join(chunks, "-")
	fmt.Println("SAS Code: ", finalSAScode, "\n")

	return sharedKey, nil
}
