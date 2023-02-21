package twins

import (
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/relab/hotstuff"
	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

type PanicInfo struct {
	Err        any
	StackTrace string
}

type ErrorInfo struct {
	panics map[string][]PanicInfo
}

func (errorInfo *ErrorInfo) Init() {
	errorInfo.panics = make(map[string][]PanicInfo)
}

func (errorInfo *ErrorInfo) AddPanic(stack string, err any) {

	simpleStack := SimplifyStack(stack)

	panicList := errorInfo.panics[simpleStack]

	if panicList == nil {
		panicList = make([]PanicInfo, 0)
	}

	panic := PanicInfo{
		Err:        err,
		StackTrace: stack,
	}

	panicList = append(panicList, panic)
}

func SimplifyStack(stack string) string {

	stackLines := strings.Split(strings.ReplaceAll(stack, "\r\n", "\n"), "\n")

	simpleStackLines := make([]string, 0)

	for _, line := range stackLines {
		if len(line) > 0 {
			if line[0] == "\t"[0] {
				simpleStackLines = append(simpleStackLines, line[1:])
			}
		}
	}

	simpleStack := strings.Join(simpleStackLines, "\n")

	return simpleStack
}

func TryExecuteScenario(t *testing.T, errorInfo *ErrorInfo, scenario Scenario, numNodes, numTwins uint8, numTicks int, consensusName string, oldMessage any, newMessage any) {

	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err)

			fmt.Printf("PANIC ERROR: %v\n", err)
		}
	}()

	result, err := ExecuteScenario(scenario, numNodes, 0, 100, "chainedhotstuff", oldMessage, newMessage)

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

func basicScenario(t *testing.T, id hotstuff.ID, cmd hotstuff.Command, view hotstuff.View, proposer hotstuff.ID, hash [32]byte) {

	errorInfo := new(ErrorInfo)
	errorInfo.Init()

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

	result, _ := ExecuteScenario(s, numNodes, 0, 100, "chainedhotstuff")

	//fmt.Print(result.NetworkLog)

	//fmt.Printf("\n\n\n\nmessages: %v\n\n\n\n", result.MessageCount)

	messageCount := result.MessageCount

	for i := 1; i <= messageCount; i++ {
		//fmt.Print("\n\n\n\n\n")
		oldMessage := i

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

		TryExecuteScenario(t, errorInfo, s, numNodes, 0, 100, "chainedhotstuff", oldMessage, newMessage)

		//fmt.Print(result.NetworkLog)
	}
}

func TestBasicScenario(t *testing.T) {
	basicScenario(t, 1, "0", 1, 1, [32]byte{})
}

func FuzzBasicScenario(f *testing.F) {
	f.Add(uint32(1), string("0"), uint64(1), uint32(1))
	f.Fuzz(func(t *testing.T, id uint32, cmd string, view uint64, proposer uint32) {
		basicScenario(t, hotstuff.ID(id), hotstuff.Command(cmd), hotstuff.View(view), hotstuff.ID(proposer), [32]byte{})
	})
}
