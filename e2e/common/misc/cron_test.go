//go:build integration
// +build integration

// To enable compilation of this file in Goland, go to "Settings -> Go -> Vendoring & Build Tags -> Custom Tags" and add "integration"

/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	. "github.com/apache/camel-k/v2/e2e/support"
	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"

	"github.com/stretchr/testify/assert"
)

func TestRunCronExample(t *testing.T) {
	t.Parallel()

	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		t.Run("cron-yaml", func(t *testing.T) {
			name := RandomizedSuffixName("cron-yaml")
			g.Expect(KamelRun(t, ctx, ns, "files/cron-yaml.yaml", "--name", name).Execute()).To(Succeed())
			g.Eventually(IntegrationCronJob(t, ctx, ns, name)).ShouldNot(BeNil())
			g.Eventually(IntegrationConditionStatus(t, ctx, ns, name, v1.IntegrationConditionReady)).Should(Equal(corev1.ConditionTrue))
			g.Eventually(IntegrationLogs(t, ctx, ns, name), TestTimeoutMedium).Should(ContainSubstring("Magicstring!"))
			g.Expect(Kamel(t, ctx, "delete", name, "-n", ns).Execute()).To(Succeed())
		})

		t.Run("cron-timer", func(t *testing.T) {
			name := RandomizedSuffixName("cron-timer")
			g.Expect(KamelRun(t, ctx, ns, "files/cron-timer.yaml", "--name", name).Execute()).To(Succeed())
			g.Eventually(IntegrationCronJob(t, ctx, ns, name), TestTimeoutLong).ShouldNot(BeNil())
			g.Eventually(IntegrationConditionStatus(t, ctx, ns, name, v1.IntegrationConditionReady)).Should(Equal(corev1.ConditionTrue))
			g.Eventually(IntegrationLogs(t, ctx, ns, name), TestTimeoutMedium).Should(ContainSubstring("Magicstring!"))
			g.Expect(Kamel(t, ctx, "delete", name, "-n", ns).Execute()).To(Succeed())
		})

		t.Run("cron-fallback", func(t *testing.T) {
			name := RandomizedSuffixName("cron-fallback")
			g.Expect(KamelRun(t, ctx, ns, "files/cron-fallback.yaml", "--name", name).Execute()).To(Succeed())
			g.Eventually(IntegrationPodPhase(t, ctx, ns, name)).Should(Equal(corev1.PodRunning))
			g.Eventually(IntegrationConditionStatus(t, ctx, ns, name, v1.IntegrationConditionReady)).Should(Equal(corev1.ConditionTrue))
			g.Eventually(IntegrationLogs(t, ctx, ns, name), TestTimeoutMedium).Should(ContainSubstring("Magicstring!"))
			g.Expect(Kamel(t, ctx, "delete", name, "-n", ns).Execute()).To(Succeed())
		})

		t.Run("cron-quartz", func(t *testing.T) {
			name := RandomizedSuffixName("cron-quartz")
			g.Expect(KamelRun(t, ctx, ns, "files/cron-quartz.yaml", "--name", name).Execute()).To(Succeed())
			g.Eventually(IntegrationPodPhase(t, ctx, ns, name)).Should(Equal(corev1.PodRunning))
			g.Eventually(IntegrationConditionStatus(t, ctx, ns, name, v1.IntegrationConditionReady)).Should(Equal(corev1.ConditionTrue))
			g.Eventually(IntegrationLogs(t, ctx, ns, name), TestTimeoutMedium).Should(ContainSubstring("Magicstring!"))
			g.Expect(Kamel(t, ctx, "delete", name, "-n", ns).Execute()).To(Succeed())
		})

		t.Run("cron-trait-yaml", func(t *testing.T) {
			name := RandomizedSuffixName("cron-trait-yaml")
			g.Expect(KamelRun(t, ctx, ns, "files/cron-trait-yaml.yaml", "--name", name, "-t", "cron.enabled=true", "-t", "cron.schedule=0/2 * * * *").Execute()).To(Succeed())
			g.Eventually(IntegrationConditionStatus(t, ctx, ns, name, v1.IntegrationConditionReady)).Should(Equal(corev1.ConditionTrue))
			g.Eventually(IntegrationCronJob(t, ctx, ns, name)).ShouldNot(BeNil())

			// Verify that `-t cron.schedule` overrides the schedule in the yaml
			//
			// kubectl get cronjobs -n test-de619ae2-eddc-4bac-86a6-53d80be030ea
			// NAME               SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
			// cron-trait-yaml    0/2 * * * *   False     0        <none>          38s

			cronJob := IntegrationCronJob(t, ctx, ns, name)()
			assert.Equal(t, "0/2 * * * *", cronJob.Spec.Schedule)
			g.Expect(Kamel(t, ctx, "delete", name, "-n", ns).Execute()).To(Succeed())
		})
	})
}
