package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeAdd
	ModeEdit
	ModeFilter
	ModeDelete
	ModeHelp
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "Normal"
	case ModeAdd:
		return "Add"
	case ModeEdit:
		return "Edit"
	case ModeFilter:
		return "Filter"
	case ModeDelete:
		return "Delete"
	case ModeHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

// Model represents the application state
type Model struct {
	list           *TodoList
	storage        *Storage
	cursor         int
	mode           Mode
	textarea       textarea.Model
	filterInput    string
	filterTags     []string
	showCompleted  bool
	editingTodoID  string
	width          int
	height         int
	message        string
	messageTimeout time.Time
}

// NewModel creates a new application model
func NewModel(list *TodoList, storage *Storage) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter your todo..."
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(80)
	ta.SetHeight(5)

	return Model{
		list:          list,
		storage:       storage,
		cursor:        0,
		mode:          ModeNormal,
		textarea:      ta,
		showCompleted: false,
		width:         80,
		height:        24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(min(msg.Width-4, 100))
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// Update textarea if in add/edit mode (for non-KeyMsg events)
	if m.mode == ModeAdd || m.mode == ModeEdit {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress handles keyboard input based on current mode
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeAdd:
		return m.handleAddMode(msg)
	case ModeEdit:
		return m.handleEditMode(msg)
	case ModeFilter:
		return m.handleFilterMode(msg)
	case ModeDelete:
		return m.handleDeleteMode(msg)
	case ModeHelp:
		return m.handleHelpMode(msg)
	}
	return m, nil
}

// handleNormalMode handles key presses in normal mode
func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	todos := m.getVisibleTodos()

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "a":
		m.mode = ModeAdd
		m.textarea.Reset()
		m.textarea.Focus()
		return m, nil

	case "e":
		if len(todos) > 0 && m.cursor < len(todos) {
			todo := todos[m.cursor]
			m.mode = ModeEdit
			m.editingTodoID = todo.ID
			m.textarea.SetValue(todo.GetFullText())
			m.textarea.Focus()
		}
		return m, nil

	case "d":
		if len(todos) > 0 && m.cursor < len(todos) {
			m.mode = ModeDelete
		}
		return m, nil

	case " ":
		if len(todos) > 0 && m.cursor < len(todos) {
			todo := todos[m.cursor]
			m.list.ToggleComplete(todo.ID)
			m.save()
			// Adjust cursor if needed
			m.adjustCursor()
		}
		return m, nil

	case "t":
		m.showCompleted = !m.showCompleted
		m.adjustCursor()
		return m, nil

	case "f":
		m.mode = ModeFilter
		m.filterInput = strings.Join(m.filterTags, " ")
		return m, nil

	case "?", "h":
		m.mode = ModeHelp
		return m, nil

	case "j", "down":
		if m.cursor < len(todos)-1 {
			m.cursor++
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "g":
		m.cursor = 0
		return m, nil

	case "G":
		if len(todos) > 0 {
			m.cursor = len(todos) - 1
		}
		return m, nil

	case "J", "ctrl+j":
		if len(todos) > 0 && m.cursor < len(todos) {
			todo := todos[m.cursor]
			if !todo.Completed {
				if m.list.MoveDown(todo.ID) {
					m.save()
					if m.cursor < len(todos)-1 {
						m.cursor++
					}
				}
			}
		}
		return m, nil

	case "K", "ctrl+k":
		if len(todos) > 0 && m.cursor < len(todos) {
			todo := todos[m.cursor]
			if !todo.Completed {
				if m.list.MoveUp(todo.ID) {
					m.save()
					if m.cursor > 0 {
						m.cursor--
					}
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// handleAddMode handles key presses in add mode
func (m Model) handleAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle special keys before passing to textarea
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.textarea.Blur()
		return m, nil

	case "ctrl+s":
		text := strings.TrimSpace(m.textarea.Value())
		if text != "" {
			m.list.AddTodo(text)
			m.save()
			m.showMessage("Todo added successfully")
		}
		m.mode = ModeNormal
		m.textarea.Blur()
		m.adjustCursor()
		return m, nil
	}

	// Check for Ctrl+J (alternative to Ctrl+S)
	if msg.Type == tea.KeyCtrlJ {
		text := strings.TrimSpace(m.textarea.Value())
		if text != "" {
			m.list.AddTodo(text)
			m.save()
			m.showMessage("Todo added successfully")
		}
		m.mode = ModeNormal
		m.textarea.Blur()
		m.adjustCursor()
		return m, nil
	}

	// Pass all other keys (including regular Enter) to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// handleEditMode handles key presses in edit mode
func (m Model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle special keys before passing to textarea
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.textarea.Blur()
		m.editingTodoID = ""
		return m, nil

	case "ctrl+s":
		text := strings.TrimSpace(m.textarea.Value())
		if text != "" {
			m.list.UpdateTodo(m.editingTodoID, text)
			m.save()
			m.showMessage("Todo updated successfully")
		}
		m.mode = ModeNormal
		m.textarea.Blur()
		m.editingTodoID = ""
		return m, nil
	}

	// Check for Ctrl+J (alternative to Ctrl+S)
	if msg.Type == tea.KeyCtrlJ {
		text := strings.TrimSpace(m.textarea.Value())
		if text != "" {
			m.list.UpdateTodo(m.editingTodoID, text)
			m.save()
			m.showMessage("Todo updated successfully")
		}
		m.mode = ModeNormal
		m.textarea.Blur()
		m.editingTodoID = ""
		return m, nil
	}

	// Pass all other keys (including regular Enter) to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// handleFilterMode handles key presses in filter mode
func (m Model) handleFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.filterInput = ""
		m.filterTags = []string{}
		m.adjustCursor()
		return m, nil

	case "enter":
		m.mode = ModeNormal
		m.filterTags = parseFilterTags(m.filterInput)
		m.adjustCursor()
		return m, nil

	case "backspace":
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
		}
		return m, nil

	default:
		if len(msg.String()) == 1 {
			m.filterInput += msg.String()
		}
		return m, nil
	}
}

