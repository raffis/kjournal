# Uninstall

Kjournal can easily be removed using the same utilities as it was installed.

## Uninstall apiserver

=== "kjournal"
    ```sh
    kjournal uninstall -n kjournal-system
    ```

=== "Kustomize"
    ```sh
    kustomize build github.com/raffis/kjournal//config/default | kubectl remove -f -
    ```

=== "Helm"
    ```sh
    helm delete kjournal
    ```