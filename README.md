# Kubernetes Local Cluster with Pulumi

A production-ready Pulumi program that automatically provisions a local multi-node Kubernetes cluster on macOS using Lima VM and Kind, complete with Calico CNI networking.

## Overview

This infrastructure-as-code solution creates a fully-functional local Kubernetes development environment with the following features:

- **Multi-node Kind cluster** (1 control-plane + 3 workers)
- **Lima VM** with Docker support (customizable resources)
- **Calico CNI** with VXLAN networking
- **Automatic configuration** of kubeconfig and shell profiles
- **Auto-start on boot** via macOS launchd
- **Persistent storage** mounts on all nodes

## Architecture

```
macOS Host
├── Lima VM (Ubuntu 24.04.3 LTS, VZ driver)
│   └── Docker Engine
│       └── Kind Cluster (Kind node images run Debian)
│           ├── Control Plane Node (tainted)
│           ├── Worker Node 1
│           ├── Worker Node 2
│           └── Worker Node 3
└── Calico CNI (VXLAN mode)
```

**Note:** The Lima VM runs **Ubuntu 24.04.3 LTS (Noble Numbat)**. The Kubernetes nodes run inside Kind containers which use Debian-based node images by design. This is standard for Kind and doesn't affect functionality.

## Prerequisites

- macOS with Apple Silicon or Intel
- [Homebrew](https://brew.sh/)
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)
- Go 1.24+
- Lima
- Kind
- kubectl

### Installation

```bash
# Install prerequisites via Homebrew
brew install pulumi go lima kind kubectl
```

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/tstark7952/pulumi-kind-cluster.git
   cd pulumi-kind-cluster
   ```

2. **Initialize Pulumi stack**
   ```bash
   pulumi stack init dev
   ```

3. **Configure resources (optional)**
   ```bash
   pulumi config set cpus 8              # Default: 8
   pulumi config set memory 16           # Default: 16 (GB)
   pulumi config set disk 500            # Default: 500 (GB)
   pulumi config set vmName myk8s-docker # Default: myk8s-docker
   pulumi config set clusterName myk8s   # Default: myk8s
   ```

4. **Deploy the cluster**
   ```bash
   pulumi up
   ```

5. **Use the cluster**
   ```bash
   export KUBECONFIG=~/.kube/myk8s-config
   kubectl get nodes
   ```

   Or use the helper script:
   ```bash
   source ~/bin/use-k8s.sh
   ```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `vmName` | Name of the Lima VM | `myk8s-docker` |
| `cpus` | Number of CPUs for the VM | `8` |
| `memory` | Memory in GB for the VM | `16` |
| `disk` | Disk size in GB for the VM | `500` |
| `clusterName` | Name of the Kind cluster | `myk8s` |
| `calicoVersion` | Calico CNI version to install | `v3.29.1` |

## What Gets Created

The Pulumi program creates the following resources:

1. **Directories** - Mount points for persistent storage (`/tmp/myk8s-*-disk`)
2. **Kind Config** - Cluster configuration with extra mounts and disabled default CNI
3. **Lima VM** - Virtual machine with Docker, using macOS Virtualization framework
4. **launchd Service** - Auto-starts the VM on macOS boot
5. **Docker Context** - Configured to use Lima VM's Docker socket
6. **Kind Cluster** - 4-node Kubernetes cluster
7. **Kubeconfig** - Exported and configured for kubectl access
8. **Shell Profiles** - Updated `.zshrc` and `.bashrc` with KUBECONFIG
9. **Calico CNI** - Networking layer with VXLAN encapsulation
10. **Control Plane Taint** - NoSchedule taint on control-plane node
11. **Helper Script** - `~/bin/use-k8s.sh` for easy cluster activation

## Features

### Idempotent Deployments

The infrastructure is fully idempotent - you can run `pulumi up` multiple times safely:
- Detects existing Lima VMs and reuses them
- Checks for existing Kind clusters before creation
- Handles partial failures gracefully with retry logic
- VM startup waits with timeout and retry mechanisms

### Comprehensive Health Checks

After deployment, the system runs 8 comprehensive health checks:
1. Lima VM status verification
2. Docker connectivity test
3. Kind cluster existence check
4. Kubernetes API connectivity
5. Node readiness status
6. System pods health
7. Calico CNI verification
8. CoreDNS health check

Each check provides clear PASS/WARN/FAIL status with detailed diagnostics.

### Automatic Context Switching

The deployment automatically configures and switches to the correct contexts:
- **Docker context** - Automatically set to `lima-myk8s-docker` via `DOCKER_CONTEXT` environment variable
- **kubectl context** - Automatically set to `kind-myk8s`

Both contexts are persisted in your shell profiles (`.zshrc`, `.bashrc`), so new terminal sessions will automatically use the correct contexts.

To activate in your current shell:
```bash
source ~/bin/use-k8s.sh
# or
source ~/.zshrc  # or ~/.bashrc
```

### Complete Cleanup

All resources have proper deletion handlers. To destroy the entire infrastructure:

```bash
pulumi destroy
```

This will completely remove:
- Kind cluster
- Lima VM and all associated files
- launchd service
- Docker context (`lima-myk8s-docker`)
- kubectl context, cluster, and user entries (`kind-myk8s`)
- Kubeconfig files and symlinks
- Shell profile entries (KUBECONFIG exports)
- Helper scripts (`~/bin/use-k8s.sh`)

After `pulumi destroy`, your system is restored to its original state with no leftover configurations.

### Persistent Storage

Each node has a dedicated mount point for persistent data:
- Control plane: `/tmp/myk8s-control-disk` → `/var/lib/disk1`
- Worker 1: `/tmp/myk8s-worker1-disk` → `/var/lib/disk1`
- Worker 2: `/tmp/myk8s-worker2-disk` → `/var/lib/disk1`
- Worker 3: `/tmp/myk8s-worker3-disk` → `/var/lib/disk1`

### Calico Networking

The cluster uses Calico CNI with VXLAN overlay networking:
- VXLAN encapsulation enabled
- IPIP mode disabled
- Automatic pod-to-pod communication across nodes

## Verification

After deployment, verify your cluster:

```bash
# Check cluster info
kubectl cluster-info

