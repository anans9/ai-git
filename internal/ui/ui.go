package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anans9/ai-git/internal/git"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// UI provides methods for user interface operations
type UI struct {
	colorEnabled bool
	interactive  bool
	spinner      *spinner.Spinner
}

// Colors for different types of output
var (
	ErrorColor     = color.New(color.FgRed, color.Bold)
	SuccessColor   = color.New(color.FgGreen, color.Bold)
	WarningColor   = color.New(color.FgYellow, color.Bold)
	InfoColor      = color.New(color.FgCyan)
	HighlightColor = color.New(color.FgMagenta, color.Bold)
	DimColor       = color.New(color.FgHiBlack)

	// Git status colors
	StagedColor    = color.New(color.FgGreen)
	ModifiedColor  = color.New(color.FgYellow)
	UntrackedColor = color.New(color.FgRed)
	DeletedColor   = color.New(color.FgRed)
	RenamedColor   = color.New(color.FgBlue)
)

// NewUI creates a new UI instance
func NewUI(colorEnabled, interactive bool) *UI {
	// Disable colors if not supported or requested
	if !colorEnabled {
		color.NoColor = true
	}

	return &UI{
		colorEnabled: colorEnabled,
		interactive:  interactive,
	}
}

// Error prints an error message
func (u *UI) Error(msg string, args ...interface{}) {
	ErrorColor.Fprintf(os.Stderr, "✗ "+msg+"\n", args...)
}

// Success prints a success message
func (u *UI) Success(msg string, args ...interface{}) {
	SuccessColor.Printf("✓ "+msg+"\n", args...)
}

// Warning prints a warning message
func (u *UI) Warning(msg string, args ...interface{}) {
	WarningColor.Printf("⚠ "+msg+"\n", args...)
}

// Info prints an info message
func (u *UI) Info(msg string, args ...interface{}) {
	InfoColor.Printf("ℹ "+msg+"\n", args...)
}

// Highlight prints highlighted text
func (u *UI) Highlight(msg string, args ...interface{}) {
	HighlightColor.Printf(msg+"\n", args...)
}

// Dim prints dimmed text
func (u *UI) Dim(msg string, args ...interface{}) {
	DimColor.Printf(msg+"\n", args...)
}

// Print prints normal text
func (u *UI) Print(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

// StartSpinner starts a loading spinner with the given message
func (u *UI) StartSpinner(msg string) {
	if u.spinner != nil {
		u.spinner.Stop()
	}

	u.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	u.spinner.Suffix = " " + msg
	if u.colorEnabled {
		u.spinner.Color("cyan")
	}
	u.spinner.Start()
}

// UpdateSpinner updates the spinner message
func (u *UI) UpdateSpinner(msg string) {
	if u.spinner != nil {
		u.spinner.Suffix = " " + msg
	}
}

// StopSpinner stops the current spinner
func (u *UI) StopSpinner() {
	if u.spinner != nil {
		u.spinner.Stop()
		u.spinner = nil
	}
}

// Confirm prompts the user for confirmation
func (u *UI) Confirm(message string) (bool, error) {
	if !u.interactive {
		return true, nil
	}

	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return false, fmt.Errorf("interrupted by user")
		}
		return false, err
	}

	return result == "y" || result == "Y", nil
}

// Select prompts the user to select from a list of options
func (u *UI) Select(label string, items []string) (int, string, error) {
	if !u.interactive {
		return 0, items[0], nil
	}

	prompt := promptui.Select{
		Label: label,
		Items: items,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	index, result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return -1, "", fmt.Errorf("interrupted by user")
		}
		return -1, "", err
	}

	return index, result, nil
}

// Input prompts the user for text input
func (u *UI) Input(label string, defaultValue string) (string, error) {
	if !u.interactive {
		return defaultValue, nil
	}

	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return "", fmt.Errorf("interrupted by user")
		}
		return "", err
	}

	return result, nil
}

// MultilineInput prompts the user for multiline text input
func (u *UI) MultilineInput(label string) (string, error) {
	if !u.interactive {
		return "", nil
	}

	u.Info("%s (Press Ctrl+D when finished):", label)

	var lines []string

	for {
		prompt := promptui.Prompt{
			Label: ">",
		}

		input, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				break
			}
			return "", err
		}

		if input == "" && len(lines) > 0 {
			break
		}

		lines = append(lines, input)
	}

	return strings.Join(lines, "\n"), nil
}

// PrintStatus prints the git status in a formatted way
func (u *UI) PrintStatus(status *git.Status) {
	if len(status.Staged) == 0 && len(status.Modified) == 0 &&
		len(status.Untracked) == 0 && len(status.Deleted) == 0 &&
		len(status.Renamed) == 0 {
		u.Success("Working directory is clean")
		return
	}

	u.Highlight("Repository Status:")
	u.Print("")

	if len(status.Staged) > 0 {
		u.Print("Changes to be committed:")
		for _, file := range status.Staged {
			StagedColor.Printf("  ✓ %s\n", file.Path)
		}
		u.Print("")
	}

	if len(status.Modified) > 0 {
		u.Print("Changes not staged for commit:")
		for _, file := range status.Modified {
			ModifiedColor.Printf("  ✎ %s\n", file.Path)
		}
		u.Print("")
	}

	if len(status.Deleted) > 0 {
		u.Print("Deleted files:")
		for _, file := range status.Deleted {
			DeletedColor.Printf("  ✗ %s\n", file.Path)
		}
		u.Print("")
	}

	if len(status.Renamed) > 0 {
		u.Print("Renamed files:")
		for _, file := range status.Renamed {
			RenamedColor.Printf("  ↻ %s\n", file.Path)
		}
		u.Print("")
	}

	if len(status.Untracked) > 0 {
		u.Print("Untracked files:")
		for _, file := range status.Untracked {
			UntrackedColor.Printf("  ? %s\n", file.Path)
		}
		u.Print("")
	}
}

