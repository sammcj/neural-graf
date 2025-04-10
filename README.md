# MCP-Graph

A lightweight, self-hosted knowledge graph system built in Go with Neo4j/Memgraph integration and MCP server support.

## Overview

MCP-Graph is a modular knowledge graph system that provides efficient data ingestion, graph-based storage, and powerful querying capabilities using Neo4j/Memgraph as the primary database. The system includes the following key components:

- **Knowledge Graph**: Core graph database functionality using Neo4j/Memgraph
- **MCP Server**: Deployed as a Model Context Protocol server (using mark3labs/mcp-go) to provide standardised LLM tool interfaces

## Features

- Clean, modular architecture with well-defined interfaces
- Flexible deployment options (standalone, containerised)
- Powerful graph querying capabilities
- MCP server integration for AI applications
- Standardised knowledge graph operations

## System Architecture

The system follows a clean, modular design with the following components:

```tree
mcp-graph/
├── cmd/
│   └── server/                 # Main application entry point
├── internal/
│   ├── api/                    # API handlers
│   ├── config/                 # Configuration management
│   ├── graph/                  # Knowledge graph implementation
│   │   └── neo4j/              # Neo4j/Memgraph implementation
│   ├── mcp/                    # MCP server
│   └── service/                # Core business logic
├── pkg/
│   ├── models/                 # Shared data models
│   └── utils/                  # Utility functions
└── scripts/                    # Deployment and tooling scripts
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Neo4j or Memgraph (can be run via Docker)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/sammcj/mcp-graph.git
   cd mcp-graph
   ```

2. Build the application:
   ```bash
   go build -o bin/mcp-graph ./cmd/server
   ```

3. Run with Docker Compose (includes Memgraph):
   ```bash
   docker-compose up -d
   ```

   Alternatively, for users who want to run Memgraph directly on their machine:
   ```bash
   ./scripts/run_with_local_memgraph.sh
   ```

## Configuration

Configuration can be provided via a YAML file or environment variables. See `.env.example` and `config.yaml.example` for available options.

## Usage

### API Endpoints

The service provides RESTful API endpoints for interacting with the knowledge graph.

### MCP Server

The MCP server can be used with compatible LLM applications like Claude Desktop or Cline. Here's how you can leverage the MCP server to interact with your knowledge graph using LLMs:

#### Connecting to the MCP Server

When configuring your MCP client to connect to the MCP-Graph server, you need to use the following URL:

```
http://localhost:3000/sse
```

For example, in your `mcp-client.json` file:

```json
{
  "connection": {
    "type": "sse",
    "url": "http://localhost:3000/sse"
  }
}
```

Or in Cline MCP settings:

```json
"mcp-graph": {
  "autoApprove": [],
  "disabled": false,
  "timeout": 60,
  "url": "http://localhost:3000/sse",
  "transportType": "sse"
}
```

Note that the `/sse` endpoint is required for the SSE connection to work properly.

#### MCP Tools for LLMs

MCP-Graph exposes the following tools to LLMs:

1. **query_graph**: Execute Cypher queries against the knowledge graph
   ```json
   {
     "query": "MATCH (n:Concept) RETURN n.name, n.id LIMIT 10",
     "params": {}
   }
   ```

2. **create_node**: Create a new node in the knowledge graph
   ```json
   {
     "type": "Document",
     "properties": {
       "title": "Understanding Knowledge Graphs",
       "content": "Knowledge graphs represent information as a network of entities and relationships...",
       "tags": ["knowledge graph", "graph database", "semantic web"]
     }
   }
   ```

3. **create_relationship**: Create a relationship between two nodes
   ```json
   {
     "from_id": "doc123",
     "to_id": "concept456",
     "type": "CONTAINS",
     "properties": {
       "confidence": 0.95,
       "extracted_by": "Claude 3 Opus"
     }
   }
   ```

4. **search_concepts**: Search for concepts in the knowledge graph
   ```json
   {
     "query": "machine learning",
     "limit": 5
   }
   ```

