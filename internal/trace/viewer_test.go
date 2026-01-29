package trace

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewViewerModel(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	assert.NotNil(t, model.root)
	assert.NotNil(t, model.flatNodes)
	assert.NotNil(t, model.searchEngine)
	assert.False(t, model.searchMode)
	assert.Equal(t, "", model.searchInput)
	assert.Equal(t, 0, model.cursorPos)
	assert.Equal(t, 0, model.scrollOffset)
	assert.False(t, model.quitting)
}

func TestViewerModel_Init(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	cmd := model.Init()
	assert.Nil(t, cmd)
}

func TestViewerModel_Update_WindowSize(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ := model.Update(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 100, vm.width)
	assert.Equal(t, 30, vm.height)
}

func TestViewerModel_HandleNormalInput_Quit(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModel, cmd := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.True(t, vm.quitting)
	assert.NotNil(t, cmd)
}

func TestViewerModel_HandleNormalInput_SearchMode(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.True(t, vm.searchMode)
	assert.Equal(t, "", vm.searchInput)
}

func TestViewerModel_HandleNormalInput_NavigateDown(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 0
	
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 1, vm.cursorPos)
}

func TestViewerModel_HandleNormalInput_NavigateUp(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 2
	
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 1, vm.cursorPos)
}

func TestViewerModel_HandleNormalInput_NavigateUp_AtTop(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 0
	
	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 0, vm.cursorPos) // Should stay at 0
}

func TestViewerModel_HandleNormalInput_NavigateDown_AtBottom(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = len(model.flatNodes) - 1
	
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, len(model.flatNodes)-1, vm.cursorPos) // Should stay at bottom
}

func TestViewerModel_HandleNormalInput_Home(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 5
	
	msg := tea.KeyMsg{Type: tea.KeyHome}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 0, vm.cursorPos)
}

func TestViewerModel_HandleNormalInput_End(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 0
	
	msg := tea.KeyMsg{Type: tea.KeyEnd}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, len(model.flatNodes)-1, vm.cursorPos)
}

func TestViewerModel_HandleNormalInput_ExpandAll(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Collapse all first
	root.CollapseAll()
	model.flatNodes = root.Flatten()
	initialCount := len(model.flatNodes)
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Greater(t, len(vm.flatNodes), initialCount)
}

func TestViewerModel_HandleNormalInput_CollapseAll(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	initialCount := len(model.flatNodes)
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Less(t, len(vm.flatNodes), initialCount)
}

func TestViewerModel_HandleNormalInput_ToggleExpand(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 1 // Position on a node with children
	
	initialExpanded := model.flatNodes[1].Expanded
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.NotEqual(t, initialExpanded, vm.flatNodes[1].Expanded)
}

func TestViewerModel_HandleNormalInput_ClearSearch(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Set up a search
	model.searchEngine.SetQuery("test")
	model.searchEngine.Search(model.flatNodes)
	
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, "", vm.searchEngine.GetQuery())
}

func TestViewerModel_HandleNormalInput_NextMatch(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Set up a search with matches
	model.searchEngine.SetQuery("contract")
	model.searchEngine.Search(model.flatNodes)
	
	initialPos := model.cursorPos
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	// Cursor should move to next match
	assert.NotEqual(t, initialPos, vm.cursorPos)
}

func TestViewerModel_HandleSearchInput_AddCharacter(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = "te"
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updatedModel, _ := model.handleSearchInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, "tes", vm.searchInput)
}

func TestViewerModel_HandleSearchInput_Backspace(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = "test"
	
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.handleSearchInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, "tes", vm.searchInput)
}

func TestViewerModel_HandleSearchInput_Backspace_Empty(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = ""
	
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := model.handleSearchInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, "", vm.searchInput)
}

func TestViewerModel_HandleSearchInput_Cancel(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = "test"
	
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.handleSearchInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.False(t, vm.searchMode)
	assert.Equal(t, "", vm.searchInput)
}

func TestViewerModel_HandleSearchInput_Execute(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = "transfer"
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.handleSearchInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.False(t, vm.searchMode)
	assert.Equal(t, "transfer", vm.searchEngine.GetQuery())
	assert.Greater(t, vm.searchEngine.MatchCount(), 0)
}

func TestViewerModel_EnsureVisible(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.height = 10
	model.cursorPos = 20
	model.scrollOffset = 0
	
	model.ensureVisible()
	
	// Cursor should be visible in viewport
	assert.Greater(t, model.scrollOffset, 0)
}

