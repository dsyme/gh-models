// Package pex provides a gh command to generate tests for prompts using PromptPex-like functionality.
package pex

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/prompt"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

// NewPexCommand returns a new command for PromptPex-like functionality
func NewPexCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pex",
		Short: "Generate tests for prompts using PromptPex-like functionality",
		Long: heredoc.Docf(`
			Generate unit tests for prompts by automatically extracting output rules and creating
			test cases. This command is inspired by Microsoft's PromptPex tool.

			The pex command analyzes prompts to:
			- Extract output rules (e.g., "output should be JSON")
			- Generate test cases to verify prompt behavior
			- Create evaluators to assess model responses
			- Export tests in the gh models eval format

			This helps developers treat prompts as functions and automatically generate
			test inputs to support unit testing across different AI models.
		`),
		Example: heredoc.Doc(`
			# Generate tests for a prompt text
			gh models pex --prompt "Generate a JSON response with name and age fields"
			
			# Generate tests for a prompt file
			gh models pex --file my_prompt.txt
			
			# Generate tests with specific model
			gh models pex --prompt "Summarize this text" --model openai/gpt-4o
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptText, _ := cmd.Flags().GetString("prompt")
			promptFile, _ := cmd.Flags().GetString("file")
			modelID, _ := cmd.Flags().GetString("model")
			outputFile, _ := cmd.Flags().GetString("output")
			org, _ := cmd.Flags().GetString("org")

			if promptText == "" && promptFile == "" {
				return fmt.Errorf("either --prompt or --file must be specified")
			}

			if promptText != "" && promptFile != "" {
				return fmt.Errorf("cannot specify both --prompt and --file")
			}

			handler := &pexCommandHandler{
				cfg:        cfg,
				client:     cfg.Client,
				promptText: promptText,
				promptFile: promptFile,
				modelID:    modelID,
				outputFile: outputFile,
				org:        org,
			}

			return handler.run(cmd.Context())
		},
	}

	cmd.Flags().String("prompt", "", "Prompt text to analyze")
	cmd.Flags().String("file", "", "File containing prompt text")
	cmd.Flags().String("model", "openai/gpt-4o", "Model to use for test generation")
	cmd.Flags().String("output", "generated_tests.yml", "Output file for generated tests")
	cmd.Flags().String("org", "", "Organization to attribute usage to")

	return cmd
}

type pexCommandHandler struct {
	cfg        *command.Config
	client     azuremodels.Client
	promptText string
	promptFile string
	modelID    string
	outputFile string
	org        string
}

func (h *pexCommandHandler) run(ctx context.Context) error {
	// Load prompt text
	promptText, err := h.loadPromptText()
	if err != nil {
		return fmt.Errorf("failed to load prompt text: %w", err)
	}

	h.cfg.WriteToOut("Analyzing prompt to extract output rules...\n")

	// Extract output rules from prompt
	rules, err := h.extractOutputRules(ctx, promptText)
	if err != nil {
		return fmt.Errorf("failed to extract output rules: %w", err)
	}

	h.cfg.WriteToOut(fmt.Sprintf("Found %d output rules\n", len(rules)))

	// Generate test cases
	h.cfg.WriteToOut("Generating test cases...\n")
	testCases, err := h.generateTestCases(ctx, promptText, rules)
	if err != nil {
		return fmt.Errorf("failed to generate test cases: %w", err)
	}

	h.cfg.WriteToOut(fmt.Sprintf("Generated %d test cases\n", len(testCases)))

	// Create evaluators based on rules
	evaluators := h.createEvaluators(rules)

	// Generate the evaluation file
	evalFile := h.createEvalFile(promptText, testCases, evaluators)

	// Save to file
	err = h.saveEvalFile(evalFile)
	if err != nil {
		return fmt.Errorf("failed to save evaluation file: %w", err)
	}

	h.cfg.WriteToOut(fmt.Sprintf("✓ Generated test file: %s\n", h.outputFile))
	h.cfg.WriteToOut("Run with: gh models eval " + h.outputFile + "\n")

	return nil
}

func (h *pexCommandHandler) loadPromptText() (string, error) {
	if h.promptText != "" {
		return h.promptText, nil
	}

	// Read from file
	content, err := util.ReadFile(h.promptFile)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}

	return string(content), nil
}

func (h *pexCommandHandler) extractOutputRules(ctx context.Context, promptText string) ([]string, error) {
	// Use LLM to extract output rules from the prompt
	analysisPrompt := `Analyze the following prompt and extract any output rules or constraints mentioned.
Look for phrases like:
- "output should be JSON"
- "respond with a list"
- "format as XML"
- "use markdown"
- "include specific fields"
- "follow a certain structure"
- "return only numbers"
- etc.

Return each rule on a separate line, starting with "RULE:". If no rules are found, return "NO_RULES".

Prompt to analyze:
` + promptText

	messages := []azuremodels.ChatMessage{
		{
			Role:    azuremodels.ChatMessageRoleSystem,
			Content: util.Ptr("You are an expert at analyzing prompts and extracting output rules and constraints."),
		},
		{
			Role:    azuremodels.ChatMessageRoleUser,
			Content: util.Ptr(analysisPrompt),
		},
	}

	req := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    h.modelID,
		Stream:   false,
	}

	resp, err := h.client.GetChatCompletionStream(ctx, req, h.org)
	if err != nil {
		return nil, err
	}

	var response strings.Builder
	for {
		completion, err := resp.Reader.Read()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, err
		}

		for _, choice := range completion.Choices {
			if choice.Delta != nil && choice.Delta.Content != nil {
				response.WriteString(*choice.Delta.Content)
			}
			if choice.Message != nil && choice.Message.Content != nil {
				response.WriteString(*choice.Message.Content)
			}
		}
	}

	// Parse rules from response
	var rules []string
	lines := strings.Split(response.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "RULE:") {
			rule := strings.TrimSpace(strings.TrimPrefix(line, "RULE:"))
			if rule != "" {
				rules = append(rules, rule)
			}
		}
	}

	// If no rules found, create some generic ones
	if len(rules) == 0 {
		rules = []string{
			"Response should be coherent and relevant",
			"Response should be in proper language",
		}
	}

	return rules, nil
}

func (h *pexCommandHandler) generateTestCases(ctx context.Context, promptText string, rules []string) ([]map[string]interface{}, error) {
	// Generate test cases that would help verify the rules
	testGenPrompt := `Based on the following prompt and its output rules, generate 3-5 test cases that would help verify the prompt works correctly.

Prompt:
` + promptText + `

Output Rules:
` + strings.Join(rules, "\n") + `

For each test case, provide:
1. An input value that would test the prompt
2. A brief description of what this test case validates

Format each test case as:
INPUT: [input value]
DESCRIPTION: [what this tests]

Generate diverse test cases that cover different scenarios and edge cases.`

	messages := []azuremodels.ChatMessage{
		{
			Role:    azuremodels.ChatMessageRoleSystem,
			Content: util.Ptr("You are an expert at generating test cases for prompts."),
		},
		{
			Role:    azuremodels.ChatMessageRoleUser,
			Content: util.Ptr(testGenPrompt),
		},
	}

	req := azuremodels.ChatCompletionOptions{
		Messages: messages,
		Model:    h.modelID,
		Stream:   false,
	}

	resp, err := h.client.GetChatCompletionStream(ctx, req, h.org)
	if err != nil {
		return nil, err
	}

	var response strings.Builder
	for {
		completion, err := resp.Reader.Read()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, err
		}

		for _, choice := range completion.Choices {
			if choice.Delta != nil && choice.Delta.Content != nil {
				response.WriteString(*choice.Delta.Content)
			}
			if choice.Message != nil && choice.Message.Content != nil {
				response.WriteString(*choice.Message.Content)
			}
		}
	}

	// Parse test cases from response
	var testCases []map[string]interface{}
	lines := strings.Split(response.String(), "\n")
	
	var currentInput string
	var currentDescription string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "INPUT:") {
			currentInput = strings.TrimSpace(strings.TrimPrefix(line, "INPUT:"))
		} else if strings.HasPrefix(line, "DESCRIPTION:") {
			currentDescription = strings.TrimSpace(strings.TrimPrefix(line, "DESCRIPTION:"))
			
			// If we have both input and description, create test case
			if currentInput != "" && currentDescription != "" {
				testCases = append(testCases, map[string]interface{}{
					"input":       currentInput,
					"description": currentDescription,
				})
				currentInput = ""
				currentDescription = ""
			}
		}
	}

	// If no test cases were generated, create a default one
	if len(testCases) == 0 {
		testCases = []map[string]interface{}{
			{
				"input":       "test input",
				"description": "Basic functionality test",
			},
		}
	}

	return testCases, nil
}

func (h *pexCommandHandler) createEvaluators(rules []string) []prompt.Evaluator {
	var evaluators []prompt.Evaluator

	for i, rule := range rules {
		evaluatorName := fmt.Sprintf("rule-%d", i+1)
		
		// Create appropriate evaluator based on rule content
		if h.containsJSONRule(rule) {
			// JSON format rule
			evaluators = append(evaluators, prompt.Evaluator{
				Name: evaluatorName,
				LLM: &prompt.LLMEvaluator{
					ModelID:      h.modelID,
					SystemPrompt: "You are evaluating whether a response follows JSON format rules.",
					Prompt:       "Does this response follow proper JSON format? Response: {{completion}}",
					Choices: []prompt.Choice{
						{Choice: "yes", Score: 1.0},
						{Choice: "no", Score: 0.0},
					},
				},
			})
		} else if h.containsLengthRule(rule) {
			// Length-based rule
			evaluators = append(evaluators, prompt.Evaluator{
				Name: evaluatorName,
				LLM: &prompt.LLMEvaluator{
					ModelID:      h.modelID,
					SystemPrompt: "You are evaluating whether a response meets length requirements.",
					Prompt:       "Does this response meet the length requirements described in: '" + rule + "'? Response: {{completion}}",
					Choices: []prompt.Choice{
						{Choice: "yes", Score: 1.0},
						{Choice: "no", Score: 0.0},
					},
				},
			})
		} else {
			// Generic rule evaluation
			evaluators = append(evaluators, prompt.Evaluator{
				Name: evaluatorName,
				LLM: &prompt.LLMEvaluator{
					ModelID:      h.modelID,
					SystemPrompt: "You are evaluating whether a response follows the specified rule.",
					Prompt:       "Does this response follow the rule: '" + rule + "'? Response: {{completion}}",
					Choices: []prompt.Choice{
						{Choice: "yes", Score: 1.0},
						{Choice: "no", Score: 0.0},
					},
				},
			})
		}
	}

	// Add a coherence evaluator
	evaluators = append(evaluators, prompt.Evaluator{
		Name: "coherence",
		Uses: "github/coherence",
	})

	return evaluators
}

func (h *pexCommandHandler) containsJSONRule(rule string) bool {
	jsonPattern := regexp.MustCompile(`(?i)json|javascript|object|notation`)
	return jsonPattern.MatchString(rule)
}

func (h *pexCommandHandler) containsLengthRule(rule string) bool {
	lengthPattern := regexp.MustCompile(`(?i)length|long|short|brief|concise|characters|words`)
	return lengthPattern.MatchString(rule)
}

func (h *pexCommandHandler) createEvalFile(promptText string, testCases []map[string]interface{}, evaluators []prompt.Evaluator) *prompt.File {
	return &prompt.File{
		Name:        "Generated PromptPex Tests",
		Description: "Automatically generated test cases for prompt validation",
		Model:       h.modelID,
		TestData:    testCases,
		Messages: []prompt.Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant that follows the given instructions precisely.",
			},
			{
				Role:    "user",
				Content: promptText + "\n\nInput: {{input}}",
			},
		},
		Evaluators: evaluators,
	}
}

func (h *pexCommandHandler) saveEvalFile(evalFile *prompt.File) error {
	// Convert to YAML and save
	yamlContent, err := evalFile.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	err = util.WriteFile(h.outputFile, []byte(yamlContent))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}