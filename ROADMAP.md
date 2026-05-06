# WasmForge Roadmap

This roadmap is a living document and may change as the project evolves.

## Near term

 - [ ] Chaos/fault-injection test suite
 - [ ] OpenTelemetry-compatible metrics and tracing support
 - [ ] Policy-as-code support for route and plugin configuration
 - [ ] gRPC support for plugin communication and control plane interactions

## Midterm

- [ ] Deterministic plugin execution and state management for better reliability and observability
- [ ] Plugin resource governance
- [ ] Capability-based permissions for plugins

## Longer term

- [ ] Distributed data-plane nodes with Raft-based consensus for state synchronization
- [ ] RBAC and multi-tenant controls
- [ ] Release channels and safer rolling updates for plugin artifacts
- [ ] First-class metrics dashboards and SLO-focused tooling

## Non-goals (for now)

- [ ] Full service mesh replacement
- [ ] Proprietary plugin SDK lock-in
