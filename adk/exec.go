package adk

import (
	"fmt"
	"iter"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// CustomAgent is a simple custom agent implementation that follows ADK patterns.
type CustomAgent struct {
	name        string
	description string
}

// NewCustomAgent creates a new custom agent instance.
func NewCustomAgent() (agent.Agent, error) {
	customAgent := &CustomAgent{
		name:        "CustomAgent",
		description: "A custom ADK agent implementation",
	}

	return agent.New(agent.Config{
		Name:        customAgent.name,
		Description: customAgent.description,
		Run:         customAgent.Run,
	})
}

// Run implements the agent.Agent interface.
// It processes the input from the invocation context and generates a response.
func (a *CustomAgent) Run(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		// Get the input content from the user
		var userInput string

		userContent := ctx.UserContent()
		if userContent != nil && len(userContent.Parts) > 0 {
			if textPart := userContent.Parts[0].Text; textPart != "" {
				userInput = textPart
			}
		}

		// Create a response based on the input
		responseText := fmt.Sprintf("CustomAgent processed: %s", userInput)
		if userInput == "" {
			responseText = "CustomAgent is ready to process your input."
		}

		// Create response content using genai
		responseContent := genai.NewContentFromText(responseText, genai.RoleModel)

		// Create and yield a response event
		event := session.NewEvent(ctx.InvocationID())
		event.LLMResponse.Content = responseContent
		event.Author = a.name

		if !yield(event, nil) {
			return
		}
	}
}
