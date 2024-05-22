package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

type NodeManager struct {
	SourceNode *Libp2pNode
	TargetNode *Libp2pNode
	Encryptor  Encryptor
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func NewNodeManager() (*NodeManager, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	publicKey := &privateKey.PublicKey

	manager := &NodeManager{
		SourceNode: &Libp2pNode{},
		TargetNode: &Libp2pNode{},
		Encryptor:  &RSAEncryptor{},
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

	manager.SourceNode.CreateNode("")
	manager.TargetNode.CreateNode("/ip4/0.0.0.0/tcp/8007")

	return manager, nil
}

func (manager *NodeManager) HandleStream(stream network.Stream) {
	fmt.Println("Open chat between source node and target node")
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go manager.readData(rw)
	go manager.writeData(rw)
}

func (manager *NodeManager) readData(rw *bufio.ReadWriter) {
	for {
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(rw, lengthBytes)
		if err != nil {
			fmt.Println("Error reading length from buffer:", err)
			return
		}

		length := int(lengthBytes[0])<<24 | int(lengthBytes[1])<<16 | int(lengthBytes[2])<<8 | int(lengthBytes[3])
		encryptedData := make([]byte, length)
		_, err = io.ReadFull(rw, encryptedData)
		if err != nil {
			fmt.Println("Error reading data from buffer:", err)
			return
		}

		plaintext, err := manager.Encryptor.Decrypt(encryptedData, manager.PrivateKey)
		if err != nil {
			fmt.Println("Error decrypting data:", err)
			return
		}

		fmt.Printf("Message: %s\n", string(plaintext))
	}
}

func (manager *NodeManager) writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadBytes('\n')
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			continue
		}

		trimmedData := strings.TrimSpace(string(sendData))
		if trimmedData == "" {
			continue
		}

		if trimmedData == "/check" {
			fmt.Printf("Source node peers: %d\n", manager.SourceNode.CountPeers())
			continue
		}

		encryptedData, err := manager.Encryptor.Encrypt(sendData, manager.PublicKey)
		if err != nil {
			fmt.Println("Error encrypting data:", err)
			continue
		}

		length := len(encryptedData)
		lengthBytes := []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
		rw.Write(lengthBytes)
		rw.Write(encryptedData)
		rw.Flush()
	}
}