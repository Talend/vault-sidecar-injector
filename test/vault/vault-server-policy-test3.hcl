# To read secrets
path "secret/test3/test-app3-svc1" {
    capabilities = ["read"]
}

path "secret/test3/test-app3-svc2" {
    capabilities = ["read"]
}

# To list secrets
path "secret/test3/test-app3-svc1/" {
    capabilities = ["list"]
}

path "secret/test3/test-app3-svc2/" {
    capabilities = ["list"]
}