# goServe

[![Docker Repository on Quay](https://quay.io/repository/vqcomms/goserve/status "Docker Repository on Quay")](https://quay.io/repository/vqcomms/goserve)

A simple golang application to serve up a configmap as well as static files.

The `data` and the `binaryData` values in the configmap are served up on the paths defined in the key.

For example `foo__bar__test.png` will be served from `/foo/bar/test.png`.

It will also override any static file in the static folder at that location.

## Order of file precedence

1. `data` in configmap
2. `binaryData` in configmap
3. static files

## Static files folder

Mount static files into the docker container at `/static`.

We will also serve `/` to `index.html` with a 301 redirect.

## Environment variables

When `CONFIGMAP_NAME` is not specified, no k8s functionality is enabled and goServe is a basic static file server.

```
- CONFIGMAP_NAME [Optional]
  <The name of the configmap to pull data from in the same namespace as the pod>

- POD_NAMESPACE [Optional]
  <To override the namespace that the pod thinks its running in (set to the namespace of the configmap if the pod is in a different namespace)>

- PORT [Optional]
  <Override the default port of 8090>
```

## Routes
```
/                     (tries to read static files, otherwise 404)
/index.html           (301 redirect to /)
/healthz              (liveness/healthcheck)
/readyz               (readiness probe)
```