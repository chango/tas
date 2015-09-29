package tas

import (
	"fmt"
	"strconv"
	"strings"
)

import (
	"github.com/chango/tas/tree"
)

//Generates a json for the data.
func TreePrinter(node *tree.Node) string {

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

			output := fmt.Sprintf("%v", val_slice)
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

// Counts the number of nodes for each timestamp
// and returns the result in json format
func TSCounters(timestamps *map[string]*tree.Node) string {

	ts_counts := make(map[string]int)

	// traverse timestampNodes
	// and calculates the number of nodes for each timestamp
	for ts, children := range *timestamps {
		leaves := *children
		for _, _ = range leaves.Children {
			if _, ok := ts_counts[ts]; ok {
				ts_counts[ts]++
			} else {
				ts_counts[ts] = 1
			}
		}
	}

	//Print ts_counts as a string in json format
	var output string
	for ts, count := range ts_counts {
		ts_int, _ := strconv.Atoi(ts)
		output += fmt.Sprintf(" {\"timestamp\":%d,\"count\":%d},", ts_int, count)
	}
	output = strings.TrimRight(output, ",")
	return output
}
