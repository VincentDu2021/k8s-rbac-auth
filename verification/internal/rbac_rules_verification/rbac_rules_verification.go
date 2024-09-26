package rbac_rules_verification

import (
	"context"
	"fmt"

	proc_rules "github.com/lamatriz/ra2-auth/k8s-rbac-verfication/internal/process_rules"
	utils "github.com/lamatriz/ra2-auth/k8s-rbac-verfication/utils"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	authorizationv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientset(path string) authorizationv1client.AuthorizationV1Interface {
	var kubeconfig *string = &path

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		utils.FatalLogger.Printf("Failed to build kubeconfig, error:\n%s", err.Error())
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		utils.FatalLogger.Printf("Failed to create clientset, error:\n%s", err.Error())
		panic(err.Error())
	}

	auth_client := clientset.AuthorizationV1()
	return auth_client
}

func processResourcesFiles(all_res_path string, rb_rule_path string) (map[string]proc_rules.ApiGroupValueType, map[string]proc_rules.ApiGroupValueType) {

	proc_rules.ParseAllApiresources(all_res_path)
	proc_rules.ParseK8sRbacYaml(rb_rule_path)
	proc_rules.FilterRules()
	rb_rules := proc_rules.RbacRulesMap
	fb_rules := proc_rules.ForbiddenRulesMap

	return rb_rules, fb_rules
}

func CreateSubjectAccessReviewList(all_res_path string, rb_rule_path string, ns string) ([]*authorizationv1.SelfSubjectAccessReview, []*authorizationv1.SelfSubjectAccessReview) {
	rb, fb := processResourcesFiles(all_res_path, rb_rule_path)
	var sar_allowed, sar_forbidden []*authorizationv1.SelfSubjectAccessReview
	for k, v := range rb {
		for kk, vv := range v.Resource {
			var namespace = ""
			if vv.Namespaced {
				namespace = ns
			}
			for _, verb := range vv.Verbs {
				sar := &authorizationv1.SelfSubjectAccessReview{
					Spec: authorizationv1.SelfSubjectAccessReviewSpec{
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Namespace:   namespace,
							Verb:        verb,
							Group:       k,
							Resource:    kk,
							Subresource: vv.SubResource,
							Name:        vv.SubResource, // the name field is the same as subresource field, based on observation
						},
					},
				}
				sar_allowed = append(sar_allowed, sar)
			}
		}
	}

	for k, v := range fb {
		for kk, vv := range v.Resource {
			var namespace = ""
			if vv.Namespaced {
				namespace = ns
			}
			for _, verb := range vv.Verbs {
				sar := &authorizationv1.SelfSubjectAccessReview{
					Spec: authorizationv1.SelfSubjectAccessReviewSpec{
						ResourceAttributes: &authorizationv1.ResourceAttributes{
							Namespace:   namespace,
							Verb:        verb,
							Group:       k,
							Resource:    kk,
							Subresource: vv.SubResource,
							Name:        vv.SubResource, // the name field is the same as subresource field, based on observation
						},
					},
				}
				sar_forbidden = append(sar_forbidden, sar)
			}
		}
	}

	return sar_allowed, sar_forbidden
}

func doSelfSubjectAccessReview(auth_client authorizationv1client.AuthorizationV1Interface, sar *authorizationv1.SelfSubjectAccessReview, expect bool) (bool, error) {

	response, err := auth_client.SelfSubjectAccessReviews().Create(context.TODO(), sar, metav1.CreateOptions{})
	if err != nil {
		utils.FatalLogger.Printf("Failed to create SelfSubjectAccessReviews, error:\n%s", err.Error())
		return false, err
	}

	name := response.Spec.ResourceAttributes.Name
	apigroup := response.Spec.ResourceAttributes.Group
	namespace := response.Spec.ResourceAttributes.Namespace
	resource := response.Spec.ResourceAttributes.Resource
	verb := response.Spec.ResourceAttributes.Verb

	utils.InfoLogger.Printf("Reviewing access for {apigroup: %s, resource: %s, name :%s, namespace: %s, verb: %s} expecting: %t\n", apigroup, resource, name, namespace, verb, expect)
	if response.Status.Allowed {
		fmt.Println("yes")
		utils.InfoLogger.Printf("Verdict: yes")
	} else {
		fmt.Println("no")
		utils.InfoLogger.Printf("Verdict: no")
		if len(response.Status.Reason) > 0 {
			utils.InfoLogger.Printf(" - %v", response.Status.Reason)
		}
		if len(response.Status.EvaluationError) > 0 {
			utils.InfoLogger.Printf(" - %v", response.Status.EvaluationError)
		}
		fmt.Println()
	}
	verdict := response.Status.Allowed
	if expect == response.Status.Allowed {
		fmt.Printf("---Review Passed, expecting %t, received %t\n", expect, verdict)
		utils.InfoLogger.Printf("Review Passed, expecting %t, received %t\n", expect, verdict)
	} else {
		fmt.Printf("+++Review Failed, expecting %t, received %t\n", expect, verdict)
		utils.ErrorLogger.Printf("Review Failed, expecting %t, received %t\n", expect, verdict)
	}
	return verdict, nil
}

func DoBatchSelfSubjectAccessReviews(path string, l []*authorizationv1.SelfSubjectAccessReview, expect bool) error {
	auth_client := getClientset(path)
	for _, sar := range l {
		if _, err := doSelfSubjectAccessReview(auth_client, sar, expect); err != nil {
			return err
		}
	}
	return nil
}
