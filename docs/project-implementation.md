# MCP-Graph Implementation Plan

This document outlines the phased implementation plan for the mcp-graph project. The plan is structured to ensure we build the system incrementally, focusing on manageable chunks of work while maximising efficiency when creating or editing files.

## Implementation Checklist

### Phase 1: Project Setup and Core Structure

- [x] **Project Initialisation**
  - [x] Initialise Go module
  - [x] Create directory structure as outlined in the project overview
  - [x] Create .gitignore file
  - [x] Set up Go workspace
  - [x] Create initial README.md with project description

- [x] **Configuration Management**
  - [x] Create configuration structures in internal/config
  - [x] Implement configuration loading from file and environment variables
  - [x] Create .env.example file with template variables
  - [x] Create config.yaml example

- [x] **Core Graph Interface**
  - [x] Define and implement the graph store interface (internal/graph/graph.go)
  - [x] Create basic models for nodes and edges

### Phase 2: Dgraph Integration

- [x] **Dgraph Implementation**
  - [x] Set up Dgraph client with connection handling
  - [x] Implement graph store interface methods for Dgraph
  - [x] Create schema management functionality
  - [x] Implement query functionality
  - [x] Add transaction support
  - [x] Write unit tests for Dgraph implementation

- [x] **Docker Integration**
  - [x] Create Dockerfile for the application
  - [x] Create docker-compose.yml with Dgraph service

### Phase 3: MCP Server Integration

- [x] **MCP Server Setup**
  - [x] Implement MCP server with mcp-go
  - [x] Create tool definitions
  - [x] Implement tool handlers
  - [x] Add resource capabilities

- [x] **Knowledge Graph Tools**
  - [x] Create query tool for graph access
  - [x] Add document creation and management tools
  - [ ] Write unit tests for MCP tools

- [x] **Server Modes**
  - [x] Implement stdio server mode
  - [x] Add SSE server mode
  - [x] Create MCP client configuration example

### Phase 4: API Layer and Service Implementation

- [x] **Core Service Layer**
  - [x] Implement knowledge manager service (internal/service)
  - [x] Create service interfaces for business logic
  - [x] Implement data validation and processing logic

- [x] **API Development**
  - [x] Set up HTTP server in cmd/server
  - [x] Implement API handlers for CRUD operations
  - [x] Create middleware for authentication and logging
  - [x] Implement error handling
  - [x] Document API endpoints

- [x] **Main Application**
  - [x] Implement application entry point (cmd/server/main.go)
  - [x] Set up graceful shutdown handling
  - [x] Configure logging and monitoring

### Phase 5: Enhancements and Future Features

- [ ] **Visual Graph Explorer**
  - [ ] Research visualisation libraries
  - [ ] Implement basic graph visualisation API
  - [ ] Create web interface for graph exploration

- [ ] **Import/Export Functionality**
  - [ ] Implement data export to JSON/CSV
  - [ ] Add import capabilities from various sources
  - [ ] Create migration utilities

- [ ] **Performance Optimisations**
  - [ ] Implement caching mechanisms
  - [ ] Add connection pooling
  - [ ] Optimise query performance

### Phase 6: Documentation and Testing

- [x] **Documentation**
  - [x] Complete README with setup and usage instructions
  - [x] Create API documentation
  - [x] Add code examples
  - [x] Document configuration options

- [ ] **Testing**
  - [ ] Write integration tests
  - [ ] Create benchmarks
  - [ ] Set up CI/CD pipeline
  - [ ] Perform security testing

## Implementation Progress

### Completed Work (as of April 8, 2025)

We have successfully implemented the core components of the MCP-Graph project:

1. **Project Structure and Configuration**
   - Created a modular Go project structure
   - Implemented configuration management using Viper
   - Set up environment variable and YAML configuration

2. **Graph Interface and Dgraph Implementation**
   - Defined a clean graph store interface
   - Implemented the interface using Dgraph
   - Added support for nodes, edges, queries, and schema management

3. **MCP Server Integration**
   - Integrated with the Model Context Protocol using mark3labs/mcp-go
   - Implemented tools for graph operations
   - Added support for both stdio and SSE server modes
   - Created an MCP client configuration for use with Cline/Claude Desktop

4. **Containerisation**
   - Created a multi-stage Dockerfile for efficient builds
   - Set up docker-compose with Dgraph services
   - Configured networking and volumes

### Next Steps

The next phase of development should focus on:

1. **API Layer Development**
   - Implementing a RESTful API for non-MCP clients
   - Adding authentication and authorization

2. **Testing**
   - Writing unit tests for the service layer
   - Writing unit tests for MCP tools
   - Adding integration tests for the MCP server

3. **Future Enhancements**
   - Developing a lightweight visual graph explorer, but we need to make sure we don't end up adding a big bloated javascript ecosystem to the project
   - Adding import/export functionality
   - Implementing performance optimisations

### Recent Progress (April 8, 2025)

We have completed Phase 4 of the implementation plan, focusing on the service layer and API implementation, and made progress on testing:

