#!/bin/bash
set -e

# Export kubeconfig to ensure Terraform can access the cluster
echo "Exporting Kubeconfig..."
microk8s config > ~/.kube/config

# Destroy Terraform
echo "Destroying Infrastructure..."
cd terraform
terraform destroy -auto-approve

# Clean up any leftover PVs or Namespaces (Force delete if stuck)
echo "Cleaning up any stuck resources..."
microk8s kubectl delete ns argocd --ignore-not-found
microk8s kubectl delete ns kafka --ignore-not-found

# Optional: Disable addons if you want a cleaner slate (commented out by default)
# microk8s disable hostpath-storage

echo "Cleanup complete!"