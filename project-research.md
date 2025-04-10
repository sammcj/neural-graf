# Project Research: Extending MCP-Graph for Software Architecture Knowledge Graphs

This document summarises the research, goals, and planning for extending the `mcp-graph` server to support the creation and querying of knowledge graphs representing software architectures.

## 1. Current State & Learnings

*   **Existing Server (`mcp-graph`):** A Go-based MCP server using Neo4j/Memgraph as the backend graph database.
*   **Core Library:** Built using the `mark3labs/mcp-go` library.
*   **Current MCP Tools:**
    *   Generic graph operations: `create_node`, `get_node`, `create_edge` (legacy), `query_knowledge_graph` (for Cypher), `upsert_schema`.
    *   Specific tools for `Document` and `Concept` types: `create_document`, `get_document`, `search_documents`, `create_concept`, `get_concept`, `link_concepts`.
*   **Technology Stack:** Go, Neo4j/Memgraph, `mcp-go`.
*   **Current Focus:** General knowledge management, demonstrated with examples around research papers and personal knowledge graphs.

## 2. Project Goals

*   **Primary Objective:** Extend `mcp-graph` to model software architectures (components, relationships, dependencies) within its knowledge graph.
*   **Target Users:** AI coding agents (like Cline) interacting via MCP.
*   **Key Use Cases:**
    *   **Discovery:** Enable agents to explore and understand unfamiliar codebases by querying the graph.
    *   **Understanding:** Help agents and users grasp complex interactions, data flows, and dependencies through graph queries and visualisations.
    *   **Documentation:** Use the graph as a dynamic, queryable source of architectural documentation.
    *   **Impact Analysis:** Allow agents to determine the potential effects of changes by querying for dependents.
    *   **Refactoring Support:** Identify coupling, circular dependencies, technical debt, and other architectural patterns/anti-patterns.

## 3. Important Considerations

*   **Language Support:** The system needs to model constructs from various languages (Go, Java, C/C++, Python, JS/TS, Swift, etc.) in a consistent way. The schema should be language-agnostic, relying on properties like `language`.
*   **Granularity:** Finding the right level of detail (e.g., Module, Component, Service, Class, Function) is crucial. Starting with Function-level seems appropriate, potentially adding `Application` later. Method-level might be too granular initially.
*   **Handling Unknowns/Uncertainty:** The graph must represent incomplete knowledge. This can be achieved through:
    *   `confidence` property on relationships.
    *   `status` property on nodes (e.g., 'stub', 'partially_analysed').
    *   `source` property indicating how information was derived ('manual', 'static-analysis', 'agent-inference').
*   **Metadata for Specific Use Cases:** Supporting tags (e.g., `refactor-candidate`, `legacy-code`, `technical-debt`) is important for tasks like software rewrites.
*   **Discovery Mechanisms:** Agents need ways to populate the graph. This involves:
    *   Using standard agent tools (`read_file`, `search_files`, `execute_command`).
    *   Leveraging language-specific parsing libraries (e.g., `go/parser`, `go/ast` for Go).
    *   Parsing documentation or configuration files.

## 4. Critical Implementation Factors & MCP Alignment

