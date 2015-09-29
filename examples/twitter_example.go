// Simulating web traffic through live tweets
// The program keeps track of names of people who retweeted Chris Hadfield's tweets
// Cmdr_Hadfield.tweet0
// Cmdr_Hadfield.tweet1
// ...
// Cmdr_Hadfield.tweet99

package main

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq3"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func inputData(data string, socket *zmq.Socket) {
	// Send data through socket

	_, err := socket.SendBytes([]byte(data), 0)
	if err != nil {
		log.Println("Error sending data")
	}
}

func createRequest(urlRequest string, retweets []string) string {
	//Create a request given a url request

	//Lets get the timestamp so we can time when this request was made
	rand_time := time.Now().Unix() + int64(rand_gen(60))
	now := strconv.FormatInt(rand_time, 10)
	handles_js, _ := json.Marshal(retweets)
	msg := fmt.Sprintf("APPEND %s %s %s", now, urlRequest, handles_js)
	return msg

}

func rand_gen(n int) int {
	//Generate random non-negative integer in [0,n)

	return rand.Intn(n)
}

func main() {

	//Generate the "tweets"
	tweets := make([]string, 100)
	for i := 0; i < 100; i++ {
		tweets[i] = "Cmdr_Hadfield.tweet" + strconv.Itoa(i)
	}

	//People on twitter
	handles := []string{"katyperry", "justinbieber", "BarackObama", "YouTube", "taylorswift13",
		"ladygaga", "britneyspears", "rihanna", "instagram", "jtimberlake",
		"twitter", "TheEllenShow", "JLo", "Cristiano", "shakira", "Oprah",
		"Pink", "ddlovato", "Harry_Styles", "selenagomez", "OfficialAdele",
		"KAKA", "onedirection", "aliciakeys", "NiallOfficial", "BrunoMars",
		"Eminem", "MileyCyrus", "NICKIMINAJ", "cnnbrk", "Real_Liam_Payne",
		"LilTunechi", "pitbull", "Louis_Tomlinson", "BillGates", "aplusk",
		"ArianaGrande", "AvrilLavigne", "Drake"}
	handles_len := len(handles)

	//Connection setup
	socket, _ := zmq.NewSocket(zmq.PUSH)
	defer socket.Close()
	zmqPort := "7450"
	zmqServer := fmt.Sprintf("tcp://localhost:%s", zmqPort)
	err := socket.Connect(zmqServer)
	if err != nil {
		log.Println("zmq server connection error", err)
	}

	//Lets say we got 1000 retweets in total
	for i := 0; i < 1000; i++ {
		tweet := tweets[rand_gen(100)]
		rand_handle := []string{handles[rand_gen(handles_len)]}
		request := createRequest(tweet, rand_handle)
		inputData(request, socket)
	}

	//Wait some time so the server can pull the request
	time.Sleep(3 * time.Second)
}
