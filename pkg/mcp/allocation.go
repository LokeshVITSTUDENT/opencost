package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opencost/opencost/core/pkg/opencost"
	"github.com/opencost/opencost/pkg/costmodel"
)

// AllocationQueryHandler manages allocation-related queries and maintains conversation context
// for AI agents requesting Kubernetes cost allocation data.
type AllocationQueryHandler struct {
	// Model provides access to the OpenCost allocation computation engine
	Model costmodel.AllocationModel
	
	// ConversationManager handles conversation state and history
	ConversationManager *ConversationManager
}

// AllocationQuery represents a comprehensive allocation query with AI-friendly parameters
type AllocationQuery struct {
	// Base query parameters
	TimeRange *opencost.Window `json:"timeRange"`
	Step      time.Duration    `json:"step,omitempty"`
	
	// Aggregation and grouping
	Aggregate []string `json:"aggregate,omitempty"` // cluster, namespace, controllerName, controller, service, etc.
	
	// Filtering
	Filter string `json:"filter,omitempty"` // OpenCost filter string
	
	// Advanced options
	Accumulate                              string                    `json:"accumulate,omitempty"`
	IncludeIdle                            bool                      `json:"includeIdle"`
	IdleByNode                             bool                      `json:"idleByNode"`
	IncludeProportionalAssetResourceCosts  bool                      `json:"includeProportionalAssetResourceCosts"`
	IncludeAggregatedMetadata              bool                      `json:"includeAggregatedMetadata"`
	ShareIdle                              bool                      `json:"shareIdle"`
	ShareTenancyCosts                      bool                      `json:"shareTenancyCosts"`
	
	// AI-specific enhancements
	ResponseFormat      string            `json:"responseFormat,omitempty"`      // "summary", "detailed", "insights", "comparison"
	CostThreshold       *float64          `json:"costThreshold,omitempty"`       // Only show results above this cost
	TopN                int               `json:"topN,omitempty"`                // Limit to top N results by cost
	IncludeRecommendations bool           `json:"includeRecommendations"`        // Include cost optimization recommendations
	CompareWithPrevious bool              `json:"compareWithPrevious"`           // Compare with previous period
	HighlightAnomalies  bool              `json:"highlightAnomalies"`            // Flag unusual cost patterns
	
	// Context from conversation
	ConversationContext *ConversationContext `json:"conversationContext,omitempty"`
}

// AllocationResponse provides AI-friendly allocation data with insights and context
type AllocationResponse struct {
	// Core data
	Data *opencost.AllocationSetRange `json:"data"`
	
	// Summary and insights for AI consumption
	Summary         AllocationSummary      `json:"summary"`
	Insights        []CostInsight          `json:"insights,omitempty"`
	Recommendations []CostRecommendation   `json:"recommendations,omitempty"`
	Trends          *TrendAnalysis         `json:"trends,omitempty"`
	
	// Response metadata
	Query            AllocationQuery `json:"query"`
	GeneratedAt      time.Time      `json:"generatedAt"`
	ProcessingTimeMs int64          `json:"processingTimeMs"`
	
	// Conversation context
	SuggestedFollowUps []string `json:"suggestedFollowUps,omitempty"`
}

// AllocationSummary provides high-level insights about allocation data
type AllocationSummary struct {
	// Totals
	TotalCost        float64 `json:"totalCost"`
	TotalCPUCost     float64 `json:"totalCPUCost"`
	TotalRAMCost     float64 `json:"totalRAMCost"`
	TotalGPUCost     float64 `json:"totalGPUCost"`
	TotalStorageCost float64 `json:"totalStorageCost"`
	TotalNetworkCost float64 `json:"totalNetworkCost"`
	
	// Breakdown by resource type
	CPUHours     float64 `json:"cpuHours"`
	RAMByteHours float64 `json:"ramByteHours"`
	GPUHours     float64 `json:"gpuHours"`
	
	// Efficiency metrics
	CPUEfficiency    float64 `json:"cpuEfficiency"`    // % of requested CPU actually used
	RAMEfficiency    float64 `json:"ramEfficiency"`    // % of requested RAM actually used
	
	// Cost distribution
	TopCostDriver      string  `json:"topCostDriver"`      // namespace, service, etc. with highest cost
	TopCostDriverValue float64 `json:"topCostDriverValue"`
	
	// Time-based insights
	CostPerDay   float64 `json:"costPerDay"`
	CostPerHour  float64 `json:"costPerHour"`
	CostTrend    string  `json:"costTrend"`    // "increasing", "decreasing", "stable"
	
	// Count metrics
	TotalAllocations    int `json:"totalAllocations"`
	NamespacesIncluded  int `json:"namespacesIncluded"`
	ClustersIncluded    int `json:"clustersIncluded"`
	
	// Idle cost insights
	IdleCost            float64 `json:"idleCost,omitempty"`
	IdlePercentage      float64 `json:"idlePercentage,omitempty"`
}

