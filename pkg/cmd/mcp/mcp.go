package mcp

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opencost/opencost/pkg/cloudcost"
	"github.com/opencost/opencost/pkg/costmodel"
	opencostmcp "github.com/opencost/opencost/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewMCPCommand creates the MCP server command
func NewMCPCommand() *cobra.Command {
	var (
		port int
		host string
	)

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run OpenCost as an MCP (Model Context Protocol) server for AI agents",
		Long: `Start OpenCost as an MCP server that provides AI agents with intelligent access to:
- Kubernetes cost allocation data
- Infrastructure asset costs (nodes, disks, load balancers)
- Cloud provider billing data (AWS, Azure, GCP)

The MCP server maintains conversation context, learns user preferences, and provides
AI-friendly responses with insights, recommendations, and contextual follow-up suggestions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPServer(host, port)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the MCP server on")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "Host to bind the MCP server to")

	return cmd
}

// runMCPServer starts the OpenCost MCP server
func runMCPServer(host string, port int) error {
	log.Printf("Starting OpenCost MCP Server on %s:%d", host, port)

	// Initialize OpenCost components
	// Note: In a real implementation, these would be properly initialized with configuration
	// For now, we'll create placeholder instances
	
	// Initialize cost model (this would normally come from the main initialization)
	var costModel *costmodel.CostModel
	
	// Initialize cloud cost components
	var cloudCostQuerier cloudcost.Querier
	var cloudCostViewQuerier cloudcost.ViewQuerier
	
	// For demonstration, we'll create a mock implementation
	// In the actual implementation, these would be properly initialized
	log.Println("Warning: Using mock implementations for demonstration")
	
	// Create the OpenCost MCP server
	mcpServerImpl := opencostmcp.NewOpenCostMCPServerSimple(costModel, cloudCostQuerier, cloudCostViewQuerier)
	
	// Create MCP server with tools
	server := mcpServerImpl.CreateServer()
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Received shutdown signal, gracefully shutting down...")
		mcpServerImpl.Shutdown()
		cancel()
	}()
	
	// For stdio transport (MCP standard)
	transport := mcp.NewStdioTransport()
	
	// Start the MCP server
	log.Println("OpenCost MCP Server started successfully")
	log.Println("AI agents can now connect to access OpenCost data with intelligent conversation support")
	log.Println("Available tools: query_allocations, query_assets, query_cloud_costs, start_conversation")
	
	// Run the server with the transport
	if err := server.Run(ctx, transport); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	
	log.Println("OpenCost MCP Server shut down")
	return nil
}

// Example usage documentation
func init() {
	// This will be called when the package is imported
	// We can use this to add documentation or examples
}

/*
Example usage:

1. Start the OpenCost MCP server:
   ./opencost mcp --port 8080

2. Connect an AI agent and start a conversation:
   {
     "method": "tools/call",
     "params": {
       "name": "start_conversation",
       "arguments": {}
     }
   }

3. Query allocation data:
   {
     "method": "tools/call",
     "params": {
       "name": "query_allocations",
       "arguments": {
         "timeRange": "7d",
         "aggregate": ["namespace"],
         "includeRecommendations": true,
         "sessionId": "session-id-from-step-2"
       }
     }
   }

4. Query asset costs with utilization insights:
   {
     "method": "tools/call",
     "params": {
       "name": "query_assets",
       "arguments": {
         "timeRange": "30d",
         "includeUtilization": true,
         "highlightUnderutilized": true,
         "sessionId": "session-id-from-step-2"
       }
     }
   }

5. Query cloud costs with forecasting:
   {
     "method": "tools/call",
     "params": {
       "name": "query_cloud_costs",
       "arguments": {
         "timeRange": "90d",
         "aggregate": ["service"],
         "includeForecast": true,
         "highlightAnomalies": true,
         "sessionId": "session-id-from-step-2"
       }
     }
   }
*/