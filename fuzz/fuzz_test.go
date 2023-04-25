package fuzz

import (
	"fmt"
	"math/rand"
	"runtime/debug"
	"testing"

	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

func TryExecuteScenario(t *testing.T, errorInfo *ErrorInfo, oldMessage any, newMessage any) {

	errorInfo.totalScenarios++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err, "TryExecuteScenario")
			errorInfo.failedScenarios++
		}
	}()

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

	result, err := ExecuteScenario(s, numNodes, 0, 100, "chainedhotstuff", oldMessage, newMessage)

	if err != nil {
		panic(err)
	}

	if !result.Safe {
		panic("Expected no safety violations")
	}

	if result.Commits != 1 {
		panic(fmt.Sprintf("Expected one commit (got %d)", result.Commits))
	}
}

func getMessagesBasicScenario() int {
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

	messageCount := result.MessageCount

	return messageCount
}

func fuzzScenario(t *testing.T, errorInfo *ErrorInfo, newMessage any) {
	TryExecuteScenario(t, errorInfo, 1, newMessage)
}

func fuzzMsgToMsg(errorInfo *ErrorInfo, fuzzMsg *FuzzMsg) any {
	errorInfo.totalMessages++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err, "fuzzMsgToMsg")
			errorInfo.failedMessages++
		}
	}()

	return fuzzMsg.Msg().ToMsg()
}

func useFuzzMessage(t *testing.T, errorInfo *ErrorInfo, fuzzMessage *FuzzMsg, seed *int64) {
	errorInfo.currentFuzzMsg = fuzzMessage
	errorInfo.currentFuzzMsgSeed = seed

	newMessage := fuzzMsgToMsg(errorInfo, fuzzMessage)

	if newMessage != nil {
		fuzzScenario(t, errorInfo, newMessage)
	}
}

// the main test
func TestFuzz(t *testing.T) {
	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	f := initFuzz()

	iterations := 1000

	for i := 0; i < iterations; i++ {
		seed := rand.Int63()
		fuzzMessage := createFuzzMessage(f, errorInfo, &seed)
		useFuzzMessage(t, errorInfo, fuzzMessage, &seed)
	}

	errorInfo.OutputInfo(t)
}

// load previously created fuzz messages from a file
// it doesn't work quite right, i blame proto.Marshal()
func TestPreviousFuzz(t *testing.T) {

	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	fuzzMsgs, err := loadFuzzMessagesFromFile("previous_messages.b64")

	if err != nil {
		panic(err)
	}

	for _, fuzzMessage := range fuzzMsgs {
		useFuzzMessage(t, errorInfo, fuzzMessage, nil)
	}

	errorInfo.OutputInfo(t)
}

// load previously created fuzz messages from a file
// it recreates the fuzz messages from a 64-bit source
func TestSeedPreviousFuzz(t *testing.T) {

	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	seeds, err := loadSeedsFromFile("previous_messages.seed")

	if err != nil {
		panic(err)
	}

	f := initFuzz()

	for _, seed := range seeds {
		fmt.Println("seperator")
		fuzzMessage := createFuzzMessage(f, errorInfo, &seed)
		useFuzzMessage(t, errorInfo, fuzzMessage, nil)
	}

	errorInfo.OutputInfo(t)
}
