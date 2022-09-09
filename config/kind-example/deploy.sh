kind create cluster --config config/kind-example/control-plane.yaml

kubectl create ns logging

helm repo add elastic https://helm.elastic.co
helm upgrade elasticsearch elastic/elasticsearch --install -n logging --set replicas=1

helm repo add fluent https://fluent.github.io/helm-charts
helm upgrade fluent-bit fluent/fluent-bit --install -n logging -f config/kind-example/fluent-bit-chart-values.yaml

helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade kubernetes-event-exporter bitnami/kubernetes-event-exporter --install -n logging -f config/kind-example/kubernetes-event-exporter-values.yaml

kubectl -n logging port-forward svc/elasticsearch-master 9200 &
curl localhost:9200/_search?pretty

kustomize build config/kind-example | kubectl -n logging apply -f -
