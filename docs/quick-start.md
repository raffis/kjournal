# Quick Start

With the kjournal cli you can fetch either container logs or audit events.
[Install](cli/install) the cli on your machine.

## Container logs
To fetch all container logs originating from pods in the namespace `mynamespace` you can simply use the diary command.
The command will start to display log streams for all containers prefixed and colored by pod and container names.

```sh
kjournal diary -n mynamespace
```

You can quick filter by naming a pod or a pod prefix.

Will stream logs from all pods starting with mypod-
```sh
kjournal diary -n mynamespace mypod-
```

## Audit events
kjournal has built-in support for kubernetes audit events. 
You can access audit event using the audit command.

This will list audit events from any resource in the mynamespace namespace.
```sh
kjournal audit -n mynamespace
```

You can list audit events for specific resource groups or a specific resource by name.

Will stream events for all deployments in the mynamespace namespace.
```sh
kjournal audit -n mynamespace deployments
```

Similary you can add a name for a specic deployment:
```sh
kjournal audit -n mynamespace deployments/mydeployment
```

### Cluster events
To access audit events which are not referring to a namespaced resource you can use the clusteraudit command.
 
```sh
kjournal clusteraudit clusterroles/myclusterrole
```

## Time window
The kjournal-apiserver looks up logs from the last 24hours and starts stream from 24h ago. 
The server default is configurable (see server configuration).
You can change the timeline in which logs are looked up by using the `--since` flag.
This works for all kjournal commands.

This will stream logs starting from 7 days ago.
```sh
kjournal diary -n mynamespace mypod- --since 7d
```

## Filter
Logs can be filtered server-side. This works for all kjournal commands.
You can use the flag `--field-selector` which supports the schema as `kubectl get`. 
Howver on top of that kjournal also supports other operators including `>`,`<` or `in()`.

```sh
kjournal diary -n mynamespace mypod- --field-selector unstructured.myLogField=xxx
```
