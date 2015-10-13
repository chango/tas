package tree

import "strings"

// For testing functions
import (
	"math"
	"strconv"
	"time"
)

type Tree struct {
	DataNode      *Node
	TimestampNode *Node
}

func (t *Tree) AddData(key string, value interface{}, timestamp string) {
	bottom := t.DataNode.AddChild(strings.Split(key, "."))
	t.TimestampNode.AddValueToChild(timestamp, key, bottom, value)
}

func (t *Tree) GetValue(key []string, tsList []string, intervalSeconds float64) interface{} {
	return t.DataNode.GetValue(key, tsList, intervalSeconds)
}

func (t *Tree) Timestamps() *map[string]*Node {
	return &t.TimestampNode.Children
}

func (t *Tree) DoGC(ts string) {
	if !t.TimestampNode.hasChild(ts) {
		return
	}

	tsNode := t.TimestampNode.GetChild(ts)
	keysToDelete := make([]string, len(tsNode.Children))
	i := 0
	for _, c := range tsNode.Children {
		// Delete timestamp node
		current := c.Children["val"].Parent
		if current != nil {
			c.Children["val"].Parent.DeleteChild(ts)
		}

		for len(current.Children) == 0 {
			key := current.Key
			current = current.Parent
			// If parent == nil, we're at the root. Don't delete the root
			if current != nil {
				current.DeleteChild(key)
			} else {
				break
			}
		}

		c.DeleteChild("val")
		if i < len(keysToDelete) {
			keysToDelete[i] = c.Key
		} else {
			keysToDelete = append(keysToDelete, c.Key)
		}
		i++
	}

	for _, c := range keysToDelete {
		tsNode.DeleteChild(c)
	}
	t.TimestampNode.DeleteChild(ts)
}

func (t *Tree) CheckGCRunning() bool {
	// Check if there is anything in the pfdTree older than 70 seconds.

	roundToNearest := 5

	currentTime := time.Now().Unix()
	gcTime := roundToNearest*int(currentTime/int64(roundToNearest)) - 65
	gcTimeStr := strconv.Itoa(gcTime)
	for _, c := range *t.Timestamps() {
		if c.Key < gcTimeStr {
			return false
		}
	}
	return true
}

func (t *Tree) GetOldestTS() int {

	//Get all the timestamp nodes map[timestamp]*node
	timestamps := t.Timestamps()
	//no time stamps so return 0
	if len(*timestamps) == 0 {
		return 0
	}
	oldestTs := math.MaxInt32
	for k, _ := range *timestamps {
		key, _ := strconv.Atoi(k)
		if key < oldestTs {
			oldestTs = key
		}

	}
	return oldestTs
}

type Node struct {
	Key      string
	Children map[string]*Node
	Parent   *Node
	Value    interface{}
}

func (n *Node) hasChild(key string) bool {
	_, ok := n.Children[key]
	return ok
}

func (n *Node) GetChild(key string) *Node {
	if !n.hasChild(key) {
		return nil
	}
	return n.Children[key]
}

func (n *Node) DeleteChild(key string) {
	delete(n.Children, key)
}

func (n *Node) HasValue() bool {
	return n.Value != nil
}

func makeNode(key string, parent *Node) *Node {
	return &Node{
		Key:      key,
		Children: make(map[string]*Node),
		Parent:   parent,
	}
}

func (n *Node) AddChild(key []string) *Node {
	if len(key) == 0 {
		return n
	}

	if !n.hasChild(key[0]) {
		n.Children[key[0]] = makeNode(key[0], n)
	}
	return n.Children[key[0]].AddChild(key[1:])
}

func (n *Node) AddValueToChild(ts string, key string, leaf *Node, value interface{}) {
	tsLeaf := n.AddChild([]string{ts, key})

	var valNode *Node
	if !tsLeaf.hasChild("val") {
		valNode = leaf.AddChild([]string{ts})
		tsLeaf.Children["val"] = valNode
	} else {
		valNode = leaf.Children[ts]
	}
	valNode.setValue(value)
}

func (n *Node) GetAllChildren() []*Node {
	nodeArr := []*Node{}
	for _, c := range n.Children {
		nodeArr = append(nodeArr, c)
	}
	return nodeArr
}

func (n *Node) GetValue(key []string, tsList []string, intervalSeconds float64) interface{} {
	if n == nil {
		return nil
	}

	if len(key) == 0 {
		return generateValue(n, tsList, intervalSeconds)
	}

	if key[0] == "*" {
		// If one of the children of the current node is a ts node,
		// calculate the value for the node. Else, continue to traverse
		for _, c := range n.Children {
			if c != nil && c.HasValue() {
				return generateValue(n, tsList, intervalSeconds)
			}
		}

		returnVal := make(map[string]interface{})
		for _, c := range n.Children {
			v := c.GetValue(key[1:], tsList, intervalSeconds)
			if v != nil {
				returnVal[c.Key] = v
			}
		}
		return returnVal
	} else {
		return n.GetChild(key[0]).GetValue(key[1:], tsList, intervalSeconds)
	}
}

func (n *Node) setValue(value interface{}) {
	if x, ok := value.(int); ok {
		if n.Value == nil {
			n.Value = 0
		}
		n.Value = n.Value.(int) + x
	} else if z, ok := value.([]interface{}); ok {
		if n.Value == nil {
			n.Value = z
		} else {
			for _, i := range z {
				n.Value = append(n.Value.([]interface{}), i)
			}
		}
	}
}

func (n *Node) GetNumChildren() int {
	//Return the number of children of the node.
	nodeArr := n.GetAllChildren()
	return len(nodeArr)
}

func (t *Tree) GetNumLeafs() int {
	//Return the number of leafs in the tree.
	if t.TimestampNode == nil {
		return 0
	}

	var NumLeafs int
	NumLeafs = 0
	for _, node := range t.TimestampNode.GetAllChildren() {
		NumLeafs += node.GetNumChildren()
	}
	return NumLeafs
}

func generateValue(n *Node, tsList []string, intervalSeconds float64) interface{} {
	if n == nil {
		return nil
	}

	var returnVal interface{}
	var numDataPoints float64 = 1
	var isInt bool = false

	for _, c := range n.Children {
		if c != nil && c.HasValue() && (tsList == nil || len(tsList) == 0 || isInArray(c.Key, &tsList)) {
			if returnVal == nil {
				returnVal = c.Value
			} else if x, ok := c.Value.(int); ok {
				isInt = true
				returnVal = returnVal.(int) + x
				numDataPoints++
			} else if y, ok := c.Value.([]interface{}); ok {
				returnVal = append(returnVal.([]interface{}), y...)
			}
		}
	}

	// For interger types, we need to find the average over time
	if isInt {
		returnVal = float64(returnVal.(int)) / (numDataPoints * intervalSeconds)
	}
	return returnVal
}

func MakeTree() *Tree {
	return &Tree{
		DataNode: &Node{
			Key:      "dataroot",
			Children: make(map[string]*Node),
		},
		TimestampNode: &Node{
			Key:      "tsroot",
			Children: make(map[string]*Node),
		},
	}
}

func isInArray(item string, l *[]string) bool {
	for _, v := range *l {
		if v == item {
			return true
		}
	}
	return false
}
