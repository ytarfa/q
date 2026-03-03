package main

import "fmt"

// Message represents a chat message for the LLM API.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// buildMessages constructs the system and user messages for the API request.
func buildMessages(question, context string, limit int) []Message {
	system := buildSystemPrompt(limit)
	user := buildUserMessage(question, context)

	return []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}
}

// buildSystemPrompt constructs the system prompt with optional character limit.
func buildSystemPrompt(limit int) string {
	if limit > 0 {
		return fmt.Sprintf(
			"Answer concisely in under %d characters. Be direct. No preamble. No markdown. Plain text only. If context is provided, use it to inform your answer.",
			limit,
		)
	}
	return "Be direct. No preamble. No markdown. Plain text only. If context is provided, use it to inform your answer."
}

// buildUserMessage constructs the user message with optional context.
func buildUserMessage(question, context string) string {
	if context == "" {
		return question
	}
	return fmt.Sprintf("Context:\n---\n%s\n---\n\n%s", context, question)
}
