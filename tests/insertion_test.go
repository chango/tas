package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq3"
	"log"
	"time"
	//"strings"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"testing"
)

import (
	"github.com/chango/tas/tas"
)

const http_port = 7451
const tcp_port = 7450

func find_link(key string) string {
	return "http://localhost:" + strconv.Itoa(http_port) + "/" + key
}

func NewTestingServer() *tas.TASServer {
	c := tas.NewDefaultTASConfig()
	svr, err := tas.NewTASServer(c)
	if err != nil {
		log.Fatal(err)
	}
	return svr
}

func TestMain(m *testing.M) {
	svr := NewTestingServer()
	go svr.Run()
}

func ReadDiagServer(t *testing.T) (map[string]interface{}, interface{}) {
	// Get the output from http://localhost:{tcp_port}/DIAG

	// HTTP GET the /DIAG page
	link := find_link("DIAG")
	resp, err := http.Get(link)

	if err != nil {
		t.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read and return the output of HTTP GET
	contents, _ := ioutil.ReadAll(resp.Body)
	data := make(map[string]interface{})
	json.Unmarshal(contents, &data)

	return data, nil
}

func ReadGetServer(key string, t *testing.T) (map[string]interface{}, interface{}) {
	// Get the output from http://localhost:{tcp_port}/GET?key={key}

	// HTTP GET the /GET page
	link := find_link("GET?key=" + key)
	resp, err := http.Get(link)
	defer resp.Body.Close()

	if err != nil {
		t.Error(err)
		return nil, err
	}

	// Read and return the output of HTTP GET
	contents, _ := ioutil.ReadAll(resp.Body)
	data := make(map[string]interface{})

	// When there is NO wild card in the key string, data[key] = value
	// ie/ data["cart.grocery.vegetables.basket"] = 14
	// When there is wild card in the key string, data is the hash map
	// ie/ data = map[cart:map[grocery:map[vegetables:map[basket:14]]]]
	var wildcard = regexp.MustCompile(`[\*.]*\*`)
	if !wildcard.MatchString(key) {
		var value interface{}
		json.Unmarshal(contents, &value)
		data[key] = value

	} else {
		json.Unmarshal(contents, &data)
	}

	return data, nil
}

func WaitForEmptyTree(t *testing.T) {
	// Wait until the tree is empty using the garbage collector

	output_diag, err_diag := ReadDiagServer(t)

	error := false

	// Shows error if cannot read from /GET page or the garbage collector isn't running
	if err_diag != nil || output_diag["gc_running"] != true {
		error = true
	} else {

		// Waiting till there is no leaf in the tree
		if int(output_diag["num_leafs"].(float64)) > 0 {
			fmt.Println("Waiting for tree to be empty using garbage collector...")
		}

		for int(output_diag["num_leafs"].(float64)) > 0 {
			time.Sleep(100 * time.Millisecond)
			output_diag, err_diag = ReadDiagServer(t)

			if err_diag != nil {
				error = true
			}
		}
	}

	if error {
		t.Error("Garbage collector isn't running")
	}
}

func WaitIfExists(key string, t *testing.T) {
	// Call WaitforEmptyTree() only if a leaf under key exists

	output, _ := ReadGetServer(key, t)
	value, _ := output[key]
	if value != nil {
		WaitForEmptyTree(t)
	}
}

func FailHandler(fail bool, t *testing.T) {
	// Printing "CASE FAIL" and "CASE PASS" at the end of each test case

	if fail {
		fmt.Printf("FAIL\n")
		t.FailNow()
	} else {
		fmt.Printf("PASS\n")
	}
}

func SocketFailHandler(err interface{}, t *testing.T) {
	// Fail handler for connecting socket

	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}

func SocketSendFailHandler(err interface{}, t *testing.T) {
	// Fail handler for sending message to the server

	if err != nil {
		t.Error("Could not send input to server")
	}

	// Wait time for the server to update the tree
	time.Sleep(5 * time.Millisecond)
}

func TestIncr(t *testing.T) {
	// Generates unique keys and checks if the leafs are added correctly

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	WaitForEmptyTree(t)

	count := 1
	level1 := []string{"veg_section", "meat", "ready_food"}
	level2 := []string{"basket1", "basket2", "basket3"}
	level3 := []string{"item1", "item2"}
	for _, key1 := range level1 {
		for _, key2 := range level2 {
			for _, key3 := range level3 {

				// key insertion
				now := time.Now().Unix()
				key := fmt.Sprintf("cart.%s.%s.%s", key1, key2, key3)
				msg := fmt.Sprintf("INCR %s %s %d", strconv.FormatInt(now, 10), key, count)
				_, err := socket.SendBytes([]byte(msg), 0)
				fmt.Printf("Sending %s...", msg)
				SocketSendFailHandler(err, t)

				// Verifying output from /DIAG and /GET
				output_diag, err_diag := ReadDiagServer(t)
				num_leafs := int(output_diag["num_leafs"].(float64))
				if err_diag != nil || num_leafs != count {
					fmt.Printf("FAIL\n")
					t.Error("Total number of leafs is", strconv.Itoa(num_leafs),
						"instead of ", count)
				} else {
					output_get, err_get := ReadGetServer(key, t)
					FailHandler((err_get != nil || int(output_get[key].(float64)) != count), t)
				}

				count++
			}
		}
	}
}

func TestIncre2(t *testing.T) {
	// INCR using the same key and checks if the value increments properly
	// value is an int array

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item4"
	WaitIfExists(key, t)

	now := time.Now().Unix()
	total := 0
	for i := 0; i < 3; i++ {

		// inserting leaf
		msg := fmt.Sprintf("INCR %s %s %d", strconv.FormatInt(now, 10), key, i)
		_, err = socket.SendBytes([]byte(msg), 0)
		SocketSendFailHandler(err, t)
		fmt.Printf("Sending %s...", msg)

		total += i

		// checking output
		output_get, err_get := ReadGetServer(key, t)
		FailHandler((err_get != nil || int(output_get[key].(float64)) != total), t)
	}
}

func TestAppend(t *testing.T) {
	// Testing the APPEND function by inserting slices of int arrays

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item5"
	WaitIfExists(key, t)

	// Inserting first slice
	now := time.Now().Unix()
	value := []int{1}
	returnVal, _ := json.Marshal(value)
	msg := fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(now, 10), key, returnVal)
	_, err = socket.SendBytes([]byte(msg), 0)
	SocketSendFailHandler(err, t)
	fmt.Printf("Sending %s...", msg)

	// Checking output
	output_get, _ := ReadGetServer(key, t)
	output_value := fmt.Sprintf("%v", output_get[key])
	fail_cond := (output_value != "[1]")
	FailHandler(fail_cond, t)

	// Inserting second slice
	value = []int{2, 3, 4}
	returnVal, _ = json.Marshal(value)
	msg = fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(now, 10), key, returnVal)
	_, err = socket.SendBytes([]byte(msg), 0)
	SocketSendFailHandler(err, t)
	fmt.Printf("Sending %s...", msg)

	// Checking output
	output_get, _ = ReadGetServer(key, t)
	output_value = fmt.Sprintf("%v", output_get[key])
	fail_cond = (output_value != "[1 2 3 4]")
	FailHandler(fail_cond, t)
}

