package mycelia

import (
	"slices"
	"testing"
)

func TestFilterByTriggerFiles(t *testing.T) {
	untriggered := Command{Command: "make generate", Description: "untriggered"}
	emptyTriggers := Command{Command: "make lint", Description: "empty-triggers", TriggerFiles: []string{}}
	npm := Command{Command: "npm install", Description: "npm", TriggerFiles: []string{"package.json", "package-lock.json"}}
	gomod := Command{Command: "go mod download", Description: "gomod", TriggerFiles: []string{"go.mod"}}
	dotnet := Command{Command: "dotnet restore", Description: "dotnet", TriggerFiles: []string{"*.csproj"}}

	all := []Command{untriggered, npm, gomod, emptyTriggers, dotnet}

	tests := []struct {
		name         string
		commands     []Command
		changedFiles []string
		want         []string // descriptions of the commands expected to run, in order
	}{
		{
			name:         "no list supplied runs everything",
			commands:     all,
			changedFiles: nil,
			want:         []string{"untriggered", "npm", "gomod", "empty-triggers", "dotnet"},
		},
		{
			name:         "empty list runs only commands without triggers",
			commands:     all,
			changedFiles: []string{},
			want:         []string{"untriggered", "empty-triggers"},
		},
		{
			name:         "list matching a trigger exactly",
			commands:     all,
			changedFiles: []string{"README.md", "package-lock.json"},
			want:         []string{"untriggered", "npm", "empty-triggers"},
		},
		{
			name:         "list matching several triggers",
			commands:     all,
			changedFiles: []string{"go.mod", "package.json"},
			want:         []string{"untriggered", "npm", "gomod", "empty-triggers"},
		},
		{
			name:         "list matching a glob trigger by base name",
			commands:     all,
			changedFiles: []string{"src/App/App.csproj"},
			want:         []string{"untriggered", "empty-triggers", "dotnet"},
		},
		{
			name:         "list not matching any trigger",
			commands:     all,
			changedFiles: []string{"main.go", "docs/package.json.md"},
			want:         []string{"untriggered", "empty-triggers"},
		},
		{
			name:         "commands without triggers with no list supplied",
			commands:     []Command{untriggered, emptyTriggers},
			changedFiles: nil,
			want:         []string{"untriggered", "empty-triggers"},
		},
		{
			name:         "commands without triggers with empty list",
			commands:     []Command{untriggered, emptyTriggers},
			changedFiles: []string{},
			want:         []string{"untriggered", "empty-triggers"},
		},
		{
			name:         "commands without triggers with non-matching list",
			commands:     []Command{untriggered, emptyTriggers},
			changedFiles: []string{"main.go"},
			want:         []string{"untriggered", "empty-triggers"},
		},
		{
			name:         "only triggered commands with empty list",
			commands:     []Command{npm, gomod},
			changedFiles: []string{},
			want:         nil,
		},
		{
			name:         "no commands configured",
			commands:     nil,
			changedFiles: nil,
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := descriptions(filterByTriggerFiles(tt.commands, tt.changedFiles))
			if !slices.Equal(got, tt.want) {
				t.Errorf("filterByTriggerFiles(changedFiles=%#v) = %v, want %v", tt.changedFiles, got, tt.want)
			}
		})
	}
}

// TestExecuteCategorySelectsCommands checks, through the public Executor API,
// whether a command is actually attempted. The configured command points at a
// binary that does not exist, so an error means it was selected and run, and
// nil means it was filtered out.
func TestExecuteCategorySelectsCommands(t *testing.T) {
	const missing = "mycelia-test-binary-that-does-not-exist"

	tests := []struct {
		name         string
		triggers     []string
		useChanged   bool
		changedFiles []string
		wantRun      bool
	}{
		{name: "ExecuteCategory runs untriggered command", wantRun: true},
		{name: "ExecuteCategory runs triggered command", triggers: []string{"go.mod"}, wantRun: true},
		{name: "nil list runs triggered command", triggers: []string{"go.mod"}, useChanged: true, changedFiles: nil, wantRun: true},
		{name: "empty list skips triggered command", triggers: []string{"go.mod"}, useChanged: true, changedFiles: []string{}, wantRun: false},
		{name: "empty list runs untriggered command", useChanged: true, changedFiles: []string{}, wantRun: true},
		{name: "matching list runs triggered command", triggers: []string{"go.mod"}, useChanged: true, changedFiles: []string{"go.mod"}, wantRun: true},
		{name: "non-matching list skips triggered command", triggers: []string{"go.mod"}, useChanged: true, changedFiles: []string{"main.go"}, wantRun: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			if err := config.AddCommand("post-pull", Command{
				Command:      missing,
				Description:  "missing binary",
				TriggerFiles: tt.triggers,
			}); err != nil {
				t.Fatalf("AddCommand: %v", err)
			}

			executor := NewExecutor(config)
			var err error
			if tt.useChanged {
				err = executor.ExecuteCategoryWithChangedFiles("post-pull", tt.changedFiles, true)
			} else {
				err = executor.ExecuteCategory("post-pull", true)
			}

			if ran := err != nil; ran != tt.wantRun {
				t.Errorf("command ran = %v (err = %v), want %v", ran, err, tt.wantRun)
			}
		})
	}
}

func descriptions(commands []Command) []string {
	var out []string
	for _, cmd := range commands {
		out = append(out, cmd.Description)
	}
	return out
}
