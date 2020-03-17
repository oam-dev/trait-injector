/*

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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
})

// SetupTest will set up a testing environment before each test
// Call this function at the start of each of your tests.
func SetupTest(ctx context.Context) {
	var stopCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})

	})

	AfterEach(func() {
		close(stopCh)
	})
}