// PrintDiff prints a diff in a formatted way
func (u *UI) PrintDiff(diff *git.Diff) {
	if len(diff.Files) == 0 {
		u.Info("No changes to display")
		return
	}

	u.Highlight("Diff Summary:")
	u.Printf(" %d files changed, %d insertions(+), %d deletions(-)",
		diff.Stats.Files, diff.Stats.Additions, diff.Stats.Deletions)
	u.Print("")

	for _, file := range diff.Files {
		u.PrintFileDiff(&file)
	}
}

// PrintFileDiff prints a single file diff
func (u *UI) PrintFileDiff(file *git.FileDiff) {
	// Print file header
	switch file.Status {
	case "A":
		SuccessColor.Printf("new file: %s\n", file.Path)
	case "M":
		ModifiedColor.Printf("modified: %s\n", file.Path)
	case "D":
		DeletedColor.Printf("deleted: %s\n", file.Path)
	case "R":
		RenamedColor.Printf("renamed: %s -> %s\n", file.OldPath, file.Path)
	default:
		u.Printf("%s: %s\n", file.Status, file.Path)
	}

	// Print diff stats for this file
	if file.Additions > 0 || file.Deletions > 0 {
		DimColor.Printf("  +%d -%d\n", file.Additions, file.Deletions)
	}

	// Print a preview of the diff content (first few lines)
	if file.Content != "" {
		lines := strings.Split(file.Content, "\n")
		maxLines := 10
		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}

		for _, line := range lines {
			if strings.HasPrefix(line, "+") {
				SuccessColor.Printf("  %s\n", line)
			} else if strings.HasPrefix(line, "-") {
				ErrorColor.Printf("  %s\n", line)
			} else {
				DimColor.Printf("  %s\n", line)
			}
		}

		if len(strings.Split(file.Content, "\n")) > maxLines {
			DimColor.Println("  ...")
		}
	}

	u.Print("")
}

// PrintBranches prints branches in a formatted way
func (u *UI) PrintBranches(branches []git.Branch) {
	if len(branches) == 0 {
		u.Info("No branches found")
		return
	}

	u.Highlight("Branches:")

	for _, branch := range branches {
		prefix := "  "
		if branch.Current {
			prefix = "* "
			SuccessColor.Printf("%s%s", prefix, branch.Name)
		} else {
			u.Printf("%s%s", prefix, branch.Name)
		}

		if branch.LastCommit != "" {
			DimColor.Printf(" (%s)", branch.LastCommit)
		}

		u.Print("")
	}
}

// PrintCommits prints commit history in a formatted way
func (u *UI) PrintCommits(commits []git.Commit) {
	if len(commits) == 0 {
		u.Info("No commits found")
		return
	}

	u.Highlight("Recent Commits:")

	for _, commit := range commits {
		HighlightColor.Printf("commit %s", commit.ShortHash)
		u.Printf("Author: %s <%s>", commit.Author, commit.Email)
		u.Printf("Date:   %s", commit.Date.Format("Mon Jan 2 15:04:05 2006 -0700"))
		u.Print("")

		// Print commit message with indentation
		lines := strings.Split(strings.TrimSpace(commit.Message), "\n")
		for _, line := range lines {
			u.Printf("    %s", line)
		}
		u.Print("")
	}
}

// PrintRemotes prints remotes in a formatted way
func (u *UI) PrintRemotes(remotes []git.Remote) {
	if len(remotes) == 0 {
		u.Info("No remotes configured")
		return
	}

	u.Highlight("Remotes:")

	for _, remote := range remotes {
		u.Printf("  %s\t%s", remote.Name, remote.URL)
	}
}

// PrintTable prints data in a table format
func (u *UI) PrintTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, header := range headers {
		HighlightColor.Printf("%-*s", widths[i]+2, header)
	}
	u.Print("")

	// Print separator
	for i := range headers {
		u.Printf("%-*s", widths[i]+2, strings.Repeat("-", widths[i]))
	}
	u.Print("")

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				u.Printf("%-*s", widths[i]+2, cell)
			}
		}
		u.Print("")
	}
}

// Printf prints formatted text
func (u *UI) Printf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// ShowProgress shows a simple progress indicator
func (u *UI) ShowProgress(current, total int, message string) {
	if total == 0 {
		return
	}

	percentage := float64(current) / float64(total) * 100
	barWidth := 30
	filled := int(float64(barWidth) * percentage / 100)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r%s [%s] %.1f%% (%d/%d)", message, bar, percentage, current, total)

	if current == total {
		fmt.Println() // New line when complete
	}
}

// ClearLine clears the current line
func (u *UI) ClearLine() {
	fmt.Print("\r\033[K")
}

// Header prints a section header
func (u *UI) Header(title string) {
	u.Print("")
	HighlightColor.Printf("=== %s ===", title)
	u.Print("")
}

// Separator prints a separator line
func (u *UI) Separator() {
	DimColor.Println(strings.Repeat("-", 50))
}

// IsInteractive returns whether the UI is in interactive mode
func (u *UI) IsInteractive() bool {
	return u.interactive
}

// SetInteractive sets the interactive mode
func (u *UI) SetInteractive(interactive bool) {
	u.interactive = interactive
}

// IsColorEnabled returns whether colors are enabled
func (u *UI) IsColorEnabled() bool {
	return u.colorEnabled && !color.NoColor
}

// SetColorEnabled sets color mode
func (u *UI) SetColorEnabled(enabled bool) {
	u.colorEnabled = enabled
	color.NoColor = !enabled
}
