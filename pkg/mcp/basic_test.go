package mcp

import (
	"testing"
	"context"
	"time"
	"strings"
)

// Simple test for conversation management basics
func TestBasicConversationManagement(t *testing.T) {
	// Test conversation context creation
	ctx := &ConversationContext{
		SessionID:           "test-session-123",
		ConversationHistory: make([]ConversationEntry, 0),
		UserPreferences:     UserPreferences{},
		CurrentContext:      QueryContext{},
		CreatedAt:           time.Now(),
		LastActivity:        time.Now(),
	}
	
	if ctx.SessionID != "test-session-123" {
		t.Error("Session ID not set correctly")
	}
	
	// Test adding conversation entry
	entry := ConversationEntry{
		Timestamp:       time.Now(),
		UserQuery:       "Show me allocation costs",
		ParsedIntent:    QueryIntent{DataType: "allocations"},
		Response:        "Sample cost data",
		ResponseSummary: "Cost breakdown provided",
	}
	
	ctx.ConversationHistory = append(ctx.ConversationHistory, entry)
	
	if len(ctx.ConversationHistory) != 1 {
		t.Error("Conversation entry not added correctly")
	}
	
	if ctx.ConversationHistory[0].UserQuery != "Show me allocation costs" {
		t.Error("User query not stored correctly")
	}
}

// Test user preferences structure
func TestUserPreferences(t *testing.T) {
	prefs := UserPreferences{
		PreferredTimeRange:    "7d",
		PreferredClusters:     []string{"prod", "dev"},
		PreferredNamespaces:   []string{"default", "kube-system"},
		PreferredCurrency:     "USD",
		PreferredAggregation:  []string{"namespace"},
		PreferredFormat:       "summary",
		IncludeIdle:          true,
		CostThresholds: CostThresholds{
			HighCostDaily:   100.0,
			HighCostHourly:  5.0,
			HighCostMonthly: 3000.0,
		},
	}
	
	if prefs.PreferredTimeRange != "7d" {
		t.Error("Preferred time range not set correctly")
	}
	
	if len(prefs.PreferredClusters) != 2 {
		t.Error("Preferred clusters not set correctly")
	}
	
	if prefs.CostThresholds.HighCostDaily != 100.0 {
		t.Error("Cost thresholds not set correctly")
	}
}

// Test query intent parsing
func TestQueryIntent(t *testing.T) {
	intent := QueryIntent{
		DataType:       "allocations",
		Aggregation:    []string{"namespace", "cluster"},
		Filters:        map[string]string{"cluster": "prod"},
		IncludeIdle:    true,
		SortBy:         "cost",
		Limit:          10,
		ResponseFormat: "detailed",
	}
	
	if intent.DataType != "allocations" {
		t.Error("Data type not set correctly")
	}
	
	if len(intent.Aggregation) != 2 {
		t.Error("Aggregation not set correctly")
	}
	
	if intent.Filters["cluster"] != "prod" {
		t.Error("Filters not set correctly")
	}
}

// Test allocation query structure
func TestAllocationQuery(t *testing.T) {
	query := AllocationQuery{
		Aggregate:                               []string{"namespace"},
		Filter:                                  "cluster:prod",
		IncludeIdle:                            true,
		IncludeProportionalAssetResourceCosts:   true,
		IncludeAggregatedMetadata:              true,
		ShareIdle:                              false,
		ResponseFormat:                         "summary",
		TopN:                                   10,
		IncludeRecommendations:                 true,
		CompareWithPrevious:                    true,
		HighlightAnomalies:                     true,
	}
	
	if len(query.Aggregate) != 1 || query.Aggregate[0] != "namespace" {
		t.Error("Aggregate not set correctly")
	}
	
	if !strings.Contains(query.Filter, "cluster:prod") {
		t.Error("Filter not set correctly")
	}
	
	if !query.IncludeRecommendations {
		t.Error("Include recommendations not set correctly")
	}
}

// Test asset query structure
func TestAssetQuery(t *testing.T) {
	query := AssetQuery{
		Filter:                 "type:node",
		AssetTypes:            []string{"node", "disk"},
		ResponseFormat:        "insights",
		TopN:                  5,
		IncludeRecommendations: true,
		IncludeUtilization:    true,
		HighlightUnderutilized: true,
		IncludeCapacityPlanning: true,
	}
	
	if len(query.AssetTypes) != 2 {
		t.Error("Asset types not set correctly")
	}
	
	if query.ResponseFormat != "insights" {
		t.Error("Response format not set correctly")
	}
	
	if !query.IncludeCapacityPlanning {
		t.Error("Include capacity planning not set correctly")
	}
}

// Test cloud cost query structure
func TestCloudCostQuery(t *testing.T) {
	query := CloudCostQuery{
		Aggregate:              []string{"service", "provider"},
		Filter:                 "provider:aws",
		Providers:             []string{"aws", "gcp"},
		Services:              []string{"EC2", "S3"},
		ResponseFormat:        "comparison",
		TopN:                  20,
		IncludeRecommendations: true,
		IncludeForecast:       true,
		HighlightAnomalies:    true,
		ShowCostBreakdown:     true,
	}
	
	if len(query.Aggregate) != 2 {
		t.Error("Aggregate not set correctly")
	}
	
	if len(query.Providers) != 2 {
		t.Error("Providers not set correctly")
	}
	
	if !query.IncludeForecast {
		t.Error("Include forecast not set correctly")
	}
}

// Test cost insight structure
func TestCostInsight(t *testing.T) {
	insight := CostInsight{
		Type:              "anomaly",
		Severity:          "high",
		Title:             "Unusual cost spike detected",
		Description:       "Pod costs increased by 300% in the last hour",
		ImpactCost:        500.0,
		PotentialSavings:  200.0,
		AffectedResources: []string{"pod/app-1", "pod/app-2"},
		RecommendedAction: "Check for resource leak or scale down",
		MetricsBefore:     map[string]float64{"cost_per_hour": 10.0},
		MetricsAfter:      map[string]float64{"cost_per_hour": 40.0},
	}
	
	if insight.Type != "anomaly" {
		t.Error("Insight type not set correctly")
	}
	
	if insight.Severity != "high" {
		t.Error("Severity not set correctly")
	}
	
	if insight.ImpactCost != 500.0 {
		t.Error("Impact cost not set correctly")
	}
	
	if len(insight.AffectedResources) != 2 {
		t.Error("Affected resources not set correctly")
	}
}

// Test cost recommendation structure  
func TestCostRecommendation(t *testing.T) {
	recommendation := CostRecommendation{
		Title:                   "Optimize pod resource requests",
		Description:             "Reduce CPU and memory requests for underutilized pods",
		Category:               "rightsizing",
		PotentialMonthlySavings: 1500.0,
		ImplementationComplexity: "medium",
		Actions: []string{
			"Analyze pod utilization metrics",
			"Reduce CPU requests by 50%",
			"Reduce memory requests by 30%",
		},
		Prerequisites:     []string{"Monitoring system in place"},
		EstimatedTimeDays: 3,
		AffectedResources: []string{"namespace/frontend"},
		SupportingData: map[string]interface{}{
			"avg_cpu_utilization": 25.0,
			"avg_memory_utilization": 40.0,
		},
	}
	
	if recommendation.Category != "rightsizing" {
		t.Error("Category not set correctly")
	}
	
	if recommendation.PotentialMonthlySavings != 1500.0 {
		t.Error("Potential savings not set correctly")
	}
	
	if len(recommendation.Actions) != 3 {
		t.Error("Actions not set correctly")
	}
}