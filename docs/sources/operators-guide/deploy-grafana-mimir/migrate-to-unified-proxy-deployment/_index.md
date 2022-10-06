---
description: "Migrate to using the unified proxy deployment for NGINX and GEM gateway"
title: "Migrate to using the unified proxy deployment for NGINX and GEM gateway"
menuTitle: "Unified proxy deployment for NGINX and GEM gateway"
weight: 110
aliases:
  - /docs/mimir/latest/operators-guide/deploying-grafana-mimir/migrate-to-unified-proxy-deployment/
---

# Migrate to using the unified proxy deployment for NGINX and GEM gateway

Since the 4.0.0 version of the `mimir-distributed` Helm chart there is a new way to deploy a reverse proxy in front of
Mimir or GEM. The new configuration lives in the `proxy` section of the Helm values. The new `proxy` configuration
allows for a zero-downtime migration from Mimir to GEM. The new `proxy` configuration also brings OpenShift
Route and horizontal autoscaling to the GEM gateway.

Under the hood, the `proxy` section also deploys an NGINX or a GEM gateway, so it supports the same features that
its predecessors have.

The introduction of the new section also deprecates the `nginx` and `gateway` sections. They will be removed in 
`mimir-distributed` release 7.0.0.

It is possible to migrate to the `proxy` configuration without downtime. The migration should take less than 30 minutes.
The rest of this article contains a procedure for migrating from the old `nignx` or `gateway` sections to `proxy`.

## Before you begin

Make sure that the version of the `mimir-distributed` Helm chart that you have installed is 4.0.0 or higher.

## Procedure

1. Scale out the `proxy` deployment

   1. Change your Helm values file to enable the `proxy` and increase its replicas. Set the number of replicas of
      the proxy deployment to the number of
      replicas that the NGINX or the GEM gateway are running with. For example, if you have deployed 10 NGINX replicas,
      add the following to your Helm values file `custom.yaml`:

      ```yaml
      proxy:
        enabled: true
        replicas: 10
      ```

   2. Deploy your changes.

      ```bash
      helm upgrade $RELEASE grafana/mimir-distributed -f custom.yaml
      ```

