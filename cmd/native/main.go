package main

import (
	"fmt"
	"os"

	"templatetest/cmd/ollama"
)

func printOut(arr []string) {
	if arr == nil {
		return
	}
	for _, v := range arr {
		fmt.Println(v)
	}
}

func output(r ollama.Result) {
	printOut(r.Errors)
	printOut(r.Debug)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: ./ollama-template-test-native <ValidateTemplate|ValidateToolCalls> <conversation_json|llm_response> <template_string>")
		os.Exit(1)
	}

	function := os.Args[1]
	inputConv := os.Args[2]
	inputTmpl := os.Args[3]

	switch function {
		case "ValidateTemplate":
			output(ollama.ParseModelTemplate(inputConv, inputTmpl))
		case "ValidateToolCalls":
			output(ollama.ParseToolCalls(inputTmpl, inputConv))
		default:
			fmt.Println("Valid function should be <ValidateTemplate|ValidateToolCalls>")
			os.Exit(1)
	}
}
