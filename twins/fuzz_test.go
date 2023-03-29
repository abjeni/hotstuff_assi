package twins

import (
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/relab/hotstuff/internal/proto/hotstuffpb"
)

type PanicInfo struct {
	Err        any
	StackTrace string
	FuzzMsg    FuzzMsg
}

type ErrorInfo struct {
	messageFile     string
	currentFuzzMsg  FuzzMsg
	errorCount      int
	panics          map[string][]PanicInfo
	totalScenarios  int
	failedScenarios int
	totalMessages   int
	failedMessages  int
}

func (errorInfo *ErrorInfo) Init() {
	errorInfo.panics = make(map[string][]PanicInfo)
}

func (errorInfo *ErrorInfo) OutputInfo() {
	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("ERROR INFO")

	messages := make([]FuzzMsg, 0)

	var i int = 0
	for key, panics := range errorInfo.panics {
		i++
		var panicInfo PanicInfo
		var panicMsgLines int = 9999999

		for _, panic := range panics {
			str := panic.FuzzMsg.ToString(0)
			lines := strings.Count(str, "\n")
			if lines < panicMsgLines {
				panicMsgLines = lines
				panicInfo = panic
			}
		}

		fmt.Println()
		fmt.Printf("ERROR NUMBER %d\n", i)
		fmt.Println(key)
		fmt.Println()
		fmt.Println("- STACK TRACE BEGIN")
		fmt.Print(panicInfo.StackTrace)
		fmt.Println("- STACK TRACE END")
		fmt.Println()
		fmt.Println("- FUZZ MESSAGE BEGIN")
		str := panicInfo.FuzzMsg.ToString(0)
		fmt.Println(str)
		fmt.Println("- FUZZ MESSAGE END")
		fmt.Println()

		messages = append(messages, panicInfo.FuzzMsg)
	}

	fmt.Printf("unique errors found: %d\n", len(errorInfo.panics))
	fmt.Printf("%d runs were errors\n", errorInfo.errorCount)
	fmt.Printf("%d of %d scenarios failed\n", errorInfo.failedScenarios, errorInfo.totalScenarios)
	fmt.Printf("%d of %d messages failed\n", errorInfo.failedMessages, errorInfo.totalMessages)
}

func (errorInfo *ErrorInfo) AddPanic(stack string, err any) {

	simpleStack := SimplifyStack(stack)
	identifier := simpleStack + "\n" + fmt.Sprint(err)

	panicList := errorInfo.panics[identifier]

	if panicList == nil {
		panicList = make([]PanicInfo, 0)
	}

	panic := PanicInfo{
		Err:        err,
		StackTrace: stack,
		FuzzMsg:    errorInfo.currentFuzzMsg,
	}

	panicList = append(panicList, panic)
	errorInfo.panics[identifier] = panicList
	errorInfo.errorCount++
}

func SimplifyStack(stack string) string {
	stackLines := strings.Split(strings.ReplaceAll(stack, "\r\n", "\n"), "\n")

	// line 9 tells us where the panic happened
	return stackLines[8]
}

func TryExecuteScenario(t *testing.T, errorInfo *ErrorInfo, oldMessage any, newMessage any) {
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
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})
	s = append(s, View{Leader: 1, Partitions: []NodeSet{allNodesSet}})

	errorInfo.totalScenarios++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err)
			errorInfo.failedScenarios++
		}
	}()

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
	messageCount := getMessagesBasicScenario()

	for oldMessage := 1; oldMessage <= messageCount; oldMessage++ {
		TryExecuteScenario(t, errorInfo, oldMessage, newMessage)
	}
}

func initFuzz() *fuzz.Fuzzer {
	nilChance := 0.1

	f := fuzz.New().NilChance(nilChance).Funcs(
		func(m *FuzzMsg, c fuzz.Continue) {
			switch c.Intn(4) {
			case 0:
				msg := ProposeMsg{}
				c.Fuzz(&msg)
				*m = &msg
			case 1:
				msg := VoteMsg{}
				c.Fuzz(&msg)
				*m = &msg
			case 2:
				msg := TimeoutMsg{}
				c.Fuzz(&msg)
				*m = &msg
			case 3:
				msg := NewViewMsg{}
				c.Fuzz(&msg)
				*m = &msg
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

func createFuzzMessage(f *fuzz.Fuzzer, errorInfo *ErrorInfo) FuzzMsg {
	var newMessage FuzzMsg
	f.Fuzz(&newMessage)
	return newMessage
}

func fuzzMsgToMsg(errorInfo *ErrorInfo, fuzzMsg FuzzMsg) any {
	errorInfo.totalMessages++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err)
			errorInfo.failedMessages++
		}
	}()

	return fuzzMsg.ToMsg()
}

func useFuzzMessage(t *testing.T, errorInfo *ErrorInfo, fuzzMessage FuzzMsg) {
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

	iterations := 100

	for i := 0; i < iterations; i++ {
		fuzzMessage := createFuzzMessage(f, errorInfo)
		useFuzzMessage(t, errorInfo, fuzzMessage)
	}

	errorInfo.OutputInfo()
}
