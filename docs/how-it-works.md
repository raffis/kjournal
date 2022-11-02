# How it works

kjournal is a kubernetes compatible apiserver.
Read more about the kubernetes apiserver aggregation [here](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/).

```mermaid
graph LR
  A[<b>kjournal-cli</b>]
  B[<b>kube-apiserver</b>]
  C[<b>kjournal-apiserver</b>]
  D[<b>longterm log storage</b></br>elasticsearch,gcloud,...]
  E[<b>log shipper</b></br>fluent-bit,fluent,filebeat,...]

  A --> B;
  B -->|API-Aggregation| C;
  C --> D;
  E ---> D;
```
## Expose longterm logs using kjournal
To close the gap again between the longterm storage and kubernetes is kjournals job. 
The kjournal-apiserver exposes a single api group `core.kjournal` with `containerlogs`, `events`, `auditevents` and `logs` as resources which makes logs accessible
to kubernetes tooling.
The kjournal-apiserver talks to the longterm storage while clients including kjournal or kubectl talk only to the kjournal-apiserver (via the kube-apiserver).

## Logging in kubernetes

### Kubernetes container logs
Kubernetes stores all container logs on the nodes for a limited time. The logs get rotated and logs from older containers
are not accessible at at all and are lost if not persisted elsewhere.

It is commonly known a good practice to gather these container logs and make them available in a longterm log storage.
These logs are then accessible using third party tooling and are out of the kubernetes toolchain.

Read more about the kubernetes [logging architecture](https://kubernetes.io/docs/concepts/cluster-administration/logging/).

### Kubernetes audit logs
The kube-apiserver can be configured to store audit events for all requests going to the apiserver.
Similar to container logs these events are usually stored in log files directly on the master node(s). 
Like for container logs these logs should be persisted into longterm storage to make them available over time.

### Kubernetes events
Kubernetes events are not logs directly. Rather they are evens emitted by reconcilers and stored as normal kubernetes resource in the backing storage (etcd).
These events live only for one hour and are removed after.
kjournal exposes the same API and makes these events available from the longterm log storage.