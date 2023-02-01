package twins

import (
	"fmt"
	"testing"

	"github.com/relab/hotstuff"
	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

func basicScenario(t *testing.T) {
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

	fmt.Print(result.NetworkLog)
	fmt.Print("\n\n\n\n\n\n\n\n")
	for i, message := range result.Messages {
		fmt.Printf("%d ---- %v\n", i, message)
	}

	//fmt.Printf("%v\n", result.Messages)

	oldMessage := result.Messages[0]

	hash := oldMessage.(hotstuff.ProposeMsg).Block.Parent()

	newMessage := hotstuff.ProposeMsg{
		ID: 1,
		Block: hotstuff.NewBlock(
			hash,
			hotstuff.NewQuorumCert(nil, 1, hash),
			"",
			1,
			1,
		),
	}

	result, err := ExecuteScenario(s, 4, 0, 100, "chainedhotstuff", oldMessage, newMessage)

	fmt.Print("\n\n\n\n\n\n\n\n")
	fmt.Print(result.NetworkLog)

	if err != nil {
		t.Fatal(err)
	}

	if !result.Safe {
		t.Errorf("Expected no safety violations")
	}

	if result.Commits != 1 {
		t.Errorf("Expected one commit, got (%d)", result.Commits)
	}
}

func TestBasicScenario(t *testing.T) {
	basicScenario(t)
}

/*
func NewBlock(parent Hash, cert QuorumCert, cmd Command, view View, proposer ID) *Block {
	b := &Block{
		parent:   parent,
		cert:     cert,
		cmd:      cmd,
		view:     view,
		proposer: proposer,
	}
	// cache the hash immediately because it is too racy to do it in Hash()
	b.hash = sha256.Sum256(b.ToBytes())
	return b
}
*/

func FuzzBasicScenario(f *testing.F) {
	f.Add()
	f.Fuzz(func(t *testing.T) {
		basicScenario(t)
	})
}
