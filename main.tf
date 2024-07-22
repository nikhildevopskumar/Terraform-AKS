provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "rg1" {
  name     = var.rgname
  location = var.location
}

module "ServicePrincipal" {
  source                 = "./modules/ServicePrincipal"
  service_principal_name = var.service_principal_name

  depends_on = [
    azurerm_resource_group.rg1
  ]
}

resource "azurerm_role_assignment" "rolespn" {

  scope                = "/subscriptions/4dc27c3b-60eb-4010-84f2-51d79521a6af"
  role_definition_name = "Contributor"
  principal_id         = module.ServicePrincipal.service_principal_object_id

  depends_on = [
    module.ServicePrincipal
  ]
}

resource "azurerm_role_assignment" "rolespn_key_vault_admin" {

  scope                = module.keyvault.keyvault_id
  role_definition_name = "Key Vault Administrator"
  principal_id         = module.ServicePrincipal.service_principal_object_id

  depends_on = [
    module.ServicePrincipal
  ]
}


module "keyvault" {
  source                      = "./modules/keyvault"
  keyvault_name               = var.keyvault_name
  location                    = var.location
  resource_group_name         = var.rgname
  service_principal_name      = var.service_principal_name
  service_principal_object_id = module.ServicePrincipal.service_principal_object_id
  service_principal_tenant_id = module.ServicePrincipal.service_principal_tenant_id

  depends_on = [
    module.ServicePrincipal
  ]
}

resource "azurerm_key_vault_secret" "example" {
  name         = module.ServicePrincipal.client_id
  value        = module.ServicePrincipal.client_secret
  key_vault_id = module.keyvault.keyvault_id

  depends_on = [
    module.keyvault
  ]
}

#create Azure Kubernetes Service
module "aks" {
  source                 = "./modules/aks/"
  service_principal_name = var.service_principal_name
  client_id              = module.ServicePrincipal.client_id
  client_secret          = module.ServicePrincipal.client_secret
  location               = var.location
  resource_group_name    = var.rgname

  depends_on = [
    module.ServicePrincipal
  ]

}

resource "local_file" "kubeconfig" {
  depends_on = [module.aks]
  filename   = "./kubeconfig"
  content    = module.aks.config

}

provider "kubernetes" {
  config_path = "${path.module}/kubeconfig"
}

# Helm provider for ArgoCD, Prometheus, and Grafana
provider "helm" {
  kubernetes {
    config_path = "${path.module}/kubeconfig"
  }
}

# ArgoCD Namespace
resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
  depends_on = [module.aks]
}

# Prometheus Namespace
resource "kubernetes_namespace" "prometheus" {
  metadata {
    name = "prometheus"
  }
  depends_on = [module.aks]
}

# Grafana Namespace
resource "kubernetes_namespace" "grafana" {
  metadata {
    name = "grafana"
  }
  depends_on = [module.aks]
}

# ArgoCD Helm Chart
resource "helm_release" "argocd" {
  name       = "argocd"
  namespace  = kubernetes_namespace.argocd.metadata[0].name
  chart      = "argo-cd"
  repository = "https://argoproj.github.io/argo-helm"
  version    = "7.3.9" # Use the latest stable version

  set {
    name  = "server.service.type"
    value = "LoadBalancer"
  }

  depends_on = [kubernetes_namespace.argocd]
}

# Prometheus Helm Chart
resource "helm_release" "prometheus" {
  name       = "prometheus"
  namespace  = kubernetes_namespace.prometheus.metadata[0].name
  chart      = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  version    = "23.3.0" # Use the latest stable version

  set {
    name  = "server.service.type"
    value = "LoadBalancer"
  }

  depends_on = [kubernetes_namespace.prometheus]
}

# Grafana Helm Chart
resource "helm_release" "grafana" {
  name       = "grafana"
  namespace  = kubernetes_namespace.grafana.metadata[0].name
  chart      = "grafana"
  repository = "https://grafana.github.io/helm-charts"
  version    = "6.51.2" # Use the latest stable version

  set {
    name  = "service.type"
    value = "LoadBalancer"
  }

  set {
    name  = "adminPassword"
    value = "SuperSecret123!"
  }

  depends_on = [kubernetes_namespace.grafana]
}
