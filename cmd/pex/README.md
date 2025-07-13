# PromptPex Integration (pex command)

The `gh models pex` command integrates PromptPex-like functionality into the GitHub Models CLI extension. This command helps developers treat prompts as functions and automatically generate test inputs to support unit testing.

## Features

- **Automatic Rule Extraction**: Analyzes prompts to extract output rules and constraints
- **Test Case Generation**: Creates diverse test cases to verify prompt behavior
- **Evaluator Creation**: Generates appropriate evaluators based on extracted rules
- **Integration with existing eval command**: Outputs tests in the same format as `gh models eval`

## Usage

### Generate tests from prompt text
```bash
gh models pex --prompt "Generate a JSON response with name and age fields"
```

### Generate tests from a file
```bash
gh models pex --file examples/pex_example.txt
```

### Specify a different model
```bash
gh models pex --prompt "Summarize this text" --model openai/gpt-4o
```

### Custom output file
```bash
gh models pex --prompt "Create a list" --output my_tests.yml
```

## How it works

1. **Rule Extraction**: The command uses an LLM to analyze the prompt and extract output rules like "should be JSON", "must include specific fields", etc.

2. **Test Generation**: Based on the extracted rules, it generates diverse test cases that would help validate the prompt works correctly.

3. **Evaluator Creation**: Creates appropriate evaluators (string-based, LLM-based, or built-in) to assess model responses against the extracted rules.

4. **File Output**: Saves the generated tests in YAML format compatible with `gh models eval`.

## Example workflow

```bash
# Generate tests for a prompt
gh models pex --file examples/pex_example.txt --output json_tests.yml

# Run the generated tests
gh models eval json_tests.yml
```

## Rule Types Supported

- **JSON formatting rules**: Detected by keywords like "JSON", "object", "JavaScript"
- **Length constraints**: Detected by keywords like "brief", "short", "concise", "length"
- **Generic rules**: Any other constraints mentioned in the prompt

## Integration with existing commands

The pex command generates test files that are fully compatible with the existing `gh models eval` command, making it easy to integrate into existing workflows.