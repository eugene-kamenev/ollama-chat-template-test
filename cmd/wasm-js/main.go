package main

import (
	"fmt"
	"syscall/js"
	"templatetest/cmd/ollama"
)

func ValidateTemplate() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		result := ollama.Result{}
		if len(args) != 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("expected 2 arguments, got %d", len(args)))
			return ollama.ToJSON(result, result)
		}

		inputConv := args[0].String()
		inputTmpl := args[1].String()

		result = ollama.ParseModelTemplate(inputConv, inputTmpl)
		return ollama.ToJSON(result, result)
	})
	return jsonFunc
}

func ValidateToolCalls() js.Func {
	jsonFunc := js.FuncOf((func(this js.Value, args[]js.Value) any {
		result := ollama.Result{}
		if len(args) != 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("expected 2 arguments, got %d", len(args)))
			return ollama.ToJSON(result, result)
		}

		inputConv := args[0].String()
		inputTmpl := args[1].String()

		result = ollama.ParseToolCalls(inputTmpl, inputConv)
		return ollama.ToJSON(result, result)
	}))
	return jsonFunc
}
 
func main() {
	js.Global().Set("ValidateTemplate", ValidateTemplate())
	js.Global().Set("ValidateToolCalls", ValidateToolCalls())
	<-make(chan struct{})
}
