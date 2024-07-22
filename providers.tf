terraform {

  required_providers {
    azuread = "~> 2.9.0"
    random  = "~> 3.1"
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.110.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}