2. Switch from `nginx` or `gateway` to `proxy` in your values file.

   1. Disable the GEM gateway or NGINX. Add or change the following in your values file:

      ```yaml
      gateway:
        enabled: false
      nginx:
        enabled: false
      ```

   2. If you are using the Ingress that the chart provides, then copy the `ingress` section from `nginx` or
      `gateway` to `proxy` and override the name. Override the name to the name of the Ingress resource that
      the chart created for NGINX or the GEM gateway.

      Reusing the name allows the `helm` command to retain the existing resource instead of deleting it and
      recreating it under a slightly different name.

      In the example that follows, the name of the Ingress
      resource was `mimir-nginx`. Use `kubectl` to get the name of the existing Ingress resource:

      ```bash
      kubectl get ingress
      ```

      ```console
      NAME          CLASS    HOSTS               ADDRESS    PORTS     AGE
      mimir-nginx   <none>   mimir.example.com   10.0.0.1   80, 443   172d
      ```

      The Helm values file should look like the following after carrying out this step:

      ```yaml
      proxy:
        ingress:
          enabled: true
          nameOverride: mimir-nginx
          hosts:
            - host: mimir.example.com
              paths:
                - path: /
                  pathType: Prefix
          tls:
            - secretName: mimir-gateway-tls
              hosts:
                - mimir.example.com
      ```

   3. Copy the `service` section from `nginx` or `gateway` to `proxy` and override the name. Override the name to
      the name of the Service resource that the chart created for NGINX or the GEM gateway.

      Reusing the name allows the `helm` command to retain the existing resource instead of deleting it and
      recreating it under a slightly different name.

      In the example that follows, the name of the Service
      resource was `mimir-nginx`. Use `kubectl` to get the name of the existing Service resource:

      ```bash
      kubecl get service
      ```

      ```console
      NAME          TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)             AGE
      mimir-nginx   ClusterIP   10.188.8.32   <none>        8080/TCP,9095/TCP   172d
      ```

      The Helm values file should look like the following after carrying out this step:

      ```yaml
      proxy:
        service:
          annotations:
            networking.istio.io/exportTo: admin
          nameOverride: mimir-nginx
      ```

   4. Move the rest of your values according the following tables:

      | Deprecated field                      | New field                             | Notes                                                                  |
      | ------------------------------------- | ------------------------------------- | ---------------------------------------------------------------------- |
      | `nginx.affinity`                      | `proxy.affinity`                      | Previously `affinity` was a string. Now it should be a YAML object.    |
      | `nginx.annotations`                   | `proxy.annotations`                   |                                                                        |
      | `nginx.autoscaling`                   | `proxy.autoscaling`                   |                                                                        |
      | `nginx.basicAuth`                     | `proxy.nginx.basicAuth`               | Nested under `proxy.nginx`.                                            |
      | `nginx.containerSecurityContext`      | `proxy.containerSecurityContext`      |                                                                        |
      | `nginx.deploymentStrategy`            | `proxy.strategy`                      | Renamed from `deploymentStrategy` to `strategy`.                       |
      | `nginx.enabled`                       | `proxy.enabled`                       |                                                                        |
      | `nginx.extraArgs`                     | `proxy.extraArgs`                     |                                                                        |
      | `nginx.extraContainers`               | `proxy.extraContainers`               |                                                                        |
      | `nginx.extraEnvFrom`                  | `proxy.extraEnvFrom`                  |                                                                        |
      | `nginx.extraEnv`                      | `proxy.env`                           | Renamed from `extraEnv` to `env`.                                      |
      | `nginx.extraVolumeMounts`             | `proxy.extraVolumeMounts`             |                                                                        |
      | `nginx.extraVolumes`                  | `proxy.extraVolumes`                  |                                                                        |
      | `nginx.image`                         | `proxy.nginx.image`                   | Nested under `proxy.nginx`.                                            |
      | `nginx.ingress`                       | `proxy.ingress`                       |                                                                        |
      | `nginx.nginxConfig`                   | `proxy.nginx.config`                  | Renamed from `nginxConfig` to `config` and nested under `proxy.nginx`. |
      | `nginx.nodeSelector`                  | `proxy.nodeSelector`                  |                                                                        |
      | `nginx.podAnnotations`                | `proxy.podAnnotations`                |                                                                        |
      | `nginx.podDisruptionBudget`           | `proxy.podDisruptionBudget`           |                                                                        |
      | `nginx.podLabels`                     | `proxy.podLabels`                     |                                                                        |
      | `nginx.podSecurityContext`            | `proxy.securityContext`               | Renamed from `podSecurityContext` to `securityContext`.                |
      | `nginx.priorityClassName`             | `proxy.priorityClassName`             |                                                                        |
      | `nginx.readinessProbe`                | `proxy.readinessProbe`                |                                                                        |
      | `nginx.replicas`                      | `proxy.replicas`                      |                                                                        |
      | `nginx.resources`                     | `proxy.resources`                     |                                                                        |
      | `nginx.route`                         | `proxy.route`                         |                                                                        |
      | `nginx.service`                       | `proxy.service`                       |                                                                        |
      | `nginx.terminationGracePeriodSeconds` | `proxy.terminationGracePeriodSeconds` |                                                                        |
      | `nginx.tolerations`                   | `proxy.tolerations`                   |                                                                        |
      | `nginx.topologySpreadConstraints`     | `proxy.topologySpreadConstraints`     |                                                                        |
      | `nginx.verboseLogging`                | `proxy.nginx.verboseLogging`          | Nested under `proxy.nginx`.                                            |

      | Deprecated field                        | New field                             | Notes                                                                                           |
      | --------------------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------------------- |
      | `gateway.affinity`                      | `proxy.affinity`                      |                                                                                                 |
      | `gateway.annotations`                   | `proxy.annotations`                   |                                                                                                 |
      | `gateway.containerSecurityContext`      | `proxy.containerSecurityContext`      |                                                                                                 |
      | `gateway.env`                           | `proxy.env`                           |                                                                                                 |
      | `gateway.extraArgs`                     | `proxy.extraArgs`                     |                                                                                                 |
      | `gateway.extraContainers`               | `proxy.extraContainers`               |                                                                                                 |
      | `gateway.extraEnvFrom`                  | `proxy.extraEnvFrom`                  |                                                                                                 |
      | `gateway.extraVolumeMounts`             | `proxy.extraVolumeMounts`             |                                                                                                 |
      | `gateway.extraVolumes`                  | `proxy.extraVolumes`                  |                                                                                                 |
      | `gateway.ingress`                       | `proxy.ingress`                       |                                                                                                 |
      | `gateway.initContainers`                | `proxy.initContainers`                |                                                                                                 |
      | `gateway.nodeSelector`                  | `proxy.nodeSelector`                  |                                                                                                 |
      | `gateway.persistence`                   | -                                     | Removed. The gateway doesn't use persistence.                                                   |
      | `gateway.podAnnotations`                | `proxy.podAnnotations`                |                                                                                                 |
      | `gateway.podDisruptionBudget`           | `proxy.podDisruptionBudget`           |                                                                                                 |
      | `gateway.podLabels`                     | `proxy.podLabels`                     |                                                                                                 |
      | `gateway.priorityClassName`             | `proxy.priorityClassName`             |                                                                                                 |
      | `gateway.readinessProbe`                | `proxy.readinessProbe`                |                                                                                                 |
      | `gateway.replicas`                      | `proxy.replicas`                      |                                                                                                 |
      | `gateway.resources`                     | `proxy.resources`                     |                                                                                                 |
      | `gateway.securityContext`               | `proxy.securityContext`               |                                                                                                 |
      | `gateway.service`                       | `proxy.service`                       |                                                                                                 |
      | `gateway.strategy`                      | `proxy.strategy`                      |                                                                                                 |
      | `gateway.terminationGracePeriodSeconds` | `proxy.terminationGracePeriodSeconds` |                                                                                                 |
      | `gateway.tolerations`                   | `proxy.tolerations`                   |                                                                                                 |
      | `gateway.topologySpreadConstraints`     | `proxy.topologySpreadConstraints`     |                                                                                                 |
      | `gateway.useDefaultProxyURLs`           | -                                     | Removed. You can use `mimir.structuredConfig` to override the routes that the GEM gateway uses. |

   5. Upgrade the Helm release with the migrated values file `custom.yaml`. This concludes the migration.

      ```bash
      helm upgrade $RELEASE grafana/mimir-distributed -f custom.yaml
      ```

