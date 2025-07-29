package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/opencost/opencost/core/pkg/opencost"
	"github.com/opencost/opencost/pkg/costmodel"
)

// AssetQueryHandler manages asset-related queries for AI agents requesting 
// Kubernetes infrastructure asset cost data (nodes, disks, load balancers, etc.)
type AssetQueryHandler struct {
	// Model provides access to the OpenCost asset computation engine  
	Model *costmodel.CostModel
	
	// ConversationManager handles conversation state and history
	ConversationManager *ConversationManager
}

// AssetQuery represents a comprehensive asset query with AI-friendly parameters
type AssetQuery struct {
	// Base query parameters
	TimeRange *opencost.Window `json:"timeRange"`
	
	// Filtering
	Filter string `json:"filter,omitempty"` // OpenCost asset filter string
	
	// Asset type filtering
	AssetTypes []string `json:"assetTypes,omitempty"` // "node", "disk", "loadbalancer", "cluster"
	
	// AI-specific enhancements
	ResponseFormat         string            `json:"responseFormat,omitempty"`         // "summary", "detailed", "insights"
	CostThreshold          *float64          `json:"costThreshold,omitempty"`          // Only show assets above this cost
	TopN                   int               `json:"topN,omitempty"`                   // Limit to top N assets by cost
	IncludeRecommendations bool              `json:"includeRecommendations"`           // Include optimization recommendations
	IncludeUtilization     bool              `json:"includeUtilization"`               // Include utilization metrics
	HighlightUnderutilized bool              `json:"highlightUnderutilized"`           // Flag underutilized assets
	IncludeCapacityPlanning bool             `json:"includeCapacityPlanning"`          // Include capacity insights
	
	// Context from conversation
	ConversationContext *ConversationContext `json:"conversationContext,omitempty"`
}

// AssetResponse provides AI-friendly asset data with insights and recommendations
type AssetResponse struct {
	// Core data
	Data *opencost.AssetSet `json:"data"`
	
	// Summary and insights for AI consumption
	Summary         AssetSummary         `json:"summary"`
	Insights        []AssetInsight       `json:"insights,omitempty"`
	Recommendations []AssetRecommendation `json:"recommendations,omitempty"`
	CapacityAnalysis *CapacityAnalysis    `json:"capacityAnalysis,omitempty"`
	
	// Response metadata
	Query            AssetQuery `json:"query"`
	GeneratedAt      time.Time  `json:"generatedAt"`
	ProcessingTimeMs int64      `json:"processingTimeMs"`
	
	// Conversation context
	SuggestedFollowUps []string `json:"suggestedFollowUps,omitempty"`
}

// AssetSummary provides high-level insights about asset costs and utilization
type AssetSummary struct {
	// Total costs by asset type
	TotalCost           float64 `json:"totalCost"`
	NodeCost            float64 `json:"nodeCost"`
	DiskCost            float64 `json:"diskCost"`
	LoadBalancerCost    float64 `json:"loadBalancerCost"`
	ClusterManagementCost float64 `json:"clusterManagementCost"`
	
	// Asset counts
	TotalAssets      int `json:"totalAssets"`
	NodeCount        int `json:"nodeCount"`
	DiskCount        int `json:"diskCount"`
	LoadBalancerCount int `json:"loadBalancerCount"`
	
	// Cost distribution
	TopCostAsset      string  `json:"topCostAsset"`      // Asset with highest cost
	TopCostAssetValue float64 `json:"topCostAssetValue"`
	TopCostAssetType  string  `json:"topCostAssetType"`  // node, disk, lb
	
	// Utilization insights (when available)
	AverageNodeUtilization  float64 `json:"averageNodeUtilization,omitempty"`
	AverageDiskUtilization  float64 `json:"averageDiskUtilization,omitempty"`
	UnderutilizedAssets     int     `json:"underutilizedAssets,omitempty"`
	OverutilizedAssets      int     `json:"overutilizedAssets,omitempty"`
	
	// Efficiency metrics
	CostPerCPUHour    float64 `json:"costPerCPUHour,omitempty"`
	CostPerRAMGBHour  float64 `json:"costPerRAMGBHour,omitempty"`
	CostPerStorageGBHour float64 `json:"costPerStorageGBHour,omitempty"`
	
	// Time-based insights
	CostPerDay   float64 `json:"costPerDay"`
	CostPerHour  float64 `json:"costPerHour"`
	CostTrend    string  `json:"costTrend"` // "increasing", "decreasing", "stable"
	
	// Waste indicators
	WastedCost        float64 `json:"wastedCost,omitempty"`        // Cost of unused/underutilized resources
	OptimizationPotential float64 `json:"optimizationPotential,omitempty"` // Potential savings
}

