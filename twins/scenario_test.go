package twins

import (
	"testing"

	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

func TestBasicScenario(t *testing.T) {
	var numNodes uint8 = 4

	allNodesSet := make(NodeSet)
	for i := 1; i <= int(numNodes); i++ {
		allNodesSet.Add(uint32(i))
	}

	s := Scenario{}
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})

	result, err := ExecuteScenario(s, numNodes, 0, 100, "chainedhotstuff")

	//fmt.Println(result.NetworkLog)

	if err != nil {
		t.Fatal(err)
	}

	if !result.Safe {
		t.Errorf("Expected no safety violations")
	}

	if result.Commits != 1 {
		t.Errorf("Expected one commit (got %d)", result.Commits)
	}
}
