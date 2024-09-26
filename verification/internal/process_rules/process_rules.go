package process_rules

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"

	utils "github.com/lamatriz/ra2-auth/k8s-rbac-verfication/utils"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

/*
1. using names of apigroup then resource as keys, the version is one of the value, here we assume within the same apigroup, the resource name is unique accross the versions.
So far this has been true based on observation, if this does not hold, however, then version must be promoted to a key type, and resourceName should become a nested map.
2. The "rbac.authorization.k8s.io/v1" spec (as of Kube 1.23.7) only applies to resource and apiGroup. However the same resource name can be in different versions within a same api-group, therefore we have to assume the
rbac rules apply to all resources in different version within the apigroup.
3. the shortNames, Versions, kind are parsed and stored regardless for debugging purpose and future support.

	{
		apigroup: {
			resourceName : {
				sub, //subresource
				[v1], //version
				sn, //short names
				kind,
				true, //namespaced
				[verb1,	verb2]
			}
		}
	}
*/
type ResourceValueType struct {
	SubResource string   `json:"subresource"`
	Versions    []string `json:"version"`
	ShortNames  []string `json:"shortNames"`
	Kind        string   `json:"kind"`
	Namespaced  bool     `json:"namespaced"`
	Verbs       []string `json:"verbs"`
}
type ApiGroupValueType struct {
	Resource map[string]ResourceValueType `json:"resources"`
}

// All resources serialized
var AllResourcesMap map[string]ApiGroupValueType

// RBAC yaml input rule serialized
var RbacRulesMap map[string]ApiGroupValueType

// Forbidden ApiGroups, Resources and Verbs
var ForbiddenRulesMap map[string]ApiGroupValueType

func ParseAllApiresources(path string) {
	f := utils.ReadFile(path)
	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string
	var header string

	for fileScanner.Scan() {
		if line := fileScanner.Text(); line[:1] != "#" {
			if line[:4] == "NAME" {
				header = line
			} else {
				fileLines = append(fileLines, line)
			}
		}
	}

	var ruler []int
	ruler = append(ruler, strings.Index(header, "SHORTNAMES"))
	ruler = append(ruler, strings.Index(header, "APIVERSION"))
	ruler = append(ruler, strings.Index(header, "NAMESPACED"))
	ruler = append(ruler, strings.Index(header, "KIND"))
	ruler = append(ruler, strings.Index(header, "VERBS"))

	AllResourcesMap = make(map[string]ApiGroupValueType)
	for _, line := range fileLines {
		name := strings.Trim(line[:ruler[0]], " ")
		short_names := strings.Trim(line[ruler[0]:ruler[1]], " ")
		apiversion := strings.Trim(line[ruler[1]:ruler[2]], " ")
		namespaced := strings.Trim(line[ruler[2]:ruler[3]], " ")
		kind := strings.Trim(line[ruler[3]:ruler[4]], " ")
		verbs := strings.Trim(line[ruler[4]:], " []")

		subresource_name := ""
		if split_names, splited := utils.SplitString(name, "/"); splited {
			subresource_name = split_names[1]
		}

		var apigroup, version string
		if index := strings.Index(apiversion, "/v"); index <= 0 {
			// "v1", or "/v1"
			apigroup = ""
			version = apiversion
		} else {
			apigroup = apiversion[:index]
			version = apiversion[index+1:]
		}

		entry, ok := AllResourcesMap[apigroup]
		if !ok {
			AllResourcesMap[apigroup] = ApiGroupValueType{}
		}

		versions := []string{version}
		if entry.Resource == nil {
			entry.Resource = make(map[string]ResourceValueType)
		} else {
			versions = entry.Resource[name].Versions
			versions = append(versions, version)
		}

		entry.Resource[name] = ResourceValueType{
			subresource_name,
			versions,
			strings.Fields(short_names),
			kind,
			namespaced == "true",
			strings.Fields(verbs),
		}
		AllResourcesMap[apigroup] = entry
	}
	f.Close()
}

