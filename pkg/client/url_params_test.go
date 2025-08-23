package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetURLParams(t *testing.T) {
	client := &Client{}

	params := map[string]string{
		"gsr_client": "test-client",
		"gsr_inst":   "test-instance",
	}

	client.SetURLParams(params)

	assert.NotNil(t, client.URLParams)
	assert.Equal(t, "test-client", client.URLParams["gsr_client"])
	assert.Equal(t, "test-instance", client.URLParams["gsr_inst"])
}

func TestSetURLParamsMultipleCalls(t *testing.T) {
	client := &Client{}

	// First set of parameters
	params1 := map[string]string{
		"gsr_client": "client1",
		"gsr_inst":   "instance1",
	}
	client.SetURLParams(params1)

	// Second set of parameters (should add/override)
	params2 := map[string]string{
		"gsr_client": "client2", // Override
		"tenant_id":  "tenant1", // Add new
	}
	client.SetURLParams(params2)

	assert.Equal(t, "client2", client.URLParams["gsr_client"]) // Overridden
	assert.Equal(t, "instance1", client.URLParams["gsr_inst"]) // Preserved
	assert.Equal(t, "tenant1", client.URLParams["tenant_id"])  // Added
}

func TestSubstituteQueryParamsEmpty(t *testing.T) {
	client := &Client{}

	query := "SELECT * FROM table WHERE id = @user_id"
	result := client.SubstituteQueryParams(query)

	// Should return original query when no parameters are set
	assert.Equal(t, query, result)
}

func TestSubstituteQueryParamsBasic(t *testing.T) {
	client := &Client{
		URLParams: map[string]string{
			"gsr_client": "test-client",
			"gsr_inst":   "test-instance",
		},
	}

	query := "SELECT * FROM hubs WHERE gsr_client = @gsr_client AND gsr_inst = @gsr_inst"
	result := client.SubstituteQueryParams(query)

	expected := "SELECT * FROM hubs WHERE gsr_client = 'test-client' AND gsr_inst = 'test-instance'"
	assert.Equal(t, expected, result)
}

func TestSubstituteQueryParamsWithQuotes(t *testing.T) {
	client := &Client{
		URLParams: map[string]string{
			"user_name": "John's Company",
		},
	}

	query := "SELECT * FROM users WHERE name = @user_name"
	result := client.SubstituteQueryParams(query)

	// Should escape single quotes
	expected := "SELECT * FROM users WHERE name = 'John''s Company'"
	assert.Equal(t, expected, result)
}

func TestSubstituteQueryParamsPartialMatch(t *testing.T) {
	client := &Client{
		URLParams: map[string]string{
			"gsr_client": "test-client",
		},
	}

	query := "SELECT * FROM hubs WHERE gsr_client = @gsr_client AND gsr_inst = @gsr_inst"
	result := client.SubstituteQueryParams(query)

	// Should only substitute available parameters
	expected := "SELECT * FROM hubs WHERE gsr_client = 'test-client' AND gsr_inst = @gsr_inst"
	assert.Equal(t, expected, result)
}

func TestSubstituteQueryParamsNoMatches(t *testing.T) {
	client := &Client{
		URLParams: map[string]string{
			"other_param": "value",
		},
	}

	query := "SELECT * FROM hubs WHERE gsr_client = @gsr_client"
	result := client.SubstituteQueryParams(query)

	// Should return original query when no parameters match
	assert.Equal(t, query, result)
}

func TestSubstituteQueryParamsMultipleOccurrences(t *testing.T) {
	client := &Client{
		URLParams: map[string]string{
			"tenant_id": "123",
		},
	}

	query := "SELECT * FROM table1 WHERE tenant_id = @tenant_id UNION SELECT * FROM table2 WHERE tenant_id = @tenant_id"
	result := client.SubstituteQueryParams(query)

	expected := "SELECT * FROM table1 WHERE tenant_id = '123' UNION SELECT * FROM table2 WHERE tenant_id = '123'"
	assert.Equal(t, expected, result)
}

func TestSubstituteQueryParamsComplexQuery(t *testing.T) {
	client := &Client{
		URLParams: map[string]string{
			"gsr_client": "flow-bi",
			"gsr_inst":   "prod",
			"user_role":  "admin",
		},
	}

	query := `
		SELECT h.*, u.name as user_name 
		FROM intf_automation.hubs h
		JOIN users u ON u.client = @gsr_client
		WHERE h.gsr_client = @gsr_client 
		  AND h.gsr_inst = @gsr_inst
		  AND u.role = @user_role
		ORDER BY h.created_at DESC
	`

	result := client.SubstituteQueryParams(query)

	// Should substitute all matching parameters
	assert.Contains(t, result, "u.client = 'flow-bi'")
	assert.Contains(t, result, "h.gsr_client = 'flow-bi'")
	assert.Contains(t, result, "h.gsr_inst = 'prod'")
	assert.Contains(t, result, "u.role = 'admin'")
}
