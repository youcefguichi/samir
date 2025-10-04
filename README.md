# Samir (Minimal Container Runtime)

A minimal container runtime experiment (namespaces, veth, bridge, cgroups, etc.).

## Understand and tinker with the project

run `make build`, then start a debugging session, I already provided `.vscode/lunch.json`. Just choose you break points and that's it.

## Current Flow
1. Prepare or extract a rootfs (see scripts/extract-rootfs-from-docker-image.sh)
3. Execute: `make run`
4. Inside the spawned shell, inspect:
   ```bash
   ip addr
   ip route
   hostname
   ps aux
   ping 8.8.8.8
   ```
## config.json
Now you should provide the container spec via `config.json` as follow:

```json
{
  "id": "abc123",
  "name": "my-container",
  "rootfs": "bundle/rootfs",
  "entrypoint": ["/bin/sh"],
  "cgroup": {
    "name": "my-container",
    "max_mem": "2000Mb",
    "min_mem": "10Mb",
    "max_cpu": "100m",
    "min_cpu": "10m"
  },
  "run_as": "root",
  "ip": "10.10.0.4/16"
}
```

## make run example output
```bash
➜  root git:(main) ✗ make run
2025/10/04 21:00:55 networking.go:55: couldn't create link, file exists
2025/10/04 21:00:55 networking.go:73: couldn't assign ip range, file exists
2025/10/04 21:00:55 container.go:146: couldn't create cgroup mkdir /sys/fs/cgroup/samir: file exists
2025/10/04 21:00:55 container.go:180: setting memory limit to 2000Mb for cgroup samir
2025/10/04 21:00:55 container.go:193: setting memory request to 10Mb for cgroup samir
2025/10/04 21:00:55 container.go:207: setting cpu limit to 100m for cgroup samir
2025/10/04 21:00:55 container.go:221: setting cpu request to 10m for cgroup samir
2025/10/04 21:00:55 container.go:239: received start signal from the parent pid [SIGNAL=1] 
my-container:/# 
```

## Disclaimer
Educational / experimental.
