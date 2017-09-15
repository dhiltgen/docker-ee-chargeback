package chargeback

import (
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Used for the CSV data
type Entry struct {
	Label        string
	Collection   model.LabelValue
	ID           model.LabelValue
	Name         model.LabelValue
	TotalSeconds float64
	Cumulative   float64
	Min          float64
	Max          float64
	Ave          float64
}

type Gatherer func(promAPI v1.API, r v1.Range, skipSystemResources bool) ([]Entry, error)

var Gatherers = []Gatherer{
	CPUMetrics,
	MemoryMetrics,
	NetworkMetrics,
	VolumeMetrics,
	ContainerStorageMetrics,
}
