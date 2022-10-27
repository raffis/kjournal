# Quick Start

!!! Note
    If you are a cluster admin a and want to deploy kjournal on your cluster(s), please refer to the apiserver quick start.

## Install CLI

=== "Brew"
    ```sh
    brew install kjournal/tap/kjournal
    ```

=== "Go"
    ```sh
    go install github.com/raffis/kjournal@latest
    ```

=== "Bash"
    ```sh
    curl -sfL https://raw.githubusercontent.com/raffis/kjournal/main/cli/install/kjournal.sh | bash
    ```

=== "Docker"
    ```sh
    docker pull ghcr.io/raffis/kjournal/cli:latest
    ```
    
You will find in the CLI installation documentation more advanced options regarding the cli installation.

## Container logs
To fetch container logs from a namespace you can simply use the `pods` command.
The command will start to print log streams from all containers prefixed and colored by pod and container names.

This will display all container logs from the namespace `mynamespace`.

```sh
kjournal pods -n mynamespace
```

You can quick filter by naming a pod or a pod prefix.

Will stream logs from all pods starting with mypod-
```sh
kjournal pods -n mynamespace mypod-
```

## Time window
The kjournal-apiserver looks up logs from the last 24 hours and starts stream from 24h ago. 
The server default is configurable (see server configuration).
You can change the window in which logs are looked up by using the `--since` flag.
This works for all kjournal commands.

This will stream logs starting from 7 days ago.
```sh
kjournal pods -n mynamespace mypod- --since 7d
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
kjournal pods -n mynamespace mypod- --field-selector payload.myLogField=xxx
```

## Audit events
kjournal has built-in support for kubernetes audit events. 
You can access audit event using the audit command.

This will stream the entire audit feed:
```sh
kjournal audit
```
