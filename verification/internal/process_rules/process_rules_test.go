package process_rules

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

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
ippools                                                 crd.projectcalico.org/v1               false        IPPool                           [delete deletecollection get list patch create update watch]
endpointslices                                          discovery.k8s.io/v1                    true         EndpointSlice                    [create delete deletecollection get list patch update watch]
endpointslices                                          discovery.k8s.io/v1beta1               true         EndpointSlice                    [create delete deletecollection get list patch update watch]`

var ar_json_str string = `{
	"": {
	  "resources": {
		"pods": {
			"subresource": "",    
			"version": ["v1"],    
			"shortNames": [     
			  "po"
			],
			"kind": "Pod",
			"namespaced": true,
			"verbs": [
			  "create",
			  "delete",
			  "deletecollection",
			  "get",
			  "list",
			  "patch",
			  "update",
			  "watch"
			]
		  },
		"pods/exec": {
			"subresource": "exec",    
			"version": ["v1"],    
			"shortNames": [],
			"kind": "PodExecOptions",
			"namespaced": true,
			"verbs": [
			  "create",
			  "get"
			]
		  },
		"configmaps": {
		  "subresource": "",    
		  "version": ["v1"],    
		  "shortNames": [     
			"cm"
		  ],
		  "kind": "ConfigMap",
		  "namespaced": true,
		  "verbs": [
			"create",
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"update",
			"watch"
		  ]
		}
	  }
	},
	"apps": {
	  "resources": {
		"statefulsets": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [
			"sts"
		  ],
		  "kind": "StatefulSet",
		  "namespaced": true,
		  "verbs": [
			"create",
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"update",
			"watch"
		  ]
		}
	  }
	},
	"authentication.k8s.io": {
	  "resources": {
		"tokenreviews": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "TokenReview",
		  "namespaced": false,
		  "verbs": [
			"create"
		  ]
		}
	  }
	},
	"crd.projectcalico.org": {
	  "resources": {
		"ippools": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "IPPool",
		  "namespaced": false,
		  "verbs": [
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"create",
			"update",
			"watch"
		  ]
		},
		"networksets": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "NetworkSet",
		  "namespaced": true,
		  "verbs": [
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"create",
			"update",
			"watch"
		  ]
		}
	  }
	},
	"discovery.k8s.io": {
	  "resources": {
	    "endpointslices": {
	  	  "subresource": "",
	  	    "version": [
	  	      "v1",
	  	      "v1beta1"
	  	    ],
	  	  "shortNames": [],
	  	  "kind": "EndpointSlice",
	  	  "namespaced": true,
	  	  "verbs": [
	  	    "create",
	  	    "delete",
	  	    "deletecollection",
	  	    "get",
	  	    "list",
	  	    "patch",
	  	    "update",
	  	    "watch"
	  	  ]
	    }
	  }
    }
  }`

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
var rb_json_str = `{
	"": {
	  "resources": {
		"configmaps": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [
			"cm"
		  ],
		  "kind": "ConfigMap",
		  "namespaced": true,
		  "verbs": [
			"get",
			"list",
			"watch",
			"delete"
		  ]
	    },
	    "pods/exec": {    
	      "subresource": "exec",    
	      "version": ["v1"],    
	      "shortNames": [],
	      "kind": "PodExecOptions",
	      "namespaced": true,
	      "verbs": [
	    	"get"
		  ]
		}
	  }
	},
	"crd.projectcalico.org": {
	  "resources": {
		"ippools": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "IPPool",
		  "namespaced": false,
		  "verbs": [
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"create",
			"update",
			"watch"
		  ]
		},
		"networksets": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "NetworkSet",
		  "namespaced": true,
		  "verbs": [
			"get",
			"watch",
			"list"
		  ]
		}
	  }
	}
  }`
var rbac_all_star_yaml_text = `#Sample Clusterrole yaml file
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
  name: namespace-admin
rules:
- apiGroups:
  - "*"
  resources:
  - "*"
  verbs:
  - "*"
