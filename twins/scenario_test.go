package twins

import (
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/relab/hotstuff"

	"github.com/relab/hotstuff/internal/proto/hotstuffpb"

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

func TestFuzz(t *testing.T) {

	nilChance := 0.1

	f := fuzz.New().NilChance(nilChance).Funcs(
		func(m *any, c fuzz.Continue) {
			switch c.Intn(5) {
			case 0:
				msg := hotstuff.ProposeMsg{}
				c.Fuzz(&msg)
				*m = msg
			case 1:
				msg := hotstuff.VoteMsg{}
				c.Fuzz(&msg)
				*m = msg
			case 2:
				msg := hotstuff.TimeoutMsg{}
				c.Fuzz(&msg)
				*m = msg
			case 3:
				msg := hotstuff.NewViewMsg{}
				c.Fuzz(&msg)
				*m = msg
			case 4:
				msg := hotstuff.CommitEvent{}
				c.Fuzz(&msg)
				*m = msg
			}
		},
		func(block **hotstuff.Block, c fuzz.Continue) {
			if c.Float64() < nilChance {
				*block = nil
				return
			}

			blockpb := hotstuffpb.Block{}
			c.Fuzz(&blockpb)
			*block = hotstuffpb.BlockFromProto(&blockpb)
		},
		func(sig *hotstuffpb.QuorumSignature, c fuzz.Continue) {
			if c.RandBool() {
				ecdsa := new(hotstuffpb.QuorumSignature_ECDSASigs)
				c.Fuzz(ecdsa)
				sig.Sig = ecdsa
			} else {
				bls12 := new(hotstuffpb.QuorumSignature_BLS12Sig)
				c.Fuzz(bls12)
				sig.Sig = bls12
			}
		},
		func(sig **hotstuffpb.QuorumSignature, c fuzz.Continue) {
			if c.Float64() < nilChance {
				*sig = nil
				return
			}

			*sig = new(hotstuffpb.QuorumSignature)
			c.Fuzz(*sig)
		},
		func(qc *hotstuff.QuorumCert, c fuzz.Continue) {
			qcpb := hotstuffpb.QuorumCert{}
			c.Fuzz(&qcpb)
			*qc = hotstuffpb.QuorumCertFromProto(&qcpb)
		},
		func(sig *hotstuff.QuorumSignature, c fuzz.Continue) {
			sigpb := hotstuffpb.QuorumSignature{}
			c.Fuzz(&sigpb)
			*sig = hotstuffpb.QuorumSignatureFromProto(&sigpb)
		},
	)

	errorInfo := new(ErrorInfo)
	errorInfo.Init()

	for i := 0; i < 100; i++ {
		var newMessage any
		f.Fuzz(&newMessage)

		fmt.Printf("%T, %v\n", newMessage, newMessage)
		fuzzScenario(t, errorInfo, newMessage)
	}

	fmt.Printf("unique errors found: %d\n", len(errorInfo.panics))

	for key := range errorInfo.panics {
		fmt.Println()
		fmt.Println(key)
		fmt.Println()
	}

	if len(errorInfo.panics) > 1 {
		panic("many unique errors")
	}

	fmt.Printf("%d of %d runs were errors\n", errorInfo.errorCount, messageCount)
}
