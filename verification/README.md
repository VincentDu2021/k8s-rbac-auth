# k8s-rbac-verfication

A tool to verify current user's RBAC in a k8s cluster.

## Usage

### Build and run

```bash
cd $GOPATH/ra2-auth/verification
make all

./bin/app.exe -h

./bin/app.exe \
    -kubeconfig ./bin/dev_config_local.yaml \
    -api_resources ./bin/prod-api-resources.txt \
    -rbac_yaml ./bin/cluster-operator-clusterrole.yaml \
    - namespace "default,smoke-test" \
    -log_file ./rbac_verification.log
```

### run unit tests

```bash
make test
# unit test with coverage
make test-coverage
```

Notice the rbac_rules_verfication_test.go does not work out-of-box unless some environment variables are set for security reasons. Read the comments in the source code for more info.

## Design

### Input

There are 3 input files required for the verification application via cli:

* `kubeconfig` : sets the kubeconfig file of the target k8s cluster
* `api_resources` : sets the api-resource.txt file, currently the expected content is the result of "kubectl api-resources -o wide" to get all the resources in apiGroups in this cluster
* `rbac_yaml` : sets the path for a `rbac.authorization.k8s.io/v1` yaml file, which can be either for `clusterrole` or `role` kind.

### Core Logic

The application uses the `kubeconfig` content to authenticate to the cluster. Then it serializes the `api_resources` and `rbac_yaml` to each resource + verb combinations and group them to 3 Sets:

* the serialized items from `api_resources` are considered in **ALL** set;
* the serilaized items from `rbac_yaml` are grouped in **ALLOWED** set which is considered as a subset of **ALL**;[^1]
* the set **FORBIDDEN** is hence computed with the result of substraction of **ALLOWED** set from **ALL** set

The application then executes `auth can-i` utility on each entry from **ALLOWED** and **FORBIDDEN** sets and compare each verdicts against the expected.
The expected result is **Yes** for **ALLOWED** set and **No** for **FORBIDDEN** set, descepency between the verdict and expect is considered as a failed verification.

## Contact

vincent1.du@intel.com

[^1]: Subresources are not handled yet so the `subset` assumption does not exactly hold, see code doc/TODO.md and code comments for more info.
