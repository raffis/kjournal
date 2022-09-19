# Quick Start

!!! Note
    If you are a cluster admin a and want to deploy kjournal on your cluster(s), please refer to the apiserver quick start.

With the kjournal cli you can fetch either container logs or audit events.
[Install](cli/install) the cli on your machine.

## Container logs
To fetch all container logs originating from pods in the namespace `mynamespace` you can simply use the diary command.
The command will start to display log streams for all containers prefixed and colored by pod and container names.

```sh
kjournal container -n mynamespace
```

You can quick filter by naming a pod or a pod prefix.

Will stream logs from all pods starting with mypod-
```sh
kjournal container -n mynamespace mypod-
```

## Time window
The kjournal-apiserver looks up logs from the last 24hours and starts stream from 24h ago. 
The server default is configurable (see server configuration).
You can change the window in which logs are looked up by using the `--since` flag.
This works for all kjournal commands.

This will stream logs starting from 7 days ago.
```sh
kjournal container -n mynamespace mypod- --since 7d
```

Alternatively you may use `--range [from]-[to]`. `--range 18h-23h` will feed logs
from 18h ago to 23h ago, basically a 5h window. 

!!! Note
    `--since` is a shortcut of `--range now-[to]`. `--since 5h` is the same as `--range now-5h`. 


## Filter
Logs can be filtered server-side. This works for all kjournal commands.
You can use the flag `--field-selector` which supports the same operators as `kubectl get` does. 
However on top of that kjournal also supports other operators including `>`,`<` or `in()`.

```sh
kjournal container -n mynamespace mypod- --field-selector payload.myLogField=xxx
```

## Audit events
kjournal has built-in support for kubernetes audit events. 
You can access audit event using the audit command.

This will stream the entire audit log feed:
```sh
kjournal audit -n mynamespace
```

You can list audit events for specific resource groups or a specific resource by name.

Will stream events for all deployments.
```sh
kjournal audit deployments
```

Similary you can add a name for a specific deployment:
```sh
kjournal audit deployments/mydeployment
```