func TestAppend2(t *testing.T) {
	// Testing the APPEND function by inserting slices of string arrays

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item6"
	WaitIfExists(key, t)

	// Inserting first slice
	now := time.Now().Unix()
	value := []string{"abc"}
	returnVal, _ := json.Marshal(value)
	msg := fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(now, 10), key, returnVal)
	_, err = socket.SendBytes([]byte(msg), 0)
	SocketSendFailHandler(err, t)
	fmt.Printf("Sending %s...", msg)

	// Checking output
	output_get, _ := ReadGetServer(key, t)
	output_value := fmt.Sprintf("%v", output_get[key])
	fail_cond := (output_value != "[abc]")
	FailHandler(fail_cond, t)

	// Inserting second slice
	value = []string{"d", "ef"}
	returnVal, _ = json.Marshal(value)
	msg = fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(now, 10), key, returnVal)
	_, err = socket.SendBytes([]byte(msg), 0)
	SocketSendFailHandler(err, t)
	fmt.Printf("Sending %s...", msg)

	// Checking output
	output_get, _ = ReadGetServer(key, t)
	output_value = fmt.Sprintf("%v", output_get[key])
	fail_cond = (output_value != "[abc d ef]")
	FailHandler(fail_cond, t)
}

func TestAppend3(t *testing.T) {
	// Testing the APPEND function by inserting string arrays and ints

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item7"
	WaitIfExists(key, t)

	// Inserting first slice
	now := time.Now().Unix()
	value := []int{1, 2}
	returnVal, _ := json.Marshal(value)
	msg := fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(now, 10), key, returnVal)
	_, err = socket.SendBytes([]byte(msg), 0)
	SocketSendFailHandler(err, t)
	fmt.Printf("Sending %s..(GOOD)\n", msg)

	// Inserting second slice
	strings := []string{"a", "bc"}
	returnVal, _ = json.Marshal(strings)
	msg = fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(now, 10), key, returnVal)
	_, err = socket.SendBytes([]byte(msg), 0)
	SocketSendFailHandler(err, t)
	fmt.Printf("Sending %s...", msg)

	// Checking output
	output_get, _ := ReadGetServer(key, t)
	output_value := fmt.Sprintf("%v", output_get[key])
	fail_cond := (output_value != "[1 2 a bc]")
	FailHandler(fail_cond, t)
}

