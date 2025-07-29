package mcp

import (
	"testing"
	"time"
)

// Test that our core types can be instantiated and used
func TestCoreTypes(t *testing.T) {
	// Test ConversationContext
	ctx := ConversationContext{
		SessionID:           "test-123",
		ConversationHistory: make([]ConversationEntry, 0),
		CreatedAt:           time.Now(),
		LastActivity:        time.Now(),
	}
	
	if ctx.SessionID != "test-123" {
		t.Error("SessionID not set correctly")
	}
	
	// Test UserPreferences
	prefs := UserPreferences{
		PreferredTimeRange: "7d",
		PreferredFormat:    "summary",
		IncludeIdle:        true,
	}
	
	if prefs.PreferredTimeRange != "7d" {
		t.Error("PreferredTimeRange not set correctly")
	}
	
	// Test QueryIntent
	intent := QueryIntent{
		DataType:       "allocations",
		ResponseFormat: "summary",
		IncludeIdle:    true,
	}
	
	if intent.DataType != "allocations" {
		t.Error("DataType not set correctly")
	}
	
	// Test ConversationEntry
	entry := ConversationEntry{
		Timestamp:       time.Now(),
		UserQuery:       "Show me costs",
		ParsedIntent:    intent,
		Response:        "Cost data here",
		ResponseSummary: "Summary of costs",
	}
	
	if entry.UserQuery != "Show me costs" {
		t.Error("UserQuery not set correctly")
	}
	
	ctx.ConversationHistory = append(ctx.ConversationHistory, entry)
	
	if len(ctx.ConversationHistory) != 1 {
		t.Error("ConversationHistory not updated correctly")
	}
}

func TestQueryContext(t *testing.T) {
	qctx := QueryContext{
		LastDataType: "allocations",
		ActiveFilters: map[string]string{
			"cluster": "prod",
			"namespace": "default",
		},
	}
	
	if qctx.LastDataType != "allocations" {
		t.Error("LastDataType not set correctly")
	}
	
	if qctx.ActiveFilters["cluster"] != "prod" {
		t.Error("ActiveFilters not set correctly")
	}
}

func TestCostThresholds(t *testing.T) {
	thresholds := CostThresholds{
		HighCostDaily:   100.0,
		HighCostHourly:  5.0,
		HighCostMonthly: 3000.0,
	}
	
	if thresholds.HighCostDaily != 100.0 {
		t.Error("HighCostDaily not set correctly")
	}
	
	if thresholds.HighCostHourly != 5.0 {
		t.Error("HighCostHourly not set correctly")
	}
}

func TestResponseFeedback(t *testing.T) {
	feedback := ResponseFeedback{
		Helpful:               true,
		Comments:              "Very useful cost breakdown",
		SuggestedImprovements: "Could include more historical data",
	}
	
	if !feedback.Helpful {
		t.Error("Helpful not set correctly")
	}
	
	if feedback.Comments != "Very useful cost breakdown" {
		t.Error("Comments not set correctly")
	}
}