//Simulating web traffic through a restaurants menu
// Lets call this BurgerKing and the relative URLS:
// burgerking/
// burgerking/Menu
// burgerking/Menu/Breakfast
// burgerking/Menu/Lunch
// burgerking/Menu/Dinner
// burgerking/Menu/Snacks
// burgerking/Offers/fresh-offers
// burgerking/Offers/signup
// burgerking/Locations
// burgerking/info/
// burgerking/info/news
// burgerking/info/company-info 


package main


import (
	"fmt"
	"time"
	"math/rand"
	"log"
	"strings"
	"strconv"
	zmq "github.com/pebbe/zmq2"
)



func inputData(data string, socket *zmq.Socket) {
	_, err := socket.SendBytes([]byte(data), 0)
	if err != nil {
		log.Println("Error sending data")
	}
}

func createRequest(urlRequest string) string {
	//Create a request given a url request

	//Lets get the timestamp so we can time when this request was made
	//Randomize the timestamp to make the STATS page useful
	now := strconv.FormatInt(time.Now().Unix() + int64(rand.Intn(10)), 10)
	// Replace the / with . because that's how the tas enters data in a tree format
	urlRequest = strings.Replace(urlRequest, "/", ".", -1)
	msg := fmt.Sprintf("INCR %s %s %d", now, urlRequest, 1)

	return msg

}
func main () {
	
	urls := [12]string{"burgerking", "burgerking/Menu",
		"burgerking/Menu/Breakfast", "burgerking/Menu/Lunch", "burgerking/Menu/Dinner",
		"burgerking/Menu/Snacks", "burgerking/Offers/fresh-offers", "burgerking/Offers/signup",
		"burgerking/Locations", "burgerking/info", "burgerking/info/news", 
		"burgerkinginfo/company-info"}


	socket, _ := zmq.NewSocket(zmq.PUSH)
	defer socket.Close()

	zmqPort := "7450"
	zmqServer := fmt.Sprintf("tcp://localhost:%s", zmqPort)
	err := socket.Connect(zmqServer)
	if err != nil {
		log.Println("zmq server connection error", err)
	}

	//Lets say we got 100 url requests
	
	for c:=0; c < 100; c++ {
		index := rand.Intn(len(urls))
		url := urls[index]
		request := createRequest(url)
		inputData(request, socket)
		//Sleep so we get a new time stamp
	}

	//Wait some time so the server can pull the request
	time.Sleep(3*time.Second)
}