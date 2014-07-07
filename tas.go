package main

import (
	"encoding/json"
	//"os"
	//"runtime/pprof"
	"fmt"
	"log"
	"net/http"
	"html/template"
	"runtime"
	"strconv"
	"strings"
	"time"
)

import (
	"./tree"
	zmq "github.com/pebbe/zmq2"
)

func GCAgent(pfdTree *tree.Tree) {
	roundToNearest := 5
	for {
		currentTime := time.Now().Unix()
		gcTime := roundToNearest*int(currentTime/int64(roundToNearest)) - 60
		gcTimeStr := strconv.Itoa(gcTime)
		if gcTime%8 == 0 {
			for _, c := range *pfdTree.Timestamps() {
				if c.Key < gcTimeStr {
					pfdTree.DoGC(c.Key)
				}
			}
		} else {
			pfdTree.DoGC(gcTimeStr)
		}
		time.Sleep(4 * time.Second)
	}
}


func main() {
	runtime.GOMAXPROCS(10)
	pfdTree := tree.MakeTree()
	socket, err := zmq.NewSocket(zmq.PULL)
	zmqPort := "7450"
	httpPort := "7451"

	defer socket.Close()
	if err != nil {
		log.Println("zmq socket error", err)
		return
	}
	zmqServer := fmt.Sprintf("tcp://*:%s", zmqPort)
	err = socket.Bind(zmqServer)
	if err != nil {
		log.Println("zmq port connect error", err)
	}
	log.Println(fmt.Sprintf("TCP listening on %s...", zmqPort))
	// Start the garbage collector
	go GCAgent(pfdTree)

	var data interface{}
	go func() {
		for {
			// incoming message format:
			// INCR/APPEND TS KEY VALUE
			rawMessage, err := socket.Recv(0)
			if err != nil {
				log.Println("zmq receive error ", err)
				continue
			}
			message := strings.SplitN(rawMessage, " ", 4)
			if message[0] == "INCR" {
				value, e := strconv.Atoi(message[3])
				if e == nil {
					pfdTree.AddData(message[2], value, message[1])
				}
			} else if message[0] == "APPEND" {
				e := json.Unmarshal([]byte(message[3]), &data)
				if e == nil {
					pfdTree.AddData(message[2], data, message[1])
				}
			}
		}
	}()

	http.HandleFunc("/GET", func(w http.ResponseWriter, r *http.Request) {
		var tsList []string
		if r.FormValue("t") != "" {
			tsList = strings.Split(r.FormValue("t"), ",")
		}
		var intervalSeconds float64 = 5.0
		if r.FormValue("i") != "" {
			intervalSeconds, _ = strconv.ParseFloat(r.FormValue("i"), 32)
		}
		val := pfdTree.GetValue(strings.Split(r.FormValue("key"), "."), tsList, intervalSeconds)
		returnVal, e := json.Marshal(val)
		if e != nil {
			returnVal = []byte("{}")
		}
		fmt.Fprint(w, string(returnVal))
	})

	http.HandleFunc("/DIAG", func(w http.ResponseWriter, r *http.Request) {
    	// Function called to get diagnostics


		//Create a map of all the diagnostics
    	mapVal := map[string]interface{}{
    		"oldest_timestamp":pfdTree.GetOldestTS(),
    		"current_time":time.Now().Unix(),
    		"gc_running":pfdTree.CheckGCRunning(),
    		"num_leafs":pfdTree.GetNumLeafs(),
    		"ts_counts":TSCounters(pfdTree.Timestamps()),

    	}
    	returnVal, e := json.Marshal(mapVal)
    	if e != nil {
    		returnVal = []byte("{}")
    	}
    	fmt.Fprint(w, string(returnVal))
    })

	http.HandleFunc("/TREE", func(w http.ResponseWriter, r *http.Request) {


		//Create a new template
		t := template.New("Tree Structure")
		// define a child template which will hold the actual tree
		Tree := t.New("Tree")
		
		//Parse the tree data to create a html version of the tree
		Tree, err = Tree.Parse(TreePrinter(pfdTree.DataNode))
		
		// Throw out error if any issue		
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
	 
		//Parse files
	    t, err = t.ParseFiles("tree-d3.html")

	    //Throw out error 
	    if err != nil {
	    	fmt.Fprint(w, "Error parsing file")
	    	return
	    }
	    
	    //Execute the template and write it out
	    err = t.ExecuteTemplate(w , "tree-d3.html", pfdTree.DataNode)

	    //Throw out error 
	    if err != nil {
	    	fmt.Fprint(w, err)
	    	return
	    }
	})

	http.HandleFunc("/STATS", func(w http.ResponseWriter, r *http.Request) {

		ts_counts := TSCounters(pfdTree.Timestamps())

		// Create a new template
		t := template.New("Timestamp Counts")
		// Define a child template which holds the timestamp counts
		Timestamps := t.New("TS_COUNTS")

		// Parse the timestamp counts data to create a html version of the bar chart
		Timestamps, err = Timestamps.Parse(ts_counts)

		// Throw out error if any issue		
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		//Parse files
		t, err = t.ParseFiles("ts-stats-d3.html")

	    //Throw out error 
	    if err != nil {
	    	fmt.Fprint(w, "Error parsing file")
	    	return
	    }

	    //Execute the template and write it out
	    err = t.ExecuteTemplate(w , "ts-stats-d3.html", ts_counts)

	    //Throw out error 
	    if err != nil {
	    	fmt.Fprint(w, err)
	    	return
	    }
	})

	log.Println(fmt.Sprintf("HTTP listening on %s...", httpPort))
	log.Fatal(http.ListenAndServe("0.0.0.0:"+httpPort, nil))
}