// CostInsight represents an AI-generated insight about cost patterns
type CostInsight struct {
	// Type of insight
	Type string `json:"type"` // "anomaly", "trend", "efficiency", "waste", "optimization"
	
	// Severity level
	Severity string `json:"severity"` // "low", "medium", "high", "critical"
	
	// Human-readable description
	Title       string `json:"title"`
	Description string `json:"description"`
	
	// Impact metrics
	ImpactCost        float64 `json:"impactCost,omitempty"`        // Cost impact of this insight
	PotentialSavings  float64 `json:"potentialSavings,omitempty"`  // Potential savings if addressed
	
	// Context
	AffectedResources []string `json:"affectedResources,omitempty"`
	RecommendedAction string   `json:"recommendedAction,omitempty"`
	
	// Supporting data
	MetricsBefore map[string]float64 `json:"metricsBefore,omitempty"`
	MetricsAfter  map[string]float64 `json:"metricsAfter,omitempty"`
}

// CostRecommendation provides actionable optimization suggestions
type CostRecommendation struct {
	// Recommendation details
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"` // "rightsizing", "scheduling", "efficiency", "waste"
	
	// Impact assessment
	PotentialMonthlySavings float64 `json:"potentialMonthlySavings"`
	ImplementationComplexity string `json:"implementationComplexity"` // "low", "medium", "high"
	
	// Implementation guidance
	Actions           []string `json:"actions"`
	Prerequisites     []string `json:"prerequisites,omitempty"`
	EstimatedTimeDays int      `json:"estimatedTimeDays,omitempty"`
	
	// Context
	AffectedResources []string               `json:"affectedResources"`
	SupportingData    map[string]interface{} `json:"supportingData,omitempty"`
}

// TrendAnalysis provides cost trend insights over time
type TrendAnalysis struct {
	// Trend direction
	Direction string `json:"direction"` // "increasing", "decreasing", "stable", "volatile"
	
	// Change metrics
	PercentChange       float64 `json:"percentChange"`       // % change over the period
	AbsoluteChange      float64 `json:"absoluteChange"`      // absolute cost change
	DailyAverageChange  float64 `json:"dailyAverageChange"`  // average daily change
	
	// Comparison periods
	CurrentPeriodCost  float64 `json:"currentPeriodCost"`
	PreviousPeriodCost float64 `json:"previousPeriodCost"`
	
	// Peak and trough analysis
	PeakCost    float64   `json:"peakCost"`
	PeakTime    time.Time `json:"peakTime"`
	TroughCost  float64   `json:"troughCost"`
	TroughTime  time.Time `json:"troughTime"`
	
	// Forecasting
	PredictedNextWeekCost  float64 `json:"predictedNextWeekCost,omitempty"`
	PredictedNextMonthCost float64 `json:"predictedNextMonthCost,omitempty"`
	
	// Pattern insights
	DayOfWeekPattern map[string]float64 `json:"dayOfWeekPattern,omitempty"` // average cost by day of week
	HourOfDayPattern map[int]float64    `json:"hourOfDayPattern,omitempty"` // average cost by hour of day
}

// QueryAllocation executes an allocation query with AI-enhanced response formatting
func (h *AllocationQueryHandler) QueryAllocation(ctx context.Context, query AllocationQuery) (*AllocationResponse, error) {
	startTime := time.Now()
	
	// Apply intelligent defaults based on conversation context
	if err := h.applyConversationDefaults(&query); err != nil {
		return nil, fmt.Errorf("failed to apply conversation defaults: %w", err)
	}
	
	// Execute the core OpenCost allocation query
	step := query.Step
	if step == 0 {
		step = time.Hour // Default to hourly resolution
	}
	
	allocSetRange, err := h.Model.QueryAllocation(
		*query.TimeRange,
		step,
		query.Aggregate,
		query.IncludeIdle,
		query.IdleByNode,
		query.IncludeProportionalAssetResourceCosts,
		query.IncludeAggregatedMetadata,
		false, // shareLoadBalancer - deprecated parameter
		opencost.AccumulateOptionNone, // accumulate option - would need parsing
		query.ShareIdle,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query allocations: %w", err)
	}
	
	// Generate AI-enhanced response
	response := &AllocationResponse{
		Data:             allocSetRange,
		Query:            query,
		GeneratedAt:      time.Now(),
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
	}
	
	// Generate summary and insights
	if err := h.generateSummary(response); err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}
	
	if query.IncludeRecommendations {
		h.generateRecommendations(response)
	}
	
	if query.CompareWithPrevious {
		if err := h.generateTrendAnalysis(response); err != nil {
			return nil, fmt.Errorf("failed to generate trend analysis: %w", err)
		}
	}
	
	if query.HighlightAnomalies {
		h.generateInsights(response)
	}
	
	// Generate suggested follow-up questions
	h.generateFollowUps(response)
	
	// Update conversation context
	if query.ConversationContext != nil {
		h.updateConversationContext(query.ConversationContext, query, response)
	}
	
	return response, nil
}