// AssetInsight represents an AI-generated insight about asset usage patterns
type AssetInsight struct {
	// Type of insight
	Type string `json:"type"` // "underutilization", "waste", "cost_spike", "capacity", "efficiency"
	
	// Severity level
	Severity string `json:"severity"` // "low", "medium", "high", "critical"
	
	// Human-readable description
	Title       string `json:"title"`
	Description string `json:"description"`
	
	// Impact metrics
	ImpactCost       float64 `json:"impactCost,omitempty"`       // Cost impact
	PotentialSavings float64 `json:"potentialSavings,omitempty"` // Potential savings
	
	// Context
	AffectedAssets    []string `json:"affectedAssets,omitempty"`
	RecommendedAction string   `json:"recommendedAction,omitempty"`
	
	// Asset-specific metrics
	UtilizationMetrics map[string]float64 `json:"utilizationMetrics,omitempty"`
	CapacityMetrics    map[string]float64 `json:"capacityMetrics,omitempty"`
}

// AssetRecommendation provides actionable infrastructure optimization suggestions
type AssetRecommendation struct {
	// Recommendation details
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"` // "rightsizing", "consolidation", "scheduling", "storage_optimization"
	
	// Impact assessment
	PotentialMonthlySavings  float64 `json:"potentialMonthlySavings"`
	ImplementationComplexity string  `json:"implementationComplexity"` // "low", "medium", "high"
	RiskLevel               string  `json:"riskLevel"`                // "low", "medium", "high"
	
	// Implementation guidance
	Actions           []string `json:"actions"`
	Prerequisites     []string `json:"prerequisites,omitempty"`
	EstimatedTimeDays int      `json:"estimatedTimeDays,omitempty"`
	
	// Asset context
	AffectedAssets []AssetRecommendationTarget `json:"affectedAssets"`
	
	// Supporting data
	CurrentMetrics  map[string]float64 `json:"currentMetrics,omitempty"`
	ProjectedMetrics map[string]float64 `json:"projectedMetrics,omitempty"`
}

// AssetRecommendationTarget specifies an asset targeted for optimization
type AssetRecommendationTarget struct {
	AssetID           string             `json:"assetId"`
	AssetType         string             `json:"assetType"` // "node", "disk", "loadbalancer"
	AssetName         string             `json:"assetName"`
	CurrentCost       float64            `json:"currentCost"`
	ProjectedCost     float64            `json:"projectedCost"`
	OptimizationAction string            `json:"optimizationAction"` // "resize", "terminate", "migrate"
	UtilizationMetrics map[string]float64 `json:"utilizationMetrics,omitempty"`
}

