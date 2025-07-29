package mcp

import (
	"time"

	"github.com/opencost/opencost/core/pkg/opencost"
)

// ConversationContext maintains the state and context for an ongoing conversation
// with an AI agent about OpenCost data. It tracks conversation history, user preferences,
// and provides intelligent defaults based on previous queries.
type ConversationContext struct {
	// SessionID uniquely identifies this conversation session
	SessionID string `json:"sessionId"`
	
	// ConversationHistory tracks previous queries and responses for context
	ConversationHistory []ConversationEntry `json:"conversationHistory"`
	
	// UserPreferences stores learned preferences from the conversation
	UserPreferences UserPreferences `json:"userPreferences"`
	
	// CurrentContext tracks the current state of the conversation
	CurrentContext QueryContext `json:"currentContext"`
	
	// CreatedAt tracks when this conversation started
	CreatedAt time.Time `json:"createdAt"`
	
	// LastActivity tracks the last interaction time
	LastActivity time.Time `json:"lastActivity"`
}

// ConversationEntry represents a single exchange in the conversation
type ConversationEntry struct {
	// Timestamp when this entry was created
	Timestamp time.Time `json:"timestamp"`
	
	// UserQuery is the original user query/request
	UserQuery string `json:"userQuery"`
	
	// ParsedIntent represents the interpreted intent and parameters
	ParsedIntent QueryIntent `json:"parsedIntent"`
	
	// Response is the data/answer provided
	Response interface{} `json:"response"`
	
	// ResponseSummary is a human-readable summary of the response
	ResponseSummary string `json:"responseSummary"`
	
	// Feedback tracks if the user was satisfied with this response
	Feedback *ResponseFeedback `json:"feedback,omitempty"`
}

// UserPreferences captures learned preferences about how the user likes to view cost data
type UserPreferences struct {
	// PreferredTimeRange is the user's typical time window for queries
	PreferredTimeRange string `json:"preferredTimeRange,omitempty"` // e.g., "7d", "30d", "1d"
	
	// PreferredClusters are clusters the user frequently asks about
	PreferredClusters []string `json:"preferredClusters,omitempty"`
	
	// PreferredNamespaces are namespaces the user frequently asks about
	PreferredNamespaces []string `json:"preferredNamespaces,omitempty"`
	
	// PreferredCurrency for cost display
	PreferredCurrency string `json:"preferredCurrency,omitempty"`
	
	// PreferredAggregation is how the user likes data grouped (namespace, cluster, service, etc.)
	PreferredAggregation []string `json:"preferredAggregation,omitempty"`
	
	// PreferredFormat indicates if user prefers summaries, details, or specific formats
	PreferredFormat string `json:"preferredFormat,omitempty"` // "summary", "detailed", "table", "chart"
	
	// CostThresholds for highlighting expensive resources
	CostThresholds CostThresholds `json:"costThresholds,omitempty"`
	
	// IncludeIdle indicates if user typically wants idle costs included
	IncludeIdle bool `json:"includeIdle"`
}

// CostThresholds define what the user considers expensive
type CostThresholds struct {
	// HighCostDaily is the daily cost threshold for flagging expensive resources
	HighCostDaily float64 `json:"highCostDaily,omitempty"`
	
	// HighCostHourly is the hourly cost threshold
	HighCostHourly float64 `json:"highCostHourly,omitempty"`
	
	// HighCostMonthly is the monthly cost threshold
	HighCostMonthly float64 `json:"highCostMonthly,omitempty"`
}

// QueryContext tracks the current conversation state
type QueryContext struct {
	// LastTimeRange used in the previous query
	LastTimeRange *opencost.Window `json:"lastTimeRange,omitempty"`
	
	// LastClusters queried
	LastClusters []string `json:"lastClusters,omitempty"`
	
	// LastNamespaces queried
	LastNamespaces []string `json:"lastNamespaces,omitempty"`
	
	// LastDataType queried (allocations, assets, cloudcosts)
	LastDataType string `json:"lastDataType,omitempty"`
	
	// ActiveFilters currently applied
	ActiveFilters map[string]string `json:"activeFilters,omitempty"`
}

// ResponseFeedback captures user satisfaction with responses
type ResponseFeedback struct {
	// Helpful indicates if the response was useful
	Helpful bool `json:"helpful"`
	
	// Comments provide additional feedback
	Comments string `json:"comments,omitempty"`
	
	// SuggestedImprovements for better responses
	SuggestedImprovements string `json:"suggestedImprovements,omitempty"`
}

// QueryIntent represents the parsed and interpreted user intent
type QueryIntent struct {
	// DataType indicates what type of data is requested (allocations, assets, cloudcosts)
	DataType string `json:"dataType"`
	
	// TimeRange for the query
	TimeRange *opencost.Window `json:"timeRange,omitempty"`
	
	// Aggregation specifies how to group the data
	Aggregation []string `json:"aggregation,omitempty"`
	
	// Filters to apply to the data
	Filters map[string]string `json:"filters,omitempty"`
	
	// IncludeIdle whether to include idle costs
	IncludeIdle bool `json:"includeIdle"`
	
	// SortBy specifies how to sort results
	SortBy string `json:"sortBy,omitempty"`
	
	// Limit specifies maximum number of results
	Limit int `json:"limit,omitempty"`
	
	// ResponseFormat specifies the desired response format
	ResponseFormat string `json:"responseFormat,omitempty"`
	
	// ComparisonType for trend analysis (previous_period, year_over_year, etc.)
	ComparisonType string `json:"comparisonType,omitempty"`
}