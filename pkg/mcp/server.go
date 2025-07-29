package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/opencost/opencost/core/pkg/opencost"
	"github.com/opencost/opencost/pkg/cloudcost"
	"github.com/opencost/opencost/pkg/costmodel"
)

// OpenCostMCPServer implements the MCP tools for OpenCost
type OpenCostMCPServer struct {
	// Core OpenCost components
	CostModel       *costmodel.CostModel
	CloudCostQuerier cloudcost.Querier
	CloudCostViewQuerier cloudcost.ViewQuerier
	
	// MCP-specific handlers
	AllocationHandler   *AllocationQueryHandler
	AssetHandler       *AssetQueryHandler
	CloudCostHandler   *CloudCostQueryHandler
	ConversationManager *ConversationManager
}

// NewOpenCostMCPServer creates a new OpenCost MCP server
func NewOpenCostMCPServer(costModel *costmodel.CostModel, cloudCostQuerier cloudcost.Querier, cloudCostViewQuerier cloudcost.ViewQuerier) *OpenCostMCPServer {
	conversationManager := NewConversationManager()
	
	server := &OpenCostMCPServer{
		CostModel:           costModel,
		CloudCostQuerier:    cloudCostQuerier,
		CloudCostViewQuerier: cloudCostViewQuerier,
		ConversationManager: conversationManager,
	}
	
	// Initialize query handlers
	server.AllocationHandler = &AllocationQueryHandler{
		Model:               costModel,
		ConversationManager: conversationManager,
	}
	
	server.AssetHandler = &AssetQueryHandler{
		Model:               costModel,
		ConversationManager: conversationManager,
	}
	
	server.CloudCostHandler = &CloudCostQueryHandler{
		Querier:             cloudCostQuerier,
		ViewQuerier:         cloudCostViewQuerier,
		ConversationManager: conversationManager,
	}
	
	return server
}

// CreateServer creates an MCP Server with OpenCost tools
func (s *OpenCostMCPServer) CreateServer() *mcp.Server {
	impl := &mcp.Implementation{
		Name:    "OpenCost MCP Server",
		Version: "1.0.0",
		Title:   "OpenCost Cost Analytics",
	}
	
	opts := &mcp.ServerOptions{
		Instructions: "OpenCost MCP Server provides AI-friendly access to Kubernetes cost allocation, asset costs, and cloud billing data with intelligent conversation support.",
	}
	
	server := mcp.NewServer(impl, opts)
	
	// Add tools
	s.addQueryAllocationsTooln(server)
	s.addQueryAssetsTooln(server)
	s.addQueryCloudCostsTooln(server)
	s.addStartConversationTooln(server)
	s.addGetConversationContextTooln(server)
	s.addUpdatePreferencesTooln(server)
	
	return server
}

// Tool schemas and handlers
func (s *OpenCostMCPServer) addQueryAllocationsTooln(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "query_allocations",
		Description: "Query Kubernetes cost allocation data with intelligent filtering, aggregation, and AI-friendly insights",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"timeRange": {
					Type:        "string",
					Description: "Time range for the query (e.g., '7d', '30d', '1h', '24h')",
				},
				"aggregate": {
					Type:        "array",
					Description: "How to group the data (e.g., ['namespace'], ['cluster', 'namespace'])",
					Items: &jsonschema.Schema{
						Type: "string",
					},
				},
				"filter": {
					Type:        "string",
					Description: "OpenCost filter expression (e.g., 'namespace:\"kube-system\"')",
				},
				"includeIdle": {
					Type:        "boolean",
					Description: "Whether to include idle costs in the results",
				},
				"responseFormat": {
					Type:        "string",
					Description: "Desired response format: 'summary', 'detailed', 'insights', 'comparison'",
				},
				"topN": {
					Type:        "integer",
					Description: "Limit results to top N items by cost",
				},
				"includeRecommendations": {
					Type:        "boolean",
					Description: "Include cost optimization recommendations",
				},
				"compareWithPrevious": {
					Type:        "boolean",
					Description: "Compare with previous time period for trend analysis",
				},
				"sessionId": {
					Type:        "string",
					Description: "Conversation session ID for context tracking",
				},
			},
		},
	}
	
	server.AddTool(tool, s.handleQueryAllocations)
}

