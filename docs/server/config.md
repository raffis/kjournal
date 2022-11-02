# Configuration

The apiserver needs to be configured with a backend storage where your logs are persisted.
Besides each resource may have custom configuration related to the backend configured.

Example config:

=== "v1alpha1"
  ```yaml
  apiVersion: config.kjournal/v1alpha1
  kind: APIServerConfig

  backend: 
    elasticsearch:
      url:
      - http://elasticsearch-master:9200

  apis:
  - resource: containerlogs
    backend:
      elasticsearch:
        index: container-*

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

A field map can be used to map a kjournal api to the long term storage representation of messages.
There might be multiple reasons you may need do define a fieldmap for kjournal rather than changing the log structure at ingest time.
For instance to support backwards compatibility or there might be other services using the current log structure.

The field map basicaly consists of one or more field maps:

```yaml
fieldMap:
  kjournal-api-field: [source-field-1, source-field-2]
  ...
```
!!! Note
    You don't need to define fields which are already at the correct path for the kjournal API.


!!! Note
    You can define one or multiple source fields. The first source field found from the storage will be mapped to the output document.
    The fields after are ignored.

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