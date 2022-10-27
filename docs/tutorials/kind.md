# Try it using kind

This tutorials rolls out a local kind cluster on your computer in order to test kjournal.
We will deploy a single node kind control-plane, a single elasticsearch instance as well as fluent-bit for shipping.

## Requirements

During this tutorial you need the following tools. Make sure you have them up-to-date.

* kubectl
* kustomize
* kind
* git
* kjournal

## Fetch the repo
First lets clone the kjournal repository where we can access the deployment files for this tutorial.

```sh
git clone https://github.com/raffis/kjournal
cd kjournal
```

## Create kind cluster

Let's create the kind cluster named kjournal first. Once the control plane is running we can continue.
The apiserver is configured to log audit logs which is required if we want to query audit logs as well.

```sh
kind create cluster --config config/kind-example/control-plane.yaml
```

## Deploy kjournal and third party components

Next we deploy elasticsearch, fluent-bit, kubernetes-event-exporter as well as the kjournal apiserver itself
from an opinated kustomize overlay. 

```sh
kustomize build config/kind-example --enable-helm | kubectl apply -f -
```

## Test kjournal

Last but not least we can test if kjournal works properly.
This command will start a watch stream for all container logs in the kube-system namespace.

```sh
kjournal pods -n kube-system -w
```