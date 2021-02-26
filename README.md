# goServe

A simple golang application to serve up a configmap in the same namespace as json as well as static files.

## Static files folder

Mount static files into the docker container at `/static`.

We will also serve `/` to `index.html` with a 301 redirect.

## Required environment variables
```
- CONFIGMAP_NAME [Required]
  <The name of the configmap to pull data from in the same namespace as the pod>

- POD_NAMESPACE [Optional]
  <To override the namespace that the pod thinks its running in (set to the namespace of the configmap if the pod is in a different namespace)>

- JSON_FILENAME [optional]
  <To override the default route and json file name served, pass like `foo.json`, default is `config.json`>
```

## Routes
```
/                     (tries to read static files, otherwise 404)
/index.html           (301 redirect to /)
/healthz              (liveness/healthcheck)
/readyz               (readiness probe)
/<JSON_FILENAME>      (the configmap data as json, default is `config.json`)
```