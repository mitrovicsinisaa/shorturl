#Shorturl API

## Requirements

This project has been configured to run in a Kubernetes (k8s) environment using
KIND. KIND will set up your cluster there and deploying and running the project locally. The project also uses Kustomize to help manage k8s configuration files.

[Install KIND](https://kind.sigs.k8s.io/docs/user/quick-start/).

[Install the K8s kubectl client](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

[Install the Kustomize client](https://kubernetes-sigs.github.io/kustomize/installation/).

## running project

Build docker container.

```
$ make shorturl-api
```

Create cluster `shorturl-cluster`.

```
$ make kind-up

```

Optionaly delete cluster

```
$ make kind-down

```

Loads the `shorturl-api` container into the k8s environment.

```
$ make kind-load
```

Deploys PODs into the cluster.

```
$ make kind-services
```

Status of cluster.

```
$ make kind-status
```

### Seed Database

After shorturl-pod status is Running, we can seed the DataBase.
To seed user data to DataBase, and create table for short urls run command:

```
make runadmin
```

### Requests

Protected API endpoints require token.

```
$ curl --user "admin@example.com:gophers" http://localhost:3000/api/token/01aad0ee-cee2-11eb-b8bc-0242ac130003
```

For easier testing set environment variable with token value `$SHORTURL_TOKEN`.

```
$ export SHORTURL_TOKEN="TOKEN STRING"
```

Create new Short URL (change "EXAMPLE_URL" with actual URL)

```
curl -X POST -H "Content-Type: application/json" -d '{"url": "EXAMPLE_URL"}' http://localhost:3000/api/shorturl
```

Delete Short URL (change "EXAMPLE_URL_CODE" with actual short URL code)

```
curl -X DELETE -H "Authorization: Bearer ${SHORTURL_TOKEN}" http://localhost:3000/api/shorturl/EXAMPLE_URL_CODE
```

List Short URL's (change OFFSET to integer value for OFFSET, and ROWS for integer number of rowe to be returned)

```
curl  -H "Authorization: Bearer ${SHORTURL_TOKEN}" http://localhost:3000/api/shorturl/OFFSET/ROWS
```

Count of visits (change "EXAMPLE_URL_CODE" with actual short URL code)

```
curl -H "Authorization: Bearer ${SHORTURL_TOKEN}" http://localhost:3000/api/shorturl/EXAMPLE_URL_CODE
```
