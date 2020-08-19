package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/etng/go-tcpee/version"
)

var addr = flag.String("addr", "", "The address to listen to; default is \"\" (all interfaces).")
var port = flag.Int("port", 8000, "The port to listen on; default is 8000.")

func main() {
	fmt.Printf("Go TCPee Server %s-%s %s(%s)\n", version.Version, version.ReleaseTag, version.CommitID, version.ShortCommitID)

	flag.Parse()

	fmt.Println("Starting...")

	src := *addr + ":" + strconv.Itoa(*port)
	listener, _ := net.Listen("tcp", src)
	fmt.Printf("Listening on %s.\n", src)

	defer listener.Close()
	clients := NewClientManager()
	go func() {
		t := time.NewTicker(time.Second * 5)
		defer t.Stop()
		for _ = range t.C {
			clients.Report()
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Some connection error: %s\n", err)
		}
		client := clients.NewClient(conn)
		client.Run()
	}
}

// NewClientManager Get a new client manager
func NewClientManager() *Clients {
	return &Clients{
		clients: make(map[string]*Client),
	}
}

// Clients is the list of clients, client can sign in/off
type Clients struct {
	sync.Mutex
	clients    map[string]*Client
	LastActive time.Time
}

// Client is an client
type Client struct {
	id      string
	conn    net.Conn
	manager *Clients
}

// Report just shows the clients statistics
func (cm *Clients) Report() {
	log.Printf("current clients count: %d\n", len(cm.clients))
	if len(cm.clients) == 0 {
		if cm.LastActive.IsZero() {
			cm.LastActive = time.Now()
		} else {
			if time.Now().Sub(cm.LastActive) > time.Second*30 {
				log.Printf("last active is %s", cm.LastActive.Format(time.RFC3339))
				log.Printf("idle too long, quit now")
				os.Exit(0)
			}
		}

	} else if !cm.LastActive.IsZero() {
		cm.LastActive = time.Time{}
	}
}

// NewClient create a new client based on the tcp connection and register it
func (cm *Clients) NewClient(conn net.Conn) *Client {
	client := &Client{
		id:      conn.RemoteAddr().String(),
		conn:    conn,
		manager: cm,
	}
	cm.Lock()
	defer cm.Unlock()
	cm.clients[client.id] = client
	return client
}

// Notify send message to everybody except the client himself
func (c *Client) Notify(body []byte) {
	for _, oc := range c.manager.clients {
		if c.id != oc.id {
			oc.Send(body)
		}
	}
}

// Send just send message to the client himself
func (c *Client) Send(body []byte) {
	c.conn.Write(body)
}

// Quit just close the client connection and sign off
func (c *Client) Quit() {
	c.conn.Close()
	delete(c.manager.clients, c.id)
}

// Run start the client event loop and do the io things(reading/writing)
func (c *Client) Run() {
	go func() {
		remoteAddr := c.conn.RemoteAddr().String()
		fmt.Println("Client connected from " + remoteAddr)
		defer c.Quit()
		scanner := bufio.NewScanner(c.conn)

		for {
			ok := scanner.Scan()

			if !ok {
				break
			}

			if !handleMessage(scanner.Text(), c) {
				break
			}
		}

		fmt.Println("Client at " + remoteAddr + " disconnected.")
	}()
}

// handleMessage handle message with client
func handleMessage(message string, c *Client) bool {
	fmt.Println("> " + message)

	if len(message) > 0 && message[0] == '/' {
		switch {
		case message == "/time":
			resp := "It is " + time.Now().String() + "\n"
			fmt.Print("< " + resp)
			c.Send([]byte(resp))

		case message == "/quit":
			fmt.Println("Quitting.")
			c.Send([]byte("I'm shutting down connection with you now.\n"))
			c.Send([]byte("Welcome back anytime!\n"))
			fmt.Println("< " + "%quit%")
			c.Send([]byte("%quit%\n"))
			//os.Exit(0)
			return false
		default:
			c.Send([]byte("Unrecognized command.\n"))
		}
	} else {
		if b, e := json.Marshal(map[string]interface{}{
			"status":     "ok",
			"sender":     c.id,
			"message":    message,
			"created_at": time.Now().Format(time.RFC3339),
		}); e == nil {
			c.Notify(b)
			c.Notify([]byte("\n"))
		} else {
			c.Notify([]byte(message + "\n"))
		}

		if b, e := json.Marshal(map[string]interface{}{
			"status":     "ok",
			"message":    message,
			"created_at": time.Now().Format(time.RFC3339),
		}); e == nil {
			c.Send(b)
			c.Send([]byte("\n"))
		} else {
			c.Send([]byte(message + "\n"))
		}

	}
	return true
}