func TestDuplicatesAppend(t *testing.T) {
	// APPEND the same value to the same key

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item8"
	WaitIfExists(key, t)

	now := time.Now().Unix()
	value := []int{5}
	msg := fmt.Sprintf("APPEND %s %s %d", strconv.FormatInt(now, 10), key, value)
	for i := 0; i < 2; i++ {

		// Inserting leaf
		_, err = socket.SendBytes([]byte(msg), 0)
		SocketSendFailHandler(err, t)
		fmt.Printf("Sending %s...", msg)

		// Checking output
		output_get, _ := ReadGetServer(key, t)
		output_value := fmt.Sprintf("%v", output_get[key])
		if i == 0 {
			FailHandler((output_value != "[5]"), t)
		}
		if i == 1 {
			FailHandler((output_value != "[5 5]"), t)
		}
	}
}

func TestParamT(t *testing.T) {
	// Testing the "t" parameter by querying t={timestamp1},{timestamp2}...

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item9"
	WaitIfExists(key, t)

	// Inserting slices
	now := time.Now().Unix()
	for i := 0; i < 3; i++ {
		ts := now + int64(i)
		value := []int{i}
		returnVal, _ := json.Marshal(value)
		msg := fmt.Sprintf("APPEND %s %s %s", strconv.FormatInt(ts, 10), key, returnVal)
		_, err = socket.SendBytes([]byte(msg), 0)
		SocketSendFailHandler(err, t)
		fmt.Printf("Sending %s..(GOOD)\n", msg)
	}

	// Checking output, t={now}
	query := fmt.Sprintf("cart.seafood.basket1.*&t=%d", now)
	output_get, _ := ReadGetServer(query, t)
	output_value := fmt.Sprintf("%v", output_get["item9"])
	fmt.Printf("Verifying output of t=%d...", now)
	fail_cond := (output_value != "[0]")
	FailHandler(fail_cond, t)

	// Checking output, t={now},{now+1}
	query = fmt.Sprintf("cart.seafood.basket1.*&t=%d,%d", now, now+int64(1))
	output_get, _ = ReadGetServer(query, t)
	output_value = fmt.Sprintf("%v", output_get["item9"])
	fmt.Printf("Verifying output of t=%d,%d...", now, now+int64(1))
	fail_cond = (output_value != "[0 1]")
	FailHandler(fail_cond, t)
}

func TestParamI(t *testing.T) {
	// Testing the "i" parameter by querying i=2...

	// Socket connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	err := socket.Connect("tcp://localhost:" + strconv.Itoa(tcp_port))
	defer socket.Close()
	SocketFailHandler(err, t)

	// Wait till the leaf under key is deleted by GC
	key := "cart.seafood.basket1.item10"
	WaitIfExists(key, t)

	now := time.Now().Unix()
	for i := 0; i < 2; i++ {

		// inserting leaf
		msg := fmt.Sprintf("INCR %s %s %d", strconv.FormatInt(now+int64(i), 10), key, i)
		_, err = socket.SendBytes([]byte(msg), 0)
		SocketSendFailHandler(err, t)
		fmt.Printf("Sending %s...(GOOD)\n", msg)
	}

	// checking output
	key = key + "&i=2"
	output_get, err_get := ReadGetServer(key, t)
	fmt.Printf("Verifying output of i=2...")
	FailHandler((err_get != nil || output_get[key].(float64) != 0.25), t)
}
