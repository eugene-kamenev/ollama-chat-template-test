const TEMPLATE_JINJA = `
{% for message in messages -%}
    {% if message['role'] == 'system' and 'tools' in message and message['tools'] is not none -%}
        {{ '<|' + message['role'] + '|>' + message['content'] + '<|tool|>' + message['tools'] + '<|/tool|>' + '<|end|>' }}
    {%- else -%}
        {{ '<|' + message['role'] + '|>' + message['content'] + '<|end|>' }}
    {%- endif %}
{%- endfor %}
{%- if add_generation_prompt -%}
    {{ '<|assistant|>' }}
{% else -%}
    {{ eos_token }}
{%- endif %}`;

const TEMPLATE_OLLAMA = `
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
`;

const LLM_OUTPUT = `
Here you can test if Ollama will respond with tool_calls when you use chat endpoint.

\`\`\`json
{"name": "found", "arguments": {"key": "value"}}}, // # valid
{"name": "found_2", "arguments": {"key": "value"}}},
{"function": "not_found", "arguments": {"key": "value"}}},
\`\`\`
{"function": "not_found_1", "arguments": {"key": "value"}}},
{"name": "not_found_2", "arguments": {"key": "value"}
{"name": "not_found_3", "parameters": {"key": "value"}}
{{"name": "not_found_4", "arguments": {"key": "value"}
`;

const TOOLS = [
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
];

const MESSAGES = [
    {
        "role": "system",
        "content": "You are a very helpful AI assistant with tools."
    },
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
];

const CHATHISTORY_OLLAMA = 
{
    messages: MESSAGES,
    tools: TOOLS
};

const JINJA_SCRIPT = `
# https://github.com/microsoft/PhiCookBook/blob/main/md/02.Application/07.FunctionCalling/Phi4/FunctionCallingBasic/README.md
messages = data['messages']
messages[0]['tools'] = json.dumps(data['tools'])
data['add_generation_prompt'] = True
data['eos_token'] = "<|endoftext|>"
`;

