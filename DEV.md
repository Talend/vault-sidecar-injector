# Notes on Go modules used by Vault Sidecar Injector

To interact with Kubernetes API, following Go modules and versions are in use:

- k8s.io/api `kubernetes-1.12.5`
- k8s.io/apimachinery `kubernetes-1.12.5`
- k8s.io/client-go `kubernetes-1.12.5`

Those modules versions will be resolved when calling `go mod vendor` or `go get` and `go.mod` file will be updated with timestamp and commid id (which may be confusing and not practical at all to determine which kubernetes version we are actually using).

E.g.

```sh
$ go get k8s.io/apimachinery@kubernetes-1.12.5
go: k8s.io/apimachinery kubernetes-1.12.5 => v0.0.0-20181128191346-49ce2735e507
go: downloading k8s.io/apimachinery v0.0.0-20181128191346-49ce2735e507
```

> **Vault Sidecar Injector supports following [Kubernetes versions](README.md#kubernetes-compatibility)**

Refer to <https://github.com/kubernetes/client-go#versioning> and <https://github.com/kubernetes/client-go/blob/master/INSTALL.md> for details on `client-go` / `api` / `apimachinery` compatibility and versions.

```text
We recommend using the v0.x.y tags for Kubernetes releases >= v1.17.0 and kubernetes-1.x.y tags for Kubernetes releases < v1.17.0
```

Because of issue <https://github.com/kubernetes/client-go/issues/741>, module `github.com/googleapis/gnostic` has been forced to version `v0.4.0`.