// handleDeleteMode handles key presses in delete confirmation mode
func (m Model) handleDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	todos := m.getVisibleTodos()

	switch msg.String() {
	case "y", "Y":
		if len(todos) > 0 && m.cursor < len(todos) {
			todo := todos[m.cursor]
			m.list.DeleteTodo(todo.ID)
			m.save()
			m.showMessage("Todo deleted")
			m.adjustCursor()
		}
		m.mode = ModeNormal
		return m, nil

	case "n", "N", "esc":
		m.mode = ModeNormal
		return m, nil
	}

	return m, nil
}

// handleHelpMode handles key presses in help mode
func (m Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?", "h", "q":
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	switch m.mode {
	case ModeNormal:
		return m.renderNormal()
	case ModeAdd:
		return m.renderAdd()
	case ModeEdit:
		return m.renderEdit()
	case ModeFilter:
		return m.renderFilter()
	case ModeDelete:
		return m.renderDelete()
	case ModeHelp:
		return m.renderHelp()
	}
	return ""
}

// renderNormal renders the normal mode view
func (m Model) renderNormal() string {
	var b strings.Builder

	// Header (takes 2 lines)
	header := headerStyle.Render(fmt.Sprintf(" td - Todo List Manager%s[%s] ", strings.Repeat(" ", max(0, m.width-45)), m.mode))
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n")

	// Todos
	todos := m.getVisibleTodos()
	if len(todos) == 0 {
		b.WriteString(dimStyle.Render("\n  No todos to display.\n  Press 'a' to add a new todo, '?' for help.\n\n"))
	} else {
		// Calculate how many lines we can show (reserve space for header, footer, and message)
		// Header: 2 lines, Footer: 3 lines (separator + 2 lines status), Message: 1 line
		// This gives us a buffer to ensure footer is always visible
		maxContentLines := m.height - 8

		linesRendered := 0
		for i, todo := range todos {
			if maxContentLines > 0 && linesRendered >= maxContentLines {
				// Show indicator that there are more items
				b.WriteString(dimStyle.Render(fmt.Sprintf("\n  ... %d more items (scroll with j/k) ...\n", len(todos)-i)))
				break
			}

			todoStr := m.renderTodo(todo, i == m.cursor)
			b.WriteString(todoStr)

			// Count approximate lines (title + desc lines + metadata + blank)
			lines := 3 // minimum: title + metadata + blank
			if todo.Description != "" {
				lines += len(strings.Split(todo.Description, "\n"))
			}
			linesRendered += lines
		}
	}

	// Always show footer - add a newline before it if needed
	footer := m.renderFooter()
	b.WriteString("\n" + strings.Repeat("─", m.width) + "\n")
	b.WriteString(footer)

	// Message
	if m.message != "" && time.Now().Before(m.messageTimeout) {
		b.WriteString("\n" + successStyle.Render(m.message))
	}

	return b.String()
}

