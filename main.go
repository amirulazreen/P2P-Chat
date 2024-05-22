package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p/core/network"
)

const protocolID = "/p2p-chat/1.0.0"

func main() {
	manager, err := NewNodeManager()
	if err != nil {
		fmt.Println("Error initializing NodeManager:", err)
		return
	}

	fmt.Println("-- SOURCE NODE INFORMATION --")
	manager.SourceNode.PrintInfo()

	fmt.Println("-- TARGET NODE INFORMATION --")
	manager.TargetNode.PrintInfo()

	manager.SourceNode.SetStreamHandler(func(stream network.Stream) {
		manager.HandleStream(stream)
	})
	manager.TargetNode.SetStreamHandler(func(stream network.Stream) {
		manager.HandleStream(stream)
	})

	targetAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/8007/p2p/%s", manager.TargetNode.Host.ID())
	if err := manager.SourceNode.ConnectToTargetNode(targetAddr); err != nil {
		panic(err)
	}

	stream, err := manager.SourceNode.Host.NewStream(context.Background(), manager.TargetNode.Host.ID(), protocolID)
	if err != nil {
		panic(err)
	}
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go manager.readData(rw)
	go manager.writeData(rw)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("Exiting the chat...")
	manager.SourceNode.Close()
	manager.TargetNode.Close()
}