func (s *OpenCostMCPServer) handleQueryAllocations(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	var args map[string]interface{}
	if params.Arguments != nil {
		if argsMap, ok := params.Arguments.(map[string]interface{}); ok {
			args = argsMap
		} else {
			// Try to unmarshal from JSON if it's not already a map
			if argBytes, err := json.Marshal(params.Arguments); err == nil {
				json.Unmarshal(argBytes, &args)
			}
		}
	}
	
	query := AllocationQuery{}
	
	// Parse time range
	if timeRangeStr, ok := args["timeRange"].(string); ok && timeRangeStr != "" {
		window, err := parseTimeRange(timeRangeStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid timeRange: %v", err)}},
			}, nil
		}
		query.TimeRange = window
	}
	
	// Parse aggregation
	if aggInterface, ok := args["aggregate"]; ok {
		if aggSlice, ok := aggInterface.([]interface{}); ok {
			query.Aggregate = make([]string, len(aggSlice))
			for i, v := range aggSlice {
				query.Aggregate[i] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	// Parse other parameters
	if filter, ok := args["filter"].(string); ok {
		query.Filter = filter
	}
	
	if includeIdle, ok := args["includeIdle"].(bool); ok {
		query.IncludeIdle = includeIdle
	}
	
	if responseFormat, ok := args["responseFormat"].(string); ok {
		query.ResponseFormat = responseFormat
	}
	
	if topN, ok := args["topN"].(float64); ok {
		query.TopN = int(topN)
	}
	
	if includeRecommendations, ok := args["includeRecommendations"].(bool); ok {
		query.IncludeRecommendations = includeRecommendations
	}
	
	if compareWithPrevious, ok := args["compareWithPrevious"].(bool); ok {
		query.CompareWithPrevious = compareWithPrevious
	}
	
	// Get conversation context if session ID provided
	if sessionID, ok := args["sessionId"].(string); ok && sessionID != "" {
		conversation, err := s.ConversationManager.GetConversation(sessionID)
		if err == nil {
			query.ConversationContext = conversation
		}
	}
	
	// Execute the query
	response, err := s.AllocationHandler.QueryAllocation(ctx, query)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error querying allocations: %v", err)}},
		}, nil
	}
	
	// Convert response to JSON
	responseBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error marshaling response: %v", err)}},
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(responseBytes)}},
	}, nil
}

func (s *OpenCostMCPServer) addQueryAssetsTooln(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "query_assets",
		Description: "Query Kubernetes infrastructure asset costs (nodes, disks, load balancers) with utilization insights",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"timeRange": {
					Type:        "string",
					Description: "Time range for the query (e.g., '7d', '30d', '1h', '24h')",
				},
				"filter": {
					Type:        "string",
					Description: "OpenCost asset filter expression",
				},
				"assetTypes": {
					Type:        "array",
					Description: "Specific asset types to include: 'node', 'disk', 'loadbalancer', 'cluster'",
					Items: &jsonschema.Schema{
						Type: "string",
					},
				},
				"sessionId": {
					Type:        "string",
					Description: "Conversation session ID for context tracking",
				},
			},
		},
	}
	
	server.AddTool(tool, s.handleQueryAssets)
}

func (s *OpenCostMCPServer) handleQueryAssets(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	// Similar implementation to handleQueryAllocations but for assets
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Asset query implementation pending"}},
	}, nil
}

func (s *OpenCostMCPServer) addQueryCloudCostsTooln(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "query_cloud_costs",
		Description: "Query cloud provider billing data (AWS, Azure, GCP) with forecasting and anomaly detection",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"timeRange": {
					Type:        "string",
					Description: "Time range for the query (e.g., '7d', '30d', '90d')",
				},
				"aggregate": {
					Type:        "array",
					Description: "How to group the data (e.g., ['service'], ['provider', 'service'])",
					Items: &jsonschema.Schema{
						Type: "string",
					},
				},
				"sessionId": {
					Type:        "string",
					Description: "Conversation session ID for context tracking",
				},
			},
		},
	}
	
	server.AddTool(tool, s.handleQueryCloudCosts)
}

func (s *OpenCostMCPServer) handleQueryCloudCosts(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	// Similar implementation to handleQueryAllocations but for cloud costs
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "Cloud cost query implementation pending"}},
	}, nil
}

func (s *OpenCostMCPServer) addStartConversationTooln(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "start_conversation",
		Description: "Start a new conversation session for context tracking and preference learning",
		InputSchema: &jsonschema.Schema{
			Type:       "object",
			Properties: map[string]*jsonschema.Schema{},
		},
	}
	
	server.AddTool(tool, s.handleStartConversation)
}

