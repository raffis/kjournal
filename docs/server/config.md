# Configuration

The apiserver needs to be configured with a backend storage where your logs are persisted.
Besides each resource may have custom configuration related to the backend configured.

Example config:

```yaml
apiVersion: config.kjournal/v1alpha1
kind: APIServerConfig

backend: 
  elasticsearch:
    url:
    - http://elasticsearch-master:9200

apis:
- resource: containerlogs
  fieldMap:
    metadata.namespace: [kubernetes.namespace_name]
    metadata.creationTimestamp: ["@timestamp"]
    pod: [kubernetes.pod_name]
    container: [kubernetes.container_name]
    payload: ["."]
  dropFields:
  - payload.@timestamp
  - payload.kubernetes
  backend:
    elasticsearch:
      index: logstash-*

- resource: events
  backend:
    elasticsearch:
        index: k8sevents-*

- resource: auditevents
  backend:
    elasticsearch:
        index: k8saudit-*

- resource: logs
  fieldMap:
    metadata.creationTimestamp: ["@timestamp"]
    payload: ["."]
  backend:
    elasticsearch:
      index: "*"
```

## Apis

Each resource can be customized including how your long-term storage logs are mapped to the kjournal API.

### Field maps

A field map can be used to map a kjournal api field to the stored field in your backend storage using dotted paths.

```yaml
resource: containerlogs
fieldMap:
    metadata.namespace: [kubernetes.namespace_name]
```

The above mapping will decode the stored log into a `containerlogs.v1alpha1.core.kournal`. 
`metadata.namespace` will be mapped to the storage field `kubernetes.namespace_name`. 

!!! Note
    You can use `.` which represents the object root. For example `payload: "."` means that the entire stored object will be mapped to the `payload` field and not just a specific path.

### Remove fields

Using drop fields allows to remove specific paths from an object. This is useful if you want to remove a specific field from a sub object which was mapped previously.

**Note**: Drop fields happens after the field mapping.

### Static filters

It may be useful to have static filters appended to all storage queries. Meaning you preselect the objects returned from the backing storage.

!!! Note
    You may use static filter to prefilter objects if you have multiple kubernetes clusters logging to the same backing storage and want kjournal on each cluster
    to only fetch its own clusters logs.