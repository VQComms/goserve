# configmap-to-http

A simple golang application to serve up a configmap in the same namespace as json

## Required environment variables
```
- CONFIGMAP_NAME [Required]
  <The name of the configmap to pull data from in the same namespace as the pod>

- POD_NAMESPACE [Optional]
  <To override the namespace that the pod thinks its running in (set to the namespace of the configmap if the pod is in a different namespace)>
```

## Routes
```
/healthz          (liveness/healthcheck)
/readyz           (readiness probe)
/config.json      (the configmap data as json)
```