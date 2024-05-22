package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func TestP2PChat(t *testing.T) {
	const protocolID = "/p2p-chat/1.0.0"

	encrypt := func(plaintext string, publicKey *rsa.PublicKey) (string, error) {
		hash := sha256.New()
		ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, []byte(plaintext), nil)
		if err != nil {
			return "", err
		}
		return string(ciphertext), nil
	}

	decrypt := func(ciphertext string, privateKey *rsa.PrivateKey) (string, error) {
		hash := sha256.New()
		decryptedText, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, []byte(ciphertext), nil)
		if err != nil {
			return "", err
		}
		return string(decryptedText), nil
	}

	createNode := func(t *testing.T, listenAddr string) host.Host {
		var opts []libp2p.Option
		if listenAddr != "" {
			opts = append(opts, libp2p.ListenAddrStrings(listenAddr))
		}

		node, err := libp2p.New(opts...)
		if err != nil {
			t.Fatal(err)
		}
		return node
	}

	connectNodes := func(t *testing.T, sourceNode host.Host, targetNode host.Host) {
		targetAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/8007/p2p/%s", targetNode.ID()))
		targetAddrInfo, _ := peer.AddrInfoFromP2pAddr(targetAddr)

		err := sourceNode.Connect(context.Background(), *targetAddrInfo)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Connected Nodes: \n source node %s \n target node %s", sourceNode.ID(), targetNode.ID())
	}

	countpeers := func(node *host.Host) int {
		return len((*node).Network().Peers())
	}

	handleStream := func(stream network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go func() {
			for {
				str, err := rw.ReadString('\n')
				if err != nil {
					return
				}
				if str != "" && str != "\n" {
					fmt.Printf("Received: %s", str)
				}
			}
		}()

		go func() {
			for {
				rw.WriteString("Hello from test\n")
				rw.Flush()
				time.Sleep(1 * time.Second)
			}
		}()
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal("Error generating RSA key pair:", err)
	}
	publicKey := &privateKey.PublicKey

	sourceNode := createNode(t, "")
	targetNode := createNode(t, "/ip4/0.0.0.0/tcp/8007")

	assert.NotNil(t, sourceNode)
	assert.NotNil(t, targetNode)

	sourceNode.SetStreamHandler(protocolID, handleStream)
	targetNode.SetStreamHandler(protocolID, handleStream)

	connectNodes(t, sourceNode, targetNode)

	time.Sleep(time.Second) 
	assert.Equal(t, 1, len(sourceNode.Network().Peers()))
	assert.Equal(t, 1, len(targetNode.Network().Peers()))

	plaintext := "Hello from test\n"
	ciphertext, err := encrypt(plaintext, publicKey)
	assert.NoError(t, err)
	decryptedText, err := decrypt(ciphertext, privateKey)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decryptedText)
	t.Log("Encrypt Decrypt Succeed")

	stream, err := sourceNode.NewStream(context.Background(), targetNode.ID(), protocolID)
	assert.NoError(t, err)
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	writeData := func(rw *bufio.ReadWriter, input string) {
		trimmedData := strings.TrimSpace(input)
		if trimmedData == "" {
			t.Log("Empty message skipped")
			return
		}

		if trimmedData == "/check" {
			t.Logf("Source node peers: %d", countpeers(&sourceNode))
			return
		}

		t.Log("Sent message:", input)
	}

	readData := func(rw *bufio.ReadWriter) {
		message, err := rw.ReadString('\n')
		assert.NoError(t, err)
		assert.Equal(t, "Hello from test\n", message)
		t.Log("Received message:", message)
	}

	// Test /check command
	writeData(rw, "/check")
	time.Sleep(time.Second)

	// Test empty message
	writeData(rw, "")
	time.Sleep(time.Second)

	// Test message
	writeData(rw, "Hello from test\n")
	readData(rw)

	// Test interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		time.Sleep(time.Second)
		sigCh <- syscall.SIGINT
	}()
	<-sigCh
	fmt.Println("Exiting the chat...")
}