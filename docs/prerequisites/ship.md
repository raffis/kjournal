# Ship logs

Both audit and container logs need to be persisted in the longterm log storage.
For this a component is required which tails these logs on the kubernetes nodes and pushes them to the storage.

There are various tools out there doing exactly this. Namely [fluent-bit](), [filebeat]() or [fluent]().

!!! Important
    kjournal needs kubernetes metadata for the container logs from the longterm storage. Both [fluent-bit](https://docs.fluentbit.io/manual/pipeline/filters/kubernetes) and [filebeat](https://www.elastic.co/guide/en/beats/filebeat/current/add-kubernetes-metadata.html) provide a kubernetes filter to attach that data to all log events.


## fluent-bit

fluent-bit usually is installed as a DaemonSet.
Following an example configuration which ships both audit events and container logs to
an elasticsearch cluster.

```ini
[INPUT]
    Name              tail
    Tag               kube.*
    Path              /var/log/containers/*.log
    Parser            cri
    DB                /var/log/flb_kube-fwd.db
    Refresh_Interval  5
    Skip_Long_Lines   On
    Mem_Buf_Limit     5MB
    Read_from_Head    True

[INPUT]
    Name              tail
    Path              /var/log/kube-apiserver-audit.log
    Parser            docker
    DB                /var/log/audit-fwd.db
    Tag               audit.*
    Refresh_Interval  5
    Mem_Buf_Limit     35MB
    Buffer_Chunk_Size 2MB
    Buffer_Max_Size   10MB
    Skip_Long_Lines   On
    Key               kubernetes-audit
    Read_from_Head    True

# Stores container logs in logstash-* indexes.
[OUTPUT]
    Name            es
    Match           kube.*
    Host            logging-es-http.logging
    Port            9200
    Time_Key        @es_ts
    Logstash_Format On
    Replace_Dots    On
    Type            _doc

# Stores audit events in a separate index pattern k8saudit-*
[OUTPUT]
    Name            es
    Match           audit.*
    Host            logging-es-http.logging
    Port            9200
    Logstash_Format On
    Replace_Dots    On
    Logstash_Prefix k8saudit
    Type            _doc

[SERVICE]
    Flush 1
    Daemon Off
    Log_Level info
    Parsers_File parsers.conf
    Parsers_File custom_parsers.conf
    HTTP_Server On
    HTTP_Listen 0.0.0.0
    HTTP_Port 2020

[FILTER]
    Name parser
    Match kube.*
    Key_Name message
    Parser docker

[FILTER]
    Name kubernetes
    Match kube.*
    Merge_Log On
    Keep_Log Off
    K8S-Logging.Parser On
    K8S-Logging.Exclude On
```