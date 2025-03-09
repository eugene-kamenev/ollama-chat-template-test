package main

import (
	"templatetest/cmd/ollama"
)

//go:wasmexport ValidateTemplate
func ValidateTemplate(template, chat string) uintptr {
	return ResultToPtr(ollama.ParseModelTemplate(chat, template))
}

//go:wasmexport ValidateToolCalls
func ValidateToolCalls(template, llmMessage string) uintptr {
	return ResultToPtr(ollama.ParseToolCalls(template, llmMessage))
}

func ResultToPtr(result ollama.Result) uintptr {
	processedJson := ollama.ToJSON(result, result)
	return WriteString(string(processedJson))
}

func main() {
}
