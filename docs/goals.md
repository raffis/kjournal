# Project goals

## Goals
A kubernetes native solution to serve logs as a kubernetes API. It may bee seen as an extension to the integrated logs API endpoint.
However the goal for kjournal is to expose logs from the longterm log storage back to kubernetes api consumers.
Logs should be searchable, watchable and consumable in a kubernete style manner.
Likewise log endpoints must RBAC protected as and consumable via a kubernetes signed authentication token.

Summary: 

- Log server which provides historical kubernetes events
- Log server which provides historical container logs
- Log server which provides historical audit logs
- Log server which provides histroical generic logs (unstructered log data)
- Expose these as kubernetes APIS
- Provider kubernertes API aggregation
- Support RBAC (which includes authentication and authorization)
- Support multiple storage backends. Especially elasticsearch. More backends are considered like Loki.

## Non goals
It is **not** the job of kjournal to feed the kubernetes logs into your longterm storage solution.
For this part various tooling exists. The job of kjournal is rather the other way around.
kjournal itself does **not*- store any data. It serves logs from a backing storage.
