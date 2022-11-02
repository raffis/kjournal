# kjournal

![Version: 0.0.11](https://img.shields.io/badge/Version-0.0.11-informational?style=flat-square) ![AppVersion: v0.0.1](https://img.shields.io/badge/AppVersion-v0.0.1-informational?style=flat-square)

A Helm chart for kjournal

**Homepage:** <https://github.com/raffis/kjournal>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| raffis | <raffael.sahli@gmail.com> |  |

## Source Code

* <https://github.com/raffis/kjournal>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| certManager.caCertDuration | string | `"43800h"` |  |
| certManager.certDuration | string | `"8760h"` |  |
| certManager.enabled | bool | `false` |  |
| customAnnotations | object | `{}` |  |
| customLabels | object | `{}` |  |
| dnsConfig | object | `{}` |  |
| extraArguments | list | `[]` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| hostNetwork.enabled | bool | `false` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/raffis/kjournal"` |  |
| image.tag | string | `nil` |  |
| listenPort | int | `8443` |  |
| logLevel | int | `4` |  |
| metricsRelistInterval | string | `"1m"` |  |
| namespaceOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podDisruptionBudget.enabled | bool | `false` |  |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podDisruptionBudget.minAvailable | string | `nil` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.fsGroup | int | `10001` |  |
| priorityClassName | string | `""` |  |
| prometheus.path | string | `""` |  |
| prometheus.port | int | `9090` |  |
| prometheus.url | string | `"http://prometheus.default.svc"` |  |
| psp.create | bool | `false` |  |
| rbac.create | bool | `true` |  |
| replicas | int | `1` |  |
| resources | object | `{}` |  |
| runAsUser | int | `10001` |  |
| service.annotations | object | `{}` |  |
| service.port | int | `443` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `nil` |  |
| strategy.rollingUpdate.maxSurge | string | `"25%"` |  |
| strategy.rollingUpdate.maxUnavailable | string | `"25%"` |  |
| strategy.type | string | `"RollingUpdate"` |  |
| tls.ca | string | `"# Public CA file that signed the APIService"` |  |
| tls.certificate | string | `"# Public key of the APIService"` |  |
| tls.enable | bool | `false` |  |
| tls.key | string | `"# Private key of the APIService"` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |

