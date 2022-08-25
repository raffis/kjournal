kind create cluster --config config/kind-example/control-plane.yaml

kubectl create ns logging

helm repo add elastic https://helm.elastic.co
helm install elasticsearch elastic/elasticsearch -n logging --set replicas=1

helm repo add fluent https://fluent.github.io/helm-charts
helm install fluent-bit fluent/fluent-bit -n logging -f config/kind-example/fluent-bit-chart-values.yaml

kubectl -n logging port-forward svc/elasticsearch-master 9200 &
curl localhost:9200/_search?pretty

kustomize build config/kind-example | kubectl -n logging apply -f -