# ra2-auth/rbac

RA-2 K8S RBAC Clusterrole, Role, and Bindings YAML files

## Files and Design Logics

```bash
cluster-operator-clusterrole.yaml
cluster-operator-clusterrolebinding.yaml
mldev-namespace-admin-rolebinding.yaml
namespace-admin-clusterrole.yaml
node-viewer-clusterrole.yaml
node-viewer-clusterrolebinding.yaml
project-lima-namespace-admin-rolebinding.yaml
project-lima-namespace-viewer-rolebinding.yaml
smoke-test-namespace-admin-rolebinding.yaml
smoke-test-namespace-viewer-rolebinding.yaml
```
* `node-viewer-clusterrole.yaml` and `node-viewer-clusterrolebinding.yaml` are created to ensure all oidc groups can view current nodes in the cluster
* `namespace-admin-clusterrole.yaml` is the clusterrole for different namespaced-oidc-groups, i.e. "smoke-test" and "mldev" are existing namespaces designated to different oidc groups, they are entitled to the same resources but in different namespaces, therefore there are "smoke-test-namespace-admin-rolebinding.yaml" and "mldev-namespace-admin-rolebinding.yaml" bindings to the clusterrole defined in "namespace-admin-clusterrole.yaml".
