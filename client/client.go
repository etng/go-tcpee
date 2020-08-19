package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/etng/go-tcpee/version"
)

var host = flag.String("host", "localhost", "The hostname or IP to connect to; defaults to \"localhost\".")
var port = flag.Int("port", 8000, "The port to connect to; defaults to 8000.")
var closed bool = false
var banner string = `
Welcome to TCPee Client
You can send message to your server after the prompt '> '
The console will print the response of the server, if the result is json, it will pretty print it for you.
You can use '/exit' or 'CTRL+C' to quit.
`

func main() {
	fmt.Printf("Go TCPee Client %s-%s %s(%s)\n", version.Version, version.ReleaseTag, version.CommitID, version.ShortCommitID)

	flag.Parse()

	dest := *host + ":" + strconv.Itoa(*port)
	fmt.Printf("Connecting to %s...\n", dest)
	conn, err := net.Dial("tcp", dest)

	if err != nil {
		if _, t := err.(*net.OpError); t {
			fmt.Println("Some problem connecting.")
		} else {
			fmt.Println("Unknown error: " + err.Error())
		}
		os.Exit(1)
	}
	fmt.Println(strings.TrimSpace(banner))
	go readConnection(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) == "/exit" {
			closed = true
			break
		}
		if closed {
			fmt.Println("connection already closed, quit")
			break
		}

		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error writing to stream.")
			break
		}
	}
}

// readConnection read from tcp connection and pretty print things it received
func readConnection(conn net.Conn) {
out:
	for !closed {
		scanner := bufio.NewScanner(conn)

		for {
			ok := scanner.Scan()
			text := scanner.Text()

			command := handleCommands(text)
			if !command {
				res := map[string]interface{}{}
				if e := json.Unmarshal([]byte(text), &res); e == nil {
					b, _ := json.MarshalIndent(res, "", " ")
					fmt.Printf("\n\n** %s\n> ", string(b))
				} else {
					fmt.Printf("\b\b** %s\n> ", text)
				}
			}

			if !ok {
				fmt.Println("Reached EOF on server connection.")
				break out
			}
		}
	}
	closed = true
}

// handleCommands parse command response the respond to command such as quit
func handleCommands(text string) bool {
	r, err := regexp.Compile("^%.*%$")
	if err == nil {
		if r.MatchString(text) {

			switch {
			case text == "%quit%":
				fmt.Println("\b\bServer is leaving. Hanging up.")
				os.Exit(0)
			}

			return true
		}
	}

	return false
}
