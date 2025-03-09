# Ollama Model Chat Template Validation Tool
- Test Ollama Chat Templates without rebuilding or redeploying the model.
- Validate tool call parsing in Ollama.
- Test Jinja2 chat templates (Web Version only).
- Compare differences between rendered templates (Web Version only).

Rendered templates can be used with the Ollama `/api/generate` endpoint with the `raw` flag enabled to avoid model rebuild/redeploy/restart.

Web Version is deployed here: https://eugene-kamenev.github.io/ollama-template-test/

## Local Build
```bash
make clean && make all
```

## Local Run Web Version
Start nginx
```bash
docker compose up -d
```
Visit http://localhost:9090. Note that Web Version also supports Jinja2 templates from tokenizer_config.json

## Run Native Executable
```bash
TEMPLATE=$(cat << 'EOF'
{{- if or .System .Tools }}<|system|>{{ if .System }}{{ .System }}{{ end }}
{{- if .Tools }}
{{- if not .System }}You may call one or more functions to assist with the user query. You are provided with function signatures.
{{- end }}<|tool|>[{{- range .Tools }}{{ .Function }}{{ end }}]<|/tool|><|end|>
{{- end }}
{{- end }}
{{- range $i, $_ := .Messages }}
{{- $last := eq (len (slice $.Messages $i)) 1 -}}
{{- if ne .Role "system" }}<|{{ .Role }}|>{{ .Content }}
{{- if .ToolCalls }}<|tool_call|>[{{ range .ToolCalls }}{"name":"{{ .Function.Name }}","arguments":{{ .Function.Arguments }}}{{ end }}]<|/tool_call|>
{{- end }}
{{- if not $last }}<|end|>
{{- end }}
{{- if and (ne .Role "assistant") $last }}<|end|><|assistant|>{{ end }}
{{- end }}
{{- end }}
EOF
)

MESSAGES=$(cat <<EOF
{
    "messages": [
        {
            "role": "user",
            "content": "What is the square root of 475695037565? And what is the sum of 44.101523499 and 500.213455?"
        },
        {
            "role": "assistant",
            "content":"",
            "tool_calls": [
                {
                    "function": {
                        "index": 0,
                        "name": "getSquareRoot",
                        "arguments": {
                            "x": 475695037565
                        }
                    }
                },
                {
                    "function": {
                        "index": 1,
                        "name": "getSum",
                        "arguments": {
                            "x": 44.101523499,
                            "y": 500.213455
                        }
                    }
                }
            ]
        },
        {
            "role": "tool",
            "content": "689706.4865324959"
        },
        {
            "role": "tool",
            "content": "544.314978499"
        }
    ],
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "getSquareRoot",
                "description": "Returns square root of a number",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "x": {
                            "type": "number"
                        }
                    },
                    "required": [
                        "x"
                    ]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "getSum",
                "description": "Returns sum of two numbers",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "y": {
                            "type": "number"
                        },
                        "x": {
                            "type": "number"
                        }
                    },
                    "required": [
                        "x",
                        "y"
                    ]
                }
            }
        }
    ]
}
EOF
)  
./bin/ollama-template-test-native "ValidateToolCalls" "$MESSAGES" "$TEMPLATE"
# or
./bin/ollama-template-test-native "ValidateTemplate" "$MESSAGES" "$TEMPLATE"
```