// renderTodo renders a single todo item
func (m Model) renderTodo(todo Todo, selected bool) string {
	var b strings.Builder

	// Checkbox and selection indicator
	prefix := "  "
	if selected {
		prefix = "> "
	}

	checkbox := "☐"
	if todo.Completed {
		checkbox = "☑"
	}

	// Title line
	titleStyle := titleStyle
	if todo.Completed {
		titleStyle = completedStyle
	}
	if selected {
		titleStyle = titleStyle.Background(lipgloss.Color("240"))
	}

	// Highlight tags in title
	title := m.highlightTags(todo.Title)

	line := prefix + checkbox + " " + title
	b.WriteString(titleStyle.Render(line) + "\n")

	// Description (if present)
	if todo.Description != "" {
		descLines := strings.Split(todo.Description, "\n")
		for _, descLine := range descLines {
			desc := m.highlightTags(descLine)
			descStyle := descriptionStyle
			if todo.Completed {
				descStyle = completedStyle
			}
			b.WriteString(descStyle.Render("    "+desc) + "\n")
		}
	}

	// Metadata line
	created := todo.CreatedAt.Local().Format("Jan 2, 3:04 PM")
	meta := fmt.Sprintf("    Added: %s", created)
	if todo.Completed && todo.CompletedAt != nil {
		completed := todo.CompletedAt.Local().Format("Jan 2, 3:04 PM")
		meta += fmt.Sprintf(" | Completed: %s", completed)
	}
	b.WriteString(dimStyle.Render(meta) + "\n\n")

	return b.String()
}

// renderAdd renders the add mode view
func (m Model) renderAdd() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf(" Add New Todo [%s] ", m.mode)) + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n\n")
	b.WriteString(m.textarea.View() + "\n\n")
	b.WriteString(dimStyle.Render("Ctrl+S or Ctrl+Enter to save, Esc to cancel") + "\n")
	return b.String()
}

// renderEdit renders the edit mode view
func (m Model) renderEdit() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf(" Edit Todo [%s] ", m.mode)) + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n\n")
	b.WriteString(m.textarea.View() + "\n\n")
	b.WriteString(dimStyle.Render("Ctrl+S or Ctrl+Enter to save, Esc to cancel") + "\n")
	return b.String()
}

// renderFilter renders the filter mode view
func (m Model) renderFilter() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf(" Filter by Tags [%s] ", m.mode)) + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n\n")
	b.WriteString("  Enter tags separated by spaces (e.g., work urgent)\n")
	b.WriteString("  Filter: " + m.filterInput + "█\n\n")
	b.WriteString(dimStyle.Render("Enter to apply filter, Esc to cancel") + "\n")
	return b.String()
}