// CapacityAnalysis provides insights about infrastructure capacity and planning
type CapacityAnalysis struct {
	// Current capacity status
	OverallCapacityUtilization float64 `json:"overallCapacityUtilization"`
	CPUCapacityUtilization     float64 `json:"cpuCapacityUtilization"`
	RAMCapacityUtilization     float64 `json:"ramCapacityUtilization"`
	StorageCapacityUtilization float64 `json:"storageCapacityUtilization"`
	
	// Capacity trends
	CapacityTrend      string    `json:"capacityTrend"` // "increasing", "decreasing", "stable"
	PeakUtilization    float64   `json:"peakUtilization"`
	PeakUtilizationTime time.Time `json:"peakUtilizationTime"`
	
	// Capacity planning recommendations
	ProjectedGrowthRate       float64 `json:"projectedGrowthRate,omitempty"`       // Monthly growth rate
	EstimatedCapacityRunout   *time.Time `json:"estimatedCapacityRunout,omitempty"` // When capacity will be exhausted
	RecommendedCapacityExpansion string `json:"recommendedCapacityExpansion,omitempty"`
	
	// Resource-specific insights
	NodeCapacityInsights []NodeCapacityInsight `json:"nodeCapacityInsights,omitempty"`
	StorageCapacityInsights []StorageCapacityInsight `json:"storageCapacityInsights,omitempty"`
}

// NodeCapacityInsight provides node-specific capacity analysis
type NodeCapacityInsight struct {
	NodeName            string  `json:"nodeName"`
	CPUUtilization      float64 `json:"cpuUtilization"`
	RAMUtilization      float64 `json:"ramUtilization"`
	CapacityStatus      string  `json:"capacityStatus"` // "underutilized", "optimal", "overutilized"
	RecommendedAction   string  `json:"recommendedAction,omitempty"`
	PotentialSavings    float64 `json:"potentialSavings,omitempty"`
}

// StorageCapacityInsight provides storage-specific capacity analysis
type StorageCapacityInsight struct {
	StorageName         string  `json:"storageName"`
	StorageType         string  `json:"storageType"` // "disk", "pv"
	UtilizationPercent  float64 `json:"utilizationPercent"`
	CapacityStatus      string  `json:"capacityStatus"` // "underutilized", "optimal", "nearly_full"
	RecommendedAction   string  `json:"recommendedAction,omitempty"`
	PotentialSavings    float64 `json:"potentialSavings,omitempty"`
}

// QueryAssets executes an asset query with AI-enhanced response formatting
func (h *AssetQueryHandler) QueryAssets(ctx context.Context, query AssetQuery) (*AssetResponse, error) {
	startTime := time.Now()
	
	// Apply intelligent defaults based on conversation context
	if err := h.applyConversationDefaults(&query); err != nil {
		return nil, fmt.Errorf("failed to apply conversation defaults: %w", err)
	}
	
	// Execute the core OpenCost asset query
	assetSet, err := h.Model.ComputeAssets(*query.TimeRange.Start(), *query.TimeRange.End())
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	
	// Apply filtering if specified
	if query.Filter != "" {
		// Note: Would need to implement filter parsing and application
		// This is a placeholder for the actual filtering logic
	}
	
	// Filter by asset types if specified
	if len(query.AssetTypes) > 0 {
		assetSet = h.filterByAssetTypes(assetSet, query.AssetTypes)
	}
	
	// Apply cost threshold filtering
	if query.CostThreshold != nil {
		assetSet = h.filterByCostThreshold(assetSet, *query.CostThreshold)
	}
	
	// Limit results if TopN specified
	if query.TopN > 0 {
		assetSet = h.limitToTopN(assetSet, query.TopN)
	}
	
	// Generate AI-enhanced response
	response := &AssetResponse{
		Data:             assetSet,
		Query:            query,
		GeneratedAt:      time.Now(),
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
	}
	
	// Generate summary and insights
	if err := h.generateAssetSummary(response); err != nil {
		return nil, fmt.Errorf("failed to generate asset summary: %w", err)
	}
	
	if query.IncludeRecommendations {
		h.generateAssetRecommendations(response)
	}
	
	if query.HighlightUnderutilized {
		h.generateUtilizationInsights(response)
	}
	
	if query.IncludeCapacityPlanning {
		if err := h.generateCapacityAnalysis(response); err != nil {
			return nil, fmt.Errorf("failed to generate capacity analysis: %w", err)
		}
	}
	
	// Generate suggested follow-up questions
	h.generateAssetFollowUps(response)
	
	// Update conversation context
	if query.ConversationContext != nil {
		h.updateAssetConversationContext(query.ConversationContext, query, response)
	}
	
	return response, nil
}

