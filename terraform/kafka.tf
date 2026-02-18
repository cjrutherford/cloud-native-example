resource "helm_release" "kafka" {
  name             = "kafka"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "kafka"
  namespace        = "kafka"
  create_namespace = true
  version          = "26.8.1"

  values = [
    yamlencode({
      kraft = {
        enabled = true
      }
      zookeeper = {
        enabled = false
      }
      controller = {
        replicaCount = 1
      }
      listeners = {
        client = {
          protocol = "PLAINTEXT"
        }
      }
      extraConfig = <<-EOT
        offsets.topic.replication.factor=1
        transaction.state.log.replication.factor=1
        transaction.state.log.min.isr=1
        group.initial.rebalance.delay.ms=0
        num.partitions=3
      EOT
    })
  ]
}