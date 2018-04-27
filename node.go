package main

// DOING: Node struct

//// P2P Network feature
// DOING: Add P2P Logic (separate module?)
// DONE: Add address
// DOING: Add Node sync
// TODO: Docker network &
//       Quicker discovering (or another approach for adding nodes to P2P network)
// TODO: Add disconnecting/removing node from P2P network (knownNodes, discoverPeers)
// TODO: P2P commands/messages, protocol, message queue?
// TODO: PingLoop

//// REST API feature
// DONE: Add REST API Logic (separate module?)
// TODO-1: Add REST API endpoint POST File
// TODO-2: & Store File to folder with name {Node.Address}_{Node.Port}
// TODO-3: & Rule for POST File: Files are never stored in the store of the peer who uploaded the file
// TODO: Add REST API endpoint GET File
// TODO: & Rule for GET File: Files are available only from the peer from which it was downloaded

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"
)

const PROTOCOL = "tcp"
const COMMAND_LENGTH = 12
const DISCOVER_INTERVAL = 30 * time.Second
const TIME_FORMAT = "15:04:05.000"

var timeStamp = func() string { return time.Now().Format(TIME_FORMAT) }

type ping struct {
	From string
}

type pong struct {
	From string
}

type Node struct {
	Id         string
	Address    string
	Port       int
	knownNodes []string
	conn       net.Conn
	closed     chan struct{}
	disconnect chan error

	// REST service
	restService *RestService
}

func NewNode(address string, port int) *Node {
	fullAddress := address + ":" + strconv.Itoa(port)
	// HACK: for terminal testing addresses are equal, we can use different ports
	var restAddress string
	if address == PEER_IP_DEFAULT {
		// REST API port = port + (5100-5000)
		restAddress = address + ":" + strconv.Itoa(port+REST_PORT-PEER_PORT_DEFAULT)
	} else {
		restAddress = address + ":" + strconv.Itoa(REST_PORT)
	}

	return &Node{
		Id:          fullAddress,
		Address:     address,
		Port:        port,
		knownNodes:  []string{fullAddress},
		closed:      make(chan struct{}),
		disconnect:  make(chan error),
		restService: NewRestService(restAddress),
	}
}

func (node *Node) Run() {
	// REST API Service
	if err := node.restService.Start(); err != nil {
		Warn.Printf("Failed to start REST API service: %s", err.Error())
	} else {
		Info.Printf("Start REST API Service %s", node.restService.address)
	}

	// P2P Network
	ln, err := net.Listen(PROTOCOL, node.Id)
	if err != nil {
		Warn.Println(err)
	}
	go node.discoverPeers()

	// FIXME: can we create permanent connection?
	// simple loop
loop:
	for {
		Debug.Println("Simple Loop")
		node.conn, err = ln.Accept()
		if err != nil {
			Debug.Println(err)
			break loop
		}
		Debug.Println("Handle Connection")
		go node.handleConnection(node.conn)

		// disconnect
		select {
		case err = <-node.disconnect:
			break loop
		}
	}

	close(node.closed)
}

func (node *Node) Disconnect(reason string) {
	select {
	case node.disconnect <- fmt.Errorf("Disconnect reason: %s", reason):
	case <-node.closed:
	}
}

func (node *Node) isKnown(address string) bool {
	for _, node := range node.knownNodes {
		if node == address {
			return true
		}
	}
	return false
}

func (node *Node) discoverPeers() {
	// HACK: wait 1 second before start discovering
	time.Sleep(time.Second)
	// discover nearest peers
	node.discoverPeersTick()

	discover := time.NewTimer(DISCOVER_INTERVAL)
	for {
		select {
		case <-discover.C:
			node.discoverPeersTick()
			discover.Reset(DISCOVER_INTERVAL)
		case <-node.closed:
			return
		}
	}

}

func (node *Node) discoverPeersTick() {
	Info.Printf("%s: Discovering start...", timeStamp())
	// prepare request
	payload := gobEncode(ping{node.Id})
	pingRequest := append(commandToBytes("ping"), payload...)

	// FIXME: Docker networks workaround
	// from x.x.x.2:5000 to x.x.x.6:5005
	lastIPFrom := 2
	lastIPTo := 6
	portFrom := 5000
	portTo := 5005
	for lastIP := lastIPFrom; lastIP <= lastIPTo; lastIP++ {
		addressBaseIdx := strings.LastIndex(node.Address, ".")
		for port := portFrom; port <= portTo; port++ {
			address := node.Address[0:addressBaseIdx+1] +
				strconv.Itoa(lastIP) + ":" + strconv.Itoa(port)
			if !node.isKnown(address) {
				// ping to this address
				node.sendPing(address, pingRequest)
			}
		}
	}
	Info.Printf("%s: Discovering finish...", timeStamp())
}

func (node *Node) handleConnection(conn net.Conn) error {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		Warn.Println(err)
		return err
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
	return nil
}

func (node *Node) handlePing(request []byte) {
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

func (node *Node) handlePong(request []byte) {
	var buffer bytes.Buffer
	var pongPayload pong

	buffer.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buffer)
	err := dec.Decode(&pongPayload)
	if err != nil {
		Warn.Println(err)
	}

	if !node.isKnown(pongPayload.From) {
		Info.Printf("%s: New Peer Node is connected: %s\n", timeStamp(), pongPayload.From)
		node.knownNodes = append(node.knownNodes, pongPayload.From)
		Info.Printf("All Known Nodes: %v\n", node.knownNodes)
	}
}

func (node *Node) sendData(address string, data []byte) bool {
	//Debug.Printf("Dial to: %s\n", address)
	conn, err := net.Dial(PROTOCOL, address)
	if err != nil {
		Debug.Printf("%s: %s is not available\n", timeStamp(), address)
		return false
	}
	defer conn.Close()

	Debug.Printf("Send data to address [%s]: %x\n", address, data)
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		Warn.Println(err)
		return false
	}
	return true
}

func (node *Node) sendPing(address string, request []byte) bool {
	//Debug.Printf("PING: %s... ", address)
	return node.sendData(address, request)
}

func (node *Node) sendPong(address string) bool {
	payload := gobEncode(pong{node.Id})
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
