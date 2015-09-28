
#Abstract

TAS stands for True Air Speed. It is a lightweight system to track and analyze real time data. 
TAS is able to handle a lot of incoming data, create visuals to analyze it and take snapshots to furthur examine the data at a specific time. 

The data in TAS is stored for approximately a minute, the idea is to track current trends and analyze real time data while keeping the system light weight. 


#Prerequisites

1. Go must be installed on your computer (*Note: TAS hasn't been tested on Go 1.5*)

2. `github.com/pebbe/zmq3` package must also be installed.


#Getting Started


###Installation

1. 
Install the zmq3 library from github (if you have not already):
```go get github.com/pebbe/zmq3```
2. Download the TAS package, run:
```go get github.com/chango/tas```

----
###Notes

There are some HTML templates that are in the `html` directory, the server will look for these to be in a directory named `html` in the root where the server is started.

----
###Example


The server has to be up and running first. For UNIX-like systems, type the following commands in terminal:
```
go run $GOPATH/src/github.com/chango/tas/examples/tas-server.go
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


