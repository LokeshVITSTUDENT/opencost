package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConversationManager manages conversation state and context for AI agents
// interacting with OpenCost data. It provides intelligent context tracking,
// preference learning, and conversation continuity.
type ConversationManager struct {
	// conversations stores active conversation contexts
	conversations map[string]*ConversationContext
	
	// conversationMutex protects concurrent access to conversations
	conversationMutex sync.RWMutex
	
	// maxConversationAge is the maximum age before a conversation is considered stale
	maxConversationAge time.Duration
	
	// maxConversations is the maximum number of concurrent conversations to maintain
	maxConversations int
	
	// cleanupInterval defines how often to clean up stale conversations
	cleanupInterval time.Duration
	
	// stopCleanup channel to stop the cleanup goroutine
	stopCleanup chan struct{}
}

// NewConversationManager creates a new conversation manager with default settings
func NewConversationManager() *ConversationManager {
	cm := &ConversationManager{
		conversations:      make(map[string]*ConversationContext),
		maxConversationAge: 24 * time.Hour, // Conversations expire after 24 hours
		maxConversations:   1000,           // Maximum 1000 concurrent conversations
		cleanupInterval:    time.Hour,      // Clean up every hour
		stopCleanup:        make(chan struct{}),
	}
	
	// Start the cleanup goroutine
	go cm.cleanupRoutine()
	
	return cm
}

// StartConversation creates a new conversation context for an AI agent
func (cm *ConversationManager) StartConversation(ctx context.Context) (*ConversationContext, error) {
	cm.conversationMutex.Lock()
	defer cm.conversationMutex.Unlock()
	
	// Check if we're at the conversation limit
	if len(cm.conversations) >= cm.maxConversations {
		// Remove the oldest conversation
		cm.removeOldestConversation()
	}
	
	// Create new conversation context
	sessionID := uuid.New().String()
	conversation := &ConversationContext{
		SessionID:           sessionID,
		ConversationHistory: make([]ConversationEntry, 0),
		UserPreferences:     UserPreferences{},
		CurrentContext:      QueryContext{},
		CreatedAt:           time.Now(),
		LastActivity:        time.Now(),
	}
	
	cm.conversations[sessionID] = conversation
	
	return conversation, nil
}

// GetConversation retrieves an existing conversation context by session ID
func (cm *ConversationManager) GetConversation(sessionID string) (*ConversationContext, error) {
	cm.conversationMutex.RLock()
	defer cm.conversationMutex.RUnlock()
	
	conversation, exists := cm.conversations[sessionID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", sessionID)
	}
	
	// Update last activity
	conversation.LastActivity = time.Now()
	
	return conversation, nil
}

// AddConversationEntry adds a new entry to the conversation history
func (cm *ConversationManager) AddConversationEntry(sessionID string, entry ConversationEntry) error {
	cm.conversationMutex.Lock()
	defer cm.conversationMutex.Unlock()
	
	conversation, exists := cm.conversations[sessionID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", sessionID)
	}
	
	entry.Timestamp = time.Now()
	conversation.ConversationHistory = append(conversation.ConversationHistory, entry)
	conversation.LastActivity = time.Now()
	
	// Limit conversation history to prevent memory bloat
	maxHistoryEntries := 100
	if len(conversation.ConversationHistory) > maxHistoryEntries {
		conversation.ConversationHistory = conversation.ConversationHistory[len(conversation.ConversationHistory)-maxHistoryEntries:]
	}
	
	// Learn from this interaction to update user preferences
	cm.learnFromInteraction(conversation, entry)
	
	return nil
}

// UpdateUserPreferences updates user preferences based on explicit feedback or learned behavior
func (cm *ConversationManager) UpdateUserPreferences(sessionID string, preferences UserPreferences) error {
	cm.conversationMutex.Lock()
	defer cm.conversationMutex.Unlock()
	
	conversation, exists := cm.conversations[sessionID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", sessionID)
	}
	
	// Merge new preferences with existing ones
	cm.mergePreferences(&conversation.UserPreferences, preferences)
	conversation.LastActivity = time.Now()
	
	return nil
}

