package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateQuery(t *testing.T) {
	t.Run("should accept valid queries", func(t *testing.T) {
		validQueries := []string{
			"SELECT * FROM users WHERE id = 1",
			"INSERT INTO users (name) VALUES ('John')",
			"UPDATE users SET name = 'Jane' WHERE id = 1",
			"DELETE FROM users WHERE id = 1",
		}
		
		for _, query := range validQueries {
			err := ValidateQuery(query)
			assert.NoError(t, err, "Query should be valid: %s", query)
		}
	})

	t.Run("should reject empty query", func(t *testing.T) {
		err := ValidateQuery("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query cannot be empty")
	})

	t.Run("should reject dangerous patterns", func(t *testing.T) {
		dangerousQueries := []string{
			"DROP TABLE users",
			"SELECT * FROM users -- comment",
			"SELECT * FROM users /* comment */",
		}
		
		for _, query := range dangerousQueries {
			err := ValidateQuery(query)
			assert.Error(t, err, "Query should be rejected: %s", query)
			assert.Contains(t, err.Error(), "potentially dangerous operations")
		}
	})
}

func TestValidateSelectQuery(t *testing.T) {
	t.Run("should accept valid SELECT queries", func(t *testing.T) {
		validQueries := []string{
			"SELECT * FROM users WHERE id = 1",
			"SELECT name FROM users WHERE active = true",
		}
		
		for _, query := range validQueries {
			err := ValidateSelectQuery(query)
			assert.NoError(t, err)
		}
	})

	t.Run("should reject non-SELECT queries", func(t *testing.T) {
		invalidQueries := []string{
			"INSERT INTO users (name) VALUES ('John')",
			"UPDATE users SET name = 'Jane' WHERE id = 1",
		}
		
		for _, query := range invalidQueries {
			err := ValidateSelectQuery(query)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "only SELECT statements are allowed")
		}
	})
}

func TestValidateUpdateQuery(t *testing.T) {
	t.Run("should accept valid UPDATE queries with WHERE clause", func(t *testing.T) {
		validQueries := []string{
			"UPDATE users SET name = 'Jane' WHERE id = 1",
			"UPDATE users SET active = false WHERE created < '2023-01-01'",
		}
		
		for _, query := range validQueries {
			err := ValidateUpdateQuery(query)
			assert.NoError(t, err)
		}
	})

	t.Run("should reject UPDATE queries without WHERE clause", func(t *testing.T) {
		invalidQuery := "UPDATE users SET name = 'Jane'"
		err := ValidateUpdateQuery(invalidQuery)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UPDATE statements must include a WHERE clause")
	})

	t.Run("should reject non-UPDATE queries", func(t *testing.T) {
		invalidQueries := []string{
			"SELECT * FROM users WHERE id = 1",
			"INSERT INTO users (name) VALUES ('John')",
		}
		
		for _, query := range invalidQueries {
			err := ValidateUpdateQuery(query)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "only UPDATE statements are allowed")
		}
	})
}

func TestValidateDeleteQuery(t *testing.T) {
	t.Run("should accept valid DELETE queries with WHERE clause", func(t *testing.T) {
		validQueries := []string{
			"DELETE FROM users WHERE id = 1",
			"DELETE FROM users WHERE active = false",
		}
		
		for _, query := range validQueries {
			err := ValidateDeleteQuery(query)
			assert.NoError(t, err)
		}
	})

	t.Run("should reject DELETE queries without WHERE clause", func(t *testing.T) {
		invalidQuery := "DELETE FROM users"
		err := ValidateDeleteQuery(invalidQuery)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DELETE statements must include a WHERE clause")
	})

	t.Run("should reject non-DELETE queries", func(t *testing.T) {
		invalidQueries := []string{
			"SELECT * FROM users WHERE id = 1",
			"INSERT INTO users (name) VALUES ('John')",
		}
		
		for _, query := range invalidQueries {
			err := ValidateDeleteQuery(query)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "only DELETE statements are allowed")
		}
	})
}