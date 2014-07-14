
#Abstract

TAS stands for True Air Speed. It is a lightweight system to track and analyze real time data. 
TAS is able to handle a lot of incoming data, create visuals to analyze it and take snapshots to furthur examine the data at a specific time. 

The data in TAS is stored for approximately a minute, the idea is to track current trends and analyze real time data while keeping the system light weight. 


#Prerequisites

1. Go 1.2.2 package must be installed on your computer.

2. zmq2 package must also be installed.


#Getting Started


###Installation

1. Download the Go 1.2.2 package:http://golang.org/dl/ 
  
  *Note: TAS hasn’t been tested on Go 1.3.*  
2. 
Install the zmq2 library from github:
```go get github.com/pebbe/zmq2 ```
3. Download the TAS package, run:
```go get github.com/chango/tas```

----
###Example


The server has to be up and running first. For UNIX-like systems, type the following commands in terminal:
```
go run $GOPATH/src/github.com/chango/tas/tas.go
```
On a separate terminal, run **hellotas.go** by switching to the examples directory and entering the following command:
```
go run $GOPATH/src/github.com/chango/tas/examples/hellotas.go
```
Open any modern web browser and go to the following addresses to see the output of TAS:

- localhost:7451/DIAG            

- localhost:7451/TREE             

- localhost:7451/STATS          

- localhost:7451/GET?key=\*.\*.\*.\*

----
#Documentation
The documentation can be found [here](./doc/documentation.md).