func (s *OpenCostMCPServer) addGetConversationContextTooln(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "get_conversation_context",
		Description: "Retrieve conversation context and learned preferences",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"sessionId": {
					Type:        "string",
					Description: "Conversation session ID",
				},
			},
		},
	}
	
	server.AddTool(tool, s.handleGetConversationContext)
}

func (s *OpenCostMCPServer) addUpdatePreferencesTooln(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "update_preferences",
		Description: "Update user preferences for future queries",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"sessionId": {
					Type:        "string",
					Description: "Conversation session ID",
				},
				"preferences": {
					Type:        "object",
					Description: "User preferences to update",
				},
			},
		},
	}
	
	server.AddTool(tool, s.handleUpdatePreferences)
}

// Shutdown gracefully shuts down the MCP server
func (s *OpenCostMCPServer) Shutdown() {
	s.ConversationManager.Stop()
}

// handleAssetsQuery processes asset queries
func (s *OpenCostMCPServer) handleAssetsQuery(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	query := AssetQuery{}
	
	// Parse time range
	if timeRangeStr, ok := args["timeRange"].(string); ok && timeRangeStr != "" {
		window, err := parseTimeRange(timeRangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeRange: %w", err)
		}
		query.TimeRange = window
	}
	
	// Parse filter
	if filter, ok := args["filter"].(string); ok {
		query.Filter = filter
	}
	
	// Parse asset types
	if assetTypesInterface, ok := args["assetTypes"]; ok {
		if assetTypesSlice, ok := assetTypesInterface.([]interface{}); ok {
			query.AssetTypes = make([]string, len(assetTypesSlice))
			for i, v := range assetTypesSlice {
				query.AssetTypes[i] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	// Parse other parameters
	if responseFormat, ok := args["responseFormat"].(string); ok {
		query.ResponseFormat = responseFormat
	}
	
	if topN, ok := args["topN"].(float64); ok {
		query.TopN = int(topN)
	}
	
	if includeRecommendations, ok := args["includeRecommendations"].(bool); ok {
		query.IncludeRecommendations = includeRecommendations
	}
	
	if includeUtilization, ok := args["includeUtilization"].(bool); ok {
		query.IncludeUtilization = includeUtilization
	}
	
	if highlightUnderutilized, ok := args["highlightUnderutilized"].(bool); ok {
		query.HighlightUnderutilized = highlightUnderutilized
	}
	
	if includeCapacityPlanning, ok := args["includeCapacityPlanning"].(bool); ok {
		query.IncludeCapacityPlanning = includeCapacityPlanning
	}
	
	// Get conversation context if session ID provided
	if sessionID, ok := args["sessionId"].(string); ok && sessionID != "" {
		conversation, err := s.ConversationManager.GetConversation(sessionID)
		if err == nil {
			query.ConversationContext = conversation
		}
	}
	
	// Execute the query
	response, err := s.AssetHandler.QueryAssets(ctx, query)
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error querying assets: %v", err),
				},
			},
		}, nil
	}
	
	// Convert response to JSON
	responseBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling response: %v", err),
				},
			},
		}, nil
	}
	
	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: string(responseBytes),
			},
		},
	}, nil
}

// handleCloudCostsQuery processes cloud cost queries
func (s *OpenCostMCPServer) handleCloudCostsQuery(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	query := CloudCostQuery{}
	
	// Parse time range
	if timeRangeStr, ok := args["timeRange"].(string); ok && timeRangeStr != "" {
		window, err := parseTimeRange(timeRangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timeRange: %w", err)
		}
		query.TimeRange = window
	}
	
	// Parse aggregation
	if aggInterface, ok := args["aggregate"]; ok {
		if aggSlice, ok := aggInterface.([]interface{}); ok {
			query.Aggregate = make([]string, len(aggSlice))
			for i, v := range aggSlice {
				query.Aggregate[i] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	// Parse filter
	if filter, ok := args["filter"].(string); ok {
		query.Filter = filter
	}
	
	// Parse providers
	if providersInterface, ok := args["providers"]; ok {
		if providersSlice, ok := providersInterface.([]interface{}); ok {
			query.Providers = make([]string, len(providersSlice))
			for i, v := range providersSlice {
				query.Providers[i] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	// Parse services
	if servicesInterface, ok := args["services"]; ok {
		if servicesSlice, ok := servicesInterface.([]interface{}); ok {
			query.Services = make([]string, len(servicesSlice))
			for i, v := range servicesSlice {
				query.Services[i] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	// Parse other parameters
	if responseFormat, ok := args["responseFormat"].(string); ok {
		query.ResponseFormat = responseFormat
	}
	
	if topN, ok := args["topN"].(float64); ok {
		query.TopN = int(topN)
	}
	
	if includeRecommendations, ok := args["includeRecommendations"].(bool); ok {
		query.IncludeRecommendations = includeRecommendations
	}
	
	if includeForecast, ok := args["includeForecast"].(bool); ok {
		query.IncludeForecast = includeForecast
	}
	
	if highlightAnomalies, ok := args["highlightAnomalies"].(bool); ok {
		query.HighlightAnomalies = highlightAnomalies
	}
	
	// Get conversation context if session ID provided
	if sessionID, ok := args["sessionId"].(string); ok && sessionID != "" {
		conversation, err := s.ConversationManager.GetConversation(sessionID)
		if err == nil {
			query.ConversationContext = conversation
		}
	}
	
	// Execute the query
	response, err := s.CloudCostHandler.QueryCloudCosts(ctx, query)
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error querying cloud costs: %v", err),
				},
			},
		}, nil
	}
	
	// Convert response to JSON
	responseBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling response: %v", err),
				},
			},
		}, nil
	}
	
	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: string(responseBytes),
			},
		},
	}, nil
}

