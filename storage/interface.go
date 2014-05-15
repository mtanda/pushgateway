package storage

import (
	"time"

	dto "github.com/prometheus/client_model/go"
)

// MetricStore is the interface to the storage layer for metrics. All its
// methods must be safe to be called concurrently.
type MetricStore interface {
	// SubmitWriteRequest submits a WriteRequest for processing. There is no
	// guarantee when a request will be processed, but it is guaranteed that
	// the requests are processed in the order of submission.
	SubmitWriteRequest(req WriteRequest)
	// GetMetricFamilies returns all the currently saved MetricFamilies. The
	// returned MetricFamilies are guaranteed to not be modified by the
	// MetricStore anymore. However, they may still be read somewhere else,
	// so the caller is not allowed to modify the returned MetricFamilies.
	GetMetricFamilies() []*dto.MetricFamily
	// GetMetricFamiliesMap returns a nested map (job -> instance ->
	// metric-name -> TimestampedMetricFamily). The MetricFamily pointed to
	// by each TimestampedMetricFamily is guaranteed to not be modified by
	// the MetricStore anymore. However, they may still be read somewhere
	// else, so the caller is not allowed to modify it. Otherwise, the
	// returned nested map is a deep copy if the internal state of the
	// MetricStore and completely owned by the caller.
	GetMetricFamiliesMap() JobToInstanceMap
	// Shutdown must only be called after the caller has made sure that
	// SubmitWriteRequests is not called anymore. (If it is called later,
	// the request might get submitted, but not processed anymore.) The
	// Shutdown method waits for the write request queue to empty, then it
	// persists the content of the MetricStore (if supported by the
	// implementation). Also, all internal goroutines are stopped. This
	// method blocks until all of that is complete. If an error is
	// encountered, it is returned (whereupon the MetricStorage is in an
	// undefinded state). If nil is returned, the MetricStore cannot be
	// "restarted" again, but it can still be used for read operations.
	Shutdown() error
}

// WriteRequest is a request to change the MetricStore, i.e. to process it, a
// write lock has to be acquired. If MetricFamilies is nil, this is a request to
// delete metrics that share the given Job and (if not empty) Instance
// labels. Otherwise, this is a request to update the MetricStore with the
// MetricFamilies. The key in MetricFamilies is the name of the mapped metric
// family. All metrics in MetricFamilies MUST have already set job and instance
// labels that are consistent with the Job and Instance fields. The Timestamp
// field marks the time the request was received from the network. It is not
// related to the timestamp_ms field in the Metric proto message.
type WriteRequest struct {
	Job, Instance  string
	Timestamp      time.Time
	MetricFamilies map[string]*dto.MetricFamily
}

type TimestampedMetricFamily struct {
	Timestamp    time.Time
	MetricFamily *dto.MetricFamily
}

type JobToInstanceMap map[string]InstanceToNameMap
type InstanceToNameMap map[string]NameToTimestampedMetricFamilyMap
type NameToTimestampedMetricFamilyMap map[string]TimestampedMetricFamily
