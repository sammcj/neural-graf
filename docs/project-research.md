# Project Research: Extending MCP-Graph for Software Architecture Knowledge Graphs

## Introduction: Vision & High-Level Workflow

**Goal:** To enhance the existing `mcp-graph` server, transforming it into a specialised tool for building and querying a knowledge graph that represents the architecture of software projects. This knowledge graph will serve as a dynamic, queryable model of codebases, enabling deeper understanding, better documentation, and more informed decision-making during development, refactoring, and maintenance.

**Target Interaction:** The primary users of this enhanced server will be AI coding agents (like Cline, Claude, etc.) interacting via the Model Context Protocol (MCP). The server will expose specialised MCP Tools allowing these agents to populate and query the graph based on their analysis of software repositories.

**How it Should Work (Agent-Driven Workflow):**

1. **Analysis Trigger:** An AI agent is tasked with analysing a software repository (or part of it).
2. **Code/Config Retrieval:** The agent uses its standard capabilities (filesystem access, file reading, searching - potentially via other MCP tools or built-in functions) to access source code, dependency manifests, configuration files, etc.
3. **Parsing & Identification:** The agent parses these files to identify key architectural elements (functions, classes, services, libraries, dependencies, calls, data stores used, etc.) based on a predefined schema.
4. **Graph Population/Update:** For each identified element or relationship, the agent calls specific MCP Tools provided by the `mcp-graph` server (e.g., `find_or_create_entity`, `find_or_create_relationship`). These tools interact with the underlying Neo4j database, ensuring data is added or updated idempotently. The agent provides details like element type, name, file path, relationships, and source of information.
5. **Querying & Visualisation:** The agent (or a user interacting with the agent) can then query the populated graph using other specialised MCP Tools (e.g., `find_neighbors`, `get_entity_details`, `get_entity_subgraph`) provided by `mcp-graph`. The agent can use the retrieved graph data to answer questions about the architecture, identify dependencies, assess impact, or generate visualisations (like Mermaid diagrams).
6. **Maintenance:** The process includes mechanisms (like timestamping and status flags) to handle code evolution and mark stale data within the graph during subsequent analysis runs.

This agent-driven approach leverages the agent's analytical capabilities while providing it with a structured way (via MCP Tools) to store, update, and retrieve architectural knowledge persistently in the Neo4j graph managed by `mcp-graph`.

---

*This document summarises the research, goals, and planning for extending the `mcp-graph` server to support the creation and querying of knowledge graphs representing software architectures.*

## 1. Current State & Learnings

* **Existing Server (`mcp-graph`):** A Go-based MCP server using Neo4j/Memgraph as the backend graph database.
* **Core Library:** Built using the `mark3labs/mcp-go` library.
* **Current MCP Tools:**
  * Generic graph operations: `create_node`, `get_node`, `create_edge` (legacy), `query_knowledge_graph` (for Cypher), `upsert_schema`.
  * Specific tools for `Document` and `Concept` types: `create_document`, `get_document`, `search_documents`, `create_concept`, `get_concept`, `link_concepts`.
* **Technology Stack:** Go, Neo4j/Memgraph, `mcp-go`.
* **Current Focus:** General knowledge management, demonstrated with examples around research papers and personal knowledge graphs.

## 2. Project Goals

* **Primary Objective:** Extend `mcp-graph` to model software architectures (components, relationships, dependencies) within its knowledge graph.
* **Target Users:** AI coding agents (like Cline) interacting via MCP.
* **Key Use Cases:**
  * **Discovery:** Enable agents to explore and understand unfamiliar codebases by querying the graph.
  * **Understanding:** Help agents and users grasp complex interactions, data flows, and dependencies through graph queries and visualisations.
  * **Documentation:** Use the graph as a dynamic, queryable source of architectural documentation.
  * **Impact Analysis:** Allow agents to determine the potential effects of changes by querying for dependents.
  * **Refactoring Support:** Identify coupling, circular dependencies, technical debt, and other architectural patterns/anti-patterns.

## 3. Important Considerations

