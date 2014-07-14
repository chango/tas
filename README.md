
#Abstract

True Air Speed aka Tree Avoidance System, a lightweight system to track and analyze real-time data.


#Prerequisites

1. Go 1.2.2 package must be installed on your computer.

2. zmq2 package must also be installed.


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
git clone https://github.com/chango/tas.git

```
----
###Example


The server has to be up and running first. For UNIX-like systems, type the following commands in terminal:
```
go run tas.go
```
On a separate terminal, run **hellotas.go** by switching to the examples directory and entering the following command:
```
go run hellotas.go
```
Open any modern web browser and go to the following addresses to see the output of TAS:
1. localhost:7451/DIAG 
2. localhost:7451/TREE
3. localhost:7451/STATS
4. localhost:7451/GET?key=*.*.*.*

----
#Documentation
The documentation can be found [here](./doc/documentation.md).


