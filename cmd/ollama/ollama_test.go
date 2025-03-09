package ollama

import (
	"testing"
)

func TestParseModelTemplate(t *testing.T) {
	inputConv := `{
		"messages": [
			{"role": "user", "content": "Hello!"}
		],
		"tools": [
			{
				"id": "test_tool",
				"name": "Test Tool"
			}
		]
	}`

	inputTmpl := `
	{{range .Messages}}
	{{.Content}}
	{{end}}
	`

	result := ParseModelTemplate(inputConv, inputTmpl)
	if len(result.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", result.Errors)
		return
	}
	if len(result.Debug) == 0 {
		t.Errorf("Missing response")
		return
	}
}

func TestParseToolCalls(t *testing.T) {
	inputTmpl := `
	{{range .ToolCalls}}
	{
		"name": "{{.Function.Name}}",
		"arguments": {{json .Function.Arguments}}
	}
	{{end}}
	`

	response := `{"name": "test", "arguments": {"key": "value"}}`

	result := ParseToolCalls(inputTmpl, response)
	if len(result.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", result.Errors)
		return
	}
	if len(result.Debug) == 0 {
		t.Errorf("Missing response")
		return
	}
}

func TestToJSON(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

 testData := testStruct{
	 Name:  "test",
	 Value: 123,
 }

 result := Result{}
 jsonStr := ToJSON(result, testData)
 if len(result.Errors) > 0 {
	 t.Errorf("Unexpected errors: %v", result.Errors)
 }
 if jsonStr == "" {
	 t.Error("Expected non-empty JSON string")
 }
}

func TestMain(m *testing.M) {
	m.Run()
}