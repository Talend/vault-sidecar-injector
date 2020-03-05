cache {
    use_auto_auth_token = true
}

listener "tcp" {
    address = "127.0.0.1:<VSI_PROXY_PORT>"
    tls_disable = true
}