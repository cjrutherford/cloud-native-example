resource "kubernetes_manifest" "application" {
    manifest = yamldecode(file("${path.module}/../k8s/argocd-app.yaml"))
    depends_on = [helm_release.argocd]
}