
#Abstract

True Air Speed aka Tree Avoidance System, a lightweight system to track and analyze real-time data.


#Prerequisites

1. Go 1.2.2 package must be installed on your computer.

2. zmq2 package must also be installed.

#Usage
TAS is useful if you are looking for code that can store information in a tree and automatically deletes the expired information.

Some applications of TAS include

- storing the users’ shopping history
- real-time display of stock market trends
- displaying live tweets

#Getting Started


###Installation

1. Download the Go 1.2.2 package:http://golang.org/dl/ 
  
  *Note: TAS hasn’t been tested on Go 1.3.*  

2. 
Install the zmq2 library from github:
```
go get github/com/pebbe/zmq2
```

3. Download the TAS package, run:
```
hg clone ssh://hg@bitbucket.org/chango/atc
cd atc/tas
```
----
###Example


The server has to be up and running first. For UNIX-like systems, type the following commands in terminal:
```
cd tas
go run tas.go
```
Create a file called hellotas.go and paste the following code:
```go
package main

import (
	
	"fmt"
	zmq "github.com/pebbe/zmq2"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

const http_port = 7451
const tcp_port = 7450

func main() {
	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	socket.Connect("tcp://localhost:"+strconv.Itoa(tcp_port))
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
	time.Sleep(2*time.Millisecond)

	// Check the result from the http://localhost:7451/GET page
	link := "http://localhost:"+strconv.Itoa(http_port)+"/GET?key=*.*.*"
	resp, _ := http.Get(link)
	defer resp.Body.Close()

	// Read and return the output of HTTP GET
	contents, _ := ioutil.ReadAll(resp.Body)
	data := make(map[string]interface{})
	json.Unmarshal(contents, &data)

	// Print the output to terminal
	fmt.Println(data)
}
```
On a separate terminal, run **hellotas.go** by changing directory to where you stored hellotas.go. Then run the code by typing the command:
```
go run hellotas.go
```
----
#Documentation
The documentation can be found [here](./documentation.md).


