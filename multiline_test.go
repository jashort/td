package main

import (
	"testing"
)

func TestMultiLineInput(t *testing.T) {
	list := NewTodoList()

	// Test that multi-line text is properly split
	multiLineText := "Title line here\nFirst description line\nSecond description line"
	todo := list.AddTodo(multiLineText)

	if todo.Title != "Title line here" {
		t.Errorf("Expected title 'Title line here', got '%s'", todo.Title)
	}

	expectedDesc := "First description line\nSecond description line"
	if todo.Description != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, todo.Description)
	}

	// Test GetFullText reconstructs properly
	fullText := todo.GetFullText()
	if fullText != multiLineText {
		t.Errorf("Expected full text '%s', got '%s'", multiLineText, fullText)
	}
}

func TestSingleLineInput(t *testing.T) {
	list := NewTodoList()

	// Test that single line text has empty description
	todo := list.AddTodo("Just a title")

	if todo.Title != "Just a title" {
		t.Errorf("Expected title 'Just a title', got '%s'", todo.Title)
	}

	if todo.Description != "" {
		t.Errorf("Expected empty description, got '%s'", todo.Description)
	}

	// Test GetFullText returns just title
	fullText := todo.GetFullText()
	if fullText != "Just a title" {
		t.Errorf("Expected full text 'Just a title', got '%s'", fullText)
	}
}
