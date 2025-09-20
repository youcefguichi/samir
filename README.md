# Samir (Minimal Container Runtime WIP)

This repository is a work‑in‑progress minimal container runtime experiment (namespaces, veth, bridge, cgroups, etc.).

### Roadmap
- [ ] Refactor the runtime package.
- [ ] Add proper CLI to interacte with samir
    - `samir run <my-api-binary> --config config.json` 
- [ ] Add teardown/cleanup for mounts & network (bridge/veth removal ..)
- [ ] Add GitHub Actions CI (build + test)
- [ ] Improve dynamic IP assignment (DHCP)
- [ ] Implement sandbox mode
    - `samir run --sandboxed <my-api-binary> --config config.json`
- [ ] Run samir as daemon
- [ ] Make samir OCI compliant


## Prerequisites
- Go 1.23+
- Linux (with user having sudo for network / namespace / cgroup ops)
- iptables (for NAT if you add it later)
- (Optional) Rootfs directory (e.g. extracted Alpine or Debian)

## Make Targets

| Target    | Description                                   |
|-----------|-----------------------------------------------|
| make build | Build the samir binary (outputs ./samir)     |
| make run  | Build then run the container demo (uses sudo) |
| make test | Run all Go tests                              |
| make clean| Remove built binary                           |


## Why sudo on run?
Creating network namespaces, moving veth pairs, mounting proc, applying cgroups, etc. require CAP_SYS_ADMIN / CAP_NET_ADMIN which you typically have only with sudo. (switch later to daemon architecture ...)

## Current Flow
1. Prepare or extract a rootfs (e.g. rootfs/samir-os).
2. Adjust code (rootfs path, networking).
3. Execute: `make run`
4. Inside the spawned shell, inspect:
   ```bash
   ip addr
   ip route
   hostname
   ps aux
   ping 8.8.8.8
   ```

## Disclaimer
Educational / experimental.