/*
1. For entries with "resource/subresources" format, the entire string is used as the resource name key
2. Not going to support group.res/specific-resource format

vdu:~/> k auth can-i list jobs.batch/bar -v 8
...
I1219 20:25:17.188213 1678894 request.go:1181] Response Body:
{"kind":"SelfSubjectAccessReview","apiVersion":"authorization.k8s.io/v1","metadata":{"creationTimestamp":null,"managedFields":

	[{"manager":"kubectl","operation":"Update","apiVersion":"authorization.k8s.io/v1","time":"2022-12-19T20:25:17Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:resourceAttributes":{".":{},"f:group":{},"f:name":{},"f:namespace":{},"f:resource":{},"f:verb":{}}}}}]},
	"spec":{"resourceAttributes":{"namespace":"default","verb":"list","group":"batch","resource":"jobs","name":"bar"}},
	"status":{"allowed":true,"reason":"RBAC: allowed by ClusterRoleBinding \"cluster-operator\" of ClusterRole \"cluster-operator\" to Group \"oidc:cluster-operator\""}}

yes
*/
func ParseK8sRbacYaml(path string) {

	f := utils.ReadFile(path)
	decoder := yamlutil.NewYAMLOrJSONDecoder(bufio.NewReader(f), 100)

	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, gkv, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			utils.FatalLogger.Printf(err.Error())
		}
		utils.InfoLogger.Printf("Processing %s", gkv.Kind)

		RbacRulesMap = make(map[string]ApiGroupValueType)
		jsonStream, err := json.Marshal(unstructuredMap["rules"])
		if err != nil {
			break
		}

		jsonDecoder := json.NewDecoder(strings.NewReader(string(jsonStream)))
		for {
			// type to Marshall k8s rbac yaml
			type ArvEntry struct {
				ApiGroups []string `json:"apiGroups"`
				Resources []string `json:"resources"`
				Verbs     []string `json:"verbs"`
			}
			var arvEntryArray []ArvEntry
			if err := jsonDecoder.Decode(&arvEntryArray); err != nil {
				break
			}
			for _, arv := range arvEntryArray {
				apiGroups := arv.ApiGroups
				resources := arv.Resources
				verbs := arv.Verbs

				// star * handler at apigroup
				var apigroup_keys []string
				if apiGroups[0] == "*" {
					apigroup_keys = make([]string, 0, len(AllResourcesMap))
					for k := range AllResourcesMap {
						apigroup_keys = append(apigroup_keys, k)
					}
				} else {
					apigroup_keys = make([]string, 0, len(apiGroups))
					for _, k := range apiGroups {
						apigroup_keys = append(apigroup_keys, k)
					}
				}

				for _, ag := range apigroup_keys {
					if ag_entry, ok := AllResourcesMap[ag]; !ok {
						log.Default().Printf("Found nonexisting apigroup: %s, skipping!\n", ag)
						continue
					} else {
						entry, ok := RbacRulesMap[ag]
						if !ok {
							entry = ApiGroupValueType{}
						}
						if entry.Resource == nil {
							entry.Resource = make(map[string]ResourceValueType)
						}
						/*
							star * handler at resource
							Find all defined resources under said apiGroup, handling case:
							resources:
							- "*"
						*/
						var resource_keys []string
						if resources[0] == "*" {
							resource_keys = make([]string, 0, len(ag_entry.Resource))
							for k := range ag_entry.Resource {
								resource_keys = append(resource_keys, k)
							}
						} else {
							resource_keys = make([]string, 0, len(resources))
							for _, k := range resources {
								resource_keys = append(resource_keys, k)
							}
						}

						for _, res := range resource_keys {
							subresource_name := ""
							if split_names, splited := utils.SplitString(res, "/"); splited {
								subresource_name = split_names[1]
							}

							if re_entry, ok := ag_entry.Resource[res]; !ok {
								utils.InfoLogger.Printf("Found nonexisting resource %s in apigroup: %s, skipping!\n", res, ag)
								//continue
							} else {
								/*
									star * handler at verbs
									Find all vaid verbs under said apiGroup and resource, hanndling case:
									resources:
									- "*"
									verbs:
									- "*"
								*/
								var res_type = ResourceValueType{
									subresource_name,
									re_entry.Versions,
									re_entry.ShortNames,
									re_entry.Kind,
									re_entry.Namespaced,
									nil,
								}
								if verbs[0] == "*" {
									res_type.Verbs = re_entry.Verbs
								} else {
									for _, verb := range verbs {
										if !slices.Contains(re_entry.Verbs, verb) {
											utils.InfoLogger.Printf("Found non-avaible verb %s for resource %s, ignoring", verb, res)
										} else {
											res_type.Verbs = append(res_type.Verbs, verb)
										}
									}
								}
								entry.Resource[res] = res_type
							}
						}
						RbacRulesMap[ag] = entry
					}
				}
			}
		}
	}
	f.Close()
}

