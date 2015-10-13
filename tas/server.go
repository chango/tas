package tas

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

import (
	"github.com/chango/tas/tree"
	"github.com/pebbe/zmq3"
)

type TASServer struct {
	config  *TASConfig
	pfdTree *tree.Tree
	socket  *zmq3.Socket
	closing bool
}

// Returns a new TAS server that is running in the background
func NewTASServer(config *TASConfig) (t *TASServer, err error) {
	t = &TASServer{
		config:  config,
		pfdTree: tree.MakeTree(),
	}
	t.socket, err = zmq3.NewSocket(zmq3.PULL)
	if err != nil {
		err = fmt.Errorf("Could not create ZMQ socket: %v", err)
		return
	}
	zmqAddress := fmt.Sprintf("tcp://%s:%s", t.config.ZMQAddress, t.config.ZMQPort)
	err = t.socket.Bind(zmqAddress)
	if err != nil {
		err = fmt.Errorf("Could not bind ZMQ socket: %v", err)
		return
	}
	go t.gcAgent()
	go t.receiver()
	go t.httpServer()
	return
}

// Blocking runner that traps SIGINT and SIGTERM to gracefully shutdown
// the TAS server.
func (t *TASServer) Run() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-c:
		tasLog.Info("[tas] Stopping server")
		t.close()
		break
	}
	tasLog.Info("[tas] Server stopped")
}

// ZMQ Receiver
func (t *TASServer) receiver() {
	var data interface{}
	var rawMessage string
	var err error

	tasLog.Info("[tas] Starting receiver")
	for {
		// incoming message format:
		// INCR/APPEND TS KEY VALUE
		if t.closing {
			return
		}
		rawMessage, err = t.socket.Recv(0)
		tasLog.Debug(rawMessage)
		if err != nil {
			tasLog.Info("[tas] ZMQ receive error ", err)
			continue
		}
		t.process(rawMessage)
	}
}

func (t *TASServer) process(rawMessage string) {
	defer func() {
		if r := recover(); r != nil {
			tasLog.Info("TAS Panic", rawMessage, r)
		}
	}()

	message := strings.SplitN(rawMessage, " ", 4)
	if message[0] == "INCR" {
		value, e := strconv.Atoi(message[3])
		if e == nil {
			t.pfdTree.AddData(message[2], value, message[1])
		}
	} else if message[0] == "APPEND" {
		e := json.Unmarshal([]byte(message[3]), &data)
		if e == nil {
			t.pfdTree.AddData(message[2], data, message[1])
		}
	}

}

// Agent that runs a GC on all the child nodes every 4 seconds
func (t *TASServer) gcAgent() {
	log.Println("[tas] Starting gcAgent")
	roundToNearest := 5
	for {
		if t.closing {
			return
		}
		currentTime := time.Now().Unix()
		gcTime := roundToNearest*int(currentTime/int64(roundToNearest)) - 60
		gcTimeStr := strconv.Itoa(gcTime)
		if gcTime%8 == 0 {
			for _, c := range *t.pfdTree.Timestamps() {
				if c.Key < gcTimeStr {
					t.pfdTree.DoGC(c.Key)
				}
			}
		} else {
			t.pfdTree.DoGC(gcTimeStr)
		}
		time.Sleep(4 * time.Second)
	}
}

// The HTTP server
func (t *TASServer) httpServer() {
	var err error
	http.HandleFunc("/GET", func(w http.ResponseWriter, r *http.Request) {
		var tsList []string
		if r.FormValue("t") != "" {
			tsList = strings.Split(r.FormValue("t"), ",")
		}
		var intervalSeconds float64 = 5.0
		if r.FormValue("i") != "" {
			intervalSeconds, _ = strconv.ParseFloat(r.FormValue("i"), 32)
		}
		val := t.pfdTree.GetValue(strings.Split(r.FormValue("key"), "."), tsList, intervalSeconds)
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
			"oldest_timestamp": t.pfdTree.GetOldestTS(),
			"current_time":     time.Now().Unix(),
			"gc_running":       t.pfdTree.CheckGCRunning(),
			"num_leafs":        t.pfdTree.GetNumLeafs(),
			"ts_counts":        TSCounters(t.pfdTree.Timestamps()),
		}
		returnVal, e := json.Marshal(mapVal)
		if e != nil {
			returnVal = []byte("{}")
		}
		fmt.Fprint(w, string(returnVal))
	})

	http.HandleFunc("/TREE", func(w http.ResponseWriter, r *http.Request) {

		//Create a new template
		templ := template.New("Tree Structure")
		// define a child template which will hold the actual tree
		Tree := templ.New("Tree")

		//Parse the tree data to create a html version of the tree
		Tree, err = Tree.Parse(TreePrinter(t.pfdTree.DataNode))

		// Throw out error if any issue
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		//Parse files
		templ, err = templ.ParseFiles("html/tree-d3.html")

		//Throw out error
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf("Error parsing file: %v", err))
			return
		}

		//Execute the template and write it out
		err = templ.ExecuteTemplate(w, "tree-d3.html", t.pfdTree.DataNode)

		//Throw out error
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
	})

	http.HandleFunc("/STATS", func(w http.ResponseWriter, r *http.Request) {

		ts_counts := TSCounters(t.pfdTree.Timestamps())

		// Create a new template
		templ := template.New("Timestamp Counts")
		// Define a child template which holds the timestamp counts
		Timestamps := templ.New("TS_COUNTS")

		// Parse the timestamp counts data to create a html version of the bar chart
		Timestamps, err = Timestamps.Parse(ts_counts)

		// Throw out error if any issue
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		//Parse files
		templ, err = templ.ParseFiles("html/ts-stats-d3.html")

		//Throw out error
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf("Error parsing file: %v", err))
			return
		}

		//Execute the template and write it out
		err = templ.ExecuteTemplate(w, "ts-stats-d3.html", ts_counts)

		//Throw out error
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templ := template.New("Main")
		templ, err := templ.ParseFiles("html/main.html")

		//Throw out error
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf("Error parsing file: %v", err))
			return
		}

		err = templ.ExecuteTemplate(w, "main.html", nil)

		//Throw out error
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf("Error parsing ex file: %v", err))
			return
		}
	})

	httpAddr := fmt.Sprintf("%s:%s", t.config.HTTPAddress, t.config.HTTPPort)
	log.Printf("[tas] HTTP listening on %s...", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

// Gracefully close TAS and its subsequent connections
func (t *TASServer) close() {
	log.Println("[tas] Closing server connections")
	if !t.closing {
		t.closing = true
		t.socket.Close()
	}
}