*   **Schema Definition:** A well-defined Neo4j schema is fundamental. The following node types, relationship types, and properties form the basis of the schema:
    *   **General Properties (Applicable to most nodes):**
        *   `id`: (string, required, unique) - System-generated unique identifier.
        *   `name`: (string, required) - Human-readable name.
        *   `description`: (string) - Optional textual description.
        *   `source`: (string enum: 'manual', 'static-analysis', 'doc-parsing', 'agent-inference', required) - How information was derived.
        *   `tags`: (list of strings) - Custom labels (e.g., `legacy`, `refactor-candidate`).
        *   `createdAt`: (datetime) - Timestamp of creation.
        *   `lastModifiedAt`: (datetime) - Timestamp of last update.
        *   `confidence`: (float, 0.0-1.0) - Confidence score for inferred data.
        *   `status`: (string enum: 'stub', 'partially_analysed', 'fully_analysed', 'deprecated') - Analysis or lifecycle status.
    *   **Node Types & Specific Properties:**
        *   `Application`: `ownerTeam` (string).
        *   `Repository`: `url` (string, required), `defaultBranch` (string).
        *   `Module`: `language` (string), `version` (string), `filePath` (string - path to definition file).
        *   `Component`: `language` (string), `filePath` (string - primary path), `ownerTeam` (string).
        *   `Service`: `language` (string), `filePath` (string), `ownerTeam` (string), `apiEndpoint` (string), `communicationProtocol` (string).
        *   `Library`: `language` (string), `version` (string), `groupId` (string), `artifactId` (string), `scope` (string).
        *   `Class`: `language` (string), `filePath` (string), `visibility` (string).
        *   `Interface`: `language` (string), `filePath` (string).
        *   `Function`: `language` (string), `filePath` (string), `signature` (string), `parameters` (list of strings), `returnType` (string), `visibility` (string).
        *   `File`: `filePath` (string, required), `language` (string), `format` (string).
        *   `DataStore`: `type` (string, required), `location` (string).
        *   `ExternalAPI`: `endpointUrl` (string), `documentationUrl` (string).
        *   `ConfigurationFile` (Sub-type of `File`): `filePath` (string, required), `format` (string, required).
    *   **Relationship Types & Specific Properties:**
        *   `CONTAINS`: (Hierarchical - no specific properties usually).
        *   `DEPENDS_ON`: `versionConstraint` (string), `scope` (string).
        *   `CALLS`: `lineNumber` (integer), `isAsync` (boolean).
        *   `IMPLEMENTS`: (No specific properties usually).
        *   `USES`: (No specific properties usually).
        *   `COMMUNICATES_WITH`: `protocol` (string), `isAsync` (boolean).
        *   `DEFINED_IN`: `startLine` (integer), `endLine` (integer).
        *   `CONFIGURED_BY`: (No specific properties usually).
        *   `PART_OF`: (No specific properties usually).
    *   **Note on File Imports:** We will rely on semantic relationships (`CALLS`, `IMPLEMENTS`, `DEPENDS_ON` between code constructs) rather than explicit `File IMPORTS File` relationships for code, as the former provides more architectural meaning. Specific `INCLUDES` relationships might be used for non-code files (e.g., HTML including CSS).
*   **MCP Primitives:**
    *   **Tools:** This is the primary mechanism for interaction. Specialised, model-controlled tools should be created. Key tools include:
        *   `find_or_create_entity`: Idempotently finds or creates nodes, updating properties on match. Essential for preventing duplicates and handling updates. Requires defining unique identifying properties for different entity types.
        *   `find_or_create_relationship`: Idempotently finds or creates relationships, updating properties on match.
        *   Basic Querying: Tools like `get_entity_details` and `find_neighbors` for exploration.
        *   Future Tools: More advanced queries like `find_dependencies`, `find_dependents`, `visualise_component`.
        These tools need clear input schemas (using `mcp-go` helpers) and return structured results or errors appropriately (`isError: true`).
    *   **Resources:** Can be used optionally to expose specific graph views or reports for user-controlled context (e.g., a diagram for a specific service).
    *   **Prompts:** Can be used optionally to define user-initiated analysis workflows (e.g., `/summarise_component`).
    *   **Sampling:** A potential future enhancement for server-side LLM analysis, but likely deferred due to current client/library support.
*   **`mcp-go` Library:** Implementation should follow the patterns established by `mcp-go` for defining and handling tools, resources, and prompts.
*   **Data Ingestion Strategy:** Define how the graph will be populated. Agent-driven analysis using standard tools combined with the new specialised graph tools seems like a good starting point. Automated static analysis integration is a potential enhancement.
*   **Error Handling:** Tool handlers must manage errors gracefully, returning meaningful error information within the `CallToolResult` for the agent to process (e.g., specific error types like `ENTITY_NOT_FOUND`).
*   **Stale Data Management:** Keeping the graph synchronised requires handling deletions or changes. The proposed "Stale Marking Strategy" involves:
    *   Updating a `lastModifiedAt` timestamp on entities/relationships touched during an analysis run (via `find_or_create_...`).
    *   After analysing a scope (e.g., a file), querying for entities expected within that scope.
    *   Marking entities found in the graph but *not* in the current analysis as 'stale' or 'deprecated' using their `status` property.
*   **Performance:** Indexes must be created in Neo4j on properties used for matching in `find_or_create_entity/relationship` (e.g., `name`, `filePath`, `url`) to ensure `MERGE` performance.

## 5. Next Steps (Planning)

1.  Finalise the detailed schema properties for core node and relationship types.
2.  Prioritise and define the first set of specialised MCP Tools to implement (likely focusing on entity/relationship creation and basic querying).
3.  Refine the initial data ingestion strategy (e.g., agent-driven workflow).
4.  Transition to ACT MODE for implementation.