// applyConversationDefaults intelligently fills in query parameters based on conversation history
func (h *AssetQueryHandler) applyConversationDefaults(query *AssetQuery) error {
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
	
	return nil
}

// Helper methods for filtering and processing asset data
// These would be implemented with the actual logic:

func (h *AssetQueryHandler) filterByAssetTypes(assetSet *opencost.AssetSet, assetTypes []string) *opencost.AssetSet {
	// Implementation would filter assets by type
	return assetSet
}

func (h *AssetQueryHandler) filterByCostThreshold(assetSet *opencost.AssetSet, threshold float64) *opencost.AssetSet {
	// Implementation would filter assets above cost threshold
	return assetSet
}

func (h *AssetQueryHandler) limitToTopN(assetSet *opencost.AssetSet, topN int) *opencost.AssetSet {
	// Implementation would limit to top N assets by cost
	return assetSet
}

// Additional helper methods would be implemented here for:
// - generateAssetSummary
// - generateAssetRecommendations
// - generateUtilizationInsights
// - generateCapacityAnalysis
// - generateAssetFollowUps
// - updateAssetConversationContext

// Placeholder implementations for now
func (h *AssetQueryHandler) generateAssetSummary(response *AssetResponse) error {
	// TODO: Implement asset summary generation
	if response.Data != nil {
		totalCost := 0.0
		nodeCount := 0
		diskCount := 0
		lbCount := 0
		
		for _, asset := range response.Data.Assets {
			totalCost += asset.TotalCost()
			switch asset.Type() {
			case opencost.AssetTypeNode:
				nodeCount++
			case opencost.AssetTypeDisk:
				diskCount++
			case opencost.AssetTypeLoadBalancer:
				lbCount++
			}
		}
		
		response.Summary = AssetSummary{
			TotalCost:         totalCost,
			TotalAssets:       len(response.Data.Assets),
			NodeCount:         nodeCount,
			DiskCount:         diskCount,
			LoadBalancerCount: lbCount,
		}
	} else {
		response.Summary = AssetSummary{
			TotalCost:   0.0,
			TotalAssets: 0,
		}
	}
	return nil
}

func (h *AssetQueryHandler) generateAssetRecommendations(response *AssetResponse) {
	// TODO: Implement asset recommendations
	response.Recommendations = []AssetRecommendation{
		{
			Title:                   "Example Asset Recommendation",
			Description:             "This is a placeholder asset recommendation",
			PotentialMonthlySavings: 75.0,
		},
	}
}

func (h *AssetQueryHandler) generateUtilizationInsights(response *AssetResponse) {
	// TODO: Implement utilization insights
	response.Insights = []AssetInsight{
		{
			Type:        "underutilization",
			Severity:    "medium",
			Title:       "Example Utilization Insight",
			Description: "This is a placeholder utilization insight",
		},
	}
}

func (h *AssetQueryHandler) generateCapacityAnalysis(response *AssetResponse) error {
	// TODO: Implement capacity analysis
	response.CapacityAnalysis = &CapacityAnalysis{
		OverallCapacityUtilization: 65.0,
		CapacityTrend:              "stable",
	}
	return nil
}

func (h *AssetQueryHandler) generateAssetFollowUps(response *AssetResponse) {
	// TODO: Implement asset follow-ups
	response.SuggestedFollowUps = []string{
		"Show me underutilized resources that could be optimized",
		"Display the capacity planning analysis",
		"Compare asset costs across different clusters",
	}
}

func (h *AssetQueryHandler) updateAssetConversationContext(context *ConversationContext, query AssetQuery, response *AssetResponse) {
	// TODO: Implement asset conversation context updates
}