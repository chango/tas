package main

import (
	
	"testing"
	"time"
	"strconv"
	"../tree"
)

func createAddDataTree(key string, val int) (*tree.Tree, string){
	//Create and add data with key and val, return the tree and the
	//time stamp of the data
	
	pfdTree := tree.MakeTree()
	now := string(time.Now().Unix())
	pfdTree.AddData(key, val, now)
	return pfdTree, now
}

//Tree Test Cases
func TestAddData(t *testing.T) {
	// Test if data is being added properly
	
	pfdTree, now := createAddDataTree("test.tes.te.t", 5)
	
	for k,_ := range pfdTree.DataNode.Children {
		if k!= "test" {
			t.Error("Error with adding data to tree")
		}
	}
	for k, _ := range pfdTree.TimestampNode.Children {
		if k != now {
			t.Error("Error with adding data to tree")
		}
	}
}

func TestGetValue(t *testing.T) {
	// Test the Tree.GetValue function

	pfdTree, now := createAddDataTree("test.tes.te.t", 5)
	val :=pfdTree.GetValue([]string{"test", "tes", "te", "t"}, []string{now}, 5)
	if val != 5 {
		t.Error("Error with getting value")
	}

}

func TestTimestamps(t *testing.T) {
	// Test the Node.Timestamps function

	pfdTree, now := createAddDataTree("test.tes.te.t", 5)
	ts := pfdTree.Timestamps()
	for key, _ := range *ts {
		if key != now {
			t.Error("Error with timestamps")
		}
	}
}

//Node Test cases
func TestGetChild(t *testing.T) {
	//Test the Node.HasChild function

	pfdTree, _ := createAddDataTree("test.tes.te.t", 5)
	node := pfdTree.DataNode.GetChild("test")
	
	if node.Key != "test" {
		t.Error("Error with getting child of a node")
	}
}

func TestDeleteChild(t *testing.T) {
	//Test the Node.DeleteChild function
	pfdTree, _ := createAddDataTree("test.tes.te.t", 5)
	pfdTree.DataNode.DeleteChild("test")
	if pfdTree.DataNode.GetChild("test") != nil {
		t.Error("Error with deleting child")
	}
}

func TestHasValue(t *testing.T) {
	//Test if the inserted node has a value

	pfdTree, now := createAddDataTree("test", 5)
	node := pfdTree.DataNode.GetChild("test").GetChild(now)
	if !node.HasValue(){
		t.Error("Error with HasValue function")
	}

	if pfdTree.DataNode.GetChild("test").HasValue() {
		t.Error("Error, node has value when it shouldn't")
	}
}

func TestAddChild(t *testing.T) {
	//Testing add child function
	pfdTree, _ := createAddDataTree("test.tes.te.t", 5)
	pfdTree.DataNode.AddChild([]string{"test2", "tes", "te", "t"})
	if pfdTree.DataNode.Children["test2"] == nil {
		t.Error("Error with adding child to a node")
	}

}

func TestGetNumLeafs(t *testing.T) {
	//Testing function GetNumLeafs
	pfdTreeA := tree.MakeTree() 
	pfdTreeB, _ := createAddDataTree("test.tes.te.t", 5)
	if pfdTreeB.GetNumLeafs() != 1 || pfdTreeA.GetNumLeafs() != 0{
		t.Error("Error with the number of leafs in the tree")
	}
}

func TestGetOldestTS(t *testing.T) {
	pfdTreeA := tree.MakeTree()
	pfdTreeB, now := createAddDataTree("test.tes.te.t", 5)
	now_int, _ := strconv.Atoi(now)
	if pfdTreeB.GetOldestTS() != now_int || pfdTreeA.GetOldestTS() != 0 {
		t.Error("Error with the oldest time stamp in the tree")
	}
}
