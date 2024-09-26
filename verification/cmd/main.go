/*
Copyright 2022 -- TBD

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"path/filepath"

	verify "github.com/lamatriz/ra2-auth/k8s-rbac-verfication/internal/rbac_rules_verification"
	utils "github.com/lamatriz/ra2-auth/k8s-rbac-verfication/utils"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var kubeconfig, api_resources, rbac_yaml, namespace, log_file *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	api_resources = flag.String("api_resources", "", `absolute path to cluster api_resource file.
	Use "kubectl api-resources -o wide > api_resources.txt" for resounce only, 
	or "scripts/k8s/print-all-res.sh" for resource and subresources to generate.`)
	rbac_yaml = flag.String("rbac_yaml", "", "absolute path to the rbac yaml file")
	namespace = flag.String("namespace", "smoke-test", "list of namespaces to be verified, separately by \",\"")
	log_file = flag.String("log_file", "./rbac_verify.log", "absolute path to the log file")
	flag.Parse()

	utils.Set_logging(*log_file)
	namespaces, _ := utils.SplitString(*namespace, ",")
	for _, ns := range namespaces {
		sar_allowed, sar_forbidden := verify.CreateSubjectAccessReviewList(*api_resources, *rbac_yaml, ns)
		if err := verify.DoBatchSelfSubjectAccessReviews(*kubeconfig, sar_allowed, true); err != nil {
			fmt.Printf("Test Error %s", err.Error())
		}
		if err := verify.DoBatchSelfSubjectAccessReviews(*kubeconfig, sar_forbidden, false); err != nil {
			fmt.Printf("Test Error %s", err.Error())
		}
	}
}