/*
Substract allowed items (recorded in 'rbac') from 'all' items, result would be "forbidden" items (recorded in 'ret'), follows this logic:

	*if an apigroup is in all but not in rbac, copy all resources owned by this apiGroup to ret;
	*else if an apigroup is found in both all and rbac, check the resources under this agigroup between all and rbac, if it is only in all not in rbac, copy the resource into the ret, under the apigroup entry
	*else if a resource under an apigroup is found both in all and rbac, check the verbs between the two, if there is verbs found only in all, but not in rbac, copy the verbs to the resource under the apigroup entry

	Loosely written rbac yaml often has more verbs than what are avaiable for a resource, i.e. for "pods/exec", the available verbs are only "[create get]", but k8s auth check does not return error when "list" or "delete"
	verbs are checked against such resource, here we try to mimic the same behavoir.
*/
func FilterRules() {
	ForbiddenRulesMap = make(map[string]ApiGroupValueType)

	for k, v := range AllResourcesMap {
		if rb_val, ok := RbacRulesMap[k]; !ok {
			// did not find the entire apiGroup, copy over.
			addToRetObject(ForbiddenRulesMap, k, v, "", nil, nil)
		} else {
			// found the apiGroup, need to verify each Resource, compare val vs. rb_val
			for kk, vv := range v.Resource {
				if rb_res, ok := rb_val.Resource[kk]; !ok {
					// Found a resource not allowed, copy the entire resource over
					addToRetObject(ForbiddenRulesMap, k, nil, kk, vv, nil)
				} else {
					// Found apiGroup, found resource, need to check the verbs are equal.
					var neg_verbs []string
					for _, verb := range vv.Verbs {
						if !slices.Contains(rb_res.Verbs, verb) {
							neg_verbs = append(neg_verbs, verb)
						}
					}
					if len(neg_verbs) > 0 {
						addToRetObject(ForbiddenRulesMap, k, nil, kk, vv, neg_verbs)
					}
				}
			}
		}
	}
}

// helper functions
func addToRetObject(ret map[string]ApiGroupValueType, k string, ag interface{}, kk string, res interface{}, verbs []string) {
	ret_entry, ok := ret[k]
	if !ok {
		ret_entry = ApiGroupValueType{}
		ret_entry.Resource = make(map[string]ResourceValueType)
	}

	if cast, ok := ag.(ApiGroupValueType); ok {
		for kk, vv := range cast.Resource {
			ret_entry.Resource[kk] = ResourceValueType{
				vv.SubResource,
				vv.Versions,
				vv.ShortNames,
				vv.Kind,
				vv.Namespaced,
				vv.Verbs,
			}
		}
	} else if cast, ok := res.(ResourceValueType); ok {
		ret_res := ResourceValueType{
			cast.SubResource,
			cast.Versions,
			cast.ShortNames,
			cast.Kind,
			cast.Namespaced,
			cast.Verbs,
		}
		if verbs != nil {
			ret_res.Verbs = verbs
		}
		ret_entry.Resource[kk] = ret_res
	}
	ret[k] = ret_entry
}
