package twins

import (
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/relab/hotstuff"

	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

type PanicInfo struct {
	Err        any
	StackTrace string
}

type ErrorInfo struct {
	errorCount int
	panics     map[string][]PanicInfo
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

	errorInfo.panics[simpleStack] = panicList

	errorInfo.errorCount++
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

	simpleStackLines = simpleStackLines[3:]

	//simpleStack := strings.Join(simpleStackLines, "\n")
	simpleStack := simpleStackLines[0]

	return simpleStack
}

func TryExecuteScenario(t *testing.T, errorInfo *ErrorInfo, scenario Scenario, numNodes, numTwins uint8, numTicks int, consensusName string, oldMessage any, newMessage any) {

	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err)

			//fmt.Printf("%s\n\n\n", stack)

			//fmt.Printf("PANIC ERROR: %v\n", err)
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

	//fmt.Println(result.NetworkLog)
}

func basicScenario(t *testing.T, newMessage any) {

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

		TryExecuteScenario(t, errorInfo, s, numNodes, 0, 100, "chainedhotstuff", oldMessage, newMessage)

		//fmt.Print(result.NetworkLog)
	}

	fmt.Printf("unique errors found: %d\n", len(errorInfo.panics))

	if len(errorInfo.panics) > 1 {
		for key := range errorInfo.panics {
			fmt.Println()
			fmt.Println(key)
			fmt.Println()
		}
		panic("many unique errors")
	}

	fmt.Printf("%d of %d runs were errors\n", errorInfo.errorCount, messageCount)
}

func TestBasicScenario(t *testing.T) {

	newMessage := hotstuff.ProposeMsg{
		ID: 1,
		Block: hotstuff.NewBlock(
			[32]byte{},
			hotstuff.NewQuorumCert(nil, 0, [32]byte{}),
			"0",
			1,
			1,
		),
	}

	basicScenario(t, newMessage)
}

func TestFuzz(t *testing.T) {

	nilChance := 0.1

	f := fuzz.New().NilChance(nilChance).Funcs(
		func(m *any, c fuzz.Continue) {
			msgs := []any{
				hotstuff.ProposeMsg{},
				hotstuff.VoteMsg{},
				hotstuff.TimeoutMsg{},
				hotstuff.NewViewMsg{},
				hotstuff.CommitEvent{},
			}
			*m = msgs[c.Intn(5)]
			c.Fuzz(m)
		},
		func(block **hotstuff.Block, c fuzz.Continue) {

			if c.Float64() < nilChance {
				*block = nil
				return
			}

			var hash [32]byte
			var cert hotstuff.QuorumCert
			var cmd hotstuff.Command
			var view hotstuff.View
			var proposer hotstuff.ID

			c.Fuzz(&hash)
			c.Fuzz(&cert)
			c.Fuzz(&cmd)
			c.Fuzz(&view)
			c.Fuzz(&proposer)
			*block = hotstuff.NewBlock(
				hash,
				cert,
				cmd,
				view,
				proposer,
			)
		},
		func(qc *hotstuff.QuorumCert, c fuzz.Continue) {
			var view hotstuff.View
			c.Fuzz(&view)
			var hash [32]byte
			c.Fuzz(&hash)

			*qc = hotstuff.NewQuorumCert(nil, view, hash)
		},
	)

	for i := 0; i < 10; i++ {
		var newMessage any
		f.Fuzz(&newMessage)

		fmt.Printf("%T, %v\n", newMessage, newMessage)
		basicScenario(t, newMessage)
	}
}
