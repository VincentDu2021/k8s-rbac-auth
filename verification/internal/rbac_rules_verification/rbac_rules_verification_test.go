/*
IMPORTANT: Set $KUBECONFIG in your environment before running the unit test. The value is "base64 -w 0 your_kube_config_yaml"
For vscode, set it in "go.testEnvVars" block of "settings.json" file
*/
package rbac_rules_verification

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"testing"

	proc_rules "github.com/lamatriz/ra2-auth/k8s-rbac-verfication/internal/process_rules"
	"github.com/lamatriz/ra2-auth/k8s-rbac-verfication/utils"
)

var api_resource_txt = `#Test All Resources File#
NAME                              SHORTNAMES            APIVERSION                             NAMESPACED   KIND                             VERBS                                                       
pods                              po                    v1                                     true         Pod                              [create delete deletecollection get list patch update watch]
pods/exec                                               v1                                     true         PodExecOptions                   [create get]
configmaps                        cm                    v1                                     true         ConfigMap                        [create delete deletecollection get list patch update watch]
statefulsets                      sts                   apps/v1                                true         StatefulSet                      [create delete deletecollection get list patch update watch]
tokenreviews                                            authentication.k8s.io/v1               false        TokenReview                      [create]
networksets                                             crd.projectcalico.org/v1               true         NetworkSet                       [delete deletecollection get list patch create update watch]
ippools                                                 crd.projectcalico.org/v1               false        IPPool                           [delete deletecollection get list patch create update watch]`

var rbac_yaml_text = `#Sample Clusterrole yaml file
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
  name: namespace-admin
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - crd.projectcalico.org
  resources:
  - ippools
  verbs:
  - "*"
- apiGroups:
  - crd.projectcalico.org
  resources:
  - networksets
  verbs:
  - get
  - watch
  - list
`
var rb, fb map[string]proc_rules.ApiGroupValueType

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	if err := utils.Set_logging("./unitest_logging.log"); err != nil {
		log.Panic(err)
	}
	dev_config_yaml_str, err := base64.StdEncoding.DecodeString(os.ExpandEnv("$KUBECONFIG"))
	if err != nil {
		utils.FatalLogger.Panicf("Failed to base64 decode kubeconfig string: %s", err.Error())
	}
	if err := os.WriteFile("./test_dev_config.yaml", []byte(dev_config_yaml_str), 0644); err != nil {
		utils.FatalLogger.Printf("Failed to create kubeconfig file, error:\n%s", err.Error())
		log.Panic(err)
	}
	if err := os.WriteFile("./test_all_api_resources.txt", []byte(api_resource_txt), 0644); err != nil {
		utils.FatalLogger.Printf("Failed to create allresource file, error:\n%s", err.Error())
		log.Panic(err)
	}
	if err := os.WriteFile("./test_clusterrole.yaml", []byte(rbac_yaml_text), 0644); err != nil {
		utils.FatalLogger.Printf("Failed to create rbac yaml file, error:\n%s", err.Error())
		log.Panic(err)
	}

}

func shutdown() {
	if err := os.Remove("./test_dev_config.yaml"); err != nil {
		log.Panic(err)
	}
	fmt.Printf("All Done.")
}
func TestGetClientset(t *testing.T) {
	if ret := getClientset("./test_dev_config.yaml"); ret == nil {
		t.Error("Failed to create clientset.")
	}
}

func TestCreateSubjectAccessReviewList(t *testing.T) {
	sar_allowed, sar_forbidden := CreateSubjectAccessReviewList("./test_all_api_resources.txt", "./test_clusterrole.yaml", "smoke-test")
	if len(sar_allowed) == 0 || len(sar_forbidden) == 0 {
		t.Errorf("Getting wrong length,  sar_allowed: %d, sar_forbidden %d", len(sar_allowed), len(sar_forbidden))
	}
	t.Logf("sar_allowed: %d, sar_forbidden %d", len(sar_allowed), len(sar_forbidden))
}

// This is a real functional test not a unit test, no PASS/FAIL criteria yet
func TestDoBatchSelfSubjectAccessReviews(t *testing.T) {
	sar_allowed, sar_forbidden := CreateSubjectAccessReviewList("./test_all_api_resources.txt", "./test_clusterrole.yaml", "smoke-test")
	if err := DoBatchSelfSubjectAccessReviews("./test_dev_config.yaml", sar_allowed, true); err != nil {
		t.Errorf("Test Error %s", err.Error())
	}
	if err := DoBatchSelfSubjectAccessReviews("./test_dev_config.yaml", sar_forbidden, false); err != nil {
		t.Errorf("Test Error %s", err.Error())
	}
}
