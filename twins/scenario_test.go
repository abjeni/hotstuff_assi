package twins

import (
	"testing"

	"github.com/relab/hotstuff"
	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

func basicScenario(t *testing.T, id hotstuff.ID, cmd hotstuff.Command, view hotstuff.View, proposer hotstuff.ID) {
	allNodesSet := make(NodeSet)
	for i := 1; i <= 4; i++ {
		allNodesSet.Add(uint32(i))
	}

	s := Scenario{}
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})

	result, _ := ExecuteScenario(s, 4, 0, 100, "chainedhotstuff", nil, nil)

	oldMessage := result.Messages[0]

	hash := oldMessage.(hotstuff.ProposeMsg).Block.Parent()

	newMessage := hotstuff.ProposeMsg{
		ID: id,
		Block: hotstuff.NewBlock(
			hash,
			hotstuff.NewQuorumCert(nil, 0, hash),
			cmd,
			view,
			proposer,
		),
	}

	result, err := ExecuteScenario(s, 4, 0, 100, "chainedhotstuff", oldMessage, newMessage)

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

func TestBasicScenario(t *testing.T) {
	basicScenario(t, 1, "0", 1, 1)
}

func FuzzBasicScenario(f *testing.F) {
	f.Add(uint32(1), string("0"), uint64(1), uint32(1))
	f.Fuzz(func(t *testing.T, id uint32, cmd string, view uint64, proposer uint32) {
		basicScenario(t, hotstuff.ID(id), hotstuff.Command(cmd), hotstuff.View(view), hotstuff.ID(proposer))
	})
}
