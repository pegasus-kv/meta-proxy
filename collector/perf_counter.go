package collector

var (
	TableWatcherEvictCounter = registerCounter(
		"table_watcher_evict_count").(Counter)
)
