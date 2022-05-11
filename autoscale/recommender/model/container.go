package model

import "time"

// ContainerUsageSample is a measure of resource usage of a container over some
// interval.
type ContainerUsageSample struct {
	// Start of the measurement interval.
	MeasureStart time.Time
	// Average CPU usage in cores or memory usage in bytes.
	Usage ResourceAmount
	// CPU or memory request at the time of measurment.
	Request ResourceAmount
	// Which resource is this sample for.
	Resource ResourceName
}

// ContainerState stores information about a single container instance.
// Each ContainerState has a pointer to the aggregation that is used for
// aggregating its usage samples.
// It holds the recent history of CPU and memory utilization.
//   Note: samples are added to intervals based on their start timestamps.
type ContainerState struct {
	// Current request.
	Request Resources
	// Start of the latest CPU usage sample that was aggregated.
	LastCPUSampleStart time.Time
	// Max memory usage observed in the current aggregation interval.
	memoryPeak ResourceAmount
	// Max memory usage estimated from an OOM event in the current aggregation interval.
	oomPeak ResourceAmount
	// End time of the current memory aggregation interval (not inclusive).
	WindowEnd time.Time
	// Start of the latest memory usage sample that was aggregated.
	lastMemorySampleStart time.Time
	// Aggregation to add usage samples to.
	aggregator ContainerStateAggregator
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////