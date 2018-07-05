/*
Copyright 2015 The Kubernetes Authors.

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

package framework

import (
	"fmt"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	api "github.com/presslabs/mysql-operator/pkg/apis/mysql/v1alpha1"
	myclientset "github.com/presslabs/mysql-operator/pkg/generated/clientset/versioned"
	orc "github.com/presslabs/mysql-operator/pkg/util/orchestrator"
)

const (
	maxKubectlExecRetries           = 5
	DefaultNamespaceDeletionTimeout = 10 * time.Minute
	orchestratorURITemplate         = "http://localhost:%d/api"
	TIMEOUT                         = 120 * time.Second
	POLLING                         = 2 * time.Second
)

var OrchestratorPort = 3000

type Framework struct {
	BaseName  string
	Namespace *core.Namespace

	ClientSet   clientset.Interface
	MyClientSet myclientset.Interface

	cleanupHandle         CleanupActionHandle
	SkipNamespaceCreation bool

	OrcClient orc.Interface
}

func NewFramework(baseName string) *Framework {
	f := &Framework{
		BaseName:              baseName,
		SkipNamespaceCreation: false,
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	// The fact that we need this feels like a bug in ginkgo.
	// https://github.com/onsi/ginkgo/issues/222
	f.cleanupHandle = AddCleanupAction(f.AfterEach)

	if f.ClientSet == nil {
		By("Creating a kubernetes client")
		var err error
		f.ClientSet, f.MyClientSet, err = KubernetesClients()
		Expect(err).NotTo(HaveOccurred())

	}

	if !f.SkipNamespaceCreation {
		By("Building a namespace api object")
		namespace, err := f.CreateNamespace(f.BaseName, map[string]string{
			"e2e-framework": f.BaseName,
		})
		Expect(err).NotTo(HaveOccurred())

		f.Namespace = namespace
	}

	f.OrcClient = orc.NewFromUri(fmt.Sprintf(orchestratorURITemplate, OrchestratorPort))

}

// AfterEach deletes the namespace, after reading its events.
func (f *Framework) AfterEach() {
	By("Run cleanup actions")
	RemoveCleanupAction(f.cleanupHandle)

	By("Delete testing namespace")
	err := DeleteNS(f.ClientSet, f.Namespace.Name, DefaultNamespaceDeletionTimeout)
	if err != nil {
		Failf(fmt.Sprintf("Can't delete namespace: %s", err))
	}
}

func (f *Framework) CreateNamespace(baseName string, labels map[string]string) (
	*core.Namespace, error) {
	return CreateTestingNS(baseName, f.ClientSet, labels)
}

// WaitForPodReady waits for the pod to flip to ready in the namespace.
func (f *Framework) WaitForPodReady(podName string) error {
	return waitTimeoutForPodReadyInNamespace(f.ClientSet, podName,
		f.Namespace.Name, PodStartTimeout)
}

func (f *Framework) ClusterEventuallyCondition(cluster *api.MysqlCluster,
	condType api.ClusterConditionType, status core.ConditionStatus) {
	Eventually(func() []api.ClusterCondition {
		c, err := f.MyClientSet.MysqlV1alpha1().MysqlClusters(f.Namespace.Name).Get(
			cluster.Name, meta.GetOptions{})
		if err != nil {
			return nil
		}
		return c.Status.Conditions
	}, TIMEOUT, POLLING).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
		"Type":   Equal(condType),
		"Status": Equal(status),
	})))

}

func (f *Framework) NodeEventuallyCondition(cluster *api.MysqlCluster, nodeName string,
	condType api.NodeConditionType, status core.ConditionStatus) {
	Eventually(func() []api.NodeCondition {
		cluster, err := f.MyClientSet.MysqlV1alpha1().MysqlClusters(cluster.Namespace).Get(
			cluster.Name, meta.GetOptions{})
		if err != nil {
			return nil
		}

		for _, ns := range cluster.Status.Nodes {
			if ns.Name == nodeName {
				return ns.Conditions
			}
		}

		return nil
	}, TIMEOUT, POLLING).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
		"Type":   Equal(condType),
		"Status": Equal(status),
	})))
}