func TSCounters(timestamps *map[string]*tree.Node) string{
	// Counts the number of nodes for each timestamp 
	// and returns the result in json format

	ts_counts := make(map[string]int)

	// traverse timestampNodes 
	// and calculates the number of nodes for each timestamp
	for ts, children := range *timestamps{
		leaves := *children
		for _, _ = range leaves.Children{
			if _, ok := ts_counts[ts]; ok{
				ts_counts[ts]++
			}else{
				ts_counts[ts] = 1
			}
		}
	}

	//Print ts_counts as a string in json format
	var output string
	for ts, count := range ts_counts{
		ts_int, _ := strconv.Atoi(ts)
		output += fmt.Sprintf(" {\"timestamp\":%d,\"count\":%d},", ts_int, count)
	}
	output = strings.TrimRight(output, ",")	
	return output
}

func TreePrinter(node *tree.Node) string {
	//Generates a json for the data.

	//To see if the node is the datanode
	var parentName string = "null"
	if node.Parent != nil {
		parentName = node.Parent.Key
	}

	//Number of children
	num_children := len(node.GetAllChildren())
	nodeName := node.Key

	//If timestamp node
	if num_children == 0 && parentName != "null" {
		//Integer value
		if val_int, ok := node.Value.(int); ok {
			nodeName = "value: " + strconv.Itoa(val_int) + " timestamp:(" + nodeName + ")" 
		// Array value 
		} else if val_slice, ok := node.Value.([]interface{}); ok {

			output  := fmt.Sprintf("%v", val_slice)
			nodeName = "value: " + output + " timestamp:(" + nodeName + ")" 

		}
	}

	var output string = "{\"name\": \"" + nodeName + "\", \"parent\": \"" + parentName + "\""
	//We check if the parentName is null just so the datanode appears even if it has no children
	if len(node.GetAllChildren()) != 0 || parentName == "null" {
		output += ", \"children\": [ "
		for _, n := range node.Children {
			output += TreePrinter(n)
			output += ","
		}
		output += "]"
	}
	output += "}"
	return output
}