// GetSuggestedQueries generates suggested follow-up queries based on conversation history
func (cm *ConversationManager) GetSuggestedQueries(sessionID string) ([]string, error) {
	cm.conversationMutex.RLock()
	defer cm.conversationMutex.RUnlock()
	
	conversation, exists := cm.conversations[sessionID]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", sessionID)
	}
	
	return cm.generateSuggestedQueries(conversation), nil
}

// EndConversation closes a conversation and optionally saves it for analytics
func (cm *ConversationManager) EndConversation(sessionID string) error {
	cm.conversationMutex.Lock()
	defer cm.conversationMutex.Unlock()
	
	conversation, exists := cm.conversations[sessionID]
	if !exists {
		return fmt.Errorf("conversation not found: %s", sessionID)
	}
	
	// Here you might save conversation data for analytics
	// cm.saveConversationAnalytics(conversation)
	
	delete(cm.conversations, sessionID)
	
	return nil
}

// GetConversationStats returns statistics about active conversations
func (cm *ConversationManager) GetConversationStats() ConversationStats {
	cm.conversationMutex.RLock()
	defer cm.conversationMutex.RUnlock()
	
	var totalEntries int
	oldestConversation := time.Now()
	
	for _, conv := range cm.conversations {
		totalEntries += len(conv.ConversationHistory)
		if conv.CreatedAt.Before(oldestConversation) {
			oldestConversation = conv.CreatedAt
		}
	}
	
	return ConversationStats{
		ActiveConversations: len(cm.conversations),
		TotalEntries:       totalEntries,
		OldestConversation: oldestConversation,
		AverageEntriesPerConversation: float64(totalEntries) / float64(len(cm.conversations)),
	}
}

// ConversationStats provides statistics about conversation management
type ConversationStats struct {
	ActiveConversations           int       `json:"activeConversations"`
	TotalEntries                 int       `json:"totalEntries"`
	OldestConversation           time.Time `json:"oldestConversation"`
	AverageEntriesPerConversation float64   `json:"averageEntriesPerConversation"`
}

// cleanupRoutine periodically cleans up stale conversations
func (cm *ConversationManager) cleanupRoutine() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cm.cleanupStaleConversations()
		case <-cm.stopCleanup:
			return
		}
	}
}

// cleanupStaleConversations removes conversations that haven't been active recently
func (cm *ConversationManager) cleanupStaleConversations() {
	cm.conversationMutex.Lock()
	defer cm.conversationMutex.Unlock()
	
	now := time.Now()
	staleThreshold := now.Add(-cm.maxConversationAge)
	
	for sessionID, conversation := range cm.conversations {
		if conversation.LastActivity.Before(staleThreshold) {
			delete(cm.conversations, sessionID)
		}
	}
}

// removeOldestConversation removes the conversation with the oldest last activity
func (cm *ConversationManager) removeOldestConversation() {
	if len(cm.conversations) == 0 {
		return
	}
	
	var oldestSessionID string
	var oldestTime time.Time = time.Now()
	
	for sessionID, conversation := range cm.conversations {
		if oldestSessionID == "" || conversation.LastActivity.Before(oldestTime) {
			oldestSessionID = sessionID
			oldestTime = conversation.LastActivity
		}
	}
	
	if oldestSessionID != "" {
		delete(cm.conversations, oldestSessionID)
	}
}

// learnFromInteraction analyzes conversation entries to learn user preferences
func (cm *ConversationManager) learnFromInteraction(conversation *ConversationContext, entry ConversationEntry) {
	intent := entry.ParsedIntent
	
	// Learn time range preferences
	if intent.TimeRange != nil {
		duration := intent.TimeRange.Duration()
		if duration > 0 {
			// Infer preferred time range based on query patterns
			switch {
			case duration <= 24*time.Hour:
				conversation.UserPreferences.PreferredTimeRange = "1d"
			case duration <= 7*24*time.Hour:
				conversation.UserPreferences.PreferredTimeRange = "7d"
			case duration <= 30*24*time.Hour:
				conversation.UserPreferences.PreferredTimeRange = "30d"
			default:
				conversation.UserPreferences.PreferredTimeRange = "30d"
			}
		}
	}
	
	// Learn aggregation preferences
	if len(intent.Aggregation) > 0 {
		conversation.UserPreferences.PreferredAggregation = intent.Aggregation
	}
	
	// Learn format preferences from response format requests
	if intent.ResponseFormat != "" {
		conversation.UserPreferences.PreferredFormat = intent.ResponseFormat
	}
	
	// Learn idle cost preferences
	conversation.UserPreferences.IncludeIdle = intent.IncludeIdle
	
	// Update current context
	conversation.CurrentContext.LastTimeRange = intent.TimeRange
	conversation.CurrentContext.LastDataType = intent.DataType
	if intent.Filters != nil {
		conversation.CurrentContext.ActiveFilters = intent.Filters
	}
}

