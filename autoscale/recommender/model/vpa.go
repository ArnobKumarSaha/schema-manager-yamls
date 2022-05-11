package model

import (
	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"time"
)

// Map from VPA annotation key to value.
type vpaAnnotationsMap map[string]string

// Map from VPA condition type to condition.
type vpaConditionsMap map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition

// Vpa (Vertical Pod Autoscaler) object is responsible for vertical scaling of
// Pods matching a given label selector.
type Vpa struct {
	ID VpaID
	// Labels selector that determines which Pods are controlled by this VPA
	// object. Can be nil, in which case no Pod is matched.
	PodSelector labels.Selector
	// Map of the object annotations (key-value pairs).
	Annotations vpaAnnotationsMap
	// Map of the status conditions (keys are condition types).
	Conditions vpaConditionsMap
	// Most recently computed recommendation. Can be nil.
	Recommendation *vpa_types.RecommendedPodResources
	// All container aggregations that contribute to this VPA.
	// TODO: Garbage collect old AggregateContainerStates.
	aggregateContainerStates aggregateContainerStatesMap
	// Pod Resource Policy provided in the VPA API object. Can be nil.
	ResourcePolicy *vpa_types.PodResourcePolicy
	// Initial checkpoints of AggregateContainerStates for containers.
	// The key is container name.
	ContainersInitialAggregateState ContainerNameToAggregateStateMap
	// UpdateMode describes how recommendations will be applied to pods
	UpdateMode *vpa_types.UpdateMode
	// Created denotes timestamp of the original VPA object creation
	Created time.Time
	// CheckpointWritten indicates when last checkpoint for the VPA object was stored.
	CheckpointWritten time.Time
	// IsV1Beta1API is set to true if VPA object has labelSelector defined as in v1beta1 api.
	IsV1Beta1API bool
	// TargetRef points to the controller managing the set of pods.
	TargetRef *autoscaling.CrossVersionObjectReference
	// PodCount contains number of live Pods matching a given VPA object.
	PodCount int
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
