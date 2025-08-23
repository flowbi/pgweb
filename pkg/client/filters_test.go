package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileRegexPatterns(t *testing.T) {
	t.Run("empty patterns", func(t *testing.T) {
		patterns, err := CompileRegexPatterns("")
		assert.NoError(t, err)
		assert.Nil(t, patterns)
	})

	t.Run("single pattern", func(t *testing.T) {
		patterns, err := CompileRegexPatterns("public")
		assert.NoError(t, err)
		assert.Len(t, patterns, 1)
	})

	t.Run("multiple patterns", func(t *testing.T) {
		patterns, err := CompileRegexPatterns("public,meta,^temp_")
		assert.NoError(t, err)
		assert.Len(t, patterns, 3)
	})

	t.Run("patterns with spaces", func(t *testing.T) {
		patterns, err := CompileRegexPatterns("public, meta , ^temp_")
		assert.NoError(t, err)
		assert.Len(t, patterns, 3)
	})

	t.Run("invalid regex", func(t *testing.T) {
		_, err := CompileRegexPatterns("public,[invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex pattern")
	})

	t.Run("empty patterns in list", func(t *testing.T) {
		patterns, err := CompileRegexPatterns("public,,meta")
		assert.NoError(t, err)
		assert.Len(t, patterns, 2)
	})
}

func TestShouldHideItem(t *testing.T) {
	patterns, err := CompileRegexPatterns("public,^temp_,_backup$")
	assert.NoError(t, err)

	t.Run("exact match", func(t *testing.T) {
		assert.True(t, shouldHideItem("public", patterns))
	})

	t.Run("prefix match", func(t *testing.T) {
		assert.True(t, shouldHideItem("temp_table", patterns))
		assert.True(t, shouldHideItem("temp_view", patterns))
	})

	t.Run("suffix match", func(t *testing.T) {
		assert.True(t, shouldHideItem("table_backup", patterns))
		assert.True(t, shouldHideItem("data_backup", patterns))
	})

	t.Run("no match", func(t *testing.T) {
		assert.False(t, shouldHideItem("users", patterns))
		assert.False(t, shouldHideItem("orders", patterns))
		assert.False(t, shouldHideItem("backup_table", patterns)) // doesn't end with _backup
	})
}

func TestFilterStringSlice(t *testing.T) {
	patterns, err := CompileRegexPatterns("public,^temp_")
	assert.NoError(t, err)

	items := []string{"public", "temp_table", "users", "orders", "temp_view", "meta"}
	filtered := FilterStringSlice(items, patterns)

	expected := []string{"users", "orders", "meta"}
	assert.Equal(t, expected, filtered)
}

func TestFilterObjectsResult(t *testing.T) {
	schemaPatterns, err := CompileRegexPatterns("public")
	assert.NoError(t, err)

	objectPatterns, err := CompileRegexPatterns("^temp_")
	assert.NoError(t, err)

	// Mock result with objects query structure: oid, schema, name, type, owner, comment
	result := &Result{
		Columns: []string{"oid", "schema", "name", "type", "owner", "comment"},
		Rows: []Row{
			{"1", "public", "users", "table", "postgres", "Users table"},
			{"2", "app", "temp_data", "table", "postgres", "Temporary data"},
			{"3", "public", "orders", "table", "postgres", "Orders table"},
			{"4", "app", "products", "table", "postgres", "Products table"},
			{"5", "app", "temp_cache", "view", "postgres", "Temporary cache"},
		},
	}

	filtered := filterObjectsResult(result, schemaPatterns, objectPatterns)

	// Should exclude: public.* (schema filter) and *temp_* (object filter)
	// Should keep: app.products only
	assert.Len(t, filtered.Rows, 1)
	assert.Equal(t, "app", filtered.Rows[0][1])      // schema
	assert.Equal(t, "products", filtered.Rows[0][2]) // name
}

func TestFilterObjectsResultNoPatterns(t *testing.T) {
	result := &Result{
		Columns: []string{"oid", "schema", "name", "type", "owner", "comment"},
		Rows: []Row{
			{"1", "public", "users", "table", "postgres", "Users table"},
		},
	}

	// No patterns should return original result
	filtered := filterObjectsResult(result, nil, nil)
	assert.Equal(t, result, filtered) // Should be the same object
}