// mergePreferences merges new preferences with existing ones, giving priority to newer values
func (cm *ConversationManager) mergePreferences(existing *UserPreferences, new UserPreferences) {
	if new.PreferredTimeRange != "" {
		existing.PreferredTimeRange = new.PreferredTimeRange
	}
	if len(new.PreferredClusters) > 0 {
		existing.PreferredClusters = new.PreferredClusters
	}
	if len(new.PreferredNamespaces) > 0 {
		existing.PreferredNamespaces = new.PreferredNamespaces
	}
	if new.PreferredCurrency != "" {
		existing.PreferredCurrency = new.PreferredCurrency
	}
	if len(new.PreferredAggregation) > 0 {
		existing.PreferredAggregation = new.PreferredAggregation
	}
	if new.PreferredFormat != "" {
		existing.PreferredFormat = new.PreferredFormat
	}
	if new.CostThresholds.HighCostDaily > 0 {
		existing.CostThresholds.HighCostDaily = new.CostThresholds.HighCostDaily
	}
	if new.CostThresholds.HighCostHourly > 0 {
		existing.CostThresholds.HighCostHourly = new.CostThresholds.HighCostHourly
	}
	if new.CostThresholds.HighCostMonthly > 0 {
		existing.CostThresholds.HighCostMonthly = new.CostThresholds.HighCostMonthly
	}
	
	// Update boolean preferences (these might need more sophisticated logic)
	existing.IncludeIdle = new.IncludeIdle
}

// generateSuggestedQueries creates contextually relevant follow-up questions
func (cm *ConversationManager) generateSuggestedQueries(conversation *ConversationContext) []string {
	suggestions := make([]string, 0)
	
	// Get the last few entries to understand context
	historyLen := len(conversation.ConversationHistory)
	if historyLen == 0 {
		// Return general getting-started suggestions
		return []string{
			"Show me the total cost breakdown for the last 7 days",
			"What are the most expensive namespaces this month?",
			"Display asset costs for our production clusters",
			"Compare cloud costs between AWS and Azure",
		}
	}
	
	lastEntry := conversation.ConversationHistory[historyLen-1]
	lastIntent := lastEntry.ParsedIntent
	
	switch lastIntent.DataType {
	case "allocations":
		suggestions = append(suggestions, []string{
			"Show me the efficiency metrics for these allocations",
			"Compare these costs with the previous period",
			"Which pods are driving the highest costs?",
			"Show me cost recommendations for optimization",
		}...)
		
	case "assets":
		suggestions = append(suggestions, []string{
			"What is the utilization of these assets?",
			"Show me underutilized resources that could be optimized",
			"Display the capacity planning analysis",
			"Compare asset costs across different clusters",
		}...)
		
	case "cloudcosts":
		suggestions = append(suggestions, []string{
			"Break down costs by cloud service",
			"Show me the cost forecast for next month",
			"Which cloud accounts have the highest spend?",
			"Display cost anomalies and unusual spending patterns",
		}...)
	}
	
	// Add time-based suggestions
	if lastIntent.TimeRange != nil {
		duration := lastIntent.TimeRange.Duration()
		if duration <= 24*time.Hour {
			suggestions = append(suggestions, "Expand the time range to see weekly trends")
		} else if duration <= 7*24*time.Hour {
			suggestions = append(suggestions, "Show me the monthly view for better trend analysis")
		}
	}
	
	// Add context-aware suggestions based on current filters
	if len(conversation.CurrentContext.ActiveFilters) > 0 {
		suggestions = append(suggestions, "Remove filters to see the complete picture")
		suggestions = append(suggestions, "Apply additional filters to drill down further")
	}
	
	// Limit suggestions to avoid overwhelming the user
	maxSuggestions := 5
	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}
	
	return suggestions
}

// Stop gracefully shuts down the conversation manager
func (cm *ConversationManager) Stop() {
	close(cm.stopCleanup)
}