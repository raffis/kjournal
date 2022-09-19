# Events

To persist the kubernetes events an exporter is needed which watches the kubernetes events and exports them 
into a configured sink.

[opsgenie/kubernetes-event-exporter](https://github.com/opsgenie/kubernetes-event-exporter) is a good tool for this job.

!!! Note
    The original project is archived, [resmoio/kubernetes-event-exporter](https://github.com/resmoio/kubernetes-event-exporter) is a maintained fork.

The following installs the exporter for elasticsearch, however there are various other sinks supported.


```sh
kubectl create ns logging
helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade kubernetes-event-exporter bitnami/kubernetes-event-exporter --install -n logging -f values.yaml
```

values.yaml
```yaml
config:
  logLevel: info
  logFormat: pretty
  receivers:
    - name: "dump"
      elasticsearch:
        hosts:
          - http://elasticsearch-master:9200
        index: k8sevents
        # Ca be used optionally for time based indices, accepts Go time formatting directives
        indexFormat: "k8sevents-{2006-01-02}"
        useEventID: true
        deDot: false
  route:
    routes:
      - match:
          - receiver: "dump"

```