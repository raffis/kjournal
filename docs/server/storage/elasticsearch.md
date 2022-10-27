# Elasticsearch

kjournal was first designed with elasticsearch as long-term storage backend. At this time it is also the only
storage backend. 

## kjournal-apiserver config flags

The following flags are used by the apiserver to configure the elasticsearch storage backend. 
You will likely need to configure these.

| Flag | Default | Description | 
|----------|:-------------:|:-------------:|
--es-allow-insecure-tls | *not-set* |Allow insecure TLS connections. Do not verify the certificate |
--es-audit-index audit-*| `` | The index pattern where the kubernetes audit documents are stored. (For example: audit-*). You may specify multiple ones comma separated |
--es-audit-timestamp-field | `@timestamp` |The index field which is used as timestamop field for the audit documents
--es-cacert | `` |Path to the CA (PEM) used to verify the server tls certificate |
--es-container-index | `logstash-*` |The index pattern where the kubernetes container logs are stored. (For example: logstash-*). You may specify multiple ones comma separated |
--es-container-namespace-field | `kubernetes.namespace_name.keyword` |The field which holds the kubernetes namespace. This field must not be indexed using any analyzers! Usually a .keyword field is wanted here |
--es-container-timestamp-field | `@timestamp` |The index field which is used as timestamop field for the audit documents |
--es-refresh-rate | `500ms` |The refresh rate to poll from elasticsearch while checking for new documents during watch requests |
--es-url | `http://localhost:9200` | Elasticsearch URL, you may add multiple ones comma separated |

## Compatibility matrix

| kjournal-apiserver | elasticsearch | 
|----------|:-------------:|
| >= v0.0 |>= v7.10 | 