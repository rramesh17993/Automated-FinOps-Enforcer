# ADR-001: Custom Kubernetes Controller vs CronJob

**Status:** Accepted  
**Date:** December 2025  
**Decision Makers:** Rajesh Ramesh  
**Context:** Cost governance automation for Kubernetes environments

---

## Context

We needed to enforce idle resource cleanup in Kubernetes non-production environments to reduce cloud costs. The system must:

- Continuously monitor workload costs and activity
- Automatically enforce policies (scale to zero idle deployments)
- Provide safety guarantees (no accidental production impact)
- Integrate natively with Kubernetes primitives
- Be maintainable and extensible

Two primary approaches were considered:
1. **Custom Kubernetes Controller** (Operator pattern)
2. **CronJob with scripts** (Scheduled task pattern)

---

## Decision

**Implement a custom Kubernetes Controller using controller-runtime** instead of a scheduled CronJob.

---

## Alternatives Considered

### Option 1: Custom Kubernetes Controller (✅ CHOSEN)

**Architecture:**
- Long-running process with reconciliation loop
- Watches Kubernetes API for resource changes
- Implements Custom Resource Definition (EnforcementPolicy)
- Uses controller-runtime framework

**Pros:**
- ✅ **Real-time reaction** to resource changes (event-driven)
- ✅ **Automatic recovery** from failures via reconciliation loop
- ✅ **Native Kubernetes integration** (RBAC, admission controllers, audit logs)
- ✅ **Scales horizontally** with leader election (multiple replicas for HA)
- ✅ **Declarative API** via CRDs (kubectl-friendly, GitOps-compatible)
- ✅ **Extensible** via webhooks, custom status, and conditions
- ✅ **Better observability** with built-in metrics and health checks
- ✅ **Idiomatic** in Kubernetes ecosystem (standard pattern)

**Cons:**
- ❌ Higher initial complexity (learning controller-runtime)
- ❌ More code to write and maintain
- ❌ Requires understanding of reconciliation patterns

### Option 2: CronJob with Scripts (❌ REJECTED)

**Architecture:**
- Scheduled pod runs periodically (e.g., every 5 minutes)
- Executes bash/Python script to query and act
- Stores state externally (ConfigMap or database)

**Pros:**
- ✅ Simple to understand (bash script + kubectl)
- ✅ Lower initial development time
- ✅ No framework dependencies

**Cons:**
- ❌ **Polling-based** (5-10 minute delays between runs)
- ❌ **No automatic recovery** (if pod crashes mid-run, state is lost)
- ❌ **Manual state management** (need external storage for cooldowns, history)
- ❌ **Less idiomatic** in Kubernetes (not following operator pattern)
- ❌ **Harder to test** (requires actual scheduler, time-dependent)
- ❌ **No natural HA** (running multiple CronJobs risks conflicts)
- ❌ **Poor observability** (logs only, no structured status)
- ❌ **Not extensible** (can't easily add webhooks or custom validation)

### Option 3: Lambda/Cloud Function (❌ REJECTED)

**Architecture:**
- Cloud provider function triggered on schedule
- Queries Kubernetes API remotely

**Cons:**
- ❌ Requires external cloud credentials and networking
- ❌ Cross-boundary latency and security concerns
- ❌ Vendor lock-in (cloud-specific)
- ❌ Harder to run locally for development
- ❌ Additional cost (function invocations)

---

## Rationale

### Why Controller Pattern Wins

1. **Kubernetes-Native:**
   - Controllers are the standard way to extend Kubernetes
   - Works seamlessly with existing K8s primitives (RBAC, NetworkPolicies, etc.)
   - Recognized pattern by any Kubernetes engineer

2. **Reconciliation Loop:**
   - Self-healing: if the system drifts, next reconciliation corrects it
   - Idempotent operations: safe to retry, safe to run multiple times
   - Handles edge cases automatically (pods restarted, policies updated, etc.)

3. **Production-Grade Features:**
   - Leader election for HA (multiple replicas, no split-brain)
   - Built-in metrics endpoint (Prometheus-compatible)
   - Health checks (liveness/readiness probes)
   - Structured logging with context

4. **Operational Excellence:**
   - `kubectl get enforcementpolicies` feels natural
   - Status subresource shows current state
   - GitOps-friendly (ArgoCD, Flux can manage policies)
   - Admission webhooks can validate policies before creation

5. **Long-Term Maintainability:**
   - Controller-runtime is battle-tested (used by 100+ operators)
   - Clear separation of concerns (API, controller, clients)
   - Easy to add new features (watch more resources, add webhooks)

### Why CronJob Falls Short

For a simple one-time script, CronJob would suffice. But for a **production cost governance system**, we need:

- **Real-time enforcement** (not "eventually within 5 minutes")
- **High availability** (can't afford downtime in cost monitoring)
- **State management** (cooldown windows, enforcement history)
- **Extensibility** (future: traffic metrics, ML predictions)

CronJob provides none of these naturally.

---

## Implementation Details

### Framework Choice: controller-runtime

- **Why not Kubebuilder?** Kubebuilder uses controller-runtime under the hood. We use controller-runtime directly for flexibility.
- **Why not Operator SDK?** Operator SDK is Ansible/Helm-focused. We need Go code for complex logic.
- **Why not from scratch (client-go)?** Controller-runtime provides critical abstractions (caching, leader election, metrics) that would take weeks to reimplement correctly.

### Language Choice: Go

- Kubernetes API types are Go-native
- Controller-runtime is Go-native
- Compiled binary (no runtime dependencies)
- Excellent concurrency primitives (goroutines for parallel processing)
- Memory-efficient (50-100MB vs Python 200-400MB)

---

## Consequences

### Positive

1. **Better reliability**: Automatic reconciliation handles edge cases
2. **Better observability**: Structured status, metrics, events
3. **Better extensibility**: Easy to add features (webhooks, new actions)
4. **Better operational experience**: Native kubectl integration
5. **Industry standard**: Recognized pattern, easier to onboard new contributors

### Negative

1. **Higher learning curve**: Team must understand controller pattern
2. **More code**: ~2000 LOC vs ~500 LOC for CronJob script
3. **Testing complexity**: Need to test reconciliation logic, not just scripts

### Mitigations

- Comprehensive documentation (DESIGN.md, RUNBOOK.md)
- Table-driven unit tests for reconciliation logic
- Clear code structure following controller-runtime best practices

---

## Success Metrics

This decision will be considered successful if:

- ✅ Controller achieves >99% uptime
- ✅ Reconciliation completes within 30 seconds for typical clusters
- ✅ New engineers can understand the codebase within 1 day
- ✅ System correctly handles edge cases (network failures, API errors, concurrent updates)

---

## References

- [Kubernetes Controller Pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
- [controller-runtime documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [The Kubebuilder Book](https://book.kubebuilder.io/)
- [Writing Controllers Best Practices](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md)

---

## Review History

- **2025-12**: Initial decision, implementation complete
- Future reviews: Quarterly reassessment based on operational experience
