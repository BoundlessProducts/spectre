package mine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const anthropicModel = "claude-haiku-4-5-20251001"

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// EnhanceWithLLM sends the mined spec to the Anthropic API and returns
// suggestions for action notes, additional invariants, and temporal properties.
// Returns nil, nil if the API key is empty (graceful no-op).
func EnhanceWithLLM(spec *MinedSpec, apiKey string) (*LLMSuggestions, error) {
	if apiKey == "" {
		return nil, nil
	}

	prompt := buildPrompt(spec)

	reqBody := anthropicRequest{
		Model:     anthropicModel,
		MaxTokens: 1024,
		System:    systemPrompt(),
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var ar anthropicResponse
	if err := json.Unmarshal(rawBody, &ar); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if ar.Error != nil {
		return nil, fmt.Errorf("API error: %s", ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	text := ar.Content[0].Text
	suggestions := parseLLMResponse(text)
	suggestions.Raw = text
	return suggestions, nil
}

func systemPrompt() string {
	return `You are an expert in formal specification and Rust software engineering.
You analyze Rust source code and suggest temporal properties, invariants, and
action descriptions for Spectre formal specifications.

Spectre spec syntax reference:
  invariant name { condition }          // safety property
  temporal name { eventually (expr) }   // liveness: expr must become true
  temporal name { always (expr) }       // safety via temporal: expr must always hold
  temporal name { WF(actionName) }      // weak fairness: action taken when enabled
  temporal name { SF(actionName) }      // strong fairness

In conditions: use = for equality (not ==), use field names directly (no self.).
Enum comparison: FieldName = EnumType.Variant

Reply with JSON only. No prose outside the JSON object.`
}

func buildPrompt(spec *MinedSpec) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Analyze this Rust struct/impl for formal spec mining.\n\n"))
	b.WriteString(fmt.Sprintf("Spec name: %s\n", spec.SpecName))

	if len(spec.Enums) > 0 {
		b.WriteString("\nEnums:\n")
		for _, e := range spec.Enums {
			b.WriteString(fmt.Sprintf("  %s { %s }\n", e.Name, strings.Join(e.Variants, ", ")))
		}
	}

	if len(spec.Fields) > 0 {
		b.WriteString("\nState fields:\n")
		for _, f := range spec.Fields {
			b.WriteString(fmt.Sprintf("  %s: %s (Rust: %s)\n", f.Name, f.SpectreType, f.RustType))
		}
	}

	if len(spec.Methods) > 0 {
		b.WriteString("\nMethods:\n")
		for _, meth := range spec.Methods {
			b.WriteString(fmt.Sprintf("  %s", meth.Name))
			if len(meth.Params) > 0 {
				ps := make([]string, len(meth.Params))
				for i, p := range meth.Params {
					ps[i] = p.Name + ": " + p.SpectreType
				}
				b.WriteString("(" + strings.Join(ps, ", ") + ")")
			}
			b.WriteString("\n")
			for _, r := range meth.Requires {
				b.WriteString(fmt.Sprintf("    assert: %s\n", r))
			}
			for _, a := range meth.Assigns {
				b.WriteString(fmt.Sprintf("    assigns: %s %s %s\n", a.Field, a.Op, a.RHSRust))
			}
		}
	}

	b.WriteString(`
Based on this analysis, suggest:
1. A one-sentence description for each action (actionNotes)
2. Up to 3 invariants that should always hold (extraInvariants)
3. Up to 3 temporal properties that reflect progress or liveness (extraTemporal)

Reply with this exact JSON structure:
{
  "actionNotes": { "methodName": "one-sentence description" },
  "extraInvariants": [
    { "name": "snake_case_name", "condition": "spectre_expr", "description": "why it matters" }
  ],
  "extraTemporal": [
    { "name": "snake_case_name", "expression": "eventually (...)", "description": "what it means" }
  ]
}`)

	return b.String()
}

// parseLLMResponse extracts suggestions from the LLM's JSON response.
// On any parse error, returns whatever was successfully extracted.
func parseLLMResponse(text string) *LLMSuggestions {
	sugg := &LLMSuggestions{ActionNotes: make(map[string]string)}

	// Find JSON object in response (model might wrap it in prose).
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return sugg
	}
	jsonStr := text[start : end+1]

	var raw struct {
		ActionNotes     map[string]string `json:"actionNotes"`
		ExtraInvariants []struct {
			Name        string `json:"name"`
			Condition   string `json:"condition"`
			Description string `json:"description"`
		} `json:"extraInvariants"`
		ExtraTemporal []struct {
			Name        string `json:"name"`
			Expression  string `json:"expression"`
			Description string `json:"description"`
		} `json:"extraTemporal"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return sugg
	}

	sugg.ActionNotes = raw.ActionNotes
	if sugg.ActionNotes == nil {
		sugg.ActionNotes = make(map[string]string)
	}

	for _, inv := range raw.ExtraInvariants {
		sugg.ExtraInvariants = append(sugg.ExtraInvariants, LLMInvariant{
			Name:        inv.Name,
			Condition:   inv.Condition,
			Description: inv.Description,
		})
	}

	for _, tp := range raw.ExtraTemporal {
		sugg.ExtraTemporal = append(sugg.ExtraTemporal, LLMTemporal{
			Name:        tp.Name,
			Expression:  tp.Expression,
			Description: tp.Description,
		})
	}

	return sugg
}
