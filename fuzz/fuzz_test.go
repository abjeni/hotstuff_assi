package fuzz

import (
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/relab/hotstuff/internal/proto/hotstuffpb"
	"github.com/relab/hotstuff/twins"

	_ "github.com/relab/hotstuff/consensus/chainedhotstuff"
)

type PanicInfo struct {
	Err        any
	StackTrace string
	FuzzMsg    string
	FuzzMsgB64 string
	LineNum    int
}

type ErrorInfo struct {
	messageFile       string
	currentFuzzMsg    *FuzzMsg
	currentFuzzMsgB64 string
	errorCount        int
	panics            map[string]PanicInfo
	totalScenarios    int
	failedScenarios   int
	totalMessages     int
	failedMessages    int
}

func (errorInfo *ErrorInfo) Init() {
	errorInfo.panics = make(map[string]PanicInfo)
}

func (errorInfo *ErrorInfo) OutputInfo() {

	b64s := ""

	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("ERROR INFO")

	var i int = 0
	for key, panicInfo := range errorInfo.panics {
		b64s += panicInfo.FuzzMsgB64 + "\n"

		i++

		fmt.Println()
		fmt.Printf("ERROR NUMBER %d\n", i)
		fmt.Println("error location")
		fmt.Println(key)
		fmt.Println()
		fmt.Println("- STACK TRACE BEGIN")
		fmt.Print(panicInfo.StackTrace)
		fmt.Println("- STACK TRACE END")
		fmt.Println()
		fmt.Println("- FUZZ MESSAGE BEGIN")
		fmt.Println(panicInfo.FuzzMsg)
		fmt.Println("- FUZZ MESSAGE END")
		fmt.Println()
	}

	saveFuzzMessagesToFile("previous_messages.b64", b64s)

	fmt.Printf("unique errors found: %d\n", len(errorInfo.panics))
	fmt.Printf("%d runs were errors\n", errorInfo.errorCount)
	fmt.Printf("%d of %d scenarios failed\n", errorInfo.failedScenarios, errorInfo.totalScenarios)
	fmt.Printf("%d of %d messages failed\n", errorInfo.failedMessages, errorInfo.totalMessages)
}

func (errorInfo *ErrorInfo) AddPanic(fullStack string, err any) {

	simpleStack := SimplifyStack(fullStack)
	identifier := simpleStack + "\n" + fmt.Sprint(err)

	errorInfo.errorCount++

	oldPanic, ok := errorInfo.panics[identifier]

	b64, err := fuzzMsgToB64(errorInfo.currentFuzzMsg)
	if err != nil {
		panic(err)
	}

	FuzzMsgString := errorInfo.currentFuzzMsg.Msg().ToString(0)
	newLines := strings.Count(FuzzMsgString, "\n")

	newPanic := PanicInfo{
		Err:        err,
		StackTrace: fullStack,
		FuzzMsg:    FuzzMsgString,
		FuzzMsgB64: b64,
		LineNum:    newLines,
	}

	oldLines := oldPanic.LineNum

	if !ok || newLines < oldLines {
		errorInfo.panics[identifier] = newPanic
	}
}

func SimplifyStack(stack string) string {
	stackLines := strings.Split(strings.ReplaceAll(stack, "\r\n", "\n"), "\n")
	// line 9 tells us where the panic happened, found through testing
	return stackLines[8]
}

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

			errorInfo.AddPanic(stack, err)
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
			if c.RandBool() {
				ecdsa := new(hotstuffpb.QuorumSignature_ECDSASigs)
				c.Fuzz(&ecdsa)
				(*sig).Sig = ecdsa
			} else {
				bls12 := new(hotstuffpb.QuorumSignature_BLS12Sig)
				c.Fuzz(&bls12)
				(*sig).Sig = bls12
			}
		},
	)

	return f
}

func createFuzzMessage(f *fuzz.Fuzzer, errorInfo *ErrorInfo) *FuzzMsg {
	newMessage := new(FuzzMsg)
	f.Fuzz(newMessage)
	return newMessage
}

func fuzzMsgToMsg(errorInfo *ErrorInfo, fuzzMsg *FuzzMsg) any {
	errorInfo.totalMessages++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err)
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

func TestFuzz(t *testing.T) {
	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	f := initFuzz()

	iterations := 1000

	for i := 0; i < iterations; i++ {
		fuzzMessage := createFuzzMessage(f, errorInfo)
		useFuzzMessage(t, errorInfo, fuzzMessage)
	}

	errorInfo.OutputInfo()
}

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
