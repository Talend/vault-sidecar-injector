# To read secrets
path "secret/test/test-app-svc" {
    capabilities = ["read"]
}

# To list secrets
path "secret/test/test-app-svc/" {
    capabilities = ["list"]
}

# Manage the transit secrets engine
path "transit/*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}