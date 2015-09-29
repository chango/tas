package main

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq3"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const http_port = 7451
const tcp_port = 7450

func main() {
	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()

	// Inserting "Hello World" to the tree with key root.level1.level2
	now := time.Now().Unix()
	timestamp := strconv.FormatInt(now, 10)

	key := "root.level1.level2"

	value := []string{"Hello World"}
	value_js, _ := json.Marshal(value)

	msg := fmt.Sprintf("APPEND %s %s %s", timestamp, key, value_js)
	socket.SendBytes([]byte(msg), 0)
	fmt.Printf("Sending %s...\n", msg)

	// Wait a few milliseconds for the server to update the tree
	time.Sleep(2 * time.Millisecond)

	// Check the result from the http://localhost:7451/GET page
	link := "http://localhost:" + strconv.Itoa(http_port) + "/GET?key=*.*.*"
	resp, _ := http.Get(link)
	defer resp.Body.Close()

	// Read and return the output of HTTP GET
	contents, _ := ioutil.ReadAll(resp.Body)
	data := make(map[string]interface{})
	json.Unmarshal(contents, &data)

	// Print the output to terminal
	fmt.Println(data)
}