// renderDelete renders the delete confirmation view
func (m Model) renderDelete() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" Delete Todo ") + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n\n")

	todos := m.getVisibleTodos()
	if m.cursor < len(todos) {
		todo := todos[m.cursor]
		b.WriteString("  Are you sure you want to delete this todo?\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", todo.Title))
	}

	b.WriteString(errorStyle.Render("  Press 'y' to confirm, 'n' or Esc to cancel") + "\n")
	return b.String()
}

// renderHelp renders the help overlay
func (m Model) renderHelp() string {
	help := `
 td - Todo List Manager - Help

 Navigation:
   j/↓         Move down
   k/↑         Move up
   g           Jump to top
   G           Jump to bottom

 Actions:
   a           Add new todo
   e           Edit selected todo
   d           Delete selected todo
   space       Toggle completion status
   t           Toggle show/hide completed items
   f           Filter by tags
   J/Ctrl+J    Move todo down in order
   K/Ctrl+K    Move todo up in order

 Add/Edit Mode:
   Enter       New line (multi-line support)
   Ctrl+S      Save todo
   Esc         Cancel

 General:
   ?/h         Show this help
   q/Ctrl+C    Quit

 Tags:
   Use #hashtags anywhere in your todo text to create tags.
   Example: "Buy groceries #personal #shopping"

 Press any key to close this help screen.
`
	return headerStyle.Render(" Help ") + "\n" +
		strings.Repeat("─", m.width) + "\n" +
		help
}

// renderFooter renders the status bar footer
func (m Model) renderFooter() string {
	active := len(m.list.GetActiveTodos())
	completed := len(m.list.GetCompletedTodos())

	status := fmt.Sprintf(" %d active", active)
	if m.showCompleted {
		status += fmt.Sprintf(", %d completed (shown)", completed)
	} else if completed > 0 {
		status += fmt.Sprintf(", %d completed (hidden)", completed)
	}

	if len(m.filterTags) > 0 {
		status += " | Filter: " + tagStyle.Render("#"+strings.Join(m.filterTags, " #"))
	}

	status += " | "

	shortcuts := "a:add e:edit d:delete space:toggle t:show-done f:filter ?:help q:quit"

	return dimStyle.Render(status + shortcuts)
}

// getVisibleTodos returns the list of todos to display based on current filters
func (m Model) getVisibleTodos() []Todo {
	return m.list.GetFilteredTodos(m.filterTags, m.showCompleted)
}

// adjustCursor ensures cursor is within valid bounds
func (m Model) adjustCursor() {
	todos := m.getVisibleTodos()
	if m.cursor >= len(todos) {
		m.cursor = max(0, len(todos)-1)
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// save persists the current todo list to storage
func (m *Model) save() {
	if err := m.storage.Save(m.list); err != nil {
		m.showMessage(fmt.Sprintf("Error saving: %v", err))
	}
}

// showMessage displays a temporary message
func (m *Model) showMessage(msg string) {
	m.message = msg
	m.messageTimeout = time.Now().Add(2 * time.Second)
}

// highlightTags returns text with tags highlighted
func (m Model) highlightTags(text string) string {
	// Simple approach: replace #tag with styled version
	parts := tagRegex.FindAllStringIndex(text, -1)
	if len(parts) == 0 {
		return text
	}

	var result strings.Builder
	lastEnd := 0
	for _, match := range parts {
		result.WriteString(text[lastEnd:match[0]])
		result.WriteString(tagStyle.Render(text[match[0]:match[1]]))
		lastEnd = match[1]
	}
	result.WriteString(text[lastEnd:])
	return result.String()
}

// parseFilterTags parses filter input into tag list
func parseFilterTags(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}

	// Split by spaces and remove # prefix if present
	parts := strings.Fields(input)
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimPrefix(part, "#")
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Strikethrough(true)

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