# View nodes
kubectl get nodes -o wide

# Check system pods
kubectl -n kube-system get pods

# Verify Calico
kubectl -n kube-system get pods -l k8s-app=calico-node
```

## Troubleshooting

### Cluster not accessible

If kubectl cannot connect:

```bash
# Verify the cluster exists in Lima
DOCKER_HOST=unix://$HOME/.lima/myk8s-docker/sock/docker.sock kind get clusters

# Re-export kubeconfig
kind export kubeconfig --name myk8s --kubeconfig ~/.kube/myk8s-config

# Set KUBECONFIG
export KUBECONFIG=~/.kube/myk8s-config
```

### Lima VM not starting

```bash
# Check VM status
limactl list

# View VM logs
limactl shell myk8s-docker

# Restart VM
limactl stop myk8s-docker
limactl start myk8s-docker
```

### Docker context issues

```bash
# Reset Docker context
docker context use default
docker context rm lima-myk8s-docker
docker context create lima-myk8s-docker --docker "host=unix://$HOME/.lima/myk8s-docker/sock/docker.sock"
docker context use lima-myk8s-docker
```

## Development

### Project Structure

```
.
├── main.go           # Main Pulumi program
├── go.mod            # Go module dependencies
├── go.sum            # Dependency checksums
├── Pulumi.yaml       # Pulumi project configuration
├── README.md         # This file
├── LICENSE           # MIT License
└── .gitignore        # Git ignore rules
```

### Modifying the Cluster

To change cluster configuration:

1. Edit `main.go` to modify the Kind config or add resources
2. Run `pulumi preview` to see planned changes
3. Run `pulumi up` to apply changes

### Adding Additional Nodes

Edit the `kindConfig` variable in `main.go`:

```go
- role: worker
  extraMounts:
  - hostPath: /tmp/myk8s-worker4-disk
    containerPath: /var/lib/disk1
```

## Stack Outputs

The Pulumi stack exports:

- `clusterName`: Name of the Kubernetes cluster
- `kubeconfigPath`: Path to the kubeconfig file

Access outputs:

```bash
pulumi stack output clusterName
pulumi stack output kubeconfigPath
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Pulumi](https://www.pulumi.com/) - Infrastructure as Code platform
- [Lima](https://lima-vm.io/) - Linux virtual machines on macOS
- [Kind](https://kind.sigs.k8s.io/) - Kubernetes in Docker
- [Calico](https://www.projectcalico.org/) - Cloud native networking and security

## Resources

- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Kind Documentation](https://kind.sigs.k8s.io/docs/)
- [Lima Documentation](https://lima-vm.io/docs/)
- [Calico Documentation](https://docs.projectcalico.org/)
