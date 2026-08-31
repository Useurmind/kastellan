# PodmanPlay Resource Design Draft

## Purpose

`PodmanPlay` is the first workload resource supported by Kastellan.

It represents a workload that is deployed to one or more external hosts using Podman's `podman kube play` functionality.

The resource intentionally uses Kubernetes-native Pod, Deployment, ConfigMap, Secret and PVC definitions as its workload specification instead of introducing a custom workload format.

The Kastellan operator is responsible for distributing workload specifications to external hosts, while the Kastellan agent is responsible for reconciling the desired state using Podman.

---

# Resource Relationships

```text
ExternalHost
      │
      ▼
ExternalHostGroup
      │
      ▼
PodmanPlay
```

Example:

```text
production-lb
 ├─ lb01
 ├─ lb02
 └─ lb03

PodmanPlay
 └─ HAProxy
```

The `PodmanPlay` definition is assigned to a host group and deployed to all members of that group.

---

# Goals

The PodmanPlay resource should:

- Be easy to understand for Kubernetes users.
- Reuse Kubernetes YAML.
- Avoid a Docker Compose dependency.
- Align with Podman's native capabilities.
- Support infrastructure workloads such as:
  - HAProxy
  - CoreDNS
  - Envoy
  - FRR
  - NGINX
- Support arbitrary containers in the future.

---

# API Definition

## Example

```yaml
apiVersion: kastellan.io/v1alpha1
kind: PodmanPlay
metadata:
  name: haproxy

spec:
  hostGroupRef:
    name: production-lb

  deploymentPolicy:
    mode: All

  workload: |
    apiVersion: v1
    kind: Pod
    metadata:
      name: haproxy

    spec:
      hostNetwork: true

      containers:
      - name: haproxy
        image: docker.io/library/haproxy:lts

        ports:
        - containerPort: 80
        - containerPort: 443

        volumeMounts:
        - name: config
          mountPath: /usr/local/etc/haproxy

      volumes:
      - name: config
        hostPath:
          path: /srv/haproxy/config
```

---

# Spec

## hostGroupRef

Defines the target host group.

```yaml
spec:
  hostGroupRef:
    name: production-lb
```

The operator resolves the host group members and schedules deployment accordingly.

---

## deploymentPolicy

Defines how the workload is deployed within the host group.

### All

Deploy to every host.

```yaml
deploymentPolicy:
  mode: All
```

Example:

```text
production-lb
 ├─ lb01
 ├─ lb02
 └─ lb03

Result:
  haproxy deployed on all hosts
```

---

### Single

Deploy to exactly one host.

```yaml
deploymentPolicy:
  mode: Single
```

Future feature.

---

### ActiveStandby

Deploy to multiple hosts but only activate on one.

```yaml
deploymentPolicy:
  mode: ActiveStandby
```

Future feature.

---

## workload

Raw Kubernetes YAML that is passed to `podman kube play`.

Supported objects depend on the capabilities of Podman.

Examples:

- Pod
- Deployment
- ConfigMap
- Secret
- PersistentVolumeClaim

---

# Reconciliation

The operator does not interpret the workload definition.

Instead:

```text
PodmanPlay
       │
       ▼
Kastellan Operator
       │
       ▼
Kastellan Agent
       │
       ▼
podman kube play
```

The agent is responsible for:

1. Receiving workload revisions.
2. Storing workload YAML.
3. Running:

```bash
podman kube play
```

4. Monitoring the resulting pod.
5. Reporting status back to Kubernetes.

---

# Ownership Model

Every deployed workload receives ownership labels.

Example:

```text
kastellan.io/managed=true
kastellan.io/resource=haproxy
kastellan.io/hostgroup=production-lb
kastellan.io/revision=12
```

This allows the agent to:

- detect existing workloads
- detect drift
- support cleanup
- support upgrades

---

# Status Model

## Example

```yaml
status:
  phase: Ready

  observedGeneration: 3

  hosts:

  - host: lb01
    phase: Ready
    revision: 3

  - host: lb02
    phase: Ready
    revision: 3

  - host: lb03
    phase: Updating
    revision: 2
```

---

# Lifecycle

## Create

```text
create PodmanPlay
        │
        ▼
operator resolves host group
        │
        ▼
agent receives workload
        │
        ▼
podman kube play
        │
        ▼
status updated
```

---

## Update

```text
update PodmanPlay
        │
        ▼
new revision created
        │
        ▼
agents reconcile
        │
        ▼
status updated
```

Initially updates should use:

```text
Replace strategy
```

Meaning:

```bash
podman kube play --replace
```

---

## Delete

```text
delete PodmanPlay
        │
        ▼
finalizer added
        │
        ▼
agent executes cleanup
        │
        ▼
resources removed
        │
        ▼
finalizer removed
```

---

# Future Extensions

The resource should remain intentionally minimal.

Future additions may include:

## Config Sources

```yaml
configMaps:
- my-config
```

## Secret Distribution

```yaml
secrets:
- tls-cert
```

## Image Pull Credentials

```yaml
imagePullSecrets:
- registry-creds
```

## Update Strategies

```yaml
updateStrategy:
  type: Replace
```

Future:

```yaml
updateStrategy:
  type: Rolling
```

## Health Checks

```yaml
healthCheck:
  timeout: 60s
```

## Placement Constraints

```yaml
placement:
  requiredLabels:
    site: dc1
```

---

# MVP Scope

The initial implementation should support only:

- ExternalHost
- ExternalHostGroup
- PodmanPlay
- Deployment to all hosts in a group
- Pod manifests
- `podman kube play`
- `podman kube play --replace`
- Status reporting
- Cleanup on deletion

Anything beyond that should be postponed until the core deployment workflow is stable.

---

# Example Resource Hierarchy

```text
ExternalHost
 ├─ lb01
 ├─ lb02
 ├─ dns01
 └─ dns02

ExternalHostGroup
 ├─ production-lb
 └─ production-dns

PodmanPlay
 ├─ haproxy
 ├─ coredns
 └─ frr
```

This provides a very small and understandable API surface while leaving enough room to evolve Kastellan into a broader external infrastructure management platform.