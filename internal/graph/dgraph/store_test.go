package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sammcj/mcp-graph/internal/graph/dgraph/mocks"
)

func TestCreateNode(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewTxn().Return(mockTxn)
	mockTxn.EXPECT().Discard(gomock.Any()).Return(nil)

	// Set up the mutation expectation
	expectedProperties := map[string]interface{}{
		"type":  "Document",
		"title": "Test Document",
	}
	expectedJSON, _ := json.Marshal(expectedProperties)
	expectedMutation := &api.Mutation{
		SetJson:   expectedJSON,
		CommitNow: true,
	}

	// Mock the response
	mockResponse := &api.Response{
		Uids: map[string]string{
			"blank-0": "0x1",
		},
	}
	mockTxn.EXPECT().Mutate(gomock.Any(), gomock.Eq(expectedMutation)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	nodeType := "Document"
	properties := map[string]interface{}{
		"title": "Test Document",
	}
	id, err := store.CreateNode(context.Background(), nodeType, properties)

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, "0x1", id)
}

func TestCreateNode_Error(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewTxn().Return(mockTxn)
	mockTxn.EXPECT().Discard(gomock.Any()).Return(nil)

	// Set up the mutation expectation with an error
	expectedProperties := map[string]interface{}{
		"type":  "Document",
		"title": "Test Document",
	}
	expectedJSON, _ := json.Marshal(expectedProperties)
	expectedMutation := &api.Mutation{
		SetJson:   expectedJSON,
		CommitNow: true,
	}

	// Mock the error response
	mockError := errors.New("mutation failed")
	mockTxn.EXPECT().Mutate(gomock.Any(), gomock.Eq(expectedMutation)).Return(nil, mockError)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	nodeType := "Document"
	properties := map[string]interface{}{
		"title": "Test Document",
	}
	_, err := store.CreateNode(context.Background(), nodeType, properties)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create node")
}

func TestGetNode(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewReadOnlyTxn().Return(mockTxn)

	// Expected query
	expectedQuery := `
		{
			node(func: uid(0x1)) {
				uid
				expand(_all_)
			}
		}
	`

	// Mock response data
	responseData := map[string][]map[string]interface{}{
		"node": {
			{
				"uid":     "0x1",
				"type":    "Document",
				"title":   "Test Document",
				"content": "Test content",
			},
		},
	}
	responseJSON, _ := json.Marshal(responseData)

	// Mock the response
	mockResponse := &api.Response{
		Json: responseJSON,
	}
	mockTxn.EXPECT().Query(gomock.Any(), gomock.Eq(expectedQuery)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	node, err := store.GetNode(context.Background(), "0x1")

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, "0x1", node["uid"])
	assert.Equal(t, "Document", node["type"])
	assert.Equal(t, "Test Document", node["title"])
	assert.Equal(t, "Test content", node["content"])
}

