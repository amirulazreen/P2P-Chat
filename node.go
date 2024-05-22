package main

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type P2PNode interface {
	CreateNode() host.Host
	ConnectToTargetNode(targetAddr string) error
	CountPeers() int
	PrintInfo()
	SetStreamHandler(handler func(network.Stream))
	Close()
}

type Libp2pNode struct {
	Host host.Host
}

func (node *Libp2pNode) CreateNode(listenAddr string) host.Host {
	var err error
	if listenAddr == "" {
		node.Host, err = libp2p.New()
	} else {
		node.Host, err = libp2p.New(libp2p.ListenAddrStrings(listenAddr))
	}
	if err != nil {
		panic(err)
	}
	return node.Host
}

func (node *Libp2pNode) ConnectToTargetNode(targetAddr string) error {
	targetMultiAddr, err := multiaddr.NewMultiaddr(targetAddr)
	if err != nil {
		return err
	}
	targetAddrInfo, err := peer.AddrInfoFromP2pAddr(targetMultiAddr)
	if err != nil {
		return err
	}
	return node.Host.Connect(context.Background(), *targetAddrInfo)
}

func (node *Libp2pNode) CountPeers() int {
	return len(node.Host.Network().Peers())
}

func (node *Libp2pNode) PrintInfo() {
	fmt.Println("ID:", node.Host.ID())
}

func (node *Libp2pNode) SetStreamHandler(handler func(network.Stream)) {
	node.Host.SetStreamHandler(protocolID, handler)
}

func (node *Libp2pNode) Close() {
	node.Host.Close()
}