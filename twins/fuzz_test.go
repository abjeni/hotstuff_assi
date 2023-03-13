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

	for key, panics := range errorInfo.panics {
		panic := panics[0]
		fmt.Println()
		fmt.Println(panic.Err)
		fmt.Println(key)
		fmt.Println()
		fmt.Println("- FUZZ MESSAGE BEGIN")
		fmt.Println(panic.FuzzMsg.ToString(0))
		fmt.Println("- FUZZ MESSAGE END")
		fmt.Println()
		fmt.Println("- STACK TRACE BEGIN")
		fmt.Print(panic.StackTrace)
		fmt.Println("- STACK TRACE END")
		fmt.Println()
	}

	fmt.Printf("unique errors found: %d\n", len(errorInfo.panics))
	fmt.Printf("%d runs were errors\n", errorInfo.errorCount)
	fmt.Printf("%d of %d scenarios failed\n", errorInfo.failedScenarios, errorInfo.totalScenarios)
	fmt.Printf("%d of %d messages failed\n", errorInfo.failedMessages, errorInfo.totalMessages)
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
		FuzzMsg:    errorInfo.currentFuzzMsg,
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
			if line[0] == '\t' {
				simpleStackLines = append(simpleStackLines, line[1:])
			}
		}
	}
	/**
	simpleStack := strings.Join(simpleStackLines, "\n")
	/*/
	simpleStackLines = simpleStackLines[3:]
	simpleStack := simpleStackLines[0]
	/**/

	return simpleStack
}

func TryExecuteScenario(t *testing.T, errorInfo *ErrorInfo, scenario Scenario, numNodes, numTwins uint8, numTicks int, consensusName string, oldMessage any, newMessage any) {
	errorInfo.totalScenarios++
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())

			errorInfo.AddPanic(stack, err)
			errorInfo.failedScenarios++
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

func fuzzScenario(t *testing.T, errorInfo *ErrorInfo, newMessage any) {

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

	for i := 1; i <= messageCount; i++ {
		oldMessage := i

		TryExecuteScenario(t, errorInfo, s, numNodes, 0, 100, "chainedhotstuff", oldMessage, newMessage)
	}
}

func getFuzzMessage(f *fuzz.Fuzzer, errorInfo *ErrorInfo) FuzzMsg {
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

func TestFuzz(t *testing.T) {
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

	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	for i := 0; i < 100; i++ {

		newFuzzMessage := getFuzzMessage(f, errorInfo)

		errorInfo.currentFuzzMsg = newFuzzMessage

		newMessage := fuzzMsgToMsg(errorInfo, newFuzzMessage)

		if newMessage != nil {
			//fmt.Printf("%T, %v\n", newMessage, newMessage)
			fuzzScenario(t, errorInfo, newMessage)
		}
	}

	errorInfo.OutputInfo()
}