func TestGetNode_NotFound(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewReadOnlyTxn().Return(mockTxn)

	// Expected query
	expectedQuery := `
		{
			node(func: uid(0x1)) {
				uid
				expand(_all_)
			}
		}
	`

	// Mock empty response
	responseData := map[string][]map[string]interface{}{
		"node": {},
	}
	responseJSON, _ := json.Marshal(responseData)

	// Mock the response
	mockResponse := &api.Response{
		Json: responseJSON,
	}
	mockTxn.EXPECT().Query(gomock.Any(), gomock.Eq(expectedQuery)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	_, err := store.GetNode(context.Background(), "0x1")

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

func TestUpdateNode(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewTxn().Return(mockTxn)
	mockTxn.EXPECT().Discard(gomock.Any()).Return(nil)

	// Set up the mutation expectation
	expectedProperties := map[string]interface{}{
		"uid":   "0x1",
		"title": "Updated Document",
	}
	expectedJSON, _ := json.Marshal(expectedProperties)
	expectedMutation := &api.Mutation{
		SetJson:   expectedJSON,
		CommitNow: true,
	}

	// Mock the response
	mockResponse := &api.Response{}
	mockTxn.EXPECT().Mutate(gomock.Any(), gomock.Eq(expectedMutation)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	properties := map[string]interface{}{
		"title": "Updated Document",
	}
	err := store.UpdateNode(context.Background(), "0x1", properties)

	// Assert the results
	assert.NoError(t, err)
}

func TestDeleteNode(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewTxn().Return(mockTxn)
	mockTxn.EXPECT().Discard(gomock.Any()).Return(nil)

	// Set up the mutation expectation
	expectedMutation := &api.Mutation{
		DelNquads: []byte(`<0x1> * * .`),
		CommitNow: true,
	}

	// Mock the response
	mockResponse := &api.Response{}
	mockTxn.EXPECT().Mutate(gomock.Any(), gomock.Eq(expectedMutation)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	err := store.DeleteNode(context.Background(), "0x1")

	// Assert the results
	assert.NoError(t, err)
}

func TestCreateEdge(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewTxn().Return(mockTxn)
	mockTxn.EXPECT().Discard(gomock.Any()).Return(nil)

	// Set up the mutation expectation
	expectedEdgeData := map[string]interface{}{
		"uid": "0x1",
		"RELATED_TO": map[string]interface{}{
			"uid": "0x2",
		},
		"RELATED_TO_props": map[string]interface{}{
			"strength": 0.8,
		},
	}
	expectedJSON, _ := json.Marshal(expectedEdgeData)
	expectedMutation := &api.Mutation{
		SetJson:   expectedJSON,
		CommitNow: true,
	}

	// Mock the response
	mockResponse := &api.Response{}
	mockTxn.EXPECT().Mutate(gomock.Any(), gomock.Eq(expectedMutation)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	properties := map[string]interface{}{
		"strength": 0.8,
	}
	id, err := store.CreateEdge(context.Background(), "0x1", "0x2", "RELATED_TO", properties)

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, "0x1-RELATED_TO-0x2", id)
}

func TestQuery(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client and transaction
	mockClient := mocks.NewMockDgraphClient(ctrl)
	mockTxn := mocks.NewMockDgraphTxn(ctrl)

	// Set up expectations
	mockClient.EXPECT().NewReadOnlyTxn().Return(mockTxn)

	// Expected query and params
	expectedQuery := `{ documents(func: type(Document)) { uid title content } }`
	expectedParams := map[string]string{
		"limit": "10",
	}

	// Mock response data
	responseData := map[string][]map[string]interface{}{
		"documents": {
			{
				"uid":     "0x1",
				"title":   "Document 1",
				"content": "Content 1",
			},
			{
				"uid":     "0x2",
				"title":   "Document 2",
				"content": "Content 2",
			},
		},
	}
	responseJSON, _ := json.Marshal(responseData)

	// Mock the response
	mockResponse := &api.Response{
		Json: responseJSON,
	}
	mockTxn.EXPECT().QueryWithVars(gomock.Any(), gomock.Eq(expectedQuery), gomock.Eq(expectedParams)).Return(mockResponse, nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	params := map[string]interface{}{
		"limit": 10,
	}
	results, err := store.Query(context.Background(), expectedQuery, params)

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "0x1", results[0]["uid"])
	assert.Equal(t, "Document 1", results[0]["title"])
	assert.Equal(t, "Content 1", results[0]["content"])
	assert.Equal(t, "0x2", results[1]["uid"])
	assert.Equal(t, "Document 2", results[1]["title"])
	assert.Equal(t, "Content 2", results[1]["content"])
}

func TestUpsertSchema(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock client
	mockClient := mocks.NewMockDgraphClient(ctrl)

	// Expected schema
	expectedSchema := `
		type: string @index(exact) .
		title: string @index(fulltext, term) .
	`

	// Expected operation
	expectedOp := &api.Operation{
		Schema: expectedSchema,
	}

	// Set up expectations
	mockClient.EXPECT().Alter(gomock.Any(), gomock.Eq(expectedOp)).Return(nil)

	// Create a store with the mock client
	store := NewDgraphStoreWithClient(mockClient)

	// Call the method being tested
	err := store.UpsertSchema(context.Background(), expectedSchema)

	// Assert the results
	assert.NoError(t, err)
}
