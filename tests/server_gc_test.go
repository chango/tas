
package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq2"
	"time"
	"strconv"
	"testing"
	"net/http"
	"encoding/json"
	"io/ioutil"

)

const zmqPort = "7450"
const httpPort ="7451"

func errorCheck(err error, message string, t *testing.T) {
	// Checks if there is an error

	if err != nil {
		fmt.Println("FAIL")
		t.Error(message)
	}else {
		fmt.Println("PASS")
	}
}

func setUpSocket() (zmq.Socket, error) {
	// Sets up the socket for sending data

	// Open the socket
	socket, err := zmq.NewSocket(zmq.PUSH)

	// Check for socket error
	zmqServer := fmt.Sprintf("tcp://localhost:%s", zmqPort)
	err = socket.Connect(zmqServer)

	
	return *socket, err
}

func getJson() (map[string]interface{}, error) {
	// Return the json object from the DIAG endpoint

	//Get the response from the server
	resp, err := http.Get("http://localhost:"+httpPort+"/DIAG")
	js_resp, err := ioutil.ReadAll(resp.Body)
	
	var js map[string]interface{}
	
	//Get the json
	err = json.Unmarshal(js_resp, &js)
    return js, err
}

func tearDown() {
	//Waits till the tree is empty

	js, _ := getJson()
	if int(js["num_leafs"].(float64))>0 { 

		time.Sleep(25*time.Second)
		fmt.Println("Waiting for GC to run")
		tearDown()

	}
}

func TestServerRunning(t *testing.T) {
	// Test to see if the server is running

	_, err := http.Get("http://localhost:"+httpPort+"/DIAG")
	errorCheck(err, "Server not running. Run <go run tas.go> to start the server", t)
}

func TestServerRecievingData(t *testing.T) {
	//Test to see if the server is recieving data

    socket, err := setUpSocket()
    errorCheck(err, "Error setting up socket", t)
    defer socket.Close()

	now := time.Now().Unix()
	// Incase there is input in the tree
	tearDown()
	msg := fmt.Sprintf("INCR %s %s %d", strconv.FormatInt(now, 10), "test.tes.te.t", 5)
	_, err_send := socket.SendBytes([]byte(msg), 0)
    
    errorCheck(err_send, "Could not send input to server", t)
  
    time.Sleep(3*time.Second)

    
	js, err := getJson()
	//Check if the server is actually running
	errorCheck(err, "Server not running. Run <go run tas.go> to start the server", t)
        
   	if js["num_leafs"].(float64) != 1 {
		t.Error("Did not enter any data")
	}
	tearDown()
}

func TestGcRunning(t *testing.T) {
	//Test the end point to see if gc is running
	
	js, err := getJson()
	//Get the json

	errorCheck(err, "Could not parse json", t)

     //Check if GC is running
   	if !js["gc_running"].(bool) {
		t.Error("Garbage Collector not running")
	}

}

func TestGcWorking(t *testing.T) {
	// Test if the GC is actually deleting data

	//Open the socket
    socket, err := setUpSocket()
    defer socket.Close()
    errorCheck(err, "Could not make connection to server to send input", t)

	//Add data
	for count:=0; count < 10; count++ {
		now := time.Now().Unix() - 60 
		msg := fmt.Sprintf("INCR %s %s %d", strconv.FormatInt(now, 10), "test.tes.te.t"+string(count), 5)
		_, err_send := socket.SendBytes([]byte(msg), 0)
		errorCheck(err_send, "Could not send input to server", t)

    }

    //Wait a minute to make sure GC runs
    // If GC doesn't run in a minute something is wrong
    time.Sleep(60*time.Second)

    //Get the json
    js, err := getJson()
	errorCheck(err, "Problem getting json, server may not be running", t)

   	if js["num_leafs"].(float64) != 0 {
		t.Error("GC did not clean data")
	}
}