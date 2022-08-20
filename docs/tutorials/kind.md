# Try it using kind

This tutorials rolls out a local kind cluster on your computer in order to test kjournal.
We will deploy a single node kind control-plane, a single elasticsearch instance as well as a fluent-bit instance
besides kjournal itself.

## Requirements

During this tutorial you need the following tools. Make sure you have them up-to-date.

* kubectl
* kustomize
* kind
* git
* curl
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

## Create dedicated logging namespace

We will deploy components into a namespace called logging during this tutorial.

```sh
kubectl create ns logging
```

## Deploy elasticsearch

Install the elastic helm repository and deploy a single es node.

```sh
helm repo add elastic https://helm.elastic.co
helm install elasticsearch elastic/elasticsearch -n logging --set replicas=0
```

## Deploy fluent-bit

Install the fluent helm repository and deploy fluent-bit which will ship all container logs as well as kubernetes audit logs 
to the elasticsearch instance we deployed just previously.


```sh
helm repo add fluent https://fluent.github.io/helm-charts
helm install fluent-bit fluent/fluent-bit -n logging -f config/kind-example/fluent-bit-chart-values.yaml
```

## Test persisting logs

At this point fluent-bit should already ship logs to elasticsearch.
This should be verified before continue. 

```sh
kubectl -n logging port-forward svc/elasticsearch-master 9200 &
curl localhost:9200/_search?pretty
```

If there are no documents something is wrong. Please inspect both fluent-bit and elasticsearch pods.

## Deploy kjournal

```sh
kustomize build config/kind-example | kubectl -n logging apply -f -
```

## Test kjournal

Last but not least we can test if kjournal works properly.
This command will start a watch stream for all container logs in the kube-system namespace.

```sh
kjournal diary -n kube-system -w
```