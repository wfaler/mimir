---
title: "Prepare Grafana Mimir for production using the Helm chart"
aliases:
- docs/mimir/latest/operators-guide/run-production-environment-with-helm/
  menuTitle: "Prepare Grafana Mimir for production using the Helm chart"
  description: "Prepare Grafana Mimir to ingest metrics in a production environment using the mimir-distributed Helm chart."
  weight: 90
---

# Prepare Grafana Mimir for production using the Helm chart

Beyond [Getting started with Grafana Mimir using the Helm chart]({{< relref "../deploy-grafana-mimir/getting-started-helm-charts" >}}),
which covers setting up Grafana Mimir on a local Kubernetes cluster or
within a low-risk development environment, you can prepare Grafana Mimir
for production.

Although the information that follows assumes that you are using Grafana Mimir in
a production environment that is customer-facing, you might need the
high-availability and horizontal-scalability features of Grafana Mimir in an
internal, development environment.

[//]: # (TODO revisit this paragraph after reading the whole topic to see if 
            the intro still needs some information)
To achieve high availability, the Helm chart schedules Kubernetes Pods
onto different Kubernetes Nodes. Also, the chart increases the scale
of the Grafana Mimir cluster.

[//]: # (TODO revisit this paragraph after reading the whole topic to see if 
            the intro still needs some information)
- using an object storage setup different from the MinIO deployment that
  comes with the mimir-distributed chart.

## Before you begin

Meet all the follow prerequisites:

