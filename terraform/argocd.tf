resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  namespace        = "argocd"
  create_namespace = true
  version          = "5.53.12" 

  values = [
    yamlencode({
        server = {
            service = {
                type = "NodePort"
            }
        }
    })]
  
  # MicroK8s specific: ensuring it doesn't fight with built-in addons if present
}