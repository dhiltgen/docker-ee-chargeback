package chargeback

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func VolumeMetrics(promAPI v1.API, r v1.Range, skipSystemResources bool) ([]Entry, error) {
	results := []Entry{}
	query := "ucp_engine_volume_disk_megabytes"
	val, err := promAPI.QueryRange(context.TODO(), query, r)
	if err != nil {
		return nil, err
	}
	if val.Type() != model.ValMatrix {
		return nil, fmt.Errorf("unexpected result type for %s - %d (not %d)", query, val.Type(), model.ValMatrix)
	}
	m := val.(model.Matrix)
	for _, stream := range m {
		collection := stream.Metric["collection"]
		// Make sure to report which instance we saw the volume from as it may be different across the cluster
		name := stream.Metric["instance"] + "/" + stream.Metric["volume"]
		if skipSystemResources && (collection == "" || collection == "/") {
			continue
		}
		start := stream.Values[0].Timestamp
		end := start
		min := stream.Values[0].Value
		max := min
		var total, ave float64
		for _, reading := range stream.Values {
			// Disk usage in MB
			val := reading.Value
			// Establish min/max values
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
			total = total + float64(val)
			// Establish start/end times
			if reading.Timestamp < start {
				start = reading.Timestamp
			}
			if reading.Timestamp > end {
				end = reading.Timestamp
			}
		}
		ave = total / float64(len(stream.Values))
		uptime := end.Sub(start)

		results = append(results, Entry{
			Label:        "VOLUME MB",
			Collection:   collection,
			Name:         name,
			TotalSeconds: uptime.Seconds(),
			Min:          float64(min),
			Max:          float64(max),
			Ave:          ave,
		})
	}
	return results, nil
}

func ContainerStorageMetrics(promAPI v1.API, r v1.Range, skipSystemResources bool) ([]Entry, error) {
	results := []Entry{}
	query := "ucp_engine_container_disk_megabytes"
	val, err := promAPI.QueryRange(context.TODO(), query, r)
	if err != nil {
		return nil, err
	}
	if val.Type() != model.ValMatrix {
		return nil, fmt.Errorf("unexpected result type for %s - %d (not %d)", query, val.Type(), model.ValMatrix)
	}
	m := val.(model.Matrix)
	for _, stream := range m {
		collection := stream.Metric["collection"]
		id := stream.Metric["container"]
		name := stream.Metric["name"]
		if skipSystemResources && (collection == "" || collection == "/") {
			continue
		}
		start := stream.Values[0].Timestamp
		end := start
		min := stream.Values[0].Value
		max := min
		var total, ave float64
		for _, reading := range stream.Values {
			// Memory usage in MB
			val := reading.Value
			// Establish min/max values
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
			total = total + float64(val)
			// Establish start/end times
			if reading.Timestamp < start {
				start = reading.Timestamp
			}
			if reading.Timestamp > end {
				end = reading.Timestamp
			}
		}
		ave = total / float64(len(stream.Values))
		uptime := end.Sub(start)

		results = append(results, Entry{
			Label:        "CONTAINER STORAGE MB",
			Collection:   collection,
			ID:           id,
			Name:         name,
			TotalSeconds: uptime.Seconds(),
			// Cumulative:
			Min: float64(min),
			Max: float64(max),
			Ave: ave,
		})
	}
	return results, nil
}