1. **Knowledge Manager Service**
   - Created a service interface with high-level operations for managing the knowledge graph
   - Implemented document operations (create, get, update, delete, search)
   - Implemented concept operations (create, get, link, search)
   - Added schema initialisation functionality

2. **MCP Server Enhancement**
   - Updated the MCP server to use the service layer
   - Added new tools for document and concept management
   - Maintained backward compatibility with existing tools

3. **RESTful API Implementation**
   - Created a RESTful API server using gorilla/mux
   - Implemented document endpoints (create, get, update, delete, search)
   - Implemented concept endpoints (create, get, link, search)
   - Added query and schema endpoints
   - Implemented logging and error handling middleware
   - Integrated the API server with the main application
   - Created comprehensive API documentation in docs/api.md

4. **Project Architecture**
   - The project now has a clean, modular architecture:
     - **Graph Interface**: Defines the core operations for the knowledge graph
     - **Dgraph Implementation**: Implements the graph interface using Dgraph
     - **Service Layer**: Provides high-level business logic and domain operations
     - **MCP Server**: Exposes the knowledge graph via the Model Context Protocol
     - **API Server**: Provides a RESTful API for non-MCP clients

5. **Testing Progress**
   - Implemented comprehensive unit tests for the Dgraph implementation:
     - Created interface abstractions for Dgraph client and transactions
     - Generated mocks using mockgen for testing
     - Added tests for all core graph operations (create, get, update, delete)
     - Included tests for edge operations and query functionality
     - Added both success and error test cases

### Running the Application

The system can be run using the following command:

```bash
go run cmd/server/main.go
```

This will start both the MCP server and the RESTful API server, allowing clients to interact with the knowledge graph through either interface.

This implementation plan provides a structured approach to building the mcp-graph system, with each phase building upon the previous ones. As we complete each phase, we'll update this checklist to track our progress.

## Database Schema & Indexing (Neo4j)

To ensure optimal performance for the software architecture knowledge graph features, especially for `MERGE` and `MATCH` operations used by the `find_or_create_entity`, `find_or_create_relationship`, `get_entity_details`, and `find_neighbors` tools, the following indexes are recommended in Neo4j. These should be created after the Neo4j instance is running and before significant data ingestion.

**Note:** The exact index names (`index_name_...`) are suggestions and can be adjusted.

```cypher
// Indexes for unique identification

// Repository by URL
CREATE INDEX index_name_repository_url IF NOT EXISTS FOR (n:Repository) ON (n.url);

// Module by definition file path
CREATE INDEX index_name_module_filepath IF NOT EXISTS FOR (n:Module) ON (n.filePath);

// File (and ConfigurationFile) by file path
CREATE INDEX index_name_file_filepath IF NOT EXISTS FOR (n:File) ON (n.filePath);
// CREATE INDEX index_name_configfile_filepath IF NOT EXISTS FOR (n:ConfigurationFile) ON (n.filePath); // If ConfigurationFile is a separate label

// Function by file path and name (Composite)
CREATE INDEX index_name_function_filepath_name IF NOT EXISTS FOR (n:Function) ON (n.filePath, n.name);
// Optional: Index on signature if used for identification
// CREATE INDEX index_name_function_signature IF NOT EXISTS FOR (n:Function) ON (n.signature);

// Class by file path and name (Composite)
CREATE INDEX index_name_class_filepath_name IF NOT EXISTS FOR (n:Class) ON (n.filePath, n.name);

// Interface by file path and name (Composite)
CREATE INDEX index_name_interface_filepath_name IF NOT EXISTS FOR (n:Interface) ON (n.filePath, n.name);

// Library by name and version (Composite) - Adjust if groupId/artifactId are primary identifiers
CREATE INDEX index_name_library_name_version IF NOT EXISTS FOR (n:Library) ON (n.name, n.version);
// CREATE INDEX index_name_library_group_artifact_version IF NOT EXISTS FOR (n:Library) ON (n.groupId, n.artifactId, n.version);

// DataStore by type and location (Composite)
CREATE INDEX index_name_datastore_type_location IF NOT EXISTS FOR (n:DataStore) ON (n.type, n.location);

// Service by name
CREATE INDEX index_name_service_name IF NOT EXISTS FOR (n:Service) ON (n.name);

// Component by name
CREATE INDEX index_name_component_name IF NOT EXISTS FOR (n:Component) ON (n.name);

// Application by name
CREATE INDEX index_name_application_name IF NOT EXISTS FOR (n:Application) ON (n.name);

// ExternalAPI by endpoint URL
CREATE INDEX index_name_externalapi_endpointurl IF NOT EXISTS FOR (n:ExternalAPI) ON (n.endpointUrl);

// General purpose indexes (Optional but potentially useful)
// CREATE INDEX index_name_node_name IF NOT EXISTS FOR (n) ON (n.name); // Generic name index - use with caution on large graphs
```

These indexes help Neo4j quickly locate nodes based on the properties used in `MATCH` and `MERGE` clauses, significantly improving query performance.
