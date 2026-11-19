package mycelia

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Executor struct {
	config *Config
}

func NewExecutor(config *Config) *Executor {
	return &Executor{config: config}
}

// ExecuteCategory runs every command configured for category, regardless of
// trigger files. Use it when no information about changed files is available
// (for example a plain "mycelia run post-pull").
func (e *Executor) ExecuteCategory(category string, autoApprove bool) error {
	return e.ExecuteCategoryWithChangedFiles(category, nil, autoApprove)
}

// ExecuteCategoryWithChangedFiles runs the commands configured for category,
// using changedFiles to decide which trigger-gated commands are needed.
//
// A nil changedFiles means "unknown": no filtering is applied and every
// command runs, exactly like ExecuteCategory. A non-nil changedFiles, even an
// empty one, is treated as the authoritative list of changed paths: commands
// without trigger files always run, and commands with trigger files run only
// if one of their triggers matches a changed path. So a non-nil empty slice
// means "nothing changed" and skips every trigger-gated command.
func (e *Executor) ExecuteCategoryWithChangedFiles(category string, changedFiles []string, autoApprove bool) error {
	commands, err := e.config.GetCommands(category)
	if err != nil {
		return err
	}

	commands = filterByTriggerFiles(commands, changedFiles)

	if len(commands) == 0 {
		fmt.Printf("No %s commands to run.\n", category)
		return nil
	}

	fmt.Printf("Found %d %s tasks:\n", len(commands), category)
	for _, cmd := range commands {
		desc := cmd.Description
		if desc == "" {
			desc = cmd.Command
		}
		fmt.Printf("  • %s\n", desc)
	}

	if !autoApprove {
		fmt.Print("Run these? [Y/n]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "n" || response == "no" {
			fmt.Println("Skipped housekeeping tasks.")
			return nil
		}
	}

	fmt.Println("Running housekeeping tasks...")
	for i, cmd := range commands {
		fmt.Printf("[%d/%d] %s\n", i+1, len(commands), cmd.Description)
		if err := e.executeCommand(cmd); err != nil {
			return fmt.Errorf("failed to execute command '%s': %w", cmd.Command, err)
		}
	}

	fmt.Println("All housekeeping tasks completed successfully!")
	return nil
}

// filterByTriggerFiles returns the commands that should run given changedFiles.
//
// If changedFiles is nil (no change information), all commands are returned
// unfiltered. Otherwise commands with no trigger files are always kept, and
// commands with trigger files are kept only if at least one trigger matches a
// changed path. An empty, non-nil changedFiles therefore keeps only the
// commands without trigger files.
func filterByTriggerFiles(commands []Command, changedFiles []string) []Command {
	if changedFiles == nil {
		return commands
	}

	var filtered []Command
	for _, cmd := range commands {
		if len(cmd.TriggerFiles) == 0 {
			filtered = append(filtered, cmd)
			continue
		}

		for _, trigger := range cmd.TriggerFiles {
			if matchesAnyChangedFile(trigger, changedFiles) {
				filtered = append(filtered, cmd)
				break
			}
		}
	}
	return filtered
}

// matchesAnyChangedFile checks if a trigger file pattern matches any file in the changed list.
func matchesAnyChangedFile(trigger string, changedFiles []string) bool {
	for _, changed := range changedFiles {
		// Exact match
		if changed == trigger {
			return true
		}
		// Glob match (e.g. "*.csproj", "prisma/schema.prisma")
		if strings.Contains(trigger, "*") {
			if matched, _ := filepath.Match(trigger, changed); matched {
				return true
			}
			if matched, _ := filepath.Match(trigger, filepath.Base(changed)); matched {
				return true
			}
		}
	}
	return false
}

func (e *Executor) executeCommand(cmd Command) error {
	workingDir := cmd.WorkingDir
	if workingDir == "" || workingDir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		workingDir = wd
	}

	parts := strings.Fields(cmd.Command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	execCmd := exec.Command(parts[0], parts[1:]...)
	execCmd.Dir = workingDir
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
