package main

// DOING: Node struct
// DOING: Add P2P Logic (separate module?)
// TODO: Add Node sync
// TODO: P2P commands, protocol?
// TODO: Add REST API Logic (separate module?)

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

const IPADDRESS = "localhost"
const PROTOCOL = "tcp"
const COMMAND_LENGTH = 12
const NEAREST = 10
const DISCOVER_INTERVAL = 15 * time.Second

type ping struct {
	From string
}

type pong struct {
	From string
}

type Node struct {
	Id         int
	Address    string
	Port       int
	knownNodes []string
	conn       net.Conn
}

func NewNode(port int) *Node {
	address := IPADDRESS + ":" + strconv.Itoa(port)
	return &Node{
		Id:         port,
		Address:    address,
		Port:       port,
		knownNodes: []string{address},
	}
}

func (node *Node) Run() {
	ln, err := net.Listen(PROTOCOL, node.Address)
	if err != nil {
		Warn.Println(err)
	}
	go node.discoverPeers()
	// simple loop
	for {
		Debug.Println("Simple Loop")
		node.conn, err = ln.Accept()
		if err != nil {
			Debug.Println(err)
			break
		}
		Debug.Println("Handle Connection")
		go node.handleConnection(node.conn)
	}
}

func (node Node) isKnown(address string) bool {
	for _, node := range node.knownNodes {
		if node == address {
			return true
		}
	}
	return false
}

func (node *Node) discoverPeers() {
	// discover nearest peers
	discover := time.NewTimer(DISCOVER_INTERVAL)

	for {
		select {
		case <-discover.C:
			for port := (node.Port - NEAREST); port <= (node.Port + NEAREST); port++ {
				address := IPADDRESS + ":" + strconv.Itoa(port)
				if !node.isKnown(address) {
					// ping to this address
					node.sendPing(address)
				}
			}
			//discover.Reset(DISCOVER_INTERVAL)

			// TODO: when Node will close
			//case <-node.closed:
			//	return
		}
	}

}

func (node Node) handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		Warn.Println(err)
	}
	Debug.Printf("Request: %x\n", request)
	command := bytesToCommand(request[:COMMAND_LENGTH])
	Info.Printf("COMMAND: %s\n", command)
	switch command {
	case "ping":
		node.handlePing(request)
	case "pong":
		node.handlePong(request)
	default:
		Info.Printf("Unknown command: %s!\n", command)
	}
	conn.Close()
}

func (node Node) handlePing(request []byte) {
	var buffer bytes.Buffer
	var pingPayload ping

	buffer.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buffer)
	err := dec.Decode(&pingPayload)
	if err != nil {
		Warn.Println(err)
	}

	node.sendPong(pingPayload.From)
}

func (node Node) handlePong(request []byte) {
	var buffer bytes.Buffer
	var pongPayload pong

	buffer.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buffer)
	err := dec.Decode(&pongPayload)
	if err != nil {
		Warn.Println(err)
	}

	if !node.isKnown(pongPayload.From) {
		Info.Printf("New Peer Node is connected: %s\n", pongPayload.From)
		node.knownNodes = append(node.knownNodes, pongPayload.From)
	}
}

func (node Node) sendData(address string, data []byte) bool {
	Debug.Printf("Dial to: %s\n", address)
	conn, err := net.Dial(PROTOCOL, address)
	if err != nil {
		Debug.Printf("%s is not available\n", address)
		return false
	}
	defer conn.Close()

	Debug.Printf("Send data: %x\n", data)
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		Warn.Println(err)
		return false
	}
	return true
}

func (node Node) sendPing(address string) bool {
	payload := gobEncode(ping{node.Address})
	request := append(commandToBytes("ping"), payload...)
	Debug.Printf("PING: %s\n", address)
	return node.sendData(address, request)
}

func (node Node) sendPong(address string) bool {
	payload := gobEncode(pong{node.Address})
	request := append(commandToBytes("pong"), payload...)
	Info.Printf("PONG: %s\n", address)
	return node.sendData(address, request)
}

////

func commandToBytes(command string) []byte {
	var bytes [COMMAND_LENGTH]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:COMMAND_LENGTH]
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		Warn.Println(err)
	}

	return buff.Bytes()
}