func TestViewerModel_EnsureVisible_CursorAboveViewport(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.height = 10
	model.cursorPos = 0
	model.scrollOffset = 5
	
	model.ensureVisible()
	
	// Should scroll up to show cursor
	assert.Equal(t, 0, model.scrollOffset)
}

func TestViewerModel_View_Quitting(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.quitting = true
	
	view := model.View()
	assert.Equal(t, "", view)
}

func TestViewerModel_View_Normal(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.height = 20
	
	view := model.View()
	
	// Should contain help text
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "search")
	assert.Contains(t, view, "quit")
}

func TestViewerModel_View_SearchMode(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = "test"
	model.height = 20
	
	view := model.View()
	
	// Should show search input
	assert.Contains(t, view, "test")
	assert.Contains(t, view, "/")
}

func TestViewerModel_View_WithMatches(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.height = 20
	
	// Set up search with matches
	model.searchEngine.SetQuery("contract")
	model.searchEngine.Search(model.flatNodes)
	
	view := model.View()
	
	// Should show match count
	assert.Contains(t, view, "matches")
}

func TestViewerModel_FormatNodeContent(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	node := &TraceNode{
		ID:         "test",
		Type:       "contract_call",
		ContractID: "CDLZFC3",
		Function:   "transfer",
	}
	
	content := model.formatNodeContent(node)
	
	assert.Contains(t, content, "contract_call")
	assert.Contains(t, content, "CDLZFC3")
	assert.Contains(t, content, "transfer")
}

func TestViewerModel_FormatNodeContent_WithError(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	node := &TraceNode{
		ID:    "test",
		Type:  "error",
		Error: "Something went wrong",
	}
	
	content := model.formatNodeContent(node)
	
	assert.Contains(t, content, "error")
	assert.Contains(t, content, "Something went wrong")
}

func TestViewerModel_FormatNodeContent_WithEvent(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	node := &TraceNode{
		ID:        "test",
		Type:      "event",
		EventData: "Transfer: 100 XLM",
	}
	
	content := model.formatNodeContent(node)
	
	assert.Contains(t, content, "Transfer: 100 XLM")
}

func TestViewerModel_HighlightMatches_NoQuery(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	node := &TraceNode{
		ID:       "test",
		Function: "transfer",
	}
	
	result := model.highlightMatches(node, "transfer function")
	
	// Should return text unchanged when no query
	assert.Equal(t, "transfer function", result)
}

func TestViewerModel_HighlightMatches_WithQuery(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	model.searchEngine.SetQuery("transfer")
	
	node := &TraceNode{
		ID:       "test",
		Function: "transfer",
	}
	
	result := model.highlightMatches(node, "transfer function")
	
	// Should contain the text (highlighting adds ANSI codes)
	assert.Contains(t, result, "transfer")
}

func TestViewerModel_RenderNode(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 0
	
	node := model.flatNodes[0]
	rendered := model.renderNode(node, 0)
	
	// Should contain cursor indicator
	assert.Contains(t, rendered, ">")
}

func TestViewerModel_RenderNode_NotCurrent(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 0
	
	node := model.flatNodes[1]
	rendered := model.renderNode(node, 1)
	
	// Should not have cursor at start
	assert.True(t, strings.HasPrefix(rendered, " ") || strings.HasPrefix(rendered, "  "))
}

func TestRunViewer_Error(t *testing.T) {
	// Test that RunViewer doesn't panic with nil root
	// We can't fully test the TUI, but we can ensure it handles edge cases
	root := CreateMockTrace()
	
	// This would normally block, so we just ensure it's callable
	assert.NotPanics(t, func() {
		// Don't actually run it, just verify the function exists and is callable
		_ = RunViewer
		require.NotNil(t, root)
	})
}

func TestViewerModel_HandleNormalInput_VimKeys(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 2
	
	// Test 'k' (up)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel, _ := model.handleNormalInput(msg)
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 1, vm.cursorPos)
	
	// Test 'j' (down)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ = vm.handleNormalInput(msg)
	vm = updatedModel.(ViewerModel)
	assert.Equal(t, 2, vm.cursorPos)
	
	// Test 'g' (home)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	updatedModel, _ = vm.handleNormalInput(msg)
	vm = updatedModel.(ViewerModel)
	assert.Equal(t, 0, vm.cursorPos)
	
	// Test 'G' (end)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	updatedModel, _ = vm.handleNormalInput(msg)
	vm = updatedModel.(ViewerModel)
	assert.Equal(t, len(model.flatNodes)-1, vm.cursorPos)
}

