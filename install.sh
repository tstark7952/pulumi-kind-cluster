#!/bin/bash
set -e

echo "=================================================="
echo "🚀 Kubernetes Cluster Installer"
echo "=================================================="

# Check for Homebrew
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew not found. Please install it from https://brew.sh"
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
brew install lima kind kubectl pulumi go || true

# Clone repository
REPO_URL="https://github.com/tstark7952/pulumi-kind-cluster.git"
INSTALL_DIR="$HOME/.local/share/pulumi-kind-cluster"

echo "📥 Cloning repository..."
rm -rf "$INSTALL_DIR"
git clone "$REPO_URL" "$INSTALL_DIR"
cd "$INSTALL_DIR"

# Setup Pulumi passphrase
echo ""
echo "🔐 Pulumi Configuration"
echo "=================================================="
echo "Enter a passphrase for Pulumi state encryption"
echo "(This will be saved to .pulumi-passphrase)"
read -sp "Passphrase: " PASSPHRASE
echo ""

# Save passphrase
echo -n "$PASSPHRASE" > .pulumi-passphrase
chmod 600 .pulumi-passphrase

export PULUMI_CONFIG_PASSPHRASE_FILE="$INSTALL_DIR/.pulumi-passphrase"

# Initialize stack
echo ""
echo "🏗️  Initializing Pulumi stack..."
pulumi stack init dev 2>/dev/null || pulumi stack select dev

# Deploy
echo ""
echo "🚀 Deploying Kubernetes cluster..."
pulumi up --yes

echo ""
echo "=================================================="
echo "✅ Installation Complete!"
echo "=================================================="
echo ""
echo "Add these to your shell profile (~/.bashrc or ~/.zshrc):"
echo "  export KUBECONFIG=~/.kube/myk8s-config"
echo "  export DOCKER_CONTEXT=lima-myk8s-docker"
echo "  export PULUMI_CONFIG_PASSPHRASE_FILE=$INSTALL_DIR/.pulumi-passphrase"
echo ""
echo "Or run: source ~/bin/use-k8s.sh"
echo ""
