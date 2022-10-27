# Resources

kjournal supports four different resource types within the core.kjournal API group.

## Container

Container logs are logs logged by containers originating from a kubernetes pod. Container logs are namespaced and associated with a pod and container name.

## Audit

The kube-apiserver has the ability to log audits. kjournal exposes the audit logs a separate resource. This is a cluster scoped resource.

## Event

Kubernetes stores events (events.core.v1.k8s.io / events.events.v1.k8s.io) in etcd, however they are only stored for 1h. etcd is not meant for long-term log storage.
It is good practice to persist the events into a long-term log storage. 
kjournal exposes the the same event api as the core kubernetes api for this purpose.

## Generic logs

Generic logs could basically be anything. Logs are served as a log resource and is a cluster scoped resource.

