package main

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Todo represents a single todo item with metadata
type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`       // First line of text
	Description string     `json:"description"` // Remaining lines
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Tags        []string   `json:"tags"`
	Order       int        `json:"order"`
}

// TodoList manages a collection of todos
type TodoList struct {
	Todos     []Todo `json:"todos"`
	NextOrder int    `json:"next_order"`
}

// NewTodoList creates a new empty todo list
func NewTodoList() *TodoList {
	return &TodoList{
		Todos:     []Todo{},
		NextOrder: 0,
	}
}

// AddTodo adds a new todo to the list
func (tl *TodoList) AddTodo(text string) *Todo {
	// Split into title and description
	title, description := splitTitleDescription(text)

	// Extract tags from full text
	tags := ExtractTags(text)

	todo := Todo{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now().UTC(),
		CompletedAt: nil,
		Tags:        tags,
		Order:       tl.NextOrder,
	}

	tl.Todos = append(tl.Todos, todo)
	tl.NextOrder++

	return &todo
}

// UpdateTodo updates an existing todo's text
func (tl *TodoList) UpdateTodo(id string, text string) bool {
	for i, todo := range tl.Todos {
		if todo.ID == id {
			title, description := splitTitleDescription(text)
			tl.Todos[i].Title = title
			tl.Todos[i].Description = description
			tl.Todos[i].Tags = ExtractTags(text)
			return true
		}
	}
	return false
}

// DeleteTodo removes a todo from the list
func (tl *TodoList) DeleteTodo(id string) bool {
	for i, todo := range tl.Todos {
		if todo.ID == id {
			tl.Todos = append(tl.Todos[:i], tl.Todos[i+1:]...)
			return true
		}
	}
	return false
}

// ToggleComplete toggles the completion status of a todo
func (tl *TodoList) ToggleComplete(id string) bool {
	for i, todo := range tl.Todos {
		if todo.ID == id {
			tl.Todos[i].Completed = !tl.Todos[i].Completed
			if tl.Todos[i].Completed {
				now := time.Now().UTC()
				tl.Todos[i].CompletedAt = &now
			} else {
				tl.Todos[i].CompletedAt = nil
			}
			return true
		}
	}
	return false
}

// MoveUp moves a todo up in order (swap with previous active todo)
func (tl *TodoList) MoveUp(id string) bool {
	// Get active todos only
	activeTodos := tl.GetActiveTodos()
	if len(activeTodos) <= 1 {
		return false
	}

	// Find the todo in active list
	idx := -1
	for i, todo := range activeTodos {
		if todo.ID == id {
			idx = i
			break
		}
	}

	if idx <= 0 {
		return false // Already at top or not found
	}

	// Swap orders
	for i := range tl.Todos {
		if tl.Todos[i].ID == activeTodos[idx].ID {
			tl.Todos[i].Order = activeTodos[idx-1].Order
		} else if tl.Todos[i].ID == activeTodos[idx-1].ID {
			tl.Todos[i].Order = activeTodos[idx].Order
		}
	}

	return true
}

// MoveDown moves a todo down in order (swap with next active todo)
func (tl *TodoList) MoveDown(id string) bool {
	// Get active todos only
	activeTodos := tl.GetActiveTodos()
	if len(activeTodos) <= 1 {
		return false
	}

	// Find the todo in active list
	idx := -1
	for i, todo := range activeTodos {
		if todo.ID == id {
			idx = i
			break
		}
	}

	if idx < 0 || idx >= len(activeTodos)-1 {
		return false // Already at bottom or not found
	}

	// Swap orders
	for i := range tl.Todos {
		if tl.Todos[i].ID == activeTodos[idx].ID {
			tl.Todos[i].Order = activeTodos[idx+1].Order
		} else if tl.Todos[i].ID == activeTodos[idx+1].ID {
			tl.Todos[i].Order = activeTodos[idx].Order
		}
	}

	return true
}

// GetActiveTodos returns all active (not completed) todos sorted by order
func (tl *TodoList) GetActiveTodos() []Todo {
	active := []Todo{}
	for _, todo := range tl.Todos {
		if !todo.Completed {
			active = append(active, todo)
		}
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].Order < active[j].Order
	})

	return active
}

// GetCompletedTodos returns all completed todos sorted by completion date (newest first)
func (tl *TodoList) GetCompletedTodos() []Todo {
	completed := []Todo{}
	for _, todo := range tl.Todos {
		if todo.Completed {
			completed = append(completed, todo)
		}
	}

	sort.Slice(completed, func(i, j int) bool {
		if completed[i].CompletedAt == nil {
			return false
		}
		if completed[j].CompletedAt == nil {
			return true
		}
		return completed[i].CompletedAt.After(*completed[j].CompletedAt)
	})

	return completed
}

// GetFilteredTodos returns todos filtered by tags (AND logic)
func (tl *TodoList) GetFilteredTodos(filterTags []string, showCompleted bool) []Todo {
	if len(filterTags) == 0 {
		// No filter, return all based on completion visibility
		if showCompleted {
			return tl.GetSortedTodos()
		}
		return tl.GetActiveTodos()
	}

	// Normalize filter tags to lowercase
	normalizedFilters := make([]string, len(filterTags))
	for i, tag := range filterTags {
		normalizedFilters[i] = strings.ToLower(tag)
	}

	// Filter todos that have ALL specified tags
	filtered := []Todo{}
	for _, todo := range tl.Todos {
		if !showCompleted && todo.Completed {
			continue
		}

		hasAllTags := true
		for _, filterTag := range normalizedFilters {
			found := false
			for _, todoTag := range todo.Tags {
				if strings.ToLower(todoTag) == filterTag {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}

		if hasAllTags {
			filtered = append(filtered, todo)
		}
	}

	// Sort filtered results: active by order, completed by date
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Completed != filtered[j].Completed {
			return !filtered[i].Completed // Active items first
		}
		if filtered[i].Completed {
			// Both completed, sort by completion date
			if filtered[i].CompletedAt == nil {
				return false
			}
			if filtered[j].CompletedAt == nil {
				return true
			}
			return filtered[i].CompletedAt.After(*filtered[j].CompletedAt)
		}
		// Both active, sort by order
		return filtered[i].Order < filtered[j].Order
	})

	return filtered
}

// GetSortedTodos returns all todos sorted (active by order, then completed by date)
func (tl *TodoList) GetSortedTodos() []Todo {
	active := tl.GetActiveTodos()
	completed := tl.GetCompletedTodos()
	return append(active, completed...)
}

// splitTitleDescription splits text into title (first line) and description (rest)
func splitTitleDescription(text string) (string, string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}

	lines := strings.Split(text, "\n")
	title := strings.TrimSpace(lines[0])

	if len(lines) == 1 {
		return title, ""
	}

	// Join remaining lines
	description := strings.Join(lines[1:], "\n")
	description = strings.TrimSpace(description)

	return title, description
}

// GetFullText returns the full text (title + description) of a todo
func (t *Todo) GetFullText() string {
	if t.Description == "" {
		return t.Title
	}
	return t.Title + "\n" + t.Description
}
