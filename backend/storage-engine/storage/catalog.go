package storage

import (
	"github.com/axiomhq/hyperloglog"
	"github.com/bits-and-blooms/bloom/v3"
	"runtime/metrics"
)

type TableName string
type Catalog struct {
	Tables map[TableName]*TableInfo
}

type Column string
type TableInfo struct {
	Schema               map[string]ColumnType
	NumOfPages           uint32
	UsedSpace            uint64 // bytes
	FreeSpace            uint64 // bytes
	TupleCountTotal      uint32
	TupleCountPerPageAvg uint16
	TupleAvgSize         uint16            // bytes
	ColumnAvgWidth       map[Column]uint16 // bytes
	Histogram            map[Column]*metrics.Float64Histogram
	UniqueCount          *hyperloglog.Sketch
	SkipPage             map[PageID]map[Column]*bloom.BloomFilter
}

type ColumnType struct {
	IsIndex bool
	Type    string
}
