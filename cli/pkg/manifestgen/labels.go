package manifestgen

// These labels can be used to track down the namespace, custom resource definitions, deployments,
// services, network policies, service accounts, cluster roles and cluster role bindings belonging to Flux.
const (
	PartOfLabelKey   = "app.kubernetes.io/part-of"
	PartOfLabelValue = "kjournal"
	InstanceLabelKey = "app.kubernetes.io/instance"
	VersionLabelKey  = "app.kubernetes.io/version"
)
