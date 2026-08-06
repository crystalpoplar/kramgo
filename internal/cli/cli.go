// Package cli wires together the kramgo command-line interface.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/crystalpoplar/kramgo/internal/ai"
	"github.com/crystalpoplar/kramgo/internal/config"
)

// Version is the application version. It may be overridden at build time via
// -ldflags "-X github.com/crystalpoplar/kramgo/internal/cli.Version=x.y.z".
var Version = "dev"

// Execute builds the root command and runs it.
func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "kramgo",
		Short: "Cross-platform CLI for AI interactions",
		Long: `kramgo is a cross-platform command-line tool for interacting with AI providers
such as OpenAI and Anthropic directly from your terminal.`,
	}

	root.AddCommand(newVersionCmd())
	root.AddCommand(newChatCmd())
	root.AddCommand(newConfigCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of kramgo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "kramgo %s\n", Version)
		},
	}
}

func newChatCmd() *cobra.Command {
	var message string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Send a message to an AI provider and print the response",
		Example: `  kramgo chat --message "What is the capital of France?"
  kramgo chat --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiKey := cfg.APIKey
			if apiKey == "" {
				// Fall back to environment variable.
				envKey := envKeyForProvider(cfg.Provider)
				apiKey = os.Getenv(envKey)
			}
			if apiKey == "" {
				return fmt.Errorf("no API key configured; set api_key in config or the %s environment variable",
					envKeyForProvider(cfg.Provider))
			}

			provider, err := ai.ProviderFactory(cfg.Provider, apiKey)
			if err != nil {
				return err
			}
			client := ai.NewClient(provider)

			if interactive {
				return runInteractive(cmd, client, cfg)
			}

			if message == "" && len(args) > 0 {
				message = strings.Join(args, " ")
			}
			if message == "" {
				return fmt.Errorf("provide a message with --message or run with --interactive")
			}

			reply, err := client.Chat(context.Background(), cfg.Model, message)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), reply)
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Message to send to the AI")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Start an interactive chat session")
	return cmd
}

func runInteractive(cmd *cobra.Command, client *ai.Client, cfg *config.Config) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Starting interactive chat session. Type 'exit' or press Ctrl+C to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(cmd.OutOrStdout(), "> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			break
		}
		reply, err := client.Chat(context.Background(), cfg.Model, input)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
			continue
		}
		fmt.Fprintln(cmd.OutOrStdout(), reply)
	}
	return scanner.Err()
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage kramgo configuration",
	}
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			maskedKey := cfg.APIKey
			if len(maskedKey) > 8 {
				maskedKey = maskedKey[:4] + strings.Repeat("*", len(maskedKey)-8) + maskedKey[len(maskedKey)-4:]
			} else if len(maskedKey) > 0 {
				maskedKey = strings.Repeat("*", len(maskedKey))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "provider: %s\nmodel:    %s\napi_key:  %s\n",
				cfg.Provider, cfg.Model, maskedKey)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value (provider, model, api_key)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			key, value := args[0], args[1]
			switch key {
			case "provider":
				cfg.Provider = value
			case "model":
				cfg.Model = value
			case "api_key":
				cfg.APIKey = value
			default:
				return fmt.Errorf("unknown config key %q; valid keys: provider, model, api_key", key)
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set %s\n", key)
			return nil
		},
	}
}

func envKeyForProvider(provider string) string {
	switch provider {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	default:
		return "OPENAI_API_KEY"
	}
}
