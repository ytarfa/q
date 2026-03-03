package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runInit handles the `q init` subcommand.
func runInit() int {
	path := configPath()
	if path == "" {
		fmt.Fprintln(os.Stderr, "Error: could not determine home directory")
		return 2
	}

	// Check if config already exists
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Config file already exists at %s\n", path)
		fmt.Print("Overwrite? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return 0
		}
	}

	reader := bufio.NewReader(os.Stdin)

	// Provider type
	fmt.Print("Provider (openai/ollama) [openai]: ")
	providerType, _ := reader.ReadString('\n')
	providerType = strings.TrimSpace(providerType)
	if providerType == "" {
		providerType = "openai"
	}
	if providerType != "openai" && providerType != "ollama" {
		fmt.Fprintf(os.Stderr, "Error: unknown provider type: %s\n", providerType)
		return 2
	}

	// Model
	defaultModel := "gpt-4o-mini"
	if providerType == "ollama" {
		defaultModel = "llama3"
	}
	fmt.Printf("Model [%s]: ", defaultModel)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}

	// API key (only for openai)
	apiKey := ""
	if providerType == "openai" {
		fmt.Print("API key: ")
		apiKey, _ = reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: API key is required for OpenAI provider")
			return 2
		}
	}

	// Build config
	cfg := Config{
		Limit: defaultLimit,
		Provider: ProviderConfig{
			Type:   providerType,
			Model:  model,
			APIKey: apiKey,
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not serialize config: %v\n", err)
		return 2
	}

	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create config directory: %v\n", err)
		return 2
	}

	// Write file with 0600 permissions
	if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not write config file: %v\n", err)
		return 2
	}

	fmt.Printf("Config written to %s\n", path)
	return 0
}