* **Language Support:** The system needs to model constructs from various languages (Go, Java, C/C++, Python, JS/TS, Swift, etc.) in a consistent way. The schema should be language-agnostic, relying on properties like `language`.
* **Granularity:** Finding the right level of detail (e.g., Module, Component, Service, Class, Function) is crucial. Starting with Function-level seems appropriate, potentially adding `Application` later. Method-level might be too granular initially.
  * **Handling Unknowns/Uncertainty:** The graph must represent incomplete knowledge. This can be achieved through:
    * `confidence` property on relationships/nodes.
    * `status` property on nodes (e.g., 'stub', 'partially_analysed').
    * `source` property indicating how information was derived ('manual', 'static-analysis', 'agent-inference').
    * **Future Consideration:** Develop strategies for resolving conflicting property values when data comes from multiple sources with varying reliability (e.g., based on source precedence or confidence scores).
* **Metadata for Specific Use Cases:** Supporting tags (e.g., `refactor-candidate`, `legacy-code`, `technical-debt`) is important for tasks like software rewrites.
* **Discovery Mechanisms:** Agents need ways to populate the graph. This involves:
  * Using standard agent tools (`read_file`, `search_files`, `execute_command`).
  * Leveraging language-specific parsing libraries (e.g., `go/parser`, `go/ast` for Go).
  * Parsing documentation or configuration files.

## 4. Critical Implementation Factors & MCP Alignment

