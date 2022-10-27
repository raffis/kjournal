# Install

You can install the pre-compiled binary (in several different ways), use Docker or compile from source.
Below you can find the steps for each of them.

## Install the pre-compiled binary

=== "Brew"
    ```sh
    brew install raffis/kjournal/kjournal
    ```

=== "Go"
    ```sh
    go install github.com/raffis/kjournal/cli@latest
    ```

=== "Bash"
    ```sh
    curl -sfL https://raw.githubusercontent.com/raffis/kjournal/main/cli/install/kjournal.sh | bash
    ```

=== "Docker"
    ```sh
    docker pull ghcr.io/raffis/kjournal/cli
    ```

### Specific version

Due server compatibility reasons (or any other) you may want to install anoher version rather than the latest.
Here version `v0.0.1` gets installed.

=== "Brew"
    ```sh
    brew install kjournal/tap/kjournal@v0.0.1
    ```

=== "Go"
    ```sh
    go install github.com/raffis/kjournal@v0.0.1
    ```

=== "Bash"
    ```sh
    curl -sfL https://raw.githubusercontent.com/raffis/kjournal/main/cli/install/kjournal.sh | VERSION=0.0.1 bash
    ```

=== "Docker"
    ```sh
    docker pull ghcr.io/raffis/kjournal/cli:v0.0.1
    ```

## Compatibility with apiserver

The kjornal cli gurantees compatibility for three minor versions with the apiserver.
The previous minor release, the same and a newer apiserver.
That said this guarantees full compatibility, a bigger version gap will likely still work.

Example:

| cli       | apiserver | Fully compatible
| v1.1.0    | v1.0.0    | `yes`
| v1.1.0    | v1.1.0    | `yes`
| v1.1.0    | v1.2.0    | `yes`
| v1.1.0    | v1.3.0    | `no`
| v1.0.0    | v1.2.0    | `no`


!!! Note
    Releases from 0.x do not offer any compatibility guarantees between different versions of the cli and the apiserver.


## Enable shell completion

=== "Bash"
    ```sh
    kjournal completion bash
    ```

=== "zsh"
    ```sh
    kjournal completion zsh
    ```

=== "Fish"
    ```sh
    kjournal completion fish
    ```

=== "Fish"
    ```sh
    kjournal completion powershell
    ```

## Bash Additional Options
You can also set the `VERSION`, `OS`,  and `ARCH` variables to specify
a version instead of using latest.

```bash
curl -sfL https://raw.githubusercontent.com/raffis/kjournal/main/cli/install/kjournal.sh |
    VERSION=__VERSION__  bash -s -- check
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
cosign verify ghcr.io/raffis/kjournal/cli
```

!!! info
    The `.pem` and `.sig` files are the image `name:tag`, replacing `/` and `:` with `-`.

## Running with Docker

You can also use the cli within a Docker container.
Example usage:

```sh
docker run 
    -v ~/.kube:/home/alpine/.kube \
    ghcr.io/raffis/kjournal/cli
```

## Compiling from source

Here you have two options:

If you want to contribute to the project, please follow the
steps on our [contributing guide](/contributing/).

If you just want to build from source for whatever reason, follow these steps:

**clone:**

```sh
git clone https://github.com/goreleaser/goreleaser
cd goreleaser
```

**get the dependencies:**

```sh
go mod tidy
```

**build:**

```sh
go build -o goreleaser .
```

**verify it works:**

```sh
./goreleaser --version
```
