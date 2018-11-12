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

package nodetasks

import "k8s.io/kops/upup/pkg/fi"

type StorageArray struct {
	Label string `json:"name,omitempty"`
}

var _ fi.Task = &StorageArray{}
var _ fi.HasDependencies = &StorageArray{}

func (s *StorageArray) GetDependencies(tasks map[string]fi.Task) []fi.Task {

}

var _ fi.HasName = &StorageArray{}

func (s *StorageArray) GetName() *string {
	return &s.Label
}

func (s *StorageArray) SetName(name string) {
	s.Label = name
}
