# MCP-Graph API Documentation

This document describes the RESTful API endpoints provided by the MCP-Graph service.

## Base URL

All API endpoints are prefixed with `/api/v1`.

## Authentication

Authentication is not currently implemented. Future versions will include authentication mechanisms.

## Response Format

All responses are in JSON format. Successful responses will have an appropriate HTTP status code (200, 201, etc.) and will contain the requested data. Error responses will have an appropriate HTTP status code (400, 404, 500, etc.) and will contain an error message in the following format:

```json
{
  "error": "Error message"
}
```

## Endpoints

### Documents

#### Create Document

Creates a new document in the knowledge graph.

- **URL**: `/api/v1/documents`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "title": "Document Title",
    "content": "Document content goes here...",
    "metadata": {
      "author": "John Doe",
      "tags": ["knowledge", "graph"],
      "created": "2025-04-08T09:00:00Z"
    }
  }
  ```
- **Response**:
  ```json
  {
    "id": "0x1234"
  }
  ```
- **Status Codes**:
  - `201 Created`: Document created successfully
  - `400 Bad Request`: Invalid request payload
  - `500 Internal Server Error`: Server error

#### Get Document

Retrieves a document by ID.

- **URL**: `/api/v1/documents/{id}`
- **Method**: `GET`
- **URL Parameters**:
  - `id`: Document ID
- **Response**:
  ```json
  {
    "id": "0x1234",
    "title": "Document Title",
    "content": "Document content goes here...",
    "metadata": {
      "author": "John Doe",
      "tags": ["knowledge", "graph"],
      "created": "2025-04-08T09:00:00Z"
    }
  }
  ```
- **Status Codes**:
  - `200 OK`: Document retrieved successfully
  - `404 Not Found`: Document not found
  - `500 Internal Server Error`: Server error

#### Update Document

Updates an existing document.

- **URL**: `/api/v1/documents/{id}`
- **Method**: `PUT`
- **URL Parameters**:
  - `id`: Document ID
- **Request Body**:
  ```json
  {
    "title": "Updated Document Title",
    "content": "Updated document content...",
    "metadata": {
      "author": "John Doe",
      "tags": ["knowledge", "graph", "updated"],
      "updated": "2025-04-08T10:00:00Z"
    }
  }
  ```
- **Response**:
  ```json
  {
    "success": true
  }
  ```
- **Status Codes**:
  - `200 OK`: Document updated successfully
  - `400 Bad Request`: Invalid request payload
  - `404 Not Found`: Document not found
  - `500 Internal Server Error`: Server error

#### Delete Document

Deletes a document by ID.

- **URL**: `/api/v1/documents/{id}`
- **Method**: `DELETE`
- **URL Parameters**:
  - `id`: Document ID
- **Response**:
  ```json
  {
    "success": true
  }
  ```
- **Status Codes**:
  - `200 OK`: Document deleted successfully
  - `404 Not Found`: Document not found
  - `500 Internal Server Error`: Server error

#### Search Documents

Searches for documents matching a query.

- **URL**: `/api/v1/documents?query={query}`
- **Method**: `GET`
- **Query Parameters**:
  - `query`: Search query
- **Response**:
  ```json
  [
    {
      "id": "0x1234",
      "title": "Document Title",
      "content": "Document content goes here...",
      "metadata": {
        "author": "John Doe",
        "tags": ["knowledge", "graph"],
        "created": "2025-04-08T09:00:00Z"
      }
    },
    {
      "id": "0x5678",
      "title": "Another Document",
      "content": "More content...",
      "metadata": {
        "author": "Jane Smith",
        "tags": ["knowledge"],
        "created": "2025-04-07T14:30:00Z"
      }
    }
  ]
  ```
- **Status Codes**:
  - `200 OK`: Search completed successfully
  - `400 Bad Request`: Missing query parameter
  - `500 Internal Server Error`: Server error

### Concepts

#### Create Concept

Creates a new concept in the knowledge graph.

- **URL**: `/api/v1/concepts`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "name": "Concept Name",
    "properties": {
      "description": "Concept description",
      "category": "Category",
      "importance": 5
    }
  }
  ```
