# ra2-auth

Design documents, rule specifications and utilities for RA-2 Authentication and RBAC

## Files

```bash
.
├───doc
├───rbac
├───scripts
│   └───oidc-client
└───verification
    ├───bin
    ├───cmd
    ├───doc
    ├───internal
    │   ├───process_rules
    │   └───rbac_rules_verification
    ├───pkg
    └───utils
```

* `doc` contains the design documents
* `rbac` contains the yaml files for rbac rules
* `scripts` contains the scripts for generating odic token and kubeconfig
* `verification` is the go source directory for `k8s-rbac-verficication` module

## Use Cases

* For k8s user, execute the `refresh-k8s-token.sh` from `scripts/oidc-client` for oidc token generation or refresh, and a .kube/abc_xyz_config kubeconfig file;
* For k8s cluster admin, set the cluster IP, cert, oidc_url etc. values properly first before distrubtion the above script to user;
* For k8s cluster admin, review the yaml files in `rbac` directory to set proper rules for user subjects, then apply to or update on k8s cluster;
* For k8s cluster admin, review the results from `k8s-rbac-verficication` on all subjects and resoruces, to ensure the RBAC rules from above are set properly, or if there is any drift from the cluster.
