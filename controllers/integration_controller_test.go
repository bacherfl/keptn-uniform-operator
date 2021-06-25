package controllers

import (
	"context"
	"fmt"
	"github.com/bacherfl/keptn-uniform-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("Integration controller", func() {

	const (
		IntegrationName      = "test-integration"
		IntegrationNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)
	Context("When deploying Keptn Integration Resource", func() {
		It("Should create correct Keptn Integration", func() {
			By("By creating Keptn Integration Resource", func() {
				ctx := context.Background()

				integration := &v1alpha1.Integration{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "uniform.smy.domain/v1alpha1",
						Kind:       "Integration",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      IntegrationName,
						Namespace: IntegrationNamespace,
					},
					Spec: v1alpha1.IntegrationSpec{
						Name:   IntegrationName,
						Events: []string{"sh.keptn.>"},
						Image:  "keptnsandbox/job-executor-service:0.1.1",
					},
				}

				Expect(k8sClient.Create(ctx, integration)).Should(Succeed())

				integrationLookupKey := types.NamespacedName{Name: IntegrationName, Namespace: IntegrationNamespace}
				createdIntegration := &v1alpha1.Integration{}

				Eventually(func() bool {
					err := k8sClient.Get(ctx, integrationLookupKey, createdIntegration)
					fmt.Println(err)
					return err == nil
				}, timeout, interval).Should(BeTrue())

				fmt.Println(createdIntegration.Spec)
				Expect(createdIntegration.Spec.Image).To(Equal("keptnsandbox/job-executor-service:0.1.1"))

				// TODO: check behavior of operator
			})
		})
	})
})
