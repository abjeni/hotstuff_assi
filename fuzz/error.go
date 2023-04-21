package fuzz

import (
	"fmt"
	"strings"
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
