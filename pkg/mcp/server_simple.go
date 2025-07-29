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

// OpenCostMCPServerSimple provides a simplified MCP server implementation
type OpenCostMCPServerSimple struct {
	CostModel            *costmodel.CostModel
	CloudCostQuerier     cloudcost.Querier
	CloudCostViewQuerier cloudcost.ViewQuerier
	ConversationManager  *ConversationManager
}

// NewOpenCostMCPServerSimple creates a new simplified MCP server
func NewOpenCostMCPServerSimple(costModel *costmodel.CostModel, cloudCostQuerier cloudcost.Querier, cloudCostViewQuerier cloudcost.ViewQuerier) *OpenCostMCPServerSimple {
	return &OpenCostMCPServerSimple{
		CostModel:            costModel,
		CloudCostQuerier:     cloudCostQuerier,
		CloudCostViewQuerier: cloudCostViewQuerier,
		ConversationManager:  NewConversationManager(),
	}
}

// CreateServer creates an MCP Server with OpenCost tools
func (s *OpenCostMCPServerSimple) CreateServer() *mcp.Server {
	impl := &mcp.Implementation{
		Name:    "OpenCost MCP Server",
		Version: "1.0.0",
		Title:   "OpenCost Cost Analytics",
	}
	
	opts := &mcp.ServerOptions{
		Instructions: "OpenCost MCP Server provides AI-friendly access to Kubernetes cost allocation, asset costs, and cloud billing data with intelligent conversation support.",
	}
	
	server := mcp.NewServer(impl, opts)
	
	// Add a simple test tool
	s.addTestTool(server)
	
	return server
}

// addTestTool adds a simple test tool to verify MCP integration
func (s *OpenCostMCPServerSimple) addTestTool(server *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "test_opencost",
		Description: "Test OpenCost MCP integration",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"message": {
					Type:        "string",
					Description: "Test message",
				},
			},
		},
	}
	
	server.AddTool(tool, s.handleTestTool)
}

// handleTestTool handles the test tool
func (s *OpenCostMCPServerSimple) handleTestTool(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	var args map[string]interface{}
	if params.Arguments != nil {
		if argsMap, ok := params.Arguments.(map[string]interface{}); ok {
			args = argsMap
		}
	}
	
	message := "Hello from OpenCost MCP Server!"
	if msg, ok := args["message"].(string); ok {
		message = fmt.Sprintf("Echo: %s", msg)
	}
	
	response := map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "OpenCost MCP Server is running",
		"tools": []string{
			"query_allocations (coming soon)",
			"query_assets (coming soon)",
			"query_cloud_costs (coming soon)",
			"start_conversation (coming soon)",
		},
	}
	
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

// Shutdown gracefully shuts down the MCP server
func (s *OpenCostMCPServerSimple) Shutdown() {
	if s.ConversationManager != nil {
		s.ConversationManager.Stop()
	}
}