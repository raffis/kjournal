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
    helm upgrade kjournal --install oci://github.com/raffis/charts/kjournal
    ```

    You may find addtional documentation regarding support chart values in the chart documentation [here](methods/helm).

!!! Note
    You need to configure the kjournal-apiserver for your logging backend. kjournal will not serve any data if no backend is configured.
    Please follow the next chapter.

!!! Warning
    It is recommended to enable certmanager support on any production cluster. See bellow. 

## Configure apiserver

A backing storage needs to be confgured in order to tell kjournal from where it can get the data.
This is the longterm storage your log shippers will send data to.

Each installation method offers couple of preconfigured installation templates to get started.
Visit the [config](/kjournal/server/config) page for more information regarding the kjournal apiserver config.

=== "kjournal"
    ```sh
    kjournal install -n kjournal-system --with-config-template=<config-template-name>
    ```

=== "Helm"
    ```sh
    helm upgrade kjournal --install oci://ghcr.io/raffis/kjournal/helm --set apiserverConfig.templateName=<config-template-name>
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
    - github.com/raffis/kjournal//config/components/config-templates/<config-template-name>
    EOT && kustomize build | kubectl apply -f -
    ```

| Template name                       | Description |  
|----------------------------------   |-------------|
| `elasticsearch-kjournal-structured` | Configures the apiserver for an elasticsearch backend. The docuements are expected to be directly compatible with the kjournal api specification. |
| `elasticsearch-fluentbit-simple`    | Configures the apiserver for an elasticsearch backend. The fields are mapped to to a document structure which is usually created by the fluent-bit kubernetes plugin without any special configuration. |
| `elasticsearch-filebeat-simple`     | Configures the apiserver for an elasticsearch backend. The fields are mapped to to a document structure which is usually created by the filebeat kubernetes plugin without any special configuration. |

## Install a specific version of the pre-compiled apiserver

=== "kjournal"
    ```sh
    kjournal install -n kjournal-system --version 0.0.1
    ```

=== "Kustomize"
    ```sh
    kustomize build github.com/raffis/kjournal?ref=v0.0.1//config/default | kubectl apply -f -
    ```

=== "Helm"
    ```sh
    helm upgrade kjournal --install oci://github.com/raffis/charts/kjournal --version 0.0.1
    ```

## Certmanager support

It is recommended to enable certmanger support for setting up a trusted certificate between the kubernetes apiserver
and the kjournal apiserver. By default the kuberntes apiserver trusts kjournal without validating the certificate.

=== "kjournal"
    ```sh
    kjournal install -n kjournal-system --with-certmanager
    ```

=== "Helm"
    ```sh
    helm upgrade kjournal --install oci://ghcr.io/raffis/kjournal/helm --set certmanager.enabled=true
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
    helm upgrade kjournal --install oci://ghcr.io/raffis/kjournal/helm --set serviceMonitor.enabled=true
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

### Binaries

All artifacts are checksummed and the checksum file is signed with [cosign](https://github.com/sigstore/cosign).

1. Download the files you want, and the `checksums.txt`, `checksum.txt.pem` and `checksums.txt.sig` files from the [releases][releases] page:
    ```sh
    wget https://github.com/raffis/kjournal/releases/download/__VERSION__/checksums.txt
    wget https://github.com/raffis/kjournal/releases/download/__VERSION__/checksums.txt.sig
    wget https://github.com/raffis/kjournal/releases/download/__VERSION__/checksums.txt.pem
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

### Container images

Likewise are the container images signed with [cosign](https://github.com/sigstore/cosign).

Verify the signatures:

```sh
cosign verify ghcr.io/raffis/kjournal/apiserver
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