SHELL := /bin/bash

# ==============================================================================
# Testing running system

# curl --user "admin@example.com:gophers" http://localhost:3000/api/token/01aad0ee-cee2-11eb-b8bc-0242ac130003
# export SHORTURL_TOKEN="COPY TOKEN STRING FROM LAST CALL"
# curl -X POST -H "Content-Type: application/json" -d '{"url": "https://github.com/mitrovicsinisaa/shorturl"}' http://localhost:3000/api/shorturl
# curl -X DELETE -H "Authorization: Bearer ${SHORTURL_TOKEN}" http://localhost:3000/api/shorturl/tb
# curl -H "Authorization: Bearer ${SHORTURL_TOKEN}" http://localhost:3000/api/shorturl/1/10
# curl -H "Authorization: Bearer ${SHORTURL_TOKEN}" http://localhost:3000/api/shorturl/tb

# ==============================================================================

# ==============================================================================

shorturl-api: 
	docker build \
		-f zarf/docker/Dockerfile.shorturl-api \
		-t shorturl-api-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +"%d.%m.%YT%H:%M:%SZ"` \
		.
# ==============================================================================
# Running from within k8s/dev

kind-up:
	kind create cluster --image kindest/node:v1.21.1@sha256:fae9a58f17f18f06aeac9772ca8b5ac680ebbed985e266f711d936e91d113bad --name shorturl-cluster --config zarf/k8s/dev/kind-config.yaml

kind-down:
	kind delete cluster --name shorturl-cluster

kind-load:
	kind load docker-image shorturl-api-amd64:1.0 --name shorturl-cluster

kind-services:
	kustomize build zarf/k8s/dev | kubectl apply -f -

kind-services-delete:
	kustomize build zarf/k8s/kind | kubectl delete -f -

kind-update: all kind-load
	kubectl rollout restart deployment shorturl-pod

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

kind-logs: 
	kubectl logs -lapp=shorturl-api --all-containers=true -f

kind-shorturl: shorturl-api
	kind load docker-image shorturl-api-amd64:1.0 --name shorturl-cluster
	kubectl delete pods -lapp=shorturl-api


# ==============================================================================

run:
	go run app/shorturl-api/main.go

runadmin:	
	go run app/admin/main.go

test: 
	go test ./... -count=1

tidy: 
	go mod tidy
	go mod vendor