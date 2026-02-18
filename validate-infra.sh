#!/bin/bash

echo "=== Checking Argo CD Application Status ==="
microk8s kubectl get application cloud-native-example -n argocd

echo "=== Checking Deployment Status ==="
microk8s kubectl get pods -n default -l app=cloud-native-example

echo "=== Checking Service Status ==="
microk8s kubectl get svc cloud-native-example -n default

echo "=== Argo CD Password (admin) ==="
microk8s kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
echo ""