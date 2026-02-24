package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sivepanda/mycelia"
	internaltui "github.com/sivepanda/mycelia/internal/tui"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command tree. The binary name is derived from
// os.Args[0] so both "mycelia" and "myc" work identically.
func NewRootCmd() *cobra.Command {
	name := filepath.Base(os.Args[0])

	rootCmd := &cobra.Command{
		Use:   name,
		Short: "Mycelia manages housekeeping commands for your project",
		Long:  `Mycelia detects package managers and manages housekeeping commands that run after git operations like pull and checkout.`,
		Run: func(cmd *cobra.Command, args []string) {
			m := internaltui.NewMenu()
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
				os.Exit(1)
			}
		},
	}

	detectCmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect package managers and build systems",
		Run: func(cmd *cobra.Command, args []string) {
			detector := mycelia.NewDetector(".")
			detected, err := detector.DetectPackages()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error detecting packages: %v\n", err)
				os.Exit(1)
			}

			if len(detected) == 0 {
				fmt.Println("No package managers or build systems detected.")
				return
			}

			fmt.Println("Detected package managers and build systems:")
			for _, pkg := range detected {
				fmt.Printf("  • %s (%s)\n", pkg.Type.Description, pkg.Path)
			}
		},
	}

	suggestCmd := &cobra.Command{
		Use:   "suggest <category>",
		Short: "Show suggested commands for a category",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			category := args[0]

			if category != "post-pull" && category != "post-checkout" {
				fmt.Fprintf(os.Stderr, "Error: Invalid category '%s'. Must be 'post-pull' or 'post-checkout'\n", category)
				os.Exit(1)
			}

			detector := mycelia.NewDetector(".")
			suggestions, err := detector.GetSuggestedCommands(category)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting suggestions: %v\n", err)
				os.Exit(1)
			}

			if len(suggestions) == 0 {
				fmt.Printf("No suggestions for %s commands.\n", category)
				return
			}

			fmt.Printf("Suggested %s commands:\n", category)
			for i, suggestion := range suggestions {
				fmt.Printf("  %d. %s\n", i+1, suggestion.Description)
				fmt.Printf("     Command: %s\n", suggestion.Command)
			}

			fmt.Println("\nTo add these commands, use:")
			fmt.Printf("  %s auto %s\n", name, category)
		},
	}

	autoCmd := &cobra.Command{
		Use:   "auto <category>",
		Short: "Auto-detect and add commands to config",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			category := args[0]

			if category != "post-pull" && category != "post-checkout" {
				fmt.Fprintf(os.Stderr, "Error: Invalid category '%s'. Must be 'post-pull' or 'post-checkout'\n", category)
				os.Exit(1)
			}

			detector := mycelia.NewDetector(".")
			detected, err := detector.DetectPackages()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error detecting packages: %v\n", err)
				os.Exit(1)
			}

			if len(detected) == 0 {
				fmt.Println("No package managers detected.")
				return
			}

			config, err := mycelia.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			count := 0
			for _, pkg := range detected {
				var triggerFiles []string
				if pkg.Type.DetectFile != "" {
					triggerFiles = append(triggerFiles, pkg.Type.DetectFile)
				}
				triggerFiles = append(triggerFiles, pkg.Type.DetectFiles...)

				commands, exists := pkg.Type.Commands[category]
				if !exists {
					continue
				}
				for _, suggestion := range commands {
					fmt.Printf("  • %s\n", suggestion.Description)
					if err := config.AddCommand(category, mycelia.Command{
						Command:      suggestion.Command,
						WorkingDir:   suggestion.WorkingDir,
						Description:  suggestion.Description,
						TriggerFiles: triggerFiles,
					}); err != nil {
						fmt.Fprintf(os.Stderr, "Error adding command: %v\n", err)
						os.Exit(1)
					}
					count++
				}
			}

			if count == 0 {
				fmt.Printf("No suggestions for %s commands.\n", category)
				return
			}

			fmt.Printf("Adding %d suggested %s commands:\n", count, category)

			if err := config.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\nSuccessfully added %d %s commands!\n", count, category)
		},
	}

	addCmd := &cobra.Command{
		Use:   "add <command>",
		Short: "Manually add a command",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			command := args[0]

			postPull, _ := cmd.Flags().GetBool("post-pull")
			postCheckout, _ := cmd.Flags().GetBool("post-checkout")
			workingDir, _ := cmd.Flags().GetString("working-dir")
			description, _ := cmd.Flags().GetString("description")

			if !postPull && !postCheckout {
				fmt.Fprintln(os.Stderr, "Error: Must specify either --post-pull or --post-checkout")
				os.Exit(1)
			}

			if postPull && postCheckout {
				fmt.Fprintln(os.Stderr, "Error: Cannot specify both --post-pull and --post-checkout")
				os.Exit(1)
			}

			config, err := mycelia.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			if workingDir == "" {
				workingDir = "."
			}

			if description == "" {
				description = command
			}

			var category string
			if postPull {
				category = "post-pull"
			} else {
				category = "post-checkout"
			}

			if err := config.AddCommand(category, mycelia.Command{
				Command:     command,
				WorkingDir:  workingDir,
				Description: description,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding command: %v\n", err)
				os.Exit(1)
			}

			if err := config.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Added %s command: %s\n", category, command)
		},
	}
	addCmd.Flags().Bool("post-pull", false, "Add command to post-pull category")
	addCmd.Flags().Bool("post-checkout", false, "Add command to post-checkout category")
	addCmd.Flags().StringP("working-dir", "d", ".", "Working directory for the command")
	addCmd.Flags().StringP("description", "m", "", "Description of the command")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured commands",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := mycelia.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			categories := []string{"post-pull", "post-checkout"}

			for _, category := range categories {
				commands, err := config.GetCommands(category)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting %s commands: %v\n", category, err)
					continue
				}

				label := strings.ReplaceAll(category, "-", " ")
				fmt.Printf("\n%s%s commands:\n", strings.ToUpper(label[:1]), label[1:])
				if len(commands) == 0 {
					fmt.Println("  (none)")
				} else {
					for i, cmd := range commands {
						fmt.Printf("  %d. %s\n", i+1, cmd.Description)
						fmt.Printf("     Command: %s\n", cmd.Command)
						if cmd.WorkingDir != "." && cmd.WorkingDir != "" {
							fmt.Printf("     Working Dir: %s\n", cmd.WorkingDir)
						}
						if len(cmd.TriggerFiles) > 0 {
							fmt.Printf("     Triggers: %s\n", strings.Join(cmd.TriggerFiles, ", "))
						}
					}
				}
			}
		},
	}

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Open config in $EDITOR",
		Run: func(cmd *cobra.Command, args []string) {
			if err := mycelia.OpenConfigInEditor(); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening config in editor: %v\n", err)
				os.Exit(1)
			}
		},
	}

	runCmd := &cobra.Command{
		Use:   "run <category>",
		Short: "Execute commands for a category",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			category := args[0]
			autoApprove, _ := cmd.Flags().GetBool("auto")

			if category != "post-pull" && category != "post-checkout" {
				fmt.Fprintf(os.Stderr, "Error: Invalid category '%s'. Must be 'post-pull' or 'post-checkout'\n", category)
				os.Exit(1)
			}

			config, err := mycelia.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			executor := mycelia.NewExecutor(config)
			if err := executor.ExecuteCategory(category, autoApprove); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing %s commands: %v\n", category, err)
				os.Exit(1)
			}
		},
	}
	runCmd.Flags().Bool("auto", false, "Run commands without confirmation")

	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Run post-pull housekeeping commands",
		Long:  fmt.Sprintf("Shortcut for \"%s run post-pull\". Executes all configured post-pull commands.", name),
		Run: func(cmd *cobra.Command, args []string) {
			autoApprove, _ := cmd.Flags().GetBool("auto")

			config, err := mycelia.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			executor := mycelia.NewExecutor(config)
			if err := executor.ExecuteCategory("post-pull", autoApprove); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing post-pull commands: %v\n", err)
				os.Exit(1)
			}
		},
	}
	pullCmd.Flags().Bool("auto", false, "Run commands without confirmation")

	checkoutCmd := &cobra.Command{
		Use:   "checkout",
		Short: "Run post-checkout housekeeping commands",
		Long:  fmt.Sprintf("Shortcut for \"%s run post-checkout\". Executes all configured post-checkout commands.", name),
		Run: func(cmd *cobra.Command, args []string) {
			autoApprove, _ := cmd.Flags().GetBool("auto")

			config, err := mycelia.LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			executor := mycelia.NewExecutor(config)
			if err := executor.ExecuteCategory("post-checkout", autoApprove); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing post-checkout commands: %v\n", err)
				os.Exit(1)
			}
		},
	}
	checkoutCmd.Flags().Bool("auto", false, "Run commands without confirmation")

	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(suggestCmd)
	rootCmd.AddCommand(autoCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(checkoutCmd)

	return rootCmd
}
