# Install

The kjournal apiserver can be deployed using you favourite continous delivery utitlities or you may build and deploy from the
source code.
Below you can find the steps for each of them.

## Install the pre-compiled apiserver

=== "kjournal"
    ```sh
    kjournal install -n kjournal-system
    ```

=== "Kustomize"
    ```sh
    kustomize build github.com/raffis/kjournal//config/default | kubectl apply -f -
    ```

=== "Helm"
    ```sh
    helm upgrade kjournal --install oci://github.com/raffis/kjournal/helm
    ```

=== "Manual"
    Download the pre-compiled binaries from the [OSS releases page][releases] and copy them to the desired location.

!!! Warning
    It is recommended to enable certmanager support on any production cluster. See bellow. 


## Certmanager support

It is recommended to enable certmanger support for setting up a trusted certificate between the kubernetes apiserver
and the kjournal apiserver. By default the kuberntes apiserver trusts kjournal without validating the certificate.

=== "kjournal"
    ```sh
    kjournal install -n kjournal-system --with-certmanager
    ```

=== "Helm"
    ```sh
    helm upgrade kjournal --install oci://github.com/raffis/kjournal/helm --set certmanager.enabled=true
    ```

=== "Kustomize"
    ```sh
    cat <<EOT >> kustomization.yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
    - github.com/raffis/kjournal//config/default
    - github.com/raffis/kjournal//config/rbac

    components:
    - github.com/raffis/kjournal//config/components/certmanager
    EOT && kustomize build | kubectl apply -f -
    ```

## Prometheus support

kjournal has support for the prometheus-operator or using prometheus scraping via annotations.

=== "kjournal"
    ```sh
    kjournal install -n kjournal-system --with-prometheus=operator/annotations
    ```

=== "Helm"
    ```sh
    helm upgrade kjournal --install oci://github.com/raffis/kjournal/helm --set serviceMonitor.enabled=true
    ```

=== "Kustomize"
    ```sh
    cat <<EOT >> kustomization.yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
    - github.com/raffis/kjournal//config/default
    - github.com/raffis/kjournal//config/rbac

    components:
    - github.com/raffis/kjournal//config/components/prometheus
    EOT && kustomize build | kubectl apply -f -
    ```


## Verifying the artifacts

### binaries

All artifacts are checksummed and the checksum file is signed with [cosign][].

1. Download the files you want, and the `checksums.txt`, `checksum.txt.pem` and `checksums.txt.sig` files from the [releases][releases] page:
    ```sh
    wget https://github.com/goreleaser/goreleaser/releases/download/__VERSION__/checksums.txt
    wget https://github.com/goreleaser/goreleaser/releases/download/__VERSION__/checksums.txt.sig
    wget https://github.com/goreleaser/goreleaser/releases/download/__VERSION__/checksums.txt.pem
    ```
1. Verify the signature:
    ```sh
    cosign verify-blob \
    --cert checksums.txt.pem \
    --signature checksums.txt.sig \
    checksums.txt
    ```
1. If the signature is valid, you can then verify the SHA256 sums match with the downloaded binary:
    ```sh
    sha256sum --ignore-missing -c checksums.txt
    ```

### docker images

Our Docker images are signed with [cosign][].

Verify the signatures:

```sh
COSIGN_EXPERIMENTAL=1 cosign verify goreleaser/goreleaser
```

!!! info
    The `.pem` and `.sig` files are the image `name:tag`, replacing `/` and `:` with `-`.

## Compile and install from source

Here you have two options:

If you want to contribute to the project, please follow the
steps on our [contributing guide](/contributing/).

If you just want to build from source for whatever reason, follow these steps:

**clone:**

```sh
git clone https://github.com/raffis/kjournal
cd kjournal
```

**build image:**

```sh
make docker-build
```

**deploy:**

```sh
make deploy
```

!!! Note
    `make deploy` uses kustomize under the hood to apply the overlay `config/default` with the just built image.