kind create cluster --config config/kind-example/control-plane.yaml
kustomize build config/kind-example --enable-helm | kubectl apply -f -
