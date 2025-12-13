#!/bin/bash
set -e

# Determine action from argument or default to 'up'
ACTION="${1:-up}"

if [[ "$ACTION" == "destroy" ]]; then
    echo "=================================================="
    echo "🗑️  Kubernetes Cluster Destroyer"
    echo "=================================================="
elif [[ "$ACTION" == "up" ]]; then
    echo "=================================================="
    echo "🚀 Kubernetes Cluster Installer"
    echo "=================================================="
else
    echo "❌ Invalid action: $ACTION"
    echo "Usage: $0 [up|destroy]"
    exit 1
fi

# Check for Homebrew
if ! command -v brew &> /dev/null; then
    echo "❌ Homebrew not found. Please install it from https://brew.sh"
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
brew install lima kind kubectl pulumi go || true

# Repository and install directory
REPO_URL="https://github.com/tstark7952/pulumi-kind-cluster.git"
INSTALL_DIR="$HOME/.local/share/pulumi-kind-cluster"

# Check if we're destroying and directory exists
if [[ "$ACTION" == "destroy" ]]; then
    if [[ ! -d "$INSTALL_DIR" ]]; then
        echo "❌ Installation directory not found: $INSTALL_DIR"
        echo "Nothing to destroy."
        exit 1
    fi

    cd "$INSTALL_DIR"

    # Check for passphrase file
    if [[ ! -f ".pulumi-passphrase" ]]; then
        echo "🔐 Pulumi passphrase file not found"
        echo "Enter your Pulumi passphrase:"
        read -sp "Passphrase: " PASSPHRASE
        echo ""
        echo -n "$PASSPHRASE" > .pulumi-passphrase
        chmod 600 .pulumi-passphrase
    fi

    export PULUMI_CONFIG_PASSPHRASE_FILE="$INSTALL_DIR/.pulumi-passphrase"

    echo ""
    echo "⚠️  WARNING: This will destroy your entire Kubernetes cluster!"
    echo "This action cannot be undone."
    read -p "Are you sure you want to continue? (yes/no): " CONFIRM

    if [[ "$CONFIRM" != "yes" ]]; then
        echo "Destroy cancelled."
        exit 0
    fi

    echo ""
    echo "🗑️  Destroying Kubernetes cluster..."
    pulumi stack select dev 2>/dev/null || true
    pulumi destroy --yes

    echo ""
    echo "🧹 Cleaning up installation directory..."
    rm -rf "$INSTALL_DIR"

    echo ""
    echo "=================================================="
    echo "✅ Cluster Destroyed Successfully!"
    echo "=================================================="
    echo ""
    echo "Remember to remove these from your shell profile:"
    echo "  export KUBECONFIG=~/.kube/myk8s-config"
    echo "  export DOCKER_CONTEXT=lima-myk8s-docker"
    echo "  export PULUMI_CONFIG_PASSPHRASE_FILE=$INSTALL_DIR/.pulumi-passphrase"
    echo ""

else
    # Deploy path
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
    echo "To destroy this cluster later, run:"
    echo "  curl -fsSL https://raw.githubusercontent.com/tstark7952/pulumi-kind-cluster/main/install.sh | bash -s destroy"
    echo ""
fi
