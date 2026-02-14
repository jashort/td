package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTodoList(t *testing.T) {
	list := NewTodoList()

	// Test adding todos
	todo1 := list.AddTodo("Buy groceries #personal")
	if todo1.Title != "Buy groceries #personal" {
		t.Errorf("Expected title 'Buy groceries #personal', got '%s'", todo1.Title)
	}
	if len(todo1.Tags) != 1 || todo1.Tags[0] != "personal" {
		t.Errorf("Expected tags [personal], got %v", todo1.Tags)
	}

	// Test multi-line todo
	todo2 := list.AddTodo("Fix bug #work #urgent\nThis is a critical bug\nNeed to fix ASAP")
	if todo2.Title != "Fix bug #work #urgent" {
		t.Errorf("Expected title 'Fix bug #work #urgent', got '%s'", todo2.Title)
	}
	if todo2.Description != "This is a critical bug\nNeed to fix ASAP" {
		t.Errorf("Expected description with newlines, got '%s'", todo2.Description)
	}
	if len(todo2.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(todo2.Tags))
	}

	// Test toggle complete
	if !list.ToggleComplete(todo1.ID) {
		t.Error("Failed to toggle completion")
	}
	for _, todo := range list.Todos {
		if todo.ID == todo1.ID && !todo.Completed {
			t.Error("Todo should be completed")
		}
	}

	// Test active/completed separation
	active := list.GetActiveTodos()
	if len(active) != 1 {
		t.Errorf("Expected 1 active todo, got %d", len(active))
	}

	completed := list.GetCompletedTodos()
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed todo, got %d", len(completed))
	}

	// Test filtering with AND logic
	list.AddTodo("Another work item #work")
	filtered := list.GetFilteredTodos([]string{"work", "urgent"}, false)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered todo (with both #work and #urgent), got %d", len(filtered))
	}
}

func TestStorage(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_todos.json")

	storage := NewStorage(testFile)
	list := NewTodoList()

	// Add some todos
	list.AddTodo("Test todo 1 #test")
	list.AddTodo("Test todo 2 #test #work")

	// Save
	if err := storage.Save(list); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load
	loadedList, err := storage.Load()
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(loadedList.Todos) != len(list.Todos) {
		t.Errorf("Expected %d todos, got %d", len(list.Todos), len(loadedList.Todos))
	}

	// Verify data integrity
	for i, todo := range loadedList.Todos {
		if todo.Title != list.Todos[i].Title {
			t.Errorf("Title mismatch at index %d: expected '%s', got '%s'",
				i, list.Todos[i].Title, todo.Title)
		}
		if len(todo.Tags) != len(list.Todos[i].Tags) {
			t.Errorf("Tags mismatch at index %d", i)
		}
	}
}

func TestTagExtraction(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"No tags here", []string{}},
		{"Single #tag", []string{"tag"}},
		{"Multiple #work #urgent tags", []string{"work", "urgent"}},
		{"Case #Test #TEST should dedupe", []string{"test"}},
		{"Special chars #work-item #test_123", []string{"work-item", "test_123"}},
	}

	for _, test := range tests {
		result := ExtractTags(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input '%s': expected %d tags, got %d",
				test.input, len(test.expected), len(result))
			continue
		}
		for i, tag := range result {
			if tag != test.expected[i] {
				t.Errorf("For input '%s': expected tag '%s', got '%s'",
					test.input, test.expected[i], tag)
			}
		}
	}
}

func TestReordering(t *testing.T) {
	list := NewTodoList()

	todo1 := list.AddTodo("First")
	todo2 := list.AddTodo("Second")
	todo3 := list.AddTodo("Third")

	// Initial order: First (0), Second (1), Third (2)
	active := list.GetActiveTodos()
	if active[0].ID != todo1.ID || active[1].ID != todo2.ID || active[2].ID != todo3.ID {
		t.Error("Initial order is incorrect")
	}

	// Move second item down (swap with third)
	list.MoveDown(todo2.ID)
	active = list.GetActiveTodos()
	if active[1].ID != todo3.ID || active[2].ID != todo2.ID {
		t.Error("MoveDown failed: expected Second and Third to swap")
	}

	// Move third item (now at position 1) up (swap with first)
	list.MoveUp(todo3.ID)
	active = list.GetActiveTodos()
	if active[0].ID != todo3.ID || active[1].ID != todo1.ID {
		t.Error("MoveUp failed: expected Third and First to swap")
	}
}

func TestCompletedAtBottom(t *testing.T) {
	list := NewTodoList()

	list.AddTodo("Active 1")
	list.AddTodo("Active 2")
	todo3 := list.AddTodo("To Complete")

	// Complete the third one
	list.ToggleComplete(todo3.ID)

	// Get sorted todos with completed visible
	sorted := list.GetSortedTodos()

	// Should be: Active 1, Active 2, To Complete
	if len(sorted) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(sorted))
	}

	// First two should be active
	if sorted[0].Completed || sorted[1].Completed {
		t.Error("First two items should be active")
	}

	// Last should be completed
	if !sorted[2].Completed {
		t.Error("Last item should be completed")
	}
}

func TestTimestampStorage(t *testing.T) {
	list := NewTodoList()
	todo := list.AddTodo("Test todo")

	// Check created at is recent (within last second)
	if time.Since(todo.CreatedAt) > time.Second {
		t.Error("CreatedAt timestamp is not recent")
	}

	// Check that it's in UTC
	if todo.CreatedAt.Location() != time.UTC {
		t.Error("CreatedAt should be stored in UTC")
	}

	// Complete the todo
	list.ToggleComplete(todo.ID)

	// Find the completed todo
	var completed *Todo
	for i := range list.Todos {
		if list.Todos[i].ID == todo.ID {
			completed = &list.Todos[i]
			break
		}
	}

	if completed.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	if completed.CompletedAt.Location() != time.UTC {
		t.Error("CompletedAt should be stored in UTC")
	}
}

func TestStoragePath(t *testing.T) {
	// Test CLI flag priority
	path, err := ResolveStoragePath("/custom/path")
	if err != nil {
		t.Fatalf("Failed to resolve path: %v", err)
	}
	if path != "/custom/path" {
		t.Errorf("Expected '/custom/path', got '%s'", path)
	}

	// Test environment variable
	os.Setenv("TD_FILE", "/env/path")
	defer os.Unsetenv("TD_FILE")

	path, err = ResolveStoragePath("")
	if err != nil {
		t.Fatalf("Failed to resolve path: %v", err)
	}
	if path != "/env/path" {
		t.Errorf("Expected '/env/path', got '%s'", path)
	}

	// Test default path
	os.Unsetenv("TD_FILE")
	path, err = ResolveStoragePath("")
	if err != nil {
		t.Fatalf("Failed to resolve path: %v", err)
	}
	home, _ := os.UserHomeDir()
	expectedDefault := filepath.Join(home, ".td", "todos.json")
	if path != expectedDefault {
		t.Errorf("Expected '%s', got '%s'", expectedDefault, path)
	}
}