const app = Vue.createApp({
  data() {
    return {
      activeTab: "diff",
      tabs: [
        { id: "diff", name: "Template Render Difference (Ollama vs Jinja2)" },
        { id: "ollama", name: "Ollama Model Chat Template Test" },
        { id: "jinja", name: "Jinja2 Chat Template Test" },
        { id: "tool", name: "Ollama Tool Call Check" },
        { id: "about", name: "About" } // new tab added
      ],
      outline: 'outline',
      contrast: 'contrast',
      ollamaTemplate: TEMPLATE_OLLAMA.trim(),
      chatHistory: JSON.stringify(CHATHISTORY_OLLAMA, null, 2),
      jinjaTemplate: TEMPLATE_JINJA.trim(),
      preRenderScript: JINJA_SCRIPT.trim(),
      ollamaOutput: "",
      jinjaOutput: "",
      outputType: "default",
      pyodide: null,
      llmOutput: LLM_OUTPUT.trim(),
      toolCheckOutput: "",
      loadingGeneral: true,
      loadingAction: false,
      diffViewOptions: {
        foldingStrategy: 'indentation',
        automaticLayout: true,
        diffCodeLens: true,
        enableSplitViewResizing: true, // updated to allow resizing
        renderSideBySide: false,
        wordWrap: true,
        experimental: {
          showMoves: true,
          useTrueInlineView: true
        }
      }
    }
  },
  methods: {
    async callValidateModelTemplate() {
      this.loadingAction = true;
      try {
        const result = await ValidateTemplate(this.chatHistory, this.ollamaTemplate);
        this.ollamaOutput = this.getOutput(result);
      } finally {
        this.loadingAction = false;
      }
    },
    async callValidateJinjaTemplate() {
      this.loadingAction = true;
      try {
        let pythonScript = `
import json
from jinja2 import Template
template = Template("""${this.jinjaTemplate}""")
data = json.loads("""${this.chatHistory}""")
${this.preRenderScript}
rendered = template.render(data)
rendered
`;
        const output = await this.pyodide.runPythonAsync(pythonScript)
        if (this.outputType === "escaped") {
            this.jinjaOutput = JSON.stringify([output]).slice(2, -2);
        } else {
            this.jinjaOutput = output;
        }
      } catch (err) {
        this.jinjaOutput = "Error rendering template: " + err;
      } finally {
        this.loadingAction = false;
      }
    },
    async callBothTemplates() {
      this.loadingAction = true;
      try {
        await this.callValidateModelTemplate();
        await this.callValidateJinjaTemplate();
      } finally {
        this.loadingAction = false;
      }
    },
    getOutput(input) {
      const outputObj = JSON.parse(input);
      if (outputObj.errors) {
        return outputObj.errors[0];
      }
      if (outputObj.debug) {
        const message = outputObj.debug.slice(-1)[0];
        if (this.outputType === "escaped") {
          return JSON.stringify([message]).slice(2, -2);
        }
        return message;
      }
      return "No output";
    },
    async setPyodide() {
        await loadPyodide().then((e) => {
            this.pyodide = e;
            return e.loadPackage("jinja2");
        });
    },
    async callToolCheck() {
      this.loadingAction = true;
      try {
        const result = await ValidateToolCalls(this.llmOutput, this.ollamaTemplate);
        this.toolCheckOutput = this.getOutput(result);
      } finally {
        this.loadingAction = false;
      }
    }
  },
  mounted() {
    if (WebAssembly) {
      if (WebAssembly && !WebAssembly.instantiateStreaming) {
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
          const source = await (await resp).arrayBuffer();
          return await WebAssembly.instantiate(source, importObject);
        };
      }
      const go = new Go();
      WebAssembly.instantiateStreaming(
        fetch("ollama-template-test-wasm-js.wasm"),
        go.importObject
      ).then((result) => {
        go.run(result.instance);
      });
    } else {
      alert("WebAssembly is not supported in your browser");
    }
    this.setPyodide().then(() => {
      this.loadingGeneral = false;
    });
  }
}).component('collapsible-textarea', {
  template: '#collapsible-textarea-template',
  props: {
    id: String,
    title: String,
    label: String,
    placeholder: String,
    modelValue: String,
    language: { type: String, default: 'text' },
    rows: { type: Number, default: 10 },
    cols: { type: Number, default: 50 }
  },
  computed: {
    internalValue: {
      get() { return this.modelValue },
      set(val) { this.$emit('update:modelValue', val); }
    }
  },
  data() {
    return {
      editorOptions: {
        automaticLayout: true,
        theme: "vs-dark"
      }
    }
  },
  methods: {
    handleToggle(event) {
      // When this accordion opens, close all others
      if (event.target.open) {
        document.querySelectorAll('.collapsible-editor').forEach(el => {
          if (el !== event.target) {
            el.removeAttribute('open');
          }
        });
      }
    }
  }
}).component('output-display', {
  template: '#output-display-template',
  props: {
    content: { type: String, default: '' },
    title: { type: String, defult: '' }
  },
  methods: {
    copyContent() {
      navigator.clipboard.writeText(this.content);
    }
  }
}).component('ollama-template', {
  props: {
    modelValue: String,
    textareaId: { type: String, default: 'template' }
  },
  template: '#ollama-template-template'
}).component('chat-history', {
  props: {
    modelValue: String,
    textareaId: { type: String, default: 'conversation' }
  },
  template: '#chat-history-template'
}).component('jinja-template', {
  props: {
    modelValue: String,
    textareaId: { type: String, default: 'jinja-template' }
  },
  template: '#jinja-template-template'
}).component('jinja-script', {
  props: {
    modelValue: String,
    textareaId: { type: String, default: 'jinja-script' }
  },
  template: '#jinja-script-template'
}).component('llm-output-textarea', {
  props: {
    modelValue: String,
    textareaId: { type: String, default: 'llm_output' }
  },
  template: '#llm-output-textarea-template'
}).use(monaco_vue, {
  paths: {
    vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.52.2/min/vs'
  }
});
app.mount("#app");
