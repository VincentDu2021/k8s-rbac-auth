# RBAC Roles - Proposal

-----------------------------

## ClusterRole and Roles

### cluster-admin

- Responsibility: Deploy and configure the cluster and its core components, and control the cluster auth and RBAC.
- Description: Full read and write access to all cluster resources, usually assigned to the group of individuals involved in the cluster deployment.
- Proposed head count: min 1, max 3

### cluster-operator

- Responsibility: Cluster's day-to-day operations, such as monitoring, namespace request, workload troubleshooting, etc.
- Description: Same as above, except the following.
    write access to clusterrole and clusterrolebindings
    write access to any clusterroles with the "system:" prefix
- Proposed head count: min 1, max 6

### cluster-viewer

- Responsibility: monitoring the cluster and reporting issues to the corresponding stake-holder
- Description: same as above, without any write access.
- Proposed head count: min 1, max TBD

### namespace-admin

- Responsibility: administrate a designated namespace which is typically used for workload run and research
- Description: same as cluster-admin, limited to a designated namespace
- Proposed head count: min 1(per namespace), max: TBD

### namespace-viewer

- optional, upon the decision of namespace-admin, same as above, without write access

## Example Role Privileges

Here the "can" part of each role is meant to show the unique privileges only in its own hierarchy, the "cannot" part shows the privileges available to 1 hierarchy above but not available to the current one. The listed items are examples, and not meant to be complete.

### cluster-admin

can: everything,

- create and edit clusterrole and bindings,
- view and approve CSR
- modify api-server flags and restart
- change api-server or external LB IP and ports
- tear down the entire cluster and rebuild

cannot:

- N/A

### cluster-operator

can:

- create and delete namespace
- create and edit roles and rolebindings in a specific namespace
- create new SC PV etc.
- create new components, such as MPI-operator, Prometheus, etc., and modify resources in the corresponding namespaces
- deploy, view and modify other k8s components and generated resources such as istio, calico, volcano, etc.
- view resources in kube-system namespace
- view, create, and edit all service accounts and secrets in all namespaces
- set resource quota for user-based namespaces

cannot:

- modify clusterrole and clusterrolebindings
- access cluster's master node, control-plane
- approve CSR
- modify resources in the kube-system namespace

### cluster-viewer

can:

- view all nodes
- view all namespaces
- view all pods and logs in all namespaces

cannot

- modify any resources
- view service accounts and secretes

### namespace-admin

can:

- run job, cronjob, workload, mpijob, etc. in the designated namespace
- view and edit service accounts and secrets created by **_cluster-admin** or **_cluster-operator_** in the designated namespace
- view and edit service accounts and secrets created by **_namespace-admin_**

cannot:

- delete namespace
- view another user-based namespace, except service-based namespace\*
- edit resource quota in the designated namespace

### namespace-viewer

can:

- view pods, logs, controlMap, PVC, etc. in the designated namespace

cannot:

- modify anything

## ClusterRole and Bindings Creation

### cluster-admin

Use admin.conf, no need to create any bindings

### cluster-operator

#### clusterrole

Can extend default "admin" clusterrole, with addition of write access to "namespace" and "resourceQuota" resources, and "Node" resources.

#### clusterrolebinding Subject

"cluster-operator" group

### cluster-viewer

#### clusterrole

default "view" clusterrole

#### clusterrolebinding Subject

"cluster-viewer" group

### namespace-admin

#### clusterrole

default "edit" clusterrole

#### rolebinding Subject

namespace-specific admin group

### namespace-viewer

#### clusterrole

default "view" clusterrole

#### rolebinding Subject

namespace-specific viewer group

## Special Consideration

### User-based vs. service-based namespaces

The purpose of a user-based namespace is to provide a group of users an isolated space to run workloads or perform research, .i.e _smoke-test_, the service-based namespace, on the other hand, is a namespace to provide core services.
For example, the _csi-wekafsplugin_ contains the PODs, DaemonSets etc. for _csi-wekaplugin_, it is created by **_cluster-admi_**n or **_cluster-operator_**, and provides services to the user-based plugin via csi-wekafs StorageClass. However, when a user observes failures due to storage-related issues, this user might want to check logs from the corresponding _csi-wekafsplugin-node_ in _csi-wekafsplugin_ namespace. This user may request either a **_cluster-viewer_** role processed by **_cluster-admin_** since it is a clusterrole, or a **_namespace-viewer_** for  _csi-wekaplugin_ namespace only processed by **_cluster-operator_** since it is a role with is namespace-specific.

The list of the service-based namespace:

- csi-wekafsplugin
- habanalabs
- volcano-system
- TBD

### Service Accounts and Secrets

By definition, Service Accounts(SA) provide an identity for processes that run in a Pod, SA can be configured with Secrets or Tokens. The service accounts and secrets at the cluster level contain Barcelona or Lamatriz organization's sensitive data, i.e. dockerhub secret. They are created by **_cluster-admin_** or **_cluster-operator_** for a namespace, hence, the **_namespace-admin_** of this namespace will also have read and write access to such SA and secret. This limitation is caused by k8s RBAC having no "deny" mechanism, resulting in rules applied to resource type and apiGroup being inclusive to all entities within the scope. Alternatively, the access to SA and secrets would be removed from **_namespace-admin_** entirely even in the designated namespace.

If this is not desired, a workaround has to be sought outside of k8s RBAC framework.

### Relationship between cluster-operator and namespace-admin

These 2 roles require more fine-grained demarcation as they both have read and write access to resources and directly interface with each other. A **_namespace-admin_** is usually granted read and write permission with all the resources or endpoints in the designated namespace, even for those created by **_cluster-admin_** or **_cluster-operator_**. This would either force the sharing of sensitive data and modifications on resources in the designated namespace. Some examples:

- access to service accounts and secrets discussed above
- mpijob Pod to process mpijob template

These resources, would either force the sharing of sensitive data between **_cluster-admin_** and  **_namespace-admin_** or allow a **_namespace-admin_** to modify resources configured and deployed by **_cluster-operator_** in the designated namespace.

The reason for using the "edit" clusterrole instead of "admin" for **_namespace-admin_** is that the "admin" clusterrole includes role and rolebinding within the designated namespace. **_namespace-admin_** can create a rolebinding to a clusterrole, i.e. **_cluster-admin_** to gain full access within the namespace(To Be Tested).

### Node visibility

K8S node/nodes are not a namespace-scoped resource, and it is not by default aggregated to clusterroles like "admin". Therefore special clusterrole to allow "kubectl get nodes" to work has been created, and namespace-admin can bind to it. Also "node/nodes" resource with all verbs(get, patch, delete, etc.) has been added into cluster-operator clusterrole to allow them to list, cordon/uncordon nodes.

## Business Logic and Workflow

TBD