* **Schema Definition:** A well-defined Neo4j schema is fundamental. The following node types, relationship types, and properties form the basis of the schema:
  * **General Properties (Applicable to most nodes):**
    * `id`: (string, required, unique) - System-generated unique identifier.
    * `name`: (string, required) - Human-readable name.
    * `description`: (string) - Optional textual description.
    * `source`: (string enum: 'manual', 'static-analysis', 'doc-parsing', 'agent-inference', required) - How information was derived.
    * `tags`: (list of strings) - Custom labels (e.g., `legacy`, `refactor-candidate`).
    * `createdAt`: (datetime) - Timestamp of creation.
    * `lastModifiedAt`: (datetime) - Timestamp of last update.
    * `confidence`: (float, 0.0-1.0) - Confidence score for inferred data.
    * `status`: (string enum: 'stub', 'partially_analysed', 'fully_analysed', 'deprecated') - Analysis or lifecycle status.
  * **Node Types & Specific Properties:**
    * `Application`: `ownerTeam` (string).
    * `Repository`: `url` (string, required), `defaultBranch` (string).
    * `Module`: `language` (string), `version` (string), `filePath` (string - path to definition file).
    * `Component`: `language` (string), `filePath` (string - primary path), `ownerTeam` (string).
    * `Service`: `language` (string), `filePath` (string), `ownerTeam` (string), `apiEndpoint` (string), `communicationProtocol` (string).
    * `Library`: `language` (string), `version` (string), `groupId` (string), `artifactId` (string), `scope` (string).
    * `Class`: `language` (string), `filePath` (string), `visibility` (string).
    * `Interface`: `language` (string), `filePath` (string).
    * `Function`: `language` (string), `filePath` (string), `signature` (string), `parameters` (list of strings), `returnType` (string), `visibility` (string).
    * `File`: `filePath` (string, required), `language` (string), `format` (string).
    * `DataStore`: `type` (string, required), `location` (string).
    * `ExternalAPI`: `endpointUrl` (string), `documentationUrl` (string).
    * `ConfigurationFile` (Sub-type of `File`): `filePath` (string, required), `format` (string, required).
  * **Relationship Types & Specific Properties:**
    * `CONTAINS`: (Hierarchical - no specific properties usually).
    * `DEPENDS_ON`: `versionConstraint` (string), `scope` (string).
    * `CALLS`: `lineNumber` (integer), `isAsync` (boolean).
    * `IMPLEMENTS`: (No specific properties usually).
    * `USES`: (No specific properties usually).
    * `COMMUNICATES_WITH`: `protocol` (string), `isAsync` (boolean).
    * `DEFINED_IN`: `startLine` (integer), `endLine` (integer).
    * `CONFIGURED_BY`: (No specific properties usually).
    * `PART_OF`: (No specific properties usually).
  * **Note on File Imports:** We will rely on semantic relationships (`CALLS`, `IMPLEMENTS`, `DEPENDS_ON` between code constructs) rather than explicit `File IMPORTS File` relationships for code, as the former provides more architectural meaning. Specific `INCLUDES` relationships might be used for non-code files (e.g., HTML including CSS).
  * **Identifying Properties for `find_or_create_entity`:** The `identifying_properties` input for this tool relies on unique keys for different entity types. Examples include:
    * `Repository`: `{ "url": "..." }`
    * `Module`: `{ "filePath": "..." }` (Path to definition file like go.mod, pom.xml)
    * `File`, `ConfigurationFile`: `{ "filePath": "..." }` (Relative path from repo root)
    * `Function`, `Class`, `Interface`: `{ "name": "...", "filePath": "..." }` (Or potentially `{ "signature": "..." }` for functions if names aren't unique within a file)
    * `Library`: `{ "name": "...", "version": "..." }` or `{ "groupId": "...", "artifactId": "...", "version": "..." }`
    * `DataStore`: `{ "type": "...", "location": "..." }`
    * `Service`, `Component`, `Application`: Often identified by `{ "name": "..." }` within a certain scope (e.g., unique within a repository or globally). Requires careful definition based on project context.
* **MCP Primitives:**
  * **Tools:** This is the primary mechanism for interaction. Specialised, model-controlled tools should be created. Key tools include:
    * `find_or_create_entity`: Idempotently finds or creates nodes, updating properties on match. Essential for preventing duplicates and handling updates. Requires defining unique identifying properties for different entity types.
    * `find_or_create_relationship`: Idempotently finds or creates relationships, updating properties on match.
    * Basic Querying: Tools like `get_entity_details` and `find_neighbors` for exploration.
    * Future Tools: More advanced queries like `find_dependencies`, `find_dependents`, `visualise_component`.
        These tools need clear input schemas (using `mcp-go` helpers) and return structured results or errors appropriately (`isError: true`).
  * **Tool Discovery & Usage:** Agents learn how to use these tools via the standard MCP `tools/list` mechanism. The server provides the tool `name`, `description`, and `inputSchema`. **Crucially, the quality and clarity of the tool and parameter descriptions are paramount** for enabling the agent to understand when and how to use the tools effectively. Descriptions should ideally include concrete usage examples where possible.
  * **Resources:** Can be used optionally to expose specific graph views or reports for user-controlled context (e.g., a diagram for a specific service).
  * **Prompts:** Can be used optionally to define user-initiated analysis workflows (e.g., `/summarise_component`).
  * **Sampling:** A potential future enhancement for server-side LLM analysis, but likely deferred due to current client/library support.
* **`mcp-go` Library:** Implementation should follow the patterns established by `mcp-go` for defining and handling tools, resources, and prompts.
* **MCP Conformance:** It is critical to maintain strict adherence to MCP standards, particularly when using the stdio transport. The server **must not** write arbitrary logs or debug information to stdout or stderr, as this will corrupt the JSON-RPC message stream expected by the client. Logging should be directed to a file or other appropriate sink, configurable perhaps via environment variables or a config file.
* **Data Ingestion Strategy (Agent-Driven):** The primary initial workflow involves:
    1. **Scope Definition:** Agent/user defines the analysis scope (repo, directory).
    2. **Information Gathering:** Agent uses standard tools (`list_files`, `read_file`, `search_files`, `execute_command`) to get code, config, etc.
    3. **Parsing & Analysis:** Agent internally parses content to identify architectural elements and relationships based on the schema.
    4. **Graph Population:** Agent uses the specialised `find_or_create_entity` and `find_or_create_relationship` tools to add/update the graph idempotently.
    5. **Stale Marking:** Agent applies the stale marking strategy after processing a scope.
    6. **Iteration:** Agent moves to the next scope.
    This relies on the agent's parsing capabilities and its understanding of the tools via their descriptions.
* **Concurrency Considerations:** While Neo4j's `MERGE` provides atomicity for individual operations, high-concurrency updates from multiple agents analysing overlapping scopes might require future investigation into more sophisticated locking or conflict resolution strategies if simple timestamp-based overwrites prove insufficient. For Phase 1, the basic `MERGE` approach is expected to be adequate.
* **Server-Side Analysis (Future Enhancement):**
  * An alternative approach involves the `mcp-graph` server performing the analysis directly. This would require a tool like `analyse_codebase(path, options)` and direct filesystem access for the server.
  * **Pros:** Potentially more efficient for bulk ingestion by avoiding numerous MCP calls. Centralises parsing logic.
  * **Cons:** Increases server complexity (filesystem access, multi-language parsing). Security considerations for server filesystem access.
  * **Multi-Language Handling:** Implementing multi-language parsing in the Go server is complex. Options include:
        1. Executing external static analysis tools (requires managing tool dependencies in the server environment).
        2. Using Go bindings for libraries like Tree-sitter (introduces CGo dependency, requires writing tree traversal logic per language).
  * **Decision:** Prioritise the agent-driven approach initially. Server-side analysis can be added later as an optional, potentially more performant, bulk ingestion method if needed.
* **Error Handling:** Tool handlers must manage errors gracefully, returning meaningful error information within the `CallToolResult` (e.g., `{ "isError": true, "content": [{"type": "text", "text": "Error: ENTITY_NOT_FOUND - Node with ID 'xyz' not found."}] }`) for the agent to process. Define consistent error structures.
* **Stale Data Management:** Keeping the graph synchronised requires handling deletions or changes. The proposed "Stale Marking Strategy" involves:
  * Updating a `lastModifiedAt` timestamp on entities/relationships touched during an analysis run (via `find_or_create_...`).
  * After analysing a scope (e.g., a file), querying for entities expected within that scope.
  * Marking entities found in the graph but *not* in the current analysis as 'stale' or 'deprecated' using their `status` property.
* **Performance & Scalability:**
  * **Indexing:** Indexes *must* be created in Neo4j on properties used for matching in `find_or_create_entity/relationship` (e.g., `:Function(name, filePath)`, `:Repository(url)`) to ensure `MERGE` performance. Consider compound indexes for frequently co-queried properties (e.g., label + name + filePath). Document required indexes clearly.
  * **Query Optimisation:** Ensure Cypher queries used within tool handlers are efficient and parameterised. Be mindful of potential performance issues with deep graph traversals (e.g., transitive dependencies) and test with realistic data sizes early.
  * **Transaction Management:** Tool handlers must ensure graph operations (especially `MERGE` in `find_or_create_...` tools) occur within proper database transactions managed by the `graph.Store` implementation to maintain data integrity.
  * **Batching (Future Enhancement):** Consider adding tools for batch creation/linking in Phase 2 for improved efficiency, particularly relevant if implementing server-side analysis.
* **Schema Evolution:** Plan for how schema changes will be managed over time (e.g., manual updates initially, potential future migration tooling).
* **Graph Complexity Management (Future Consideration):** As the graph grows, consider strategies like logical partitioning (e.g., by application), context-scoped queries, and advanced pruning/archiving for stale/deprecated elements.

## 5. Implementation Plan & Next Steps

1. **Phase 1 (Core Functionality - Agent-Driven):** `[COMPLETED]`
    * [x] Implement the defined Neo4j schema (node labels, properties, relationship types) via Go structs (`internal/graph/schema.go`).
    * [x] Create required Neo4j indexes for performance (e.g., on identifying properties). Documented in `docs/project-implementation.md`.
    * [x] Implement the prioritised MCP Tools using `mcp-go` (`internal/mcp/server.go`) and corresponding `graph.Store` methods (`internal/graph/neo4j/store.go`):
        * [x] `find_or_create_entity` (handling updates)
        * [x] `find_or_create_relationship` (handling updates)
        * [x] `get_entity_details`
        * [x] `find_neighbors`
    * [x] Ensure high-quality tool descriptions and input schemas in `internal/mcp/server.go`.
    * [x] Implement robust error handling within tools (basic argument validation and graph error propagation).
    * [x] Document the expected agent-driven workflow (in this document).
    * [ ] Implement a basic stale-marking mechanism. *(Deferred - Requires agent-side logic or a dedicated server tool as noted)*
2. **Phase 2 (Enhancements):** `[IN PROGRESS]`
    * [x] Implement more advanced query tools (`find_dependencies`, `find_dependents`).
    * [x] Implement a visualisation helper tool (`get_entity_subgraph`) that returns structured node/relationship data for a given entity and depth, enabling agents to generate diagrams (e.g., Mermaid).
    * [x] Implement batch operations to allow the AI coding agent to provide multiple entities/relationships in a single call, significantly reducing inference cost and speeding up the process.
        * Added `batch_find_or_create_entities` tool for creating/updating multiple entities in a single operation
        * Added `batch_find_or_create_relationships` tool for creating/updating multiple relationships in a single operation
        * Implemented parallel processing with concurrency limits for optimal performance
        * Provided detailed error handling that reports individual failures without failing the entire batch

---

**Future Considerations:** `[MAYBE]`

* Implement MCP Prompts for common query/analysis patterns (Query Templates).
* Optionally implement server-side analysis (`analyse_codebase`) if needed for bulk ingestion performance, addressing multi-language parsing challenges.
* Optionally implement MCP Resources for user-controlled views/reports.
* Refine stale data handling (e.g., dedicated tool, advanced pruning).
* Extend schema to include additional elements (e.g., Events, API Contracts, IaC).
