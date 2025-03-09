package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"text/template/parse"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/template"
)

type InputData struct {
	Messages []api.Message `json:"messages"`
	Tools    []api.Tool    `json:"tools,omitempty"`
}

type Result struct {
	Debug 	[]string `json:"debug,omitempty"`
	Errors 	[]string `json:"errors,omitempty"`
}

func ToJSON(result Result, v any) string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		result.Errors = append(result.Errors, "Error converting to JSON: " + err.Error())
		return ""
	}
	return buf.String()
}

// originally from ollama server.Model
func ParseObjects(s string) []map[string]any {
	var objs []map[string]any
	for offset := 0; offset < len(s); {
		var obj map[string]any
		decoder := json.NewDecoder(strings.NewReader(s[offset:]))
		if err := decoder.Decode(&obj); errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		} else if syntax := &(json.SyntaxError{}); errors.As(err, &syntax) {
			// skip over any syntax errors
			offset += int(syntax.Offset)
		} else if unmarshalType := &(json.UnmarshalTypeError{}); errors.As(err, &unmarshalType) {
			// skip over any unmarshalable types
			offset += int(unmarshalType.Offset)
		} else if err != nil {
			return nil
		} else {
			offset += int(decoder.InputOffset())
			objs = append(objs, obj)
		}
	}

	return objs
}

// originally from ollama server.Model
func ParseToolCalls(inputTmpl, s string) Result {
	result := Result{}
	input, err := template.Parse(inputTmpl)
	if err != nil {
		result.Errors = append(result.Errors, "Error parsing template: " + err.Error())
		return result
	}

	// Create a subtree from the node that ranges over .ToolCalls
	tmpl := input.Subtree(func(n parse.Node) bool {
		if t, ok := n.(*parse.RangeNode); ok {
			return slices.Contains(template.Identifiers(t.Pipe), "ToolCalls")
		}
		return false
	})

	if tmpl == nil {
		result.Errors = append(result.Errors, "Error: RangeNode 'ToolCalls' was not found in the model template.")
		return result
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, map[string][]api.ToolCall{
		"ToolCalls": {
			{
				Function: api.ToolCallFunction{
					Name: "@@name@@",
					Arguments: api.ToolCallFunctionArguments{
						"@@argument@@": 1,
					},
				},
			},
		},
	}); err != nil {
		result.Errors = append(result.Errors, "Failed to find matches for ToolCall. " + err.Error())
		return result
	}

	templateObjects := ParseObjects(b.String())
	if len(templateObjects) == 0 {
		result.Errors = append(result.Errors, "Error: Unable to extract template objects from the model. " + b.String())
		return result
	}

	// Find the keys that correspond to the name and arguments fields
	var name, arguments string
	for k, v := range templateObjects[0] {
		switch v.(type) {
		case string:
			name = k
		case map[string]any:
			arguments = k
		}
	}

	if name == "" || arguments == "" {
		result.Errors = append(result.Errors, "Error: Could not determine 'name' or 'arguments' fields in the model template. " + b.String())
		return result;	 
	}

	responseObjects := ParseObjects(s)
	if len(responseObjects) == 0 {
		result.Errors = append(result.Errors, "Error: No valid JSON objects found in the input string. " + b.String())
		return result;
	}

	// Collect all nested objects
	var collect func(any) []map[string]any
	collect = func(obj any) (all []map[string]any) {
		switch o := obj.(type) {
		case map[string]any:
			all = append(all, o)
			for _, v := range o {
				all = append(all, collect(v)...)
			}
		case []any:
			for _, v := range o {
				all = append(all, collect(v)...)
			}
		}
		return all
	}

	var objs []map[string]any
	for _, p := range responseObjects {
		objs = append(objs, collect(p)...)
	}

	var toolCalls []api.ToolCall
	for _, kv := range objs {
		n, nok := kv[name].(string)
		a, aok := kv[arguments].(map[string]any)
		if nok && aok {
			toolCalls = append(toolCalls, api.ToolCall{
				Function: api.ToolCallFunction{
					Name:      n,
					Arguments: a,
				},
			})
		}
	}
	
	if len(toolCalls) > 0 {
		parsedToolsJSON := ToJSON(result, toolCalls)
		result.Debug = append(result.Debug, parsedToolsJSON)
	} else {
		result.Errors = append(result.Errors, "No Tool Calls has been found")
	}
	return result;
}

func ParseModelTemplate(inputConv, inputTmpl string) Result {
	result := Result{}
	var inputData InputData
	if err := json.Unmarshal([]byte(inputConv), &inputData); err != nil {
		result.Errors = append(result.Errors, "Error parsing JSON: " + err.Error())
		return result
	}

	tmpl, err := template.Parse(inputTmpl)
	if err != nil {
		result.Errors = append(result.Errors, "Error parsing template: " + err.Error())
		return result
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, template.Values{
		Messages: inputData.Messages,
		Tools:    inputData.Tools,
	})
	if err != nil {
		result.Errors = append(result.Errors, "Failed to execute template: " + err.Error())
		return result
	}
	result.Debug = append(result.Debug, buf.String())
	return result
}