`

/*
TODO: notice the in rbac.yaml

	resources:
	- pods/exec

vs. the emtpy subresource name in forbidden roles. this is due to the lack of support on fine-grained subresources.

	    "pods": {
		  "subresource": "",

Need to fix the code then fix the unit test here.
*/
var fb_json_str string = `{
	"": {
	  "resources": {
		"configmaps": {
	      "subresource": "",
		  "version": ["v1"],
		  "shortNames": [
			"cm"
		  ],
		  "kind": "ConfigMap",
		  "namespaced": true,
		  "verbs": [
			"create",
			"deletecollection",
			"patch",
			"update"
		  ]
		},
	    "pods": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [
		    "po"
		  ],
		  "kind": "Pod",
		  "namespaced": true,
		  "verbs": [
			"create",
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"update",
			"watch"
		  ]
		},
	    "pods/exec": { 
	      "subresource": "exec",    
	      "version": ["v1"],    
	      "shortNames": [],
	      "kind": "PodExecOptions",
	      "namespaced": true,
	      "verbs": [
	        "create"
		  ]
		}
	  }
	},
	"apps": {
	  "resources": {
		"statefulsets": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [
			"sts"
		  ],
		  "kind": "StatefulSet",
		  "namespaced": true,
		  "verbs": [
			"create",
			"delete",
			"deletecollection",
			"get",
			"list",
			"patch",
			"update",
			"watch"
		  ]
		}
	  }
	},
	"authentication.k8s.io": {
	  "resources": {
		"tokenreviews": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "TokenReview",
		  "namespaced": false,
		  "verbs": [
			"create"
		  ]
		}
	  }
	},
	"crd.projectcalico.org": {
	  "resources": {
		"networksets": {
		  "subresource": "",
		  "version": ["v1"],
		  "shortNames": [],
		  "kind": "NetworkSet",
		  "namespaced": true,
		  "verbs": [
			"delete",
			"deletecollection",
			"patch",
			"create",
			"update"
		  ]
		}
	  }
	},
	"discovery.k8s.io": {
		"resources": {
		  "endpointslices": {
			  "subresource": "",
				"version": [
				  "v1",
				  "v1beta1"
				],
			  "shortNames": [],
			  "kind": "EndpointSlice",
			  "namespaced": true,
			  "verbs": [
				"create",
				"delete",
				"deletecollection",
				"get",
				"list",
				"patch",
				"update",
				"watch"
			  ]
		  }
		}
	  }
	}`

var ar, rb, fb map[string]ApiGroupValueType

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
	if err := os.WriteFile("./test_all_api_resources.txt", []byte(api_resource_txt), 0644); err != nil {
		log.Panic(err)
	}
	if err := os.WriteFile("./test_clusterrole.yaml", []byte(rbac_yaml_text), 0644); err != nil {
		log.Panic(err)
	}
	if err := os.WriteFile("./test_all_star_clusterrole.yaml", []byte(rbac_all_star_yaml_text), 0644); err != nil {
		log.Panic(err)
	}
	json.Unmarshal([]byte(ar_json_str), &ar)
	json.Unmarshal([]byte(rb_json_str), &rb)
	json.Unmarshal([]byte(fb_json_str), &fb)
}

func shutdown() {
	if err := os.Remove("./test_all_api_resources.txt"); err != nil {
		log.Panic(err)
	}
	if err := os.Remove("./test_clusterrole.yaml"); err != nil {
		log.Panic(err)
	}
	if err := os.Remove("./test_all_star_clusterrole.yaml"); err != nil {
		log.Panic(err)
	}
	fmt.Printf("All Done.")
}

func TestParseAllApiResources(t *testing.T) {

	ParseAllApiresources("./test_all_api_resources.txt")
	if !reflect.DeepEqual(AllResourcesMap, ar) {
		t.Error("Maps are not equal!")
		t.Log("Expected:\n")
		utils.PrettyPrintJson(ar)
		t.Log("Received:\n")
		utils.PrettyPrintJson(AllResourcesMap)
	}
}

func TestParseK8sRbacYaml(t *testing.T) {
	ParseAllApiresources("./test_all_api_resources.txt")
	ParseK8sRbacYaml("./test_clusterrole.yaml")

	if !reflect.DeepEqual(RbacRulesMap, rb) {
		t.Error("Maps are not equal!")
		t.Log("Expected:\n")
		utils.PrettyPrintJson(rb)
		t.Log("Received:\n")
		utils.PrettyPrintJson(RbacRulesMap)
	}
}

func TestFilterRules(t *testing.T) {
	ParseAllApiresources("./test_all_api_resources.txt")
	ParseK8sRbacYaml("./test_clusterrole.yaml")

	FilterRules()
	if !reflect.DeepEqual(ForbiddenRulesMap, fb) {
		t.Error("Maps are not equal!")
		t.Log("Expected:\n")
		utils.PrettyPrintJson(fb)
		t.Log("Received:\n")
		utils.PrettyPrintJson(ForbiddenRulesMap)
	}

	// test all * rbac yaml file
	RbacRulesMap = make(map[string]ApiGroupValueType)
	ForbiddenRulesMap = make(map[string]ApiGroupValueType)
	fb = make(map[string]ApiGroupValueType)

	ParseK8sRbacYaml("./test_all_star_clusterrole.yaml")
	FilterRules()

	//nothing should be forbidden, ForbiddenRulesMap is empty now.
	if !reflect.DeepEqual(ForbiddenRulesMap, fb) {
		t.Error("Maps are not equal!")
		t.Log("Expected:\n")
		utils.PrettyPrintJson(fb)
		t.Log("Received:\n")
		utils.PrettyPrintJson(ForbiddenRulesMap)
	}
	//all should be allowed
	if !reflect.DeepEqual(AllResourcesMap, RbacRulesMap) {
		t.Error("Maps are not equal!")
		t.Log("Expected:\n")
		utils.PrettyPrintJson(AllResourcesMap)
		t.Log("Received:\n")
		utils.PrettyPrintJson(RbacRulesMap)
	}
}

func TestAddToRetObject(t *testing.T) {
	ret := map[string]ApiGroupValueType{}
	k := "crd.projectcalico.org"
	kk := "ippools"
	all_verbs := []string{
		"delete",
		"deletecollection",
		"get",
		"list",
		"patch",
		"create",
		"update",
		"watch",
	}

	addToRetObject(ret, k, ar[k], "", nil, nil)
	if r, ok := ret[k]; !ok {
		t.Errorf("Failed to add %s ApiGroup.", k)
	} else if rr, ok := r.Resource[kk]; !ok {
		t.Errorf("Failed to add %s resource to %s Apigroup", kk, k)
	} else if verbs := rr.Verbs; len(verbs) != 8 {
		t.Errorf("Failed to add verbs for %s resource for %s ApiGroup.", kk, k)
	} else {

		if !reflect.DeepEqual(rr.Verbs, all_verbs) {
			t.Errorf("Failed to add verbs correctly for %s resource for %s ApiGroup. Expect:\n%s\ngot:\n%s", kk, k, strings.Join(rr.Verbs, ", "), strings.Join(verbs, ", "))
		}
	}

	ret = make(map[string]ApiGroupValueType)
	addToRetObject(ret, k, nil, kk, ar[k].Resource[kk], nil)
	if r, ok := ret[k]; !ok {
		t.Errorf("Failed to add %s ApiGroup.", k)
	} else if rr, ok := r.Resource[kk]; !ok {
		t.Errorf("Failed to add %s resource to %s Apigroup", kk, k)
	} else if verbs := rr.Verbs; len(verbs) != 8 {
		t.Errorf("Failed to add verbs for %s resource for %s ApiGroup.", kk, k)
	} else {
		if !reflect.DeepEqual(rr.Verbs, all_verbs) {
			t.Errorf("Failed to add verbs correctly for %s resource for %s ApiGroup. Expect:\n%s\ngot:\n%s", kk, k, strings.Join(rr.Verbs, ", "), strings.Join(verbs, ", "))
		}
	}

	ret = make(map[string]ApiGroupValueType)
	addToRetObject(ret, k, nil, kk, ar[k].Resource[kk], []string{"get", "list", "watch"})
	if r, ok := ret[k]; !ok {
		t.Errorf("Failed to add %s ApiGroup.", k)
	} else if rr, ok := r.Resource[kk]; !ok {
		t.Errorf("Failed to add %s resource to %s Apigroup", kk, k)
	} else if verbs := rr.Verbs; len(verbs) != 3 {
		t.Errorf("Failed to add verbs for %s resource for %s ApiGroup.", kk, k)
	} else {
		verbs := []string{
			"get",
			"list",
			"watch",
		}
		if !reflect.DeepEqual(rr.Verbs, verbs) {
			t.Errorf("Failed to add verbs correctly for %s resource for %s ApiGroup. Expect:\n %s\ngot:\n%s", kk, k, strings.Join(rr.Verbs, ", "), strings.Join(verbs, ", "))
		}
	}
}
