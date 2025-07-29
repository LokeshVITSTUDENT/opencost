# OpenCost MCP Server Implementation

This implementation provides a Model Context Protocol (MCP) server for OpenCost, enabling AI agents to intelligently access Kubernetes cost allocation, asset costs, and cloud billing data.

## Architecture Overview

### Core Components

1. **Conversation Management** (`conversation.go`)
   - Tracks conversation sessions with AI agents
   - Learns user preferences from interaction patterns
   - Provides context-aware query suggestions
   - Manages conversation history and cleanup

2. **Data Structures** (`types.go`)
   - Rich type definitions for AI-friendly interactions
   - Conversation context tracking
   - User preference learning
   - Query intent parsing

3. **Query Handlers** (`allocation.go`, `asset.go`, `cloudcost.go`)
   - Allocation costs with efficiency metrics and recommendations
   - Infrastructure asset costs with utilization insights
   - Cloud provider costs with forecasting and anomaly detection

4. **MCP Server** (`server.go`, `server_simple.go`)
   - Model Context Protocol integration
   - Tool definitions and JSON schema validation
   - Error handling and response formatting

## Key Features

### Intelligent Conversation Management

The `ConversationManager` maintains state across AI agent interactions:

```go
type ConversationContext struct {
    SessionID           string
    ConversationHistory []ConversationEntry
    UserPreferences     UserPreferences
    CurrentContext      QueryContext
    CreatedAt           time.Time
    LastActivity        time.Time
}
```

**Conversation Learning**: The system learns from user interactions to provide better defaults:
- Preferred time ranges (7d, 30d, etc.)
- Common aggregation patterns (namespace, cluster)
- Response format preferences (summary, detailed, insights)
- Cost thresholds for highlighting expensive resources

### AI-Enhanced Data Structures

Each data type (allocations, assets, cloud costs) includes:

#### Allocation Queries
- **Rich Parameters**: Time ranges, aggregation, filtering, idle costs
- **AI Insights**: Cost efficiency metrics, trend analysis, anomaly detection
- **Recommendations**: Rightsizing, scheduling optimizations, waste reduction
- **Context Awareness**: Previous queries inform default parameters

#### Asset Queries  
- **Infrastructure Focus**: Nodes, disks, load balancers, cluster management
- **Utilization Metrics**: CPU, memory, storage efficiency
- **Capacity Planning**: Growth projections, capacity constraints
- **Optimization**: Underutilized resource identification

#### Cloud Cost Queries
- **Multi-Provider**: AWS, Azure, GCP, Oracle
- **Service Breakdown**: Detailed cost attribution by cloud service
- **Forecasting**: Predictive cost modeling with confidence intervals
- **Anomaly Detection**: Unusual spending pattern identification

### Response Enhancement

All responses include:
- **Summaries**: High-level insights for quick understanding
- **Recommendations**: Actionable optimization suggestions  
- **Trends**: Historical analysis and future projections
- **Follow-ups**: Contextual next-step suggestions

## Implementation Highlights

### Conversation Context Tracking

```go
type UserPreferences struct {
    PreferredTimeRange    string
    PreferredClusters     []string
    PreferredNamespaces   []string
    PreferredAggregation  []string
    PreferredFormat       string
    CostThresholds        CostThresholds
    IncludeIdle          bool
}
```

The system automatically learns preferences:
- Time range patterns from repeated queries
- Aggregation preferences from user behavior
- Cost sensitivity from threshold interactions
- Format preferences from response feedback

### AI-Friendly Insights

```go
type CostInsight struct {
    Type              string  // "anomaly", "trend", "efficiency", "waste"
    Severity          string  // "low", "medium", "high", "critical"  
    Title             string
    Description       string
    ImpactCost        float64
    PotentialSavings  float64
    AffectedResources []string
    RecommendedAction string
}
```

### Recommendation Engine

```go
type CostRecommendation struct {
    Title                   string
    Description             string
    Category               string  // "rightsizing", "scheduling", "efficiency"
    PotentialMonthlySavings float64
    ImplementationComplexity string  // "low", "medium", "high"
    Actions                []string
    Prerequisites          []string
    EstimatedTimeDays      int
}
```

## MCP Integration

### Tools Available to AI Agents

1. **`query_allocations`**: Kubernetes cost allocation analysis
2. **`query_assets`**: Infrastructure asset cost and utilization  
3. **`query_cloud_costs`**: Cloud provider billing analysis
4. **`start_conversation`**: Begin conversation session
5. **`get_conversation_context`**: Retrieve learned preferences
6. **`update_preferences`**: Modify user preferences

### Example Tool Usage

```json
{
  "method": "tools/call",
  "params": {
    "name": "query_allocations",
    "arguments": {
      "timeRange": "7d",
      "aggregate": ["namespace"],
      "includeRecommendations": true,
      "sessionId": "session-123"
    }
  }
}
```

## Usage

### Starting the MCP Server

```bash
./opencost mcp --port 8080
```

### AI Agent Integration

The server uses stdio transport for MCP communication, allowing AI agents to:
1. Start conversations and track context
2. Query cost data with intelligent defaults
3. Receive recommendations and insights
4. Learn from user feedback over time

## Benefits for AI Agents

1. **Contextual Intelligence**: Learns from conversation patterns
2. **Rich Insights**: Goes beyond raw data to provide actionable intelligence
3. **Optimization Focus**: Proactive cost optimization recommendations
4. **Conversation Continuity**: Maintains context across multiple interactions
5. **Adaptive Responses**: Tailors output format and detail level to user preferences

## Future Enhancements

1. **Advanced Analytics**: Machine learning for better anomaly detection
2. **Integration Testing**: Comprehensive test suite with real OpenCost APIs
3. **Performance Optimization**: Caching and query optimization
4. **Extended Providers**: Support for additional cloud providers
5. **Custom Dashboards**: AI-generated cost dashboards based on conversation patterns

This implementation provides a sophisticated foundation for AI agents to interact with OpenCost data, with intelligent conversation management and rich, actionable insights that go far beyond traditional API access.