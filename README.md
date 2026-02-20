# pulumi-kind-cluster

Pulumi program that provisions a multi-node Kind cluster inside a Lima VM on macOS. Handles the full stack — VM creation, Docker context, Kind cluster, Calico CNI, kubeconfig, and launchd auto-start — in a single `pulumi up`.

**What gets created:** Lima VM (Ubuntu 24.04, VZ driver) → Docker → Kind (1 control-plane + 3 workers) → Calico CNI (VXLAN) → persistent storage mounts per node

## Requirements

```bash
brew install pulumi go lima kind kubectl
```

## Deploy

```bash
git clone https://github.com/justin-oleary/pulumi-kind-cluster.git
cd pulumi-kind-cluster

echo "your-passphrase" > .pulumi-passphrase
export PULUMI_CONFIG_PASSPHRASE_FILE="$(pwd)/.pulumi-passphrase"

pulumi stack init dev
pulumi up
```

Or one-liner:

```bash
curl -fsSL https://raw.githubusercontent.com/justin-oleary/pulumi-kind-cluster/main/install.sh | bash
```

## Use

```bash
export KUBECONFIG=~/.kube/myk8s-config
kubectl get nodes
kubectl -n kube-system get pods
```

## Destroy

```bash
pulumi destroy
```

Removes the Kind cluster, Lima VM, launchd service, Docker context, kubectl context, kubeconfig entries, and shell profile changes. Clean slate.

## Configuration

| Parameter | Default | Description |
|---|---|---|
| `vmName` | `myk8s-docker` | Lima VM name |
| `cpus` | `8` | VM CPU count |
| `memory` | `16` | VM memory in GB |
| `disk` | `500` | VM disk in GB |
| `clusterName` | `myk8s` | Kind cluster name |
| `calicoVersion` | `v3.29.1` | Calico CNI version |

```bash
pulumi config set cpus 16
pulumi config set memory 32
```

## Troubleshooting

**Cluster not reachable:**

```bash
# verify cluster exists inside Lima
DOCKER_HOST=unix://$HOME/.lima/myk8s-docker/sock/docker.sock kind get clusters

# re-export kubeconfig
kind export kubeconfig --name myk8s --kubeconfig ~/.kube/myk8s-config
```

**Lima VM not starting:**

```bash
limactl list
limactl stop myk8s-docker && limactl start myk8s-docker
```

**Docker context broken:**

```bash
docker context rm lima-myk8s-docker
docker context create lima-myk8s-docker \
  --docker "host=unix://$HOME/.lima/myk8s-docker/sock/docker.sock"
docker context use lima-myk8s-docker
```

## License

MIT
