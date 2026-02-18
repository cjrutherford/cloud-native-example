#!/bin/bash
set -e

# Ensure MicroK8s is running
echo "Checking MicroK8s status..."
microk8s status --wait-ready

echo "Microk8s is ready. ensuring the local storage addon is enabled..."
microk8s enable storage

# Export kubeconfig for Terraform providers
echo "Exporting Kubeconfig..."
microk8s config > ~/.kube/config

echo "Waiting for storage to be ready..."
while ! microk8s kubectl get storageclass microk8s-hostpath &>/dev/null; do
  echo "Waiting for storage class to be available..."
  sleep 5
done

# Initialize Terraform
echo "Initializing Terraform..."
cd terraform
terraform init

# Apply Terraform
echo "Applying Terraform configuration..."
terraform apply -auto-approve

echo "Deploying Argo CD..."
cd ..
microk8s kubectl apply -n argocd -f ./k8s/argocd-app.yaml

echo "Deployment complete! You can access Argo CD via NodePort or port-forwarding."