- **Response**:
  ```json
  {
    "id": "0x9abc"
  }
  ```
- **Status Codes**:
  - `201 Created`: Concept created successfully
  - `400 Bad Request`: Invalid request payload
  - `500 Internal Server Error`: Server error

#### Get Concept

Retrieves a concept by ID.

- **URL**: `/api/v1/concepts/{id}`
- **Method**: `GET`
- **URL Parameters**:
  - `id`: Concept ID
- **Response**:
  ```json
  {
    "id": "0x9abc",
    "name": "Concept Name",
    "properties": {
      "description": "Concept description",
      "category": "Category",
      "importance": 5
    }
  }
  ```
- **Status Codes**:
  - `200 OK`: Concept retrieved successfully
  - `404 Not Found`: Concept not found
  - `500 Internal Server Error`: Server error

#### Link Concepts

Creates a relationship between two concepts.

- **URL**: `/api/v1/concepts/link`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "fromId": "0x9abc",
    "toId": "0xdef0",
    "relationshipType": "RELATED_TO",
    "properties": {
      "strength": 0.8,
      "description": "These concepts are strongly related"
    }
  }
  ```
- **Response**:
  ```json
  {
    "id": "0x9abc-RELATED_TO-0xdef0"
  }
  ```
- **Status Codes**:
  - `201 Created`: Relationship created successfully
  - `400 Bad Request`: Invalid request payload
  - `404 Not Found`: One or both concepts not found
  - `500 Internal Server Error`: Server error

#### Search Concepts

Searches for concepts matching a query.

- **URL**: `/api/v1/concepts?query={query}`
- **Method**: `GET`
- **Query Parameters**:
  - `query`: Search query
- **Response**:
  ```json
  [
    {
      "id": "0x9abc",
      "name": "Concept Name",
      "properties": {
        "description": "Concept description",
        "category": "Category",
        "importance": 5
      }
    },
    {
      "id": "0xdef0",
      "name": "Another Concept",
      "properties": {
        "description": "Another description",
        "category": "Different Category",
        "importance": 3
      }
    }
  ]
  ```
- **Status Codes**:
  - `200 OK`: Search completed successfully
  - `400 Bad Request`: Missing query parameter
  - `500 Internal Server Error`: Server error

### Query

#### Execute Query

Executes a custom query against the knowledge graph.

- **URL**: `/api/v1/query`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "query": "{ documents(func: type(Document)) @filter(anyoftext(title, \"knowledge\")) { uid title content } }",
    "params": {
      "param1": "value1",
      "param2": 123
    }
  }
  ```
- **Response**: The response format depends on the query, but will be a JSON array of objects.
- **Status Codes**:
  - `200 OK`: Query executed successfully
  - `400 Bad Request`: Invalid request payload
  - `500 Internal Server Error`: Server error

### Schema

#### Update Schema

Updates or creates the graph schema.

- **URL**: `/api/v1/schema`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
    "schema": "type: string @index(exact) .\ntitle: string @index(fulltext, term) .\ncontent: string @index(fulltext) .\n..."
  }
  ```
- **Response**:
  ```json
  {
    "success": true
  }
  ```
- **Status Codes**:
  - `200 OK`: Schema updated successfully
  - `400 Bad Request`: Invalid request payload
  - `500 Internal Server Error`: Server error

## Error Handling

The API uses standard HTTP status codes to indicate the success or failure of a request. In case of an error, the response body will contain an error message explaining what went wrong.

## Rate Limiting

Rate limiting is not currently implemented. Future versions may include rate limiting to prevent abuse.

## Versioning

The API is versioned using the URL path (e.g., `/api/v1`). This allows for backward compatibility as the API evolves.
