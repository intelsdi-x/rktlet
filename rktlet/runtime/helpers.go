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

package runtime

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/rkt/lib"
	"github.com/golang/glog"

	"github.com/coreos/rkt/networking/netinfo"
	runtimeApi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

const (
	kubernetesLabelKeyPrefix              = "k8s.io/label/"
	kubernetesAnnotationKeyPrefix         = "k8s.io/annotation/"
	kubernetesReservedAnnoImageNameKey    = "k8s.io/reserved/image-name"
	kubernetesReservedAnnoAttemptCountKey = "k8s.io/reserved/attempt-count"
)

func parseContainerID(containerID string) (uuid, containerName string, err error) {
	values := strings.Split(containerID, ":")
	if len(values) != 2 {
		return "", "", fmt.Errorf("invalid container ID %q", containerID)
	}
	return values[0], values[1], nil
}

func buildContainerID(uuid, containerName string) (containerID string) {
	return fmt.Sprintf("%s:%s", uuid, containerName)
}

func toContainerStatus(uuid string, app *rkt.App) *runtimeApi.ContainerStatus {
	var status runtimeApi.ContainerStatus

	id := buildContainerID(uuid, app.Name)
	status.Id = &id

	status.Metadata = &runtimeApi.ContainerMetadata{
		Name:    &app.Name,
		Attempt: getAttemptCount(app.Annotations),
	}

	state := runtimeApi.ContainerState_UNKNOWN
	switch app.State {
	case rkt.AppStateUnknown:
		state = runtimeApi.ContainerState_UNKNOWN
	case rkt.AppStateCreated:
		state = runtimeApi.ContainerState_CREATED
	case rkt.AppStateRunning:
		state = runtimeApi.ContainerState_RUNNING
	case rkt.AppStateExited:
		state = runtimeApi.ContainerState_EXITED
	default:
		state = runtimeApi.ContainerState_UNKNOWN
	}

	status.State = &state
	status.CreatedAt = app.CreatedAt
	status.StartedAt = app.StartedAt
	status.FinishedAt = app.FinishedAt
	status.ExitCode = app.ExitCode
	status.ImageRef = &app.ImageID
	status.Image = &runtimeApi.ImageSpec{Image: getImageName(app.Annotations)}

	status.Labels = getKubernetesLabels(app.Annotations)
	status.Annotations = getKubernetesAnnotations(app.Annotations)

	// TODO: Make sure mount name is unique.
	for _, mnt := range app.Mounts {
		status.Mounts = append(status.Mounts, &runtimeApi.Mount{
			Name:          &mnt.Name,
			ContainerPath: &mnt.ContainerPath,
			HostPath:      &mnt.HostPath,
			Readonly:      &mnt.ReadOnly,
			// TODO: Selinux relabeling.
		})
	}

	return &status
}

func getKubernetesLabels(annotations map[string]string) map[string]string {
	ret := make(map[string]string)
	for key, value := range annotations {
		if strings.HasPrefix(key, kubernetesLabelKeyPrefix) {
			ret[strings.TrimPrefix(key, kubernetesLabelKeyPrefix)] = value
		}
	}
	return ret
}

func getKubernetesAnnotations(annotations map[string]string) map[string]string {
	ret := make(map[string]string)
	for key, value := range annotations {
		if strings.HasPrefix(key, kubernetesAnnotationKeyPrefix) {
			ret[strings.TrimPrefix(key, kubernetesAnnotationKeyPrefix)] = value
		}
	}
	return ret
}

func getAttemptCount(annotations map[string]string) *uint32 {
	var cnt uint32 = 0
	attemptCountStr := annotations[kubernetesReservedAnnoAttemptCountKey]
	if attemptCountStr == "" {
		return &cnt
	}

	attemptCount, err := strconv.Atoi(attemptCountStr)
	if err != nil {
		glog.Warningf("rkt: cannot parse attempt count %q: %v, assume zero", attemptCountStr, err)
		return &cnt
	}

	cnt = uint32(attemptCount)
	return &cnt
}

func getImageName(annotations map[string]string) *string {
	name := annotations[kubernetesReservedAnnoImageNameKey]
	return &name
}

// TODO remove this once https://github.com/coreos/rkt/pull/3194 is merged.
type Pod struct {
	UUID string `json:"name"`
	// State is defined in pkg/pod/pods.go
	State    string            `json:"state"`
	Networks []netinfo.NetInfo `json:"networks,omitempty"`
	// TODO(yifan): Decide if we want to include detailed app info.
	AppNames []string `json:"app_names,omitempty"`
}
