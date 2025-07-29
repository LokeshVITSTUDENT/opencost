package mcp

import (
	"testing"
	"context"
	"time"
)

func TestConversationManager(t *testing.T) {
	cm := NewConversationManager()
	defer cm.Stop()
	
	// Test starting a conversation
	ctx := context.Background()
	conversation, err := cm.StartConversation(ctx)
	if err != nil {
		t.Fatalf("Failed to start conversation: %v", err)
	}
	
	if conversation.SessionID == "" {
		t.Error("Session ID should not be empty")
	}
	
	if conversation.CreatedAt.IsZero() {
		t.Error("Created time should be set")
	}
	
	// Test retrieving a conversation
	retrieved, err := cm.GetConversation(conversation.SessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve conversation: %v", err)
	}
	
	if retrieved.SessionID != conversation.SessionID {
		t.Error("Retrieved conversation should have same session ID")
	}
	
	// Test adding conversation entry
	entry := ConversationEntry{
		UserQuery: "Show me allocation costs for the last 7 days",
		ParsedIntent: QueryIntent{
			DataType: "allocations",
			ResponseFormat: "summary",
		},
		Response: "Sample response",
		ResponseSummary: "Cost data for 7 days",
	}
	
	err = cm.AddConversationEntry(conversation.SessionID, entry)
	if err != nil {
		t.Fatalf("Failed to add conversation entry: %v", err)
	}
	
	// Verify the entry was added
	updated, err := cm.GetConversation(conversation.SessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated conversation: %v", err)
	}
	
	if len(updated.ConversationHistory) != 1 {
		t.Errorf("Expected 1 conversation entry, got %d", len(updated.ConversationHistory))
	}
	
	if updated.ConversationHistory[0].UserQuery != entry.UserQuery {
		t.Error("Conversation entry user query doesn't match")
	}
}

func TestUserPreferenceLearning(t *testing.T) {
	cm := NewConversationManager()
	defer cm.Stop()
	
	ctx := context.Background()
	conversation, err := cm.StartConversation(ctx)
	if err != nil {
		t.Fatalf("Failed to start conversation: %v", err)
	}
	
	// Simulate user interactions that should update preferences
	entries := []ConversationEntry{
		{
			UserQuery: "Show me costs for last 7 days by namespace",
			ParsedIntent: QueryIntent{
				DataType: "allocations",
				Aggregation: []string{"namespace"},
				ResponseFormat: "summary",
			},
		},
		{
			UserQuery: "Show me weekly costs by namespace",
			ParsedIntent: QueryIntent{
				DataType: "allocations",
				Aggregation: []string{"namespace"},
				ResponseFormat: "summary",
			},
		},
	}
	
	for _, entry := range entries {
		err = cm.AddConversationEntry(conversation.SessionID, entry)
		if err != nil {
			t.Fatalf("Failed to add conversation entry: %v", err)
		}
	}
	
	// Check that preferences were learned
	updated, err := cm.GetConversation(conversation.SessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve conversation: %v", err)
	}
	
	prefs := updated.UserPreferences
	
	// Should have learned namespace aggregation preference
	if len(prefs.PreferredAggregation) == 0 || prefs.PreferredAggregation[0] != "namespace" {
		t.Error("Should have learned namespace aggregation preference")
	}
	
	// Should have learned summary format preference
	if prefs.PreferredFormat != "summary" {
		t.Error("Should have learned summary format preference")
	}
}

func TestTimeRangeParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1d", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"30d", 30 * 24 * time.Hour, false},
		{"1h", time.Hour, false},
		{"24h", 24 * time.Hour, false},
		{"invalid", 0, true},
	}
	
	for _, test := range tests {
		window, err := parseTimeRange(test.input)
		
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s", test.input)
			}
			continue
		}
		
		if err != nil {
			t.Errorf("Unexpected error for input %s: %v", test.input, err)
			continue
		}
		
		if window == nil {
			t.Errorf("Window should not be nil for input %s", test.input)
			continue
		}
		
		duration := window.Duration()
		if duration != test.expected {
			t.Errorf("Expected duration %v for input %s, got %v", test.expected, test.input, duration)
		}
	}
}

func TestConversationStats(t *testing.T) {
	cm := NewConversationManager()
	defer cm.Stop()
	
	ctx := context.Background()
	
	// Start multiple conversations
	conv1, _ := cm.StartConversation(ctx)
	conv2, _ := cm.StartConversation(ctx)
	
	// Add entries to conversations
	entry := ConversationEntry{
		UserQuery: "Test query",
		ParsedIntent: QueryIntent{DataType: "allocations"},
	}
	
	cm.AddConversationEntry(conv1.SessionID, entry)
	cm.AddConversationEntry(conv1.SessionID, entry)
	cm.AddConversationEntry(conv2.SessionID, entry)
	
	stats := cm.GetConversationStats()
	
	if stats.ActiveConversations != 2 {
		t.Errorf("Expected 2 active conversations, got %d", stats.ActiveConversations)
	}
	
	if stats.TotalEntries != 3 {
		t.Errorf("Expected 3 total entries, got %d", stats.TotalEntries)
	}
	
	expectedAvg := 1.5 // 3 entries / 2 conversations
	if stats.AverageEntriesPerConversation != expectedAvg {
		t.Errorf("Expected average %.1f, got %.1f", expectedAvg, stats.AverageEntriesPerConversation)
	}
}