#### Example LLM Interactions

1. **Knowledge Extraction**:
   ```
   User: "I just read a paper about transformer architectures in NLP. Can you help me add it to my knowledge graph?"

   LLM: "I'll help you add this to your knowledge graph. Let me extract the key information first.

   [Uses MCP tools to create a Document node for the paper, extract key concepts like 'transformers', 'attention mechanisms', 'NLP', and create relationships between them]

   I've added the paper to your knowledge graph along with 5 key concepts and their relationships. You can now query this information or connect it to other related papers in your collection."
   ```

2. **Knowledge Exploration**:
   ```
   User: "What concepts in my knowledge graph are related to reinforcement learning?"

   LLM: [Uses MCP tools to query the graph]

   "I found several concepts related to reinforcement learning in your knowledge graph:

   1. Q-Learning (connected to 3 papers)
   2. Policy Gradients (connected to 2 papers and 1 course)
   3. Deep Q Networks (connected to 1 paper and 2 projects)
   4. Multi-Agent Systems (connected to 1 paper)

   Would you like me to explore any of these concepts in more detail?"
   ```

3. **Knowledge Synthesis**:
   ```
   User: "Can you help me prepare for my presentation on graph databases by summarizing what I have in my knowledge graph?"

   LLM: [Uses MCP tools to query for relevant nodes and relationships]

   "Based on your knowledge graph, here's a summary for your presentation:

   Your graph contains 15 documents related to graph databases, covering Neo4j, Memgraph, and ArangoDB. The main concepts include property graphs, graph algorithms, and query languages.

   Key papers in your collection focus on:
   1. Performance benchmarks between different graph databases
   2. Applications in recommendation systems
   3. Integration with machine learning pipelines

   I've also found connections to your personal projects where you've applied graph databases, which could make for good practical examples in your presentation."
   ```

## Example Use Cases

MCP-Graph can be used for a wide range of applications that benefit from graph-based knowledge representation. Here are some creative examples:

### Research Knowledge Management

Create a personal or team research knowledge base that connects papers, authors, concepts, and findings:

```cypher
// Create a research paper node
CREATE (p:Document {
  id: "paper123",
  title: "Advances in Graph Neural Networks",
  authors: ["Alice Smith", "Bob Jones"],
  year: 2023,
  url: "https://example.com/paper123"
})

// Create concept nodes
CREATE (c1:Concept {id: "gnn", name: "Graph Neural Networks"})
CREATE (c2:Concept {id: "attention", name: "Attention Mechanisms"})

// Connect papers to concepts
MATCH (p:Document {id: "paper123"}), (c:Concept {id: "gnn"})
CREATE (p)-[:CONTAINS]->(c)

MATCH (p:Document {id: "paper123"}), (c:Concept {id: "attention"})
CREATE (p)-[:CONTAINS]->(c)

// Query for papers about a specific concept
MATCH (p:Document)-[:CONTAINS]->(c:Concept {name: "Graph Neural Networks"})
RETURN p.title, p.authors, p.year
```

### Personal Knowledge Graph

Build a personal knowledge graph that connects your notes, ideas, projects, and learning resources:

```cypher
// Create nodes for different types of information
CREATE (n:Note {id: "note1", title: "Project Ideas", content: "..."})
CREATE (b:Book {id: "book1", title: "Graph Algorithms", author: "Mark Newman"})
CREATE (c:Course {id: "course1", title: "Machine Learning", provider: "Stanford"})
CREATE (p:Project {id: "proj1", title: "Home Automation System", status: "In Progress"})

// Create concept nodes
CREATE (c1:Concept {id: "ml", name: "Machine Learning"})
CREATE (c2:Concept {id: "iot", name: "Internet of Things"})

// Connect everything
MATCH (n:Note {id: "note1"}), (c:Concept {id: "ml"})
CREATE (n)-[:REFERENCES_TO]->(c)

MATCH (p:Project {id: "proj1"}), (c:Concept {id: "iot"})
CREATE (p)-[:RELATED_TO]->(c)

// Find all resources related to a concept
MATCH (resource)-[:REFERENCES_TO|RELATED_TO]->(c:Concept {name: "Machine Learning"})
RETURN resource.title, labels(resource)
```

