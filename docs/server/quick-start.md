# Quick start

Welcome to the quick start for the kjournal apiseverver.
This guide should make it as easy as possible to help you deploy kjournal to your kubernetes cluster(s).

## Prerequisites

Your cluster must have some prerequisites installed in order to support kjournal.

### Supported log storage

kjournal only works with a longterm log storage backend which it has support for.

Currently kjournal only supports:

* elasticsearch

### Ship logs

You must have a log shipper installed on the cluster which ships the container logs as well as kubernetes audit log events (and optionally other logs)
to a supporter kjournal storage backend.

See prerequisites for each of them if your cluster does not meet these requirements:

* Containers logs
* Kubernetes audit events
* Kubernetes events

!!! Note
    You don't need all of them if you want to use kjournal for example only for container logs.

## Configure apiserver

## Deploy apiserver

Please see the installation guide for more options regarding the apiserver installation.