// handleStartConversation creates a new conversation session
func (s *OpenCostMCPServer) handleStartConversation(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	conversation, err := s.ConversationManager.StartConversation(ctx)
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error starting conversation: %v", err),
				},
			},
		}, nil
	}
	
	responseBytes, err := json.MarshalIndent(map[string]interface{}{
		"sessionId":  conversation.SessionID,
		"createdAt":  conversation.CreatedAt,
		"message":    "Conversation started successfully. I can help you analyze your Kubernetes costs, infrastructure assets, and cloud spending.",
		"suggestions": []string{
			"Show me the cost breakdown for the last 7 days",
			"What are the most expensive namespaces?",
			"Display asset utilization and optimization opportunities",
			"Analyze cloud costs and show spending trends",
		},
	}, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling response: %v", err),
				},
			},
		}, nil
	}
	
	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: string(responseBytes),
			},
		},
	}, nil
}

// handleGetConversationContext retrieves conversation context
func (s *OpenCostMCPServer) handleGetConversationContext(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	sessionID, ok := args["sessionId"].(string)
	if !ok || sessionID == "" {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: "sessionId is required",
				},
			},
		}, nil
	}
	
	conversation, err := s.ConversationManager.GetConversation(sessionID)
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error getting conversation: %v", err),
				},
			},
		}, nil
	}
	
	responseBytes, err := json.MarshalIndent(conversation, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling response: %v", err),
				},
			},
		}, nil
	}
	
	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: string(responseBytes),
			},
		},
	}, nil
}

// handleUpdatePreferences updates user preferences
func (s *OpenCostMCPServer) handleUpdatePreferences(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
	sessionID, ok := args["sessionId"].(string)
	if !ok || sessionID == "" {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: "sessionId is required",
				},
			},
		}, nil
	}
	
	preferencesInterface, ok := args["preferences"]
	if !ok {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: "preferences are required",
				},
			},
		}, nil
	}
	
	// Convert preferences to UserPreferences struct
	preferencesBytes, err := json.Marshal(preferencesInterface)
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling preferences: %v", err),
				},
			},
		}, nil
	}
	
	var preferences UserPreferences
	if err := json.Unmarshal(preferencesBytes, &preferences); err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error unmarshaling preferences: %v", err),
				},
			},
		}, nil
	}
	
	err = s.ConversationManager.UpdateUserPreferences(sessionID, preferences)
	if err != nil {
		return &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolResultContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Error updating preferences: %v", err),
				},
			},
		}, nil
	}
	
	return &mcp.ToolResult{
		Content: []mcp.ToolResultContent{
			{
				Type: "text",
				Text: "Preferences updated successfully",
			},
		},
	}, nil
}

// ListResources returns available resources (not used in this implementation)
func (s *OpenCostMCPServer) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	return []mcp.Resource{}, nil
}

// ReadResource reads a specific resource (not used in this implementation)
func (s *OpenCostMCPServer) ReadResource(ctx context.Context, uri string) (*mcp.ResourceContent, error) {
	return nil, fmt.Errorf("resource reading not supported")
}

// Shutdown gracefully shuts down the MCP server
func (s *OpenCostMCPServer) Shutdown() {
	log.Println("Shutting down OpenCost MCP Server...")
	s.ConversationManager.Stop()
}