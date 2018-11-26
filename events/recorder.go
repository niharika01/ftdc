// Recorder
//
// The Recorder interface provides an interface for
// workloads and operations to collect data about their internal
// state, without requiring workloads to be concerned with data
// retention, storage, or compression. The implementations of Recorder
// provide different strategies for data collection and persistence
// so that tests can easily change data collection strategies without
// modifying the test.
package events

import "time"

// Recorder describes an interface that tests can use to track metrics
// and events during performance testing or normal
// operation. Implementations of recorder wrap an FTDC collector and
// will write data out to the collector for reporting purposes. The
// types produced by the collector use the Performance or
// PerformanceHDR types in this package.
//
// Choose the implementation of Recorder that will capture all
// required data used by your test with sufficient resolution for use
// later. Additionally, consider the data volume produced by the
// recorder.
type Recorder interface {
	// The Inc<> operations add values to the specified counters
	// tracked by the collector. There is an additional
	// "iteration" counter that the recorder tracks based on the
	// number of times that Begin/Record are called.
	//
	// In general, ops should refer to the number of logical
	// operations collected. This differs from the iteration
	// count, in the case of workloads that comprise of multiple
	// logical operations.
	//
	// Use size to record, typically, the number of bytes
	// processed or generated by the operation. Use this in
	// combination with logical operations to be able to explore
	// the impact of data size on overall performance. Finally use
	// Error count to tract the number of errors encountered
	// during the event.
	IncOps(int)
	IncSize(int)
	IncError(int)

	// The Set<> operations replace existing values for the state,
	// workers, and failed gauges. Workers should typically report
	// the number of active threads. The meaning of state depends
	// on the test requirements but can describe phases of an
	// experiment or operation. Use SetFailed to flag a test as
	// failed during the operation.
	SetState(int)
	SetWorkers(int)
	SetFailed(bool)

	// The Begin and Record methods mark the beginning and end of
	// a tests's iteration. Typically calling record records the
	// duration specified as its argument and increments the
	// counter for number of iterations. Additionally there is a
	// "total duration" value captured which represents the total
	// time taken in the iteration in addition to the operation
	// latency.
	//
	// The Flush method writes any unflushed material if the
	// collector's Record method does not. In all cases Flush
	// reports all errors since the last flush call, and resets
	// the internal error tracking and unsets the tracked starting
	// time. Generally you should call Flush once at the end of
	// every test run, and fail if there are errors reported.
	//
	// The Reset method set's the tracked starting time, like
	// Begin, but does not record any other values, as some
	// recorders use begin to persist the previous iteration.
	Begin()
	Record(time.Duration)
	Flush() error
	Reset()
}