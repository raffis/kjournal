# kjournal

![Version: 1.0.3](https://img.shields.io/badge/Version-1.0.3-informational?style=flat-square) ![AppVersion: 0.0.10](https://img.shields.io/badge/AppVersion-0.0.10-informational?style=flat-square)

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
| apiserverConfig.config | string | `""` |  |
| apiserverConfig.existingConfigMap | string | `nil` |  |
| apiserverConfig.templateName | string | `"elasticsearch-kjournal-structured"` |  |
| certManager.caCertDuration | string | `"43800h"` |  |
| certManager.certDuration | string | `"8760h"` |  |
| certManager.enabled | bool | `false` |  |
| customAnnotations | object | `{}` |  |
| customLabels | object | `{}` |  |
| dnsConfig | object | `{}` |  |
| env | list | `[]` |  |
| extraArguments | list | `[]` |  |
| extraVolumeMounts | list | `[]` |  |
| extraVolumes | list | `[]` |  |
| hostNetwork.enabled | bool | `false` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/raffis/kjournal/apiserver"` |  |
| image.tag | string | `nil` |  |
| listenPort | int | `8443` |  |
| livenessProbe.httpGet.path | string | `"/healthz"` |  |
| livenessProbe.httpGet.port | string | `"https"` |  |
| livenessProbe.httpGet.scheme | string | `"HTTPS"` |  |
| livenessProbe.initialDelaySeconds | int | `5` |  |
| livenessProbe.timeoutSeconds | int | `5` |  |
| namespaceOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podDisruptionBudget.enabled | bool | `false` |  |
| podDisruptionBudget.maxUnavailable | int | `1` |  |
| podDisruptionBudget.minAvailable | string | `nil` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.fsGroup | int | `10001` |  |
| priorityClassName | string | `""` |  |
| rbac.create | bool | `true` |  |
| readinessProbe.httpGet.path | string | `"/healthz"` |  |
| readinessProbe.httpGet.port | string | `"https"` |  |
| readinessProbe.httpGet.scheme | string | `"HTTPS"` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.timeoutSeconds | int | `5` |  |
| replicas | int | `1` |  |
| resources.limits.cpu | string | `"1"` |  |
| resources.limits.memory | string | `"200Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"20Mi"` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| service.annotations | object | `{}` |  |
| service.port | int | `443` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `nil` |  |
| serviceMonitor.enabled | bool | `false` |  |
| strategy.rollingUpdate.maxSurge | string | `"25%"` |  |
| strategy.rollingUpdate.maxUnavailable | string | `"25%"` |  |
| strategy.type | string | `"RollingUpdate"` |  |
| tls.ca | string | `"# Public CA file that signed the APIService"` |  |
| tls.certificate | string | `"# Public key of the APIService"` |  |
| tls.enable | bool | `false` |  |
| tls.key | string | `"# Private key of the APIService"` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |

