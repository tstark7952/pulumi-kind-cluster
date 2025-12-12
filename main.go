package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration
		conf := config.New(ctx, "")
		vmName := conf.Get("vmName")
		if vmName == "" {
			vmName = "myk8s-docker"
		}
		cpus := conf.GetInt("cpus")
		if cpus == 0 {
			cpus = 8
		}
		memory := conf.GetInt("memory")
		if memory == 0 {
			memory = 16
		}
		disk := conf.GetInt("disk")
		if disk == 0 {
			disk = 500
		}
		clusterName := conf.Get("clusterName")
		if clusterName == "" {
			clusterName = "myk8s"
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// Create mount directories but don't create dependency chain
		createDirs, err := local.NewCommand(ctx, "create-dirs", &local.CommandArgs{
			Create: pulumi.String("mkdir -p /tmp/myk8s-{control,worker1,worker2,worker3}-disk"),
		})
		if err != nil {
			return err
		}

		// Create Kind cluster config without dependency chain
		kindConfigPath := "./kind-config.yaml"
		kindConfig := `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true  # Disable kindnet CNI
nodes:
- role: control-plane
  extraMounts:
  - hostPath: /tmp/myk8s-control-disk
    containerPath: /var/lib/disk1
- role: worker
  extraMounts:
  - hostPath: /tmp/myk8s-worker1-disk
    containerPath: /var/lib/disk1
- role: worker
  extraMounts:
  - hostPath: /tmp/myk8s-worker2-disk
    containerPath: /var/lib/disk1
- role: worker
  extraMounts:
  - hostPath: /tmp/myk8s-worker3-disk
    containerPath: /var/lib/disk1
`
		createKindConfig, err := local.NewCommand(ctx, "create-kind-config", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf("cat <<EOF > %s\n%s\nEOF", kindConfigPath, kindConfig)),
			Delete: pulumi.String(fmt.Sprintf("rm -f %s", kindConfigPath)),
		})
		if err != nil {
			return err
		}

		// Only create dependencies when truly necessary - VM needs dirs and config
		limaVm, err := local.NewCommand(ctx, "lima-vm", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf("limactl start --tty=false --name %s template:docker --cpus %d --memory %d --disk %d --vm-type vz",
				vmName, cpus, memory, disk)),
			Delete: pulumi.String(fmt.Sprintf(`
				# First, try to delete any Kind cluster that might be running in this VM
				DOCKER_HOST=unix://$HOME/.lima/%s/sock/docker.sock kind delete cluster --name %s 2>/dev/null || true
				
				# Stop the VM first (required before deletion)
				limactl stop %s 2>/dev/null || true
				
				# Delete the VM with force flag to ensure it's removed
				limactl delete --force %s 2>/dev/null || true
				
				# Clean up any leftover sockets and temp files
				rm -rf $HOME/.lima/%s/sock/* 2>/dev/null || true
				
				echo "Lima VM %s cleanup completed"
			`, vmName, clusterName, vmName, vmName, vmName, vmName)),
		}, pulumi.DependsOn([]pulumi.Resource{createDirs, createKindConfig}))
		if err != nil {
			return err
		}

		// These operations depend only on VM creation and can run in parallel
		// Create launchd plist only depends on VM
		launchdPlistPath := filepath.Join(homeDir, "Library", "LaunchAgents", fmt.Sprintf("dev.lima.%s.plist", vmName))
		launchdPlist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.lima.%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>/opt/homebrew/bin/limactl</string>
        <string>start</string>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>`, vmName, vmName)

		createPlist, err := local.NewCommand(ctx, "create-launchd-plist", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				cat <<EOF > %s
%s
EOF
				launchctl load %s
			`, launchdPlistPath, launchdPlist, launchdPlistPath)),
			Delete: pulumi.String(fmt.Sprintf(`
				# Unload and remove the launchd plist
				launchctl unload %s 2>/dev/null || true
				rm -f %s 2>/dev/null || true
			`, launchdPlistPath, launchdPlistPath)),
		}, pulumi.DependsOn([]pulumi.Resource{limaVm}))
		if err != nil {
			return err
		}

		// Setup Docker context - only depends on VM
		dockerContext, err := local.NewCommand(ctx, "setup-docker", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				docker context rm lima-%s 2>/dev/null || true
				docker context create lima-%s --docker "host=unix:///Users/$USER/.lima/%s/sock/docker.sock" || true
				docker context use lima-%s || true
				echo "Current Docker context: $(docker context show)"
			`, vmName, vmName, vmName, vmName)),
			Delete: pulumi.String(fmt.Sprintf(`
				# Reset Docker context to default during cleanup
				docker context use default 2>/dev/null || true
				docker context rm lima-%s 2>/dev/null || true
			`, vmName)),
		}, pulumi.DependsOn([]pulumi.Resource{limaVm}))
		if err != nil {
			return err
		}

		// Create Kind cluster - depends on both plist and docker context
		createCluster, err := local.NewCommand(ctx, "create-kind-cluster", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				DOCKER_HOST=unix://$HOME/.lima/%s/sock/docker.sock kind create cluster --name %s --config %s
			`, vmName, clusterName, kindConfigPath)),
			Delete: pulumi.String(fmt.Sprintf(`
				# Delete the Kind cluster
				DOCKER_HOST=unix://$HOME/.lima/%s/sock/docker.sock kind delete cluster --name %s 2>/dev/null || true
			`, vmName, clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{createPlist, dockerContext}))
		if err != nil {
			return err
		}

		// Export kubeconfig first and set it up properly
		kubeconfigPath := filepath.Join(homeDir, ".kube", fmt.Sprintf("%s-config", clusterName))
		defaultKubeconfigPath := filepath.Join(homeDir, ".kube", "config")

		exportKubeconfig, err := local.NewCommand(ctx, "export-kubeconfig", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				# Create .kube directory if it doesn't exist
				mkdir -p %s/.kube

				# Export kubeconfig to a specific file
				echo "Exporting kubeconfig to %s"
				DOCKER_HOST=unix://$HOME/.lima/%s/sock/docker.sock kind export kubeconfig --name %s --kubeconfig %s

				# Make sure the kubeconfig file is accessible
				chmod 600 %s

				# Create a symlink to the default location if it doesn't exist or is empty
				if [ ! -f %s ] || [ ! -s %s ]; then
					ln -sf %s %s
					echo "Created symlink from %s to %s"
				fi

				# Export the KUBECONFIG environment variable for this session
				export KUBECONFIG=%s

				# Fix the kubeconfig if it has localhost references (often causes connection issues)
				# Replace localhost with 127.0.0.1 which is more reliable
				sed -i.bak 's|server: https://localhost:|server: https://127.0.0.1:|g' %s

				# Automatically set kubectl context to the new cluster
				kubectl config use-context kind-%s

				# Verify the kubeconfig is valid
				echo "Testing kubectl configuration..."
				kubectl version --client || true
				echo "Current kubectl context: $(kubectl config current-context)"
			`, homeDir, kubeconfigPath, vmName, clusterName, kubeconfigPath,
				kubeconfigPath, defaultKubeconfigPath, defaultKubeconfigPath,
				kubeconfigPath, defaultKubeconfigPath, kubeconfigPath, defaultKubeconfigPath,
				kubeconfigPath, kubeconfigPath, clusterName)),
			Delete: pulumi.String(fmt.Sprintf(`
				# Remove kubectl context
				kubectl config delete-context kind-%s 2>/dev/null || true
				kubectl config delete-cluster kind-%s 2>/dev/null || true
				kubectl config delete-user kind-%s 2>/dev/null || true

				# Remove the kubeconfig file during cleanup
				rm -f %s 2>/dev/null || true
				rm -f %s.bak 2>/dev/null || true

				# Remove symlink if it points to our config
				if [ -L %s ] && [ "$(readlink %s)" = "%s" ]; then
					rm -f %s 2>/dev/null || true
				fi
			`, clusterName, clusterName, clusterName, kubeconfigPath, kubeconfigPath,
				defaultKubeconfigPath, defaultKubeconfigPath, kubeconfigPath, defaultKubeconfigPath)),
		}, pulumi.DependsOn([]pulumi.Resource{createCluster}))
		if err != nil {
			return err
		}

		// Add kubeconfig to shell profiles to make it persistent
		updateProfiles, err := local.NewCommand(ctx, "update-shell-profiles", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				# Add KUBECONFIG to shell profiles
				echo "Updating shell profiles..."
				if ! grep -q "export KUBECONFIG=%s" ~/.zshrc 2>/dev/null; then
					echo "export KUBECONFIG=%s" >> ~/.zshrc
					echo "Updated .zshrc"
				fi
				if ! grep -q "export KUBECONFIG=%s" ~/.bashrc 2>/dev/null; then
					echo "export KUBECONFIG=%s" >> ~/.bashrc
					echo "Updated .bashrc"
				fi

				# Create a convenient activation script
				mkdir -p ~/bin
				cat <<EOF > ~/bin/use-k8s.sh
#!/bin/bash
export KUBECONFIG=%s
echo "Kubernetes context set to %s"
kubectl cluster-info
EOF
				chmod +x ~/bin/use-k8s.sh
				echo "Created activation script at ~/bin/use-k8s.sh"
			`, kubeconfigPath, kubeconfigPath, kubeconfigPath, kubeconfigPath,
				kubeconfigPath, clusterName)),
			Delete: pulumi.String(fmt.Sprintf(`
				# Remove KUBECONFIG from shell profiles
				sed -i.bak '/export KUBECONFIG=%s/d' ~/.zshrc 2>/dev/null || true
				sed -i.bak '/export KUBECONFIG=%s/d' ~/.bashrc 2>/dev/null || true
				rm -f ~/.zshrc.bak ~/.bashrc.bak 2>/dev/null || true

				# Remove activation script
				rm -f ~/bin/use-k8s.sh 2>/dev/null || true
			`, kubeconfigPath, kubeconfigPath)),
		}, pulumi.DependsOn([]pulumi.Resource{exportKubeconfig}))
		if err != nil {
			return err
		}

		// Apply operations to the cluster - with proper KUBECONFIG
		// 1. Taint control plane node
		_, err = local.NewCommand(ctx, "taint-control-plane", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				export KUBECONFIG=%s
				echo "Applying taints to control plane node..."
				kubectl taint nodes %s-control-plane node-role.kubernetes.io/control-plane:NoSchedule --overwrite || true
			`, kubeconfigPath, clusterName)),
			Environment: pulumi.StringMap{
				"KUBECONFIG": pulumi.String(kubeconfigPath),
			},
		}, pulumi.DependsOn([]pulumi.Resource{exportKubeconfig}))
		if err != nil {
			return err
		}

		// 2. Install Calico
		installCalico, err := local.NewCommand(ctx, "install-calico", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				export KUBECONFIG=%s
				echo "Installing Calico CNI..."
				kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.2/manifests/calico.yaml
				kubectl set env -n kube-system ds/calico-node CALICO_IPV4POOL_VXLAN=Always
				kubectl set env -n kube-system ds/calico-node CALICO_IPV4POOL_IPIP=Off
			`, kubeconfigPath)),
			Delete: pulumi.String(fmt.Sprintf(`
				export KUBECONFIG=%s
				echo "Removing Calico CNI..."
				kubectl delete -f https://raw.githubusercontent.com/projectcalico/calico/v3.28.2/manifests/calico.yaml --ignore-not-found=true 2>/dev/null || true
			`, kubeconfigPath)),
			Environment: pulumi.StringMap{
				"KUBECONFIG": pulumi.String(kubeconfigPath),
			},
		}, pulumi.DependsOn([]pulumi.Resource{exportKubeconfig}))
		if err != nil {
			return err
		}

		// 3. Wait for Calico to be ready
		waitForCalico, err := local.NewCommand(ctx, "wait-for-calico", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				export KUBECONFIG=%s
				echo "Waiting for Calico pods to be ready..."
				
				timeout=120
				interval=3
				elapsed=0
				while [ $elapsed -lt $timeout ]; do
					# Use kubectl wait for efficiency
					if kubectl wait --for=condition=ready pods -l k8s-app=calico-node -n kube-system --timeout=3s 2>/dev/null; then
						echo "All Calico pods are ready!"
						break
					fi
					
					# Fallback to manual checking if kubectl wait fails
					ready_pods=$(kubectl -n kube-system get pods -l k8s-app=calico-node -o jsonpath='{.items[*].status.containerStatuses[*].ready}' | tr ' ' '\n' | grep -c "true" || echo "0")
					desired_pods=$(kubectl -n kube-system get pods -l k8s-app=calico-node --no-headers | wc -l | tr -d ' ')
					
					if [ "$ready_pods" -eq "$desired_pods" ] && [ "$desired_pods" -ge 1 ]; then
						echo "All Calico pods are ready ($ready_pods/$desired_pods)."
						break
					fi
					
					echo "Waiting for Calico pods... ($ready_pods/$desired_pods ready)"
					sleep $interval
					elapsed=$((elapsed + interval))
				done
				
				if [ $elapsed -ge $timeout ]; then
					echo "Warning: Timed out waiting for Calico pods to be ready"
					kubectl -n kube-system get pods -l k8s-app=calico-node
				fi
			`, kubeconfigPath)),
			Environment: pulumi.StringMap{
				"KUBECONFIG": pulumi.String(kubeconfigPath),
			},
		}, pulumi.DependsOn([]pulumi.Resource{installCalico}))
		if err != nil {
			return err
		}

		// Create K8s provider with explicit kubeconfig path
		k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
			Kubeconfig: pulumi.String(kubeconfigPath),
		}, pulumi.DependsOn([]pulumi.Resource{waitForCalico}))
		if err != nil {
			return err
		}

		// Final verification
		_, err = local.NewCommand(ctx, "verify-cluster", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`
				# Ensure KUBECONFIG is set
				export KUBECONFIG=%s
				
				echo "====================================================================="
				echo "Verifying Kubernetes cluster..."
				
				# Check if kubectl can connect to the cluster
				if kubectl cluster-info; then
					echo "✅ Successfully connected to Kubernetes cluster"
				else
					echo "❌ Failed to connect to Kubernetes cluster"
					echo "Troubleshooting steps:"
					echo "1. Check if cluster exists in Lima VM:"
					DOCKER_HOST=unix://$HOME/.lima/%s/sock/docker.sock kind get clusters
					
					echo "2. Your kubeconfig is at: %s"
					echo "3. Try running: source ~/bin/use-k8s.sh"
					echo "4. Or explicitly: kubectl --kubeconfig=%s get nodes"
				fi
				
				# Show cluster resources
				echo ""
				echo "Kubernetes Nodes:"
				kubectl get nodes -o wide
				
				echo ""
				echo "Kubernetes System Pods:"
				kubectl -n kube-system get pods
				
				echo ""
				echo "====================================================================="
				echo "🎉 Setup complete! Your Kubernetes cluster is ready to use."
				echo ""
				echo "To use kubectl with this cluster, do ONE of the following:"
				echo "  1. In a new terminal: source ~/.bashrc  (or ~/.zshrc)"
				echo "  2. In this terminal: export KUBECONFIG=%s"
				echo "  3. Run the helper script: source ~/bin/use-k8s.sh"
				echo ""
				echo "Cluster Name: %s"
				echo "Lima VM Name: %s"
				echo "====================================================================="
			`, kubeconfigPath, vmName, kubeconfigPath, kubeconfigPath, kubeconfigPath, clusterName, vmName)),
			Environment: pulumi.StringMap{
				"KUBECONFIG": pulumi.String(kubeconfigPath),
			},
		}, pulumi.DependsOn([]pulumi.Resource{waitForCalico, updateProfiles, k8sProvider}))
		if err != nil {
			return err
		}

		// Export stack outputs
		ctx.Export("clusterName", pulumi.String(clusterName))
		ctx.Export("kubeconfigPath", pulumi.String(kubeconfigPath))

		return nil
	})
}
