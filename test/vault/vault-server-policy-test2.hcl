# To read secrets
path "secret/test2/test-app2-svc" {
    capabilities = ["read"]
}

# To list secrets
path "secret/test2/test-app2-svc/" {
    capabilities = ["list"]
}