### AI-Assisted Content Creation

Use MCP-Graph with LLMs to create and manage content with rich semantic connections:

1. Ingest articles, blog posts, and research papers
2. Use LLMs to extract key concepts, entities, and relationships
3. Store the structured knowledge in the graph
4. Query the graph to generate new content outlines, find connections between topics, or identify knowledge gaps

### Recommendation System

Build a recommendation system that suggests content based on user interests and behaviour:

```cypher
// Create user nodes with interests
CREATE (u1:User {id: "user1", name: "Jane"})
CREATE (u2:User {id: "user2", name: "Mike"})

// Create content nodes
CREATE (a1:Article {id: "article1", title: "Introduction to Graph Databases"})
CREATE (a2:Article {id: "article2", title: "Neo4j vs Memgraph Comparison"})

// Create topic nodes
CREATE (t1:Topic {id: "graphdb", name: "Graph Databases"})
CREATE (t2:Topic {id: "neo4j", name: "Neo4j"})
CREATE (t3:Topic {id: "memgraph", name: "Memgraph"})

// Connect users to interests
MATCH (u:User {id: "user1"}), (t:Topic {id: "graphdb"})
CREATE (u)-[:INTERESTED_IN]->(t)

// Connect articles to topics
MATCH (a:Article {id: "article1"}), (t:Topic {id: "graphdb"})
CREATE (a)-[:ABOUT]->(t)

MATCH (a:Article {id: "article2"}), (t:Topic {id: "neo4j"})
CREATE (a)-[:ABOUT]->(t)

MATCH (a:Article {id: "article2"}), (t:Topic {id: "memgraph"})
CREATE (a)-[:ABOUT]->(t)

// Recommend articles based on user interests
MATCH (u:User {id: "user1"})-[:INTERESTED_IN]->(t:Topic)<-[:ABOUT]-(a:Article)
RETURN a.title as recommendation
```

### Semantic Search with LLMs

Implement semantic search over your knowledge graph by combining MCP-Graph with LLMs:

1. Store documents and their relationships in the graph
2. When a user submits a query, use an LLM to convert it to a Cypher query
3. Execute the query against the graph database
4. Use the LLM to summarise and present the results in natural language

### Collaborative Knowledge Base

Create a team knowledge base where multiple users can contribute and connect information:

```cypher
// Create team and user nodes
CREATE (team:Team {id: "team1", name: "Engineering"})
CREATE (u1:User {id: "user1", name: "Alice"})
CREATE (u2:User {id: "user2", name: "Bob"})

// Create document nodes
CREATE (d1:Document {
  id: "doc1",
  title: "System Architecture",
  content: "...",
  created_at: datetime()
})

CREATE (d2:Document {
  id: "doc2",
  title: "API Documentation",
  content: "...",
  created_at: datetime()
})

// Connect users to team
MATCH (u:User {id: "user1"}), (t:Team {id: "team1"})
CREATE (u)-[:MEMBER_OF]->(t)

MATCH (u:User {id: "user2"}), (t:Team {id: "team1"})
CREATE (u)-[:MEMBER_OF]->(t)

// Connect documents to creators
MATCH (u:User {id: "user1"}), (d:Document {id: "doc1"})
CREATE (d)-[:CREATED_BY]->(u)

MATCH (u:User {id: "user2"}), (d:Document {id: "doc2"})
CREATE (d)-[:CREATED_BY]->(u)

// Find all documents created by team members
MATCH (t:Team {id: "team1"})<-[:MEMBER_OF]-(u:User)<-[:CREATED_BY]-(d:Document)
RETURN d.title, u.name as creator
```

## Future Enhancements

- Visual graph explorer
- Import/export functionality
- Advanced query capabilities
- Performance optimisations

## License

[MIT](LICENSE)