// applyConversationDefaults intelligently fills in query parameters based on conversation history
func (h *AllocationQueryHandler) applyConversationDefaults(query *AllocationQuery) error {
	if query.ConversationContext == nil {
		return nil
	}
	
	ctx := query.ConversationContext
	prefs := ctx.UserPreferences
	
	// Apply time range defaults
	if query.TimeRange == nil && prefs.PreferredTimeRange != "" {
		window, err := parseTimeRange(prefs.PreferredTimeRange)
		if err == nil {
			query.TimeRange = window
		}
	}
	
	// Apply default time range if still not set
	if query.TimeRange == nil {
		end := time.Now()
		start := end.Add(-7 * 24 * time.Hour) // Default to last 7 days
		query.TimeRange = opencost.NewWindow(&start, &end)
	}
	
	// Apply aggregation defaults
	if len(query.Aggregate) == 0 && len(prefs.PreferredAggregation) > 0 {
		query.Aggregate = prefs.PreferredAggregation
	}
	
	// Apply default aggregation if still not set
	if len(query.Aggregate) == 0 {
		query.Aggregate = []string{"namespace"} // Default to namespace aggregation
	}
	
	// Apply idle cost preferences
	query.IncludeIdle = prefs.IncludeIdle
	
	return nil
}

// parseTimeRange converts a human-readable time range to an opencost.Window
func parseTimeRange(timeRange string) (*opencost.Window, error) {
	end := time.Now()
	var start time.Time
	
	switch strings.ToLower(timeRange) {
	case "1d", "1day", "day":
		start = end.Add(-24 * time.Hour)
	case "7d", "7days", "week":
		start = end.Add(-7 * 24 * time.Hour)
	case "30d", "30days", "month":
		start = end.Add(-30 * 24 * time.Hour)
	case "90d", "90days", "quarter":
		start = end.Add(-90 * 24 * time.Hour)
	case "1h", "1hour", "hour":
		start = end.Add(-time.Hour)
	case "24h":
		start = end.Add(-24 * time.Hour)
	default:
		return nil, fmt.Errorf("unsupported time range: %s", timeRange)
	}
	
	return opencost.NewWindow(&start, &end), nil
}

// Additional helper methods would be implemented here for:
// - generateSummary
// - generateRecommendations  
// - generateTrendAnalysis
// - generateInsights
// - generateFollowUps
// - updateConversationContext

// Placeholder implementations for now
func (h *AllocationQueryHandler) generateSummary(response *AllocationResponse) error {
	// TODO: Implement allocation summary generation
	if response.Data != nil && len(response.Data.Allocations()) > 0 {
		totalCost := 0.0
		for _, alloc := range response.Data.Allocations() {
			for _, allocation := range alloc.Allocations {
				totalCost += allocation.TotalCost()
			}
		}
		response.Summary = AllocationSummary{
			TotalCost:        totalCost,
			TotalAllocations: len(response.Data.Allocations()),
		}
	} else {
		response.Summary = AllocationSummary{
			TotalCost:        0.0,
			TotalAllocations: 0,
		}
	}
	return nil
}

func (h *AllocationQueryHandler) generateRecommendations(response *AllocationResponse) {
	// TODO: Implement allocation recommendations
	response.Recommendations = []CostRecommendation{
		{
			Title:                   "Example Recommendation",
			Description:             "This is a placeholder recommendation",
			PotentialMonthlySavings: 50.0,
		},
	}
}

func (h *AllocationQueryHandler) generateTrendAnalysis(response *AllocationResponse) error {
	// TODO: Implement allocation trend analysis
	response.Trends = &TrendAnalysis{
		Direction:      "stable",
		PercentChange:  0.0,
		AbsoluteChange: 0.0,
	}
	return nil
}

func (h *AllocationQueryHandler) generateInsights(response *AllocationResponse) {
	// TODO: Implement allocation insights
	response.Insights = []CostInsight{
		{
			Type:        "info",
			Severity:    "low",
			Title:       "Example Insight",
			Description: "This is a placeholder insight",
		},
	}
}

func (h *AllocationQueryHandler) generateFollowUps(response *AllocationResponse) {
	// TODO: Implement allocation follow-ups
	response.SuggestedFollowUps = []string{
		"Show me the efficiency metrics for these allocations",
		"Compare these costs with the previous period",
		"Which pods are driving the highest costs?",
	}
}

func (h *AllocationQueryHandler) updateConversationContext(context *ConversationContext, query AllocationQuery, response *AllocationResponse) {
	// TODO: Implement conversation context updates
}