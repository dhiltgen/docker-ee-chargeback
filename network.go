package chargeback

import (
	"context"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func NetworkMetrics(promAPI v1.API, r v1.Range, skipSystemResources bool) ([]Entry, error) {
	res1, err := network(promAPI, r, skipSystemResources, "rx")
	if err != nil {
		return nil, err
	}
	res2, err := network(promAPI, r, skipSystemResources, "tx")
	if err != nil {
		return nil, err
	}
	return append(res1, res2...), nil

}

func network(promAPI v1.API, r v1.Range, skipSystemResources bool, direction string) ([]Entry, error) {
	results := []Entry{}
	query := "ucp_engine_network_" + direction + "_bytes"
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

		startVal := stream.Values[0].Value
		endVal := startVal
		start := stream.Values[0].Timestamp
		end := start
		for _, reading := range stream.Values {
			val := reading.Value
			// Establish start/end times
			if reading.Timestamp < start {
				start = reading.Timestamp
				startVal = val
			}
			if reading.Timestamp > end {
				end = reading.Timestamp
				endVal = val
			}
		}
		uptime := end.Sub(start)

		results = append(results, Entry{
			Label:        "NETWORK " + strings.ToUpper(direction) + " MB",
			Collection:   collection,
			ID:           id,
			Name:         name,
			TotalSeconds: uptime.Seconds(),
			Cumulative:   float64(endVal-startVal) / 1024.0 / 1024.0,
		})

	}
	return results, nil
}
