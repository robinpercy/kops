/*
Copyright 2016 The Kubernetes Authors.

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

package kops

import (
	"fmt"
	"github.com/blang/semver"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi/utils"
)

const RoleLabelName = "kubernetes.io/role"
const RoleMasterLabelValue = "master"
const RoleNodeLabelValue = "node"

// NodeLabels are defined in the InstanceGroup, but set flags on the kubelet config.
// We have a conflict here: on the one hand we want an easy to use abstract specification
// for the cluster, on the other hand we don't want two fields that do the same thing.
// So we make the logic for combining a KubeletConfig part of our core logic.
// NodeLabels are set on the instanceGroup.  We might allow specification of them on the kubelet
// config as well, but for now the precedence is not fully specified.
// (Today, NodeLabels on the InstanceGroup are merged in to NodeLabels on the KubeletConfig in the Cluster).
// In future, we will likely deprecate KubeletConfig in the Cluster, and move it into componentconfig,
// once that is part of core k8s.

// BuildKubeletConfigSpec returns the kubeletconfig for the specified instanceGroup
func BuildKubeletConfigSpec(cluster *Cluster, instanceGroup *InstanceGroup) (*KubeletConfigSpec, error) {
	// Merge KubeletConfig for NodeLabels
	c := &KubeletConfigSpec{}
	if instanceGroup.Spec.Role == InstanceGroupRoleMaster {
		utils.JsonMergeStruct(c, cluster.Spec.MasterKubelet)
	} else {
		utils.JsonMergeStruct(c, cluster.Spec.Kubelet)
	}

	if instanceGroup.Spec.Role == InstanceGroupRoleMaster {
		if c.NodeLabels == nil {
			c.NodeLabels = make(map[string]string)
		}
		c.NodeLabels[RoleLabelName] = RoleMasterLabelValue
	}

	for k, v := range instanceGroup.Spec.NodeLabels {
		if c.NodeLabels == nil {
			c.NodeLabels = make(map[string]string)
		}
		c.NodeLabels[k] = v
	}

	if instanceGroup.Spec.Kubelet != nil {
		utils.JsonMergeStruct(c, instanceGroup.Spec.Kubelet)
	}

	sv, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		return c, fmt.Errorf("Failed to lookup kubernetes version: %v", err)
	}

	// --register-with-taints was available in the first 1.6.0 alpha, no need to rely on semver's pre/build ordering
	sv.Pre = nil
	sv.Build = nil
	if sv.GTE(semver.Version{Major: 1, Minor: 6, Patch: 0, Pre: nil, Build: nil}) {
		for i, t := range instanceGroup.Spec.Taints {
			if c.Taints == nil {
				c.Taints = make([]string, len(instanceGroup.Spec.Taints))
			}
			c.Taints[i] = t
		}

		// Enable scheduling since it can be controlled via taints.
		// For pre-1.6.0 clusters, this is handled by tainter.go
		registerSchedulable := true
		c.RegisterSchedulable = &registerSchedulable

	} else if len(instanceGroup.Spec.Taints) > 0 && !hasDefaultTaints(instanceGroup) {
		return c, fmt.Errorf("User-specified taints are not supported before kubernetes version 1.6.0")
	}

	return c, nil
}

func hasDefaultTaints(instanceGroup *InstanceGroup) bool {
	return instanceGroup.Spec.Role == InstanceGroupRoleMaster &&
		len(instanceGroup.Spec.Taints) == 1 &&
		instanceGroup.Spec.Taints[0] == TaintNoScheduleMaster

}