# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-12-31

### Added

- Initial release of FinOps Enforcer
- Core controller with policy reconciliation loop
- Custom Resource Definition (CRD) for EnforcementPolicy
- OpenCost integration for real-time cost data
- Policy engine with declarative YAML configuration
- Enforcement engine with scale-to-zero action
- Slack notification system with interactive buttons
- Prometheus metrics exporter
- Safety guardrails:
  - Namespace allowlisting
  - Cooldown windows
  - Max actions per run limit
  - Dry-run mode
  - Annotation-based exclusions
- Helm chart for easy deployment
- Comprehensive documentation:
  - README with quick start
  - DESIGN.md with architecture details
  - POLICIES.md with configuration reference
  - RUNBOOK.md with operational procedures
  - CONTRIBUTING.md with development guidelines
- Sample policies for common use cases
- Kubernetes RBAC with minimal required permissions
- Docker image with distroless base
- Makefile for development workflow

### Features

- Automatically detects idle Kubernetes deployments
- Pauses idle workloads by scaling to zero
- Preserves original replica count for easy reactivation
- Sends Slack notifications with one-click reactivation
- Tracks estimated monthly savings
- Supports wildcard namespace patterns
- Label-based resource filtering
- Scheduled enforcement windows (timezone-aware)
- Leader election for high availability
- Comprehensive audit logging
- Policy status tracking

### Metrics

- `finops_paused_resources_total` - Currently paused resources
- `finops_estimated_savings_usd` - Projected monthly savings
- `finops_policy_matches_total` - Policy evaluation matches
- `finops_actions_taken_total` - Enforcement actions executed
- `finops_reactivations_total` - User-initiated reactivations
- `finops_false_positives_total` - Resources reactivated within 1 hour
- `finops_policy_evaluation_duration_seconds` - Policy evaluation time
- `finops_reconciliation_duration_seconds` - Reconciliation loop time
- `finops_opencost_api_errors_total` - OpenCost API failures

### Documentation

- Production-grade README with clear value proposition
- Detailed design document explaining architecture and trade-offs
- Policy reference guide with examples and best practices
- Operational runbook with troubleshooting procedures
- Contributing guidelines for open-source collaboration

### Known Limitations

- Only supports `scaleToZero` action type (delete not implemented)
- Traffic threshold detection requires manual annotation (no automatic metrics)
- Utilization threshold not yet implemented
- Single-cluster only (no multi-cluster support)
- Slack-only notifications (no email, PagerDuty, etc.)

### Security

- Runs as non-root user (65532)
- Read-only root filesystem
- Minimal RBAC permissions (no delete rights)
- No access to production namespaces by default
- Secrets management for Slack webhook

[Unreleased]: https://github.com/yourusername/finops-enforcer/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yourusername/finops-enforcer/releases/tag/v0.1.0
