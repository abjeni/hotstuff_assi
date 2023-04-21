package fuzz

import (
	"fmt"
	"runtime/debug"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/relab/hotstuff/internal/proto/hotstuffpb"
	"github.com/relab/hotstuff/twins"

	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

func TryExecuteScenario(t *testing.T, errorInfo *ErrorInfo, oldMessage any, newMessage any) {
	var numNodes uint8 = 4

	allNodesSet := make(twins.NodeSet)
	for i := 1; i <= int(numNodes); i++ {
		allNodesSet.Add(uint32(i))
	}

	s := twins.Scenario{}
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})

	errorInfo.totalScenarios++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err, "TryExecuteScenario")
			errorInfo.failedScenarios++
		}
	}()

	result, err := twins.ExecuteScenario(s, numNodes, 0, 100, "chainedhotstuff", oldMessage, newMessage)

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

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func getMessagesBasicScenario() int {
	var numNodes uint8 = 4

	allNodesSet := make(twins.NodeSet)
	for i := 1; i <= int(numNodes); i++ {
		allNodesSet.Add(uint32(i))
	}

	s := twins.Scenario{}
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})
	s = append(s, twins.View{Leader: 1, Partitions: []twins.NodeSet{allNodesSet}})

	result, _ := twins.ExecuteScenario(s, numNodes, 0, 100, "chainedhotstuff")

	messageCount := result.MessageCount

	return messageCount
}

func fuzzScenario(t *testing.T, errorInfo *ErrorInfo, newMessage any) {
	TryExecuteScenario(t, errorInfo, 1, newMessage)
}

func initFuzz() *fuzz.Fuzzer {
	nilChance := 0.1

	f := fuzz.New().NilChance(nilChance).Funcs(
		func(m *FuzzMsg, c fuzz.Continue) {
			switch c.Intn(4) {
			case 0:
				msg := FuzzMsg_ProposeMsg{
					ProposeMsg: &ProposeMsg{},
				}
				c.Fuzz(msg.ProposeMsg)
				m.Message = &msg
			case 1:
				msg := FuzzMsg_VoteMsg{
					VoteMsg: &VoteMsg{},
				}
				c.Fuzz(msg.VoteMsg)
				m.Message = &msg
			case 2:
				msg := FuzzMsg_TimeoutMsg{
					TimeoutMsg: &TimeoutMsg{},
				}
				c.Fuzz(msg.TimeoutMsg)
				m.Message = &msg
			case 3:
				msg := FuzzMsg_NewViewMsg{
					NewViewMsg: &NewViewMsg{},
				}
				c.Fuzz(msg.NewViewMsg)
				m.Message = &msg
			}
		},
		func(sig **hotstuffpb.QuorumSignature, c fuzz.Continue) {
			if c.Float64() < nilChance {
				*sig = nil
				return
			}

			*sig = new(hotstuffpb.QuorumSignature)
			switch c.Intn(2) {
			case 0:
				ecdsa := new(hotstuffpb.QuorumSignature_ECDSASigs)
				c.Fuzz(&ecdsa)
				(*sig).Sig = ecdsa
			case 1:
				bls12 := new(hotstuffpb.QuorumSignature_BLS12Sig)
				c.Fuzz(&bls12)
				(*sig).Sig = bls12
			}
		},
	)

	return f
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

func useFuzzMessage(t *testing.T, errorInfo *ErrorInfo, fuzzMessage *FuzzMsg) {
	errorInfo.currentFuzzMsg = fuzzMessage

	newMessage := fuzzMsgToMsg(errorInfo, fuzzMessage)

	if newMessage != nil {
		fuzzScenario(t, errorInfo, newMessage)
	}
}

func createFuzzMessage(f *fuzz.Fuzzer, errorInfo *ErrorInfo) *FuzzMsg {
	newMessage := new(FuzzMsg)
	f.Fuzz(newMessage)
	return newMessage
}

func TestFuzz(t *testing.T) {
	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	f := initFuzz()

	iterations := 100

	for i := 0; i < iterations; i++ {
		fuzzMessage := createFuzzMessage(f, errorInfo)
		useFuzzMessage(t, errorInfo, fuzzMessage)
	}

	errorInfo.OutputInfo()
}

// it doesn't work quite right, i blame proto.Marshal()
func TestPreviousFuzz(t *testing.T) {

	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	fuzzMsgs, err := loadFuzzMessagesFromFile("previous_messages.b64")

	if err != nil {
		panic(err)
	}

	for _, fuzzMessage := range fuzzMsgs {
		useFuzzMessage(t, errorInfo, fuzzMessage)
	}

	errorInfo.OutputInfo()
}

func TestFrequencyErrorFuzz(t *testing.T) {

	frequency := make(map[string]int, 0)

	f := initFuzz()
	for j := 0; j < 1000; j++ {
		errorInfo := new(ErrorInfo)
		errorInfo.Init()

		iterations := 1

		for i := 0; i < iterations; i++ {
			fuzzMessage := createFuzzMessage(f, errorInfo)
			useFuzzMessage(t, errorInfo, fuzzMessage)
		}

		for key := range errorInfo.panics {
			frequency[key]++
		}
	}

	for key, val := range frequency {
		fmt.Println(key)
		fmt.Println(val)
		fmt.Println()
	}

}
