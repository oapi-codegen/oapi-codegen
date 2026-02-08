package output

import (
	"encoding/json"
	"testing"
)

// TestPromptOneOfHasPromptField verifies that the 'prompt' field is not lost
// in nested allOf/oneOf structures.
// https://github.com/oapi-codegen/oapi-codegen/issues/1710
func TestPromptOneOfHasPromptField(t *testing.T) {
	// Test ChatPrompt variant (PromptOneOf0) - prompt is []ChatMessage
	chatType := "chat"
	chatPrompt := PromptOneOf0{
		Type: &chatType,
		Prompt: []ChatMessage{
			{Role: "user", Content: "hello"},
		},
	}

	if chatPrompt.Type == nil || *chatPrompt.Type != "chat" {
		t.Error("ChatPrompt variant should have Type='chat'")
	}
	if len(chatPrompt.Prompt) != 1 {
		t.Errorf("ChatPrompt.Prompt should have 1 message, got %d", len(chatPrompt.Prompt))
	}
	if chatPrompt.Prompt[0].Role != "user" {
		t.Errorf("ChatPrompt.Prompt[0].Role = %q, want %q", chatPrompt.Prompt[0].Role, "user")
	}

	// Test TextPrompt variant (PromptOneOf1) - prompt is string
	textType := "text"
	textPrompt := PromptOneOf1{
		Type:   &textType,
		Prompt: "Hello, world!",
	}

	if textPrompt.Type == nil || *textPrompt.Type != "text" {
		t.Error("TextPrompt variant should have Type='text'")
	}
	if textPrompt.Prompt != "Hello, world!" {
		t.Errorf("TextPrompt.Prompt = %q, want %q", textPrompt.Prompt, "Hello, world!")
	}
}

func TestPromptJSONRoundTrip(t *testing.T) {
	// Test chat prompt variant
	chatType := "chat"
	chatVariant := Prompt{
		PromptOneOf0: &PromptOneOf0{
			Type: &chatType,
			Prompt: []ChatMessage{
				{Role: "user", Content: "test message"},
			},
		},
	}

	data, err := json.Marshal(chatVariant)
	if err != nil {
		t.Fatalf("Marshal chat variant failed: %v", err)
	}

	var decoded Prompt
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal chat variant failed: %v", err)
	}

	if decoded.PromptOneOf0 == nil {
		t.Fatal("Expected PromptOneOf0 to be set after unmarshal")
	}
	if len(decoded.PromptOneOf0.Prompt) != 1 {
		t.Errorf("Expected 1 message, got %d", len(decoded.PromptOneOf0.Prompt))
	}
}

func TestTextPromptHasPromptField(t *testing.T) {
	// Verify TextPrompt (from allOf) has the prompt field
	tp := TextPrompt{
		Prompt:  "my prompt",
		Name:    "test",
		Version: 1,
	}

	if tp.Prompt != "my prompt" {
		t.Errorf("TextPrompt.Prompt = %q, want %q", tp.Prompt, "my prompt")
	}
}

func TestChatPromptHasPromptField(t *testing.T) {
	// Verify ChatPrompt (from allOf) has the prompt field
	cp := ChatPrompt{
		Prompt: []ChatMessage{
			{Role: "assistant", Content: "hello"},
		},
		Name:    "test",
		Version: 1,
	}

	if len(cp.Prompt) != 1 {
		t.Errorf("ChatPrompt.Prompt should have 1 message, got %d", len(cp.Prompt))
	}
}