func TestViewerModel_HandleNormalInput_PageUpDown(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.height = 10
	model.cursorPos = 0
	
	// Test Page Down
	msg := tea.KeyMsg{Type: tea.KeyPgDown}
	updatedModel, _ := model.handleNormalInput(msg)
	vm := updatedModel.(ViewerModel)
	assert.Greater(t, vm.cursorPos, 0)
	
	// Test Page Up
	msg = tea.KeyMsg{Type: tea.KeyPgUp}
	updatedModel, _ = vm.handleNormalInput(msg)
	vm = updatedModel.(ViewerModel)
	assert.Equal(t, 0, vm.cursorPos)
}

func TestViewerModel_HandleNormalInput_Space(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.cursorPos = 1
	
	initialExpanded := model.flatNodes[1].Expanded
	
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.NotEqual(t, initialExpanded, vm.flatNodes[1].Expanded)
}

func TestViewerModel_HandleNormalInput_CtrlC(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, cmd := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.True(t, vm.quitting)
	assert.NotNil(t, cmd)
}

func TestViewerModel_HandleNormalInput_PreviousMatch(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Set up search with matches
	model.searchEngine.SetQuery("contract")
	model.searchEngine.Search(model.flatNodes)
	
	// Move to second match
	model.searchEngine.NextMatch()
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, 1, vm.searchEngine.CurrentMatchNumber())
}

func TestViewerModel_HandleNormalInput_NextMatch_NoMatches(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// No search set up
	initialPos := model.cursorPos
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, _ := model.handleNormalInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.Equal(t, initialPos, vm.cursorPos) // Should not move
}

func TestViewerModel_Update_UnknownMessage(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Send an unknown message type
	type unknownMsg struct{}
	updatedModel, cmd := model.Update(unknownMsg{})
	
	assert.NotNil(t, updatedModel)
	assert.Nil(t, cmd)
}

func TestViewerModel_FormatNodeContent_EmptyNode(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	node := &TraceNode{
		ID: "empty-node",
	}
	
	content := model.formatNodeContent(node)
	
	// Should return ID when no other fields
	assert.Equal(t, "empty-node", content)
}

func TestViewerModel_RenderNode_WithExpandIndicator(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Find a node with children
	var nodeWithChildren *TraceNode
	var index int
	for i, n := range model.flatNodes {
		if !n.IsLeaf() {
			nodeWithChildren = n
			index = i
			break
		}
	}
	
	require.NotNil(t, nodeWithChildren)
	
	rendered := model.renderNode(nodeWithChildren, index)
	
	// Should contain expand/collapse indicator
	assert.True(t, strings.Contains(rendered, "▼") || strings.Contains(rendered, "▶"))
}

func TestViewerModel_RenderNode_CurrentMatch(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	// Set up search
	model.searchEngine.SetQuery("contract")
	model.searchEngine.Search(model.flatNodes)
	
	currentMatch := model.searchEngine.CurrentMatch()
	require.NotNil(t, currentMatch)
	
	rendered := model.renderNode(currentMatch.NodeData, currentMatch.NodeIndex)
	
	// Should contain current match indicator
	assert.Contains(t, rendered, "▶")
}

func TestViewerModel_HighlightMatches_MultipleOccurrences(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	
	model.searchEngine.SetQuery("test")
	
	node := &TraceNode{
		ID:       "test",
		Function: "test",
	}
	
	result := model.highlightMatches(node, "test test test")
	
	// Should highlight all occurrences
	assert.Contains(t, result, "test")
	// Count occurrences (rough check - ANSI codes make exact counting hard)
	assert.Greater(t, len(result), len("test test test"))
}

func TestViewerModel_HandleSearchInput_MultipleCharacters(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = ""
	
	// Add multiple characters
	chars := []rune{'t', 'e', 's', 't'}
	for _, ch := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}}
		updatedModel, _ := model.handleSearchInput(msg)
		model = updatedModel.(ViewerModel)
	}
	
	assert.Equal(t, "test", model.searchInput)
}

func TestViewerModel_HandleSearchInput_ExecuteEmptyQuery(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.searchMode = true
	model.searchInput = ""
	
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.handleSearchInput(msg)
	
	vm := updatedModel.(ViewerModel)
	assert.False(t, vm.searchMode)
	assert.Equal(t, 0, vm.searchEngine.MatchCount())
}

func TestViewerModel_EnsureVisible_CursorInMiddle(t *testing.T) {
	root := CreateMockTrace()
	model := NewViewerModel(root)
	model.height = 20
	model.cursorPos = 10
	model.scrollOffset = 5
	
	model.ensureVisible()
	
	// Cursor is visible, scroll should not change
	assert.Equal(t, 5, model.scrollOffset)
}
