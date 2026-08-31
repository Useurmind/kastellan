# Kastellan

## Overview

Kastellan is a Kubernetes operator for managing infrastructure services running on external Linux hosts.

The primary use case is managing external load balancers and DNS servers that require information from the Kubernetes API while avoiding direct API access from those systems.

Kastellan acts as a control plane that observes Kubernetes resources, transforms them into infrastructure-specific desired state, and distributes that state to remote agents running on external hosts.

---

# Problem Statement

Many Kubernetes environments operate critical infrastructure outside the cluster:

- Load balancers (HAProxy, Envoy, NGINX)
- DNS servers (CoreDNS, BIND, PowerDNS)
- Routers and network appliances

These systems typically need Kubernetes information such as:

- Services
- Endpoints
- Ingresses
- Gateway resources

Today they often access the Kubernetes API directly, requiring:

- API credentials
- RBAC permissions
- Connectivity to the API server
- Local caching and watch logic

This increases operational complexity and expands the security boundary.

---

# Goals

## Initial Goals

- Manage external load balancer configuration.
- Manage external DNS configuration.
- Use Kubernetes as the source of truth.
- Require only outbound connectivity from managed hosts.
- Avoid direct Kubernetes API access from external systems.
- Support multiple external hosts for redundancy.

## Future Goals

- Generic Docker workload deployment.

---

# High-Level Architecture

```text
+--------------------+
| Kubernetes Cluster |
+--------------------+
          |
          v
+--------------------+
| Kastellan Operator |
+--------------------+
          ^
          |
       mTLS
          |
          v
+--------------------+
| Kastellan Agent    |
+--------------------+
          |
          +--> HAProxy
          |
          +--> CoreDNS
          |
          +--> Other Services
```

## Key Principle

External systems do not talk to Kubernetes directly.

Instead:

```text
Kubernetes
    ↓
Kastellan Operator
    ↓
Kastellan Agent
    ↓
Infrastructure Service
```

The operator is the only component that interacts with the Kubernetes API. External systems receive only the configuration they need.

---

# Agent Model

Each managed Linux host runs a single Kastellan Agent container.

### Responsibilities

- Establish outbound connection to the operator.
- Authenticate using mTLS.
- Receive desired configuration.
- Validate configuration.
- Apply configuration.
- Report health and status.

The agent acts as the only component with access to local infrastructure services.

---

# Initial Use Case: External Load Balancer

A Service is created:

```yaml
apiVersion: v1
kind: Service
spec:
  type: LoadBalancer
  loadBalancerClass: Kastellan.io/external
```

### Flow

1. Operator watches the Service.
2. Operator discovers Service backends.
3. Operator generates load balancer configuration.
4. Configuration is sent to load-balancer agents.
5. Agent validates configuration.
6. Agent reloads HAProxy.
7. Agent reports status.
8. Operator updates the Service status.

---

# Initial Use Case: External DNS

A Service or dedicated resource requests a DNS name:

```yaml
metadata:
  annotations:
    Kastellan.io/dns-name: app.example.internal
```

### Flow

1. Operator determines the target address.
2. Operator generates DNS records.
3. Records are sent to DNS agents.
4. Agent updates the local DNS service.
5. Agent reports success.

---

# Core Resources

## ExternalHost

Represents a connected external host.

```yaml
kind: ExternalHost
spec:
  role: load-balancer
```

### Example Roles

- load-balancer
- dns
- infrastructure

---

## ExternalHostGroup

Groups multiple hosts that provide the same service.

```yaml
kind: ExternalHostGroup
spec:
  role: load-balancer
```

Example:

```text
production-lb
 ├─ lb01
 ├─ lb02
 └─ lb03
```

This allows Kastellan to distribute configuration to multiple redundant hosts.

---

# MVP Scope

## Infrastructure

- ExternalHost
- ExternalHostGroup
- Agent enrollment
- mTLS authentication
- Outbound-only connectivity

## Load Balancing

- HAProxy adapter
- TCP services
- Service type `LoadBalancer`
- Static VIPs
- NodePort backends

## DNS

- CoreDNS adapter
- A records
- Service-based DNS annotations

---

# Out of Scope

The following are intentionally excluded from the initial release:

- Generic Docker workloads
- BGP integration
- VRRP / Keepalived
- Service scheduling
- GPU workloads
- Remote shell access
- Multi-cluster support
- Dynamic infrastructure provisioning

---

# Long-Term Vision

Kastellan establishes Kubernetes as the control plane for external infrastructure.

Future versions may support:

- Generic container workloads
- Additional load-balancer implementations
- Additional DNS providers
- Routing and network appliances
- Edge deployments
- GPU and AI infrastructure

The initial focus, however, remains on DNS and load-balancer integration.

---

# Vision Statement

> Kastellan allows Kubernetes to manage external infrastructure services through outbound-connected agents without exposing the Kubernetes API to those systems.