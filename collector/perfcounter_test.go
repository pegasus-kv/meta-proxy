package collector

import (
	"fmt"
	"testing"
)

func TestPrometheus(t *testing.T) {
	pfcType = "prometheus"

	TableWatcherEvictCounter.Incr()
	fmt.Printf("123444")
}
