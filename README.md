# Mycelia
Configure default commands to run following a pull or checkout -- no more having to remember to run the same 3 commands over and over!

## Installation
### From Source

```bash
git clone https://github.com/sivepanda/mycelia.git
cd mycelia
go build -o mycelia ./cmd/mycelia
go build -o myc ./cmd/myc
```

Optionally install system-wide:
```bash
sudo mv mycelia myc /usr/local/bin/
```

## Commands
`mycelia` or `myc` -- Open the interactive TUI menu  
`mycelia detect` -- Detect package managers and build systems in the current directory  
`mycelia suggest <category>` -- Show suggested commands for a category (`post-pull` or `post-checkout`)  
`mycelia auto <category>` -- Auto-detect and add suggested commands to your config  
`mycelia add <command>` -- Manually add a command (`--post-pull` or `--post-checkout`, with optional `--working-dir` and `--description`)  
`mycelia list` -- List all configured commands  
`mycelia edit` -- Open the config file in `$EDITOR`  
`mycelia run <category>` -- Execute commands for a category  
`mycelia pull` -- Shortcut for `mycelia run post-pull`  
`mycelia checkout` -- Shortcut for `mycelia run post-checkout`  

All `run`, `pull`, and `checkout` commands support `--auto` to skip confirmation.  

## TUI Primitives

Mycelia's TUI is built on a set of **headless primitives** in the `tui` package (`github.com/sivepanda/mycelia/tui`). These are generic, reusable components that never handle key input directly — the parent BubbleTea model drives them via imperative method calls. This makes them easy to compose into custom UIs.

### Selector[T]

A generic multi-select list that works with any item type.

```go
import primitives "github.com/sivepanda/mycelia/tui"

type Fruit struct { Name string; Emoji string }

fruits := []Fruit{{Name: "Apple", Emoji: "🍎"}, {Name: "Banana", Emoji: "🍌"}}

sel := primitives.NewSelector(fruits, func(f Fruit) string {
    return f.Name
}).WithDetail(func(f Fruit) string {
    return f.Emoji
}).WithSelected(true)
```

**Builder methods:** `WithStyles(SelectorStyles)`, `WithDetail(func(T) string)`, `WithSelected(bool)`, `WithWidth(int)`

**Imperative control:** `CursorUp()`, `CursorDown()`, `Toggle()`, `SetItems([]T)`

**State inspection:** `Selected() []T`, `Items() []SelectorItem[T]`, `HasSelected() bool`, `Len() int`, `Cursor() int`

**Rendering:** `View() string`

To customize appearance, pass a `SelectorStyles` struct via `WithStyles()`:
```go
styles := primitives.SelectorStyles{
    Cursor:      "> ",
    CursorBlank: "  ",
    Checked:     "[x]",
    Unchecked:   "[ ]",
    ItemStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
    ActiveStyle:  lipgloss.NewStyle().Bold(true),
    DetailStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
    ActiveDetail: lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa")),
}
sel = sel.WithStyles(styles)
```

### Form

A headless multi-field text input form. Fields are defined declaratively — no subclassing needed.

```go
form := primitives.NewForm([]primitives.FormField{
    {Key: "name", Label: "Name:", Placeholder: "Enter your name"},
    {Key: "email", Label: "Email:", Placeholder: "user@example.com", CharLimit: 100},
})
```

**Builder methods:** `WithStyles(FormStyles)`

**Imperative control:** `NextField()`, `PrevField()`, `UpdateInput(tea.Msg) (Form, tea.Cmd)`, `Reset()`

**State inspection:** `Values() map[string]string`, `Value(key string) string`, `FocusedField() int`

**Rendering:** `View() string`

### Setup

A headless setup wizard that composes `Selector` and `Form` internally. It walks through package detection, category selection, command selection, and optional manual input.

```go
setup := primitives.NewSetup().WithStyles(myStyles)
cmd := setup.Init() // starts spinner + async detection
```

**Imperative control:** `CursorUp()`, `CursorDown()`, `Toggle()`, `Submit() (Setup, tea.Cmd)`, `Back()`, `EnterManualInput()`, `UpdateInput(tea.Msg)`, `FormNextField()`, `FormPrevField()`, `SetSize(w, h int)`

**State inspection:** `State() SetupState`, `Done() bool`, `Err() error`, `AddedCount() int`

**Async messages (handled by `Update`):** `DetectionCompleteMsg`, `SuggestionsLoadedMsg`, `CommandsAddedMsg`

### Extending the Primitives

All three primitives follow the same headless pattern:

1. **Create** with a constructor (`NewSelector`, `NewForm`, `NewSetup`)
2. **Configure** with builder methods (`WithStyles`, `WithDetail`, etc.)
3. **Drive** from your parent model's `Update` by calling imperative methods (`CursorUp`, `Toggle`, `Submit`, etc.)
4. **Render** by calling `View()` from your parent model's `View`

Because they never capture key input, you have full control over keybindings and can compose multiple primitives in a single view. The `internal/tui/menu.go` file demonstrates this pattern — it embeds a `Setup` primitive and drives it from a top-level `Menu` model with custom key bindings.

## Contributing to `autodetect.json`
`Autodetect.json` is a list of package managers/libraries that often require some sort of syncing after a pull or checkout. This list is used to suggest potential post- pull and checkout commands, however, it is always the user's choice on whether or not they use them. Users can also add their own custom commands to their config file, so additions to autodetect must be general and widely used libraries.  

Feel free to make a pull request!

### Keys
`name` The name of the package manager/library  
`detectFile` File that when detected, can be used to infer a potential package manager/library used  
`detectFiles` Files that only when detected *together*, can be used to infer a potential package manager/library used **CANNOT BE USED WITH detectFile**  
`description` Description of the command/library/pkgman  
`commands` contains post-pull and post-checkout commands. **Both post-pull and post-checkout use the same schema.**  
`post-pull` contains associated post-pull command  
`post-checkout` contains associated post-checkout command  
`command` bash command to run  
`workingDir` directory to run bash command  
`description` description of bash command  

### Examples
Below is a simple example of an addition to autodetect:
```json
{
    "name": "npm",
    "detectFile": "package.json",
    "description": "Node.js (npm)",
    "commands": {
      "post-pull": [
        {
          "command": "npm install",
          "workingDir": ".",
          "description": "Install npm dependencies"
        }
      ],
      "post-checkout": [
        {
          "command": "npm install",
          "workingDir": ".",
          "description": "Install npm dependencies"
        }
      ]
}
```

Alternatively, if multiple files must be detected to infer, you can add multiple `detectFiles`:
```json
    "name": "prisma-pnpm",
    "detectFiles": [
      "prisma/schema.prisma",
      "pnpm-lock.yaml"
    ],
    "excludes": [
      "prisma-npm"
    ],
    "description": "Prisma ORM (pnpm)",
    "commands": {
      "post-pull": [
        {
          "command": "pnpm prisma generate",
          "workingDir": ".",
          "description": "Generate Prisma Client (pnpm)"
        },
        {
          "command": "pnpm prisma migrate deploy",
          "workingDir": ".",
          "description": "Apply Prisma migrations (pnpm)"
        }
      ],
      "post-checkout": [
        {
          "command": "pnpm prisma generate",
          "workingDir": ".",
          "description": "Generate Prisma Client (pnpm)"
        },
        {
          "command": "pnpm prisma migrate deploy",
          "workingDir": ".",
          "description": "Apply Prisma migrations (pnpm)"
        }
      ]
    }
```