- You are familiar with [Helm](https://helm.sh/docs/intro/quickstart/).

  Add the grafana Helm repository to your local environment or to your CI/CD tooling:

  ```bash
  helm repo add grafana https://grafana.github.io/helm-charts
  helm repo update
  ```
- You have an external object storage that is different from the MinIO
  object storage that `mimir-distributed` deploys, because the MinIO
  deployment in the Helm chart is only intended for getting started and is not  
  intended for production use.

  Instead, use any S3-compatible service, such as GCS, Azure Blob
  Storage, OpenStack Swift.

## Plan capacity

The `mimir-distributed` Helm chart comes with two sizing plans:

- For 1M series: [small.yaml](https://github.com/grafana/mimir/blob/main/operations/helm/charts/mimir-distributed/small.yaml)
- For 10M series: [large.yaml](https://github.com/grafana/mimir/blob/main/operations/helm/charts/mimir-distributed/large.yaml)

These sizing plans are estimated based on experience from operating Grafana
Mimir at Grafana Labs. The ideal size for your cluster depends on your
usage patterns. Therefore, use the sizing plans as starting
point for sizing your Grafana Mimir cluster, rather than as strict guidelines.
To get a better idea of how to plan capacity, refer to the YAML comments at
the beginning of `small.yaml` and `large.yaml` files, which relate to read and write workloads.
See also [Planning Grafana Mimir capacity]({{< relref "../run-production-environment/planning-capacity.md" >}}).

To use a sizing plan, copy it from the [mimir](https://github.com/grafana/mimir/blob/main/operations/helm/charts/mimir-distributed)
GitHub repository, and pass it as a values file to the `helm` command.

For example:

```bash
helm install mimir-prod grafana/mimir-distributed -f ./small.yaml
```

### Conform to fault-tolerance requirements

As part of _Pod scheduling_, the `small.yaml` and `large.yaml` files add Pod
anti-affinity rules so that no two ingester Pods, nor two store-gateway
Pods, are scheduled on any given Kubernetes Node. This increases fault
tolerance of the Mimir cluster.

You must create and add Nodes, such that the number of Nodes is equal to or
larger than either the number of ingester Pods or the number of store-gateway Pods,
whichever one is larger. Expressed as a formula, it reads as follows:

```
number_of_nodes >= max(number_of_ingesters_pods, number_of_store_gateway_pods)
```

For more information about the failure modes of either the ingester or store-gateway
component, refer to [Ingesters failure and data loss]({{< relref "../architecture/components/ingester/#ingesters-failure-and-data-loss">}})
or [Store-gateway: Blocks sharding and replication]({{< relref "../architecture/components/store-gateway/#blocks-sharding-and-replication">}}).

## Decide whether you need geographical redundancy, fast rolling updates, or both.

You can use a rolling update strategy to apply configuration changes to
Grafana Mimir, and to upgrade Grafana Mimir to a newer version. A
rolling update results in no downtime to Grafana Mimir.

The Helm chart performs a rolling update for you. To make sure that rolling updates are faster,
configure the Helm chart to deploy Grafana Mimir with zone-aware replication.

### New installations

Grafana Mimir supports [replication across availability zones]({{< relref "../configure/configuring-zone-aware-replication/">}})
within your Kubernetes cluster.
This further increases fault tolerance of the Mimir cluster. Even if you
do not currently have multiple zones across your Kubernetes cluster, you
can avoid having to extraneously migrate your cluster when you start using
multiple zones.

For `mimir-distributed` Helm chart v4.0 or higher, zone-awareness is enabled by
default for new installations.

To benefit from zone-awareness, choose the node selectors for your different
zones. For convenience, you can use the following YAML configuration snippet
as a starting point:

[//]: # (TODO: check if this is actually the correct yaml after github.com/grafana/mimir/pull/2778 is merged)

```yaml
ingester:
  zoneAwareReplication:
    enabled: true
    topologyKey: kubernetes.io/hostname
    zones:
      - name: zone-a
        nodeSelector:
          topology.kubernetes.io/zone: zone-a
      - name: zone-b
        nodeSelector:
          topology.kubernetes.io/zone: zone-b
      - name: zone-c
        nodeSelector:
          topology.kubernetes.io/zone: zone-c

store_gateway:
  zoneAwareReplication:
    enabled: true
    topologyKey: kubernetes.io/hostname
    zones:
      - name: zone-a
        nodeSelector:
          topology.kubernetes.io/zone: zone-a
      - name: zone-b
        nodeSelector:
          topology.kubernetes.io/zone: zone-b
      - name: zone-c
        nodeSelector:
          topology.kubernetes.io/zone: zone-c
```

### Existing installations

If you are upgrading from a previous `mimir-distributed` Helm chart version
to v4.0, then refer to the [migration guide]({{< relref "../../migration-guide/migrating-from-single-zone-with-helm" >}}) to configure
zone-aware replication.

## Configure Mimir to use object storage

By default, the `mimir-distributed` Helm chart deploys a small MinIO object
storage, which is not intended nor optimized for large workloads.
To use Grafana Mimir in production, you must replace the default object storage
with an Amazon S3 compatible service, Google Cloud Storage, Microsoft® Azure Blob Storage,
or OpenStack Swift. Alternatively, to deploy MinIO yourself, see [MinIO High
Performance Object Storage](https://min.io/docs/minio/kubernetes/upstream/index.html).

**After choosing an object storage service, configure Grafana Mimir to use it:**

1. Add the following YAML to your values file, if you are not using the sizing
   plans that are mentioned in [Plan capacity](#plan-capacity):

   ```yaml
   minio:
     enabled: false
   ```

2. Prepare the credentials and bucket names for the object storage. The article
   [Configure Grafana Mimir object storage backend]({{< relref "../configure/configure-object-storage-backend" >}})
   has examples for the different types of object storage that Mimir supports.

3. Add the object storage configuration to the Helm chart values. Nest the object storage configuration under
   `mimir.structuredConfig`. This example uses S3:

   ```yaml
   mimir:
     structuredConfig:
       common:
         storage:
           backend: s3
           s3:
             endpoint: s3.us-east-2.amazonaws.com
             region: us-east
             secret_access_key: "${AWS_SECRET_ACCESS_KEY}" # This is a secret injected via an environment variable
             access_key_id: "${AWS_ACCESS_KEY_ID}" # This is a secret injected via an environment variable

       blocks_storage:
         s3:
           bucket_name: mimir-blocks
       alertmanager_storage:
         s3:
           bucket_name: mimir-alertmanager
       ruler_storage:
         s3:
           bucket_name: mimir-ruler

       # The admin_client configuration applies only to GEM deployments
       #admin_client:
       #  storage:
       #    s3:
       #      bucket_name: gem-admin
   ```


----------------------------
- Comply with security needs, which Grafana Mimir does for you. (compliance)
- Monitoring the health of your Mimir cluster. (metamonitoring)
------------------------------

## Security

Grafana Mimir does not require any special permissions from the hosts it runs on. Because of this it can be deployed
in environments that enforce the [Restricted security policy](https://kubernetes.io/docs/concepts/security/pod-security-standards/).

In Kubernetes >=1.23 the Restricted policy may be enforced via a namespace label on the namespace where Mimir will be installed:

```
pod-security.kubernetes.io/enforce: restricted
```

In Kubernetes versions prior to 1.23, the mimir-distributed chart provides a
[PodSecurityPolicy resource](https://v1-24.docs.kubernetes.io/docs/concepts/security/pod-security-policy/)
which enforces a lot of the recommendations of the Restricted policy that the namespace label enforces.
To enable the PodSecurityPolicy admission controller for your Kubernetes cluster refer to
[How do I turn on an admission controller?](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#how-do-i-turn-on-an-admission-controller)
in the Kubernetes documentation.

## Metamonitoring

You can use ready Grafana dashboards, and Prometheus alerting and recording rules to monitor Mimir itself.
See [Installing Grafana Mimir dashboards and alerts]({{< relref "../monitor-grafana-mimir/installing-dashboards-and-alerts/">}})
for more details.

The mimir-distributed Helm chart makes it easy to collect metrics and logs from Mimir. It takes care of assigning the
right labels so that the dashboards and alerts work out of the box. The chart ships metrics to a Prometheus-compatible
remote and logs to a Loki cluster.

If you are using the latest mimir-distributed Helm chart:

1. Download the Grafana Agent Operator CRDs from https://github.com/grafana/agent/tree/main/production/operator/crds
2. Install the CRDs in your cluster

   ```bash
   kubectl apply -f production/operator/crds/
   ```

3. Add the following to your values file:

   ```yaml
   metaMonitoring:
     serviceMonitor:
       enabled: true
     grafanaAgent:
       enabled: true
       installOperator: true

       logs:
         remote:
           url: "https://example.com/loki/api/v1/push"
           auth:
             username: 12345

       metrics:
         remote:
           url: "https://mimir-nginx.mimir.svc.cluster.local./api/v1/push"
           headers:
             X-Scope-OrgID: metamonitoring
   ```

The article [Collecting metrics and logs from Grafana Mimir]({{< relref "../monitor-grafana-mimir/collecting-metrics-and-logs/">}})
goes into greater detail of how to set up the credentials for this.

## Configure clients to write metrics to Mimir

Refer to [Configure Prometheus to write to Grafana Mimir]({{< relref "../deploy-grafana-mimir/getting-started-helm-charts/#configure-prometheus-to-write-to-grafana-mimir">}})
and [Configure Grafana Agent to write to Grafana Mimir]({{< relref "../deploy-grafana-mimir/getting-started-helm-charts/#configure-grafana-agent-to-write-to-grafana-mimir">}})
for details on how to configure each client to remote-write metrics to Mimir.

### High availability setup

It is possible to set up redundant groups of clients to write metrics to Mimir. Refer to
[Configuring mimir-distributed Helm Chart for high-availability deduplication with Consul]({{< relref "../configure/setting-ha-helm-deduplication-consul">}})
for instructions on setting up a Consul instance, configuring Mimir, and configuring clients.
