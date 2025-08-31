package statements

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/flowbi/pgweb/pkg/cache"
)

var (
	//go:embed sql/databases.sql
	Databases string

	//go:embed sql/schemas.sql
	Schemas string

	//go:embed sql/info.sql
	Info string

	//go:embed sql/info_simple.sql
	InfoSimple string

	//go:embed sql/estimated_row_count.sql
	EstimatedTableRowCount string

	//go:embed sql/table_indexes.sql
	TableIndexes string

	//go:embed sql/table_constraints.sql
	tableConstraintsEmbedded string

	TableConstraints string

	//go:embed sql/table_info.sql
	TableInfo string

	//go:embed sql/table_info_cockroach.sql
	TableInfoCockroach string

	//go:embed sql/table_schema.sql
	TableSchema string

	//go:embed sql/materialized_view.sql
	MaterializedView string

	//go:embed sql/objects.sql
	Objects string

	//go:embed sql/tables_stats.sql
	TablesStats string

	//go:embed sql/function.sql
	Function string

	//go:embed sql/settings.sql
	Settings string

	// Activity queries for specific PG versions
	Activity = map[string]string{
		"default": "SELECT * FROM pg_stat_activity WHERE datname = current_database()",
		"9.1":     "SELECT datname, current_query, waiting, query_start, procpid as pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database() and usename = current_user",
		"9.2":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database() and usename = current_user",
		"9.3":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database() and usename = current_user",
		"9.4":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database() and usename = current_user",
		"9.5":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database() and usename = current_user",
		"9.6":     "SELECT datname, query, state, wait_event, wait_event_type, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database() and usename = current_user",
	}

	// Cache for external SQL files
	sqlFileCache *cache.Cache
)

func init() {
	sqlFileCache = cache.New(30 * time.Minute)
	TableConstraints = loadTableConstraintsSQL()
}

func loadTableConstraintsSQL() string {
	externalPath := filepath.Join("/tmp/queries", "table_constraints.sql")

	// Check cache first
	cacheKey := cache.GenerateKey("sql_file", externalPath)
	if cached, found := sqlFileCache.Get(cacheKey); found {
		return cached.(string)
	}

	// Check if external file exists and get its mod time
	if stat, err := os.Stat(externalPath); err == nil {
		// Check if we have cached this file with its mod time
		modTimeCacheKey := cache.GenerateKey("sql_file_modtime", externalPath, stat.ModTime().String())
		if cached, found := sqlFileCache.Get(modTimeCacheKey); found {
			// Cache the content with the general key as well
			content := cached.(string)
			sqlFileCache.Set(cacheKey, content, 30*time.Minute)
			return content
		}

		// Read and cache the file
		if data, err := os.ReadFile(externalPath); err == nil {
			content := string(data)
			log.Printf("Using external table_constraints.sql from: %s", externalPath)

			// Cache with both keys
			sqlFileCache.Set(cacheKey, content, 30*time.Minute)
			sqlFileCache.Set(modTimeCacheKey, content, 30*time.Minute)
			return content
		}
	}

	// Cache the embedded fallback
	sqlFileCache.Set(cacheKey, tableConstraintsEmbedded, 30*time.Minute)
	return tableConstraintsEmbedded
}
