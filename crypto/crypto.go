package crypto

import (
	"net"
	"errors"
	"fmt"
	"bufio"
	"strings"
	"io"

	"crypto/rand"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/sha512"
	"golang.org/x/crypto/hkdf"

	"encoding/base64"
	"encoding/hex"

	"privy/types"
)

func InitRatchet(sharedKey []byte, isHost bool) (*types.SendingRatchet, *types.ReceivingRatchet, error){
	h2cReader := hkdf.New(sha512.New, sharedKey, nil, []byte("privy-host-to-client"))
	hostToClientStart := make([]byte, 32)

	_, err := io.ReadFull(h2cReader, hostToClientStart)
	if err != nil {
		return nil, nil, err
	}

	c2hReader := hkdf.New(sha512.New, sharedKey, nil, []byte("privy-client-to-host"))
	clientToHostStart := make([]byte, 32)

	_, err = io.ReadFull(c2hReader, clientToHostStart)
	if err != nil {
		return nil, nil, err
	}

	var mySendKey, myRecvdKey []byte

	if isHost {
		mySendKey = hostToClientStart
		myRecvdKey = clientToHostStart
	} else {
		mySendKey = clientToHostStart
		myRecvdKey = hostToClientStart
	}

	sendRatchet := &types.SendingRatchet{
		ChainKey: mySendKey,
		SequenceNum: 0,
	}

	recvRatchet := &types.ReceivingRatchet{
		ChainKey: myRecvdKey,
		ExpectedSeqNum: 0,
		SkippedKeys: make(map[int][]byte),
	}

	return sendRatchet, recvRatchet, nil
}

func StepRatchet(ChainKey []byte) ([]byte, []byte, error){
	kdfReader := hkdf.New(sha512.New, ChainKey, nil, []byte("privy-msg-step"))

	nextChainKey := make([]byte, 32)
	msgKey := make([]byte, 32)

	_, err := io.ReadFull(kdfReader, nextChainKey)
	if err != nil {
		return nil, nil, err
	}

	_, err = io.ReadFull(kdfReader, msgKey)
	if err != nil {
		return nil, nil, err
	}

	return nextChainKey, msgKey, nil
}

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