## Examples

The examples that follow show how your Helm values file changes after migrating from an NGINX or a GEM gateway
setup to a proxy setup.

### GEM gateway

The Helm values file before starting the migration:

```yaml
gateway:
  replicas: 4

  service:
    annotations:
      networking.istio.io/exportTo: admin
    port: 80
    legacyPort: 8080

  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 10%

  affinity: 
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: noisyNeighbour
              operator: In
              values:
                - 'true'
        topologyKey: 'kubernetes.io/hostname'

  extraArgs: 
    log.level: debug

  resources:
    requests:
      cpu: '1'
      memory: 3Gi

  extraEnvFrom:
    - configMapRef:
        name: special-config

  ingress:
    enabled: true
    hosts:
      - host: mimir.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-gateway-tls
        hosts:
          - mimir.example.com
```

The Helm values file after finishing the migration:

```yaml
gateway:
  enabled: false
  
proxy:
  enabled: true
  replicas: 4

  service:
    annotations:
      networking.istio.io/exportTo: admin
    port: 80
    legacyPort: 8080

  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 10%

  affinity: 
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: noisyNeighbour
              operator: In
              values:
                - 'true'
        topologyKey: 'kubernetes.io/hostname'

  extraArgs: 
    log.level: debug

  resources:
    requests:
      cpu: '1'
      memory: 3Gi

  extraEnvFrom:
    - configMapRef:
       name: special-config

  ingress:
    enabled: true
    nameOverride: mimir-gateway
    hosts:
      - host: mimir.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-gateway-tls
        hosts:
          - mimir.example.com
```

### NGINX


The Helm values file before starting the migration:

```yaml
nginx:
  enabled: true
  replicas: 4

  deploymentStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 10%
 
  affinity: |
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: noisyNeighbour
              operator: In
              values:
                - 'true'
        topologyKey: 'kubernetes.io/hostname'

  extraEnv:
    - name: SPECIAL_TYPE_KEY
      valueFrom:
        configMapKeyRef:
          name: special-config
          key: SPECIAL_TYPE
          
  basicAuth:
    enabled: true
    username: user
    password: pass

  image:
    tag: 1.25-alpine

  nginxConfig:
    logFormat: |-
      main '$remote_addr - $remote_user [$time_local]  $status '
      '"$request" $body_bytes_sent "$http_referer" '
      '"$http_user_agent" "$http_x_forwarded_for"';
  
  podSecurityContext:
    readOnlyRootFilesystem: true
    
  ingress:
    enabled: true
    hosts:
      - host: mimir.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-gateway-tls
        hosts:
          - mimir.example.com
```

The Helm values file after finishing the migration:

```yaml
nginx:
  enabled: false
  
proxy:
  enabled: true
  replicas: 4

  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 10%

  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: noisyNeighbour
              operator: In
              values:
                - 'true'
        topologyKey: 'kubernetes.io/hostname'

  env:
    - name: SPECIAL_TYPE_KEY
      valueFrom:
        configMapKeyRef:
          name: special-config
          key: SPECIAL_TYPE

  nginx:
    basicAuth:
      enabled: true
      username: user
      password: pass

    image:
      tag: 1.25-alpine

    nginxConfig:
      logFormat: |-
        main '$remote_addr - $remote_user [$time_local]  $status '
        '"$request" $body_bytes_sent "$http_referer" '
        '"$http_user_agent" "$http_x_forwarded_for"';

  securityContext:
    readOnlyRootFilesystem: true

  ingress:
    enabled: true
    nameOverride: mimir-nginx
    hosts:
      - host: mimir.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: mimir-gateway-tls
        hosts:
          - mimir.example.com
```
