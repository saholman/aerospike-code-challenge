package main_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	main "github.com/saholman/aerospike-code-challenge"
)

var _ = Describe("Main", func() {
	When("Run is called", func() {
		var (
			clientset kubernetes.Interface
			ctx       context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
			clientset = fake.NewSimpleClientset()
			err := main.Run(ctx, clientset)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			err := main.Cleanup(ctx, clientset)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should have created the aerospike namespace", func() {
			ns, err := clientset.CoreV1().Namespaces().Get(ctx, "aerospike", metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ns.ObjectMeta.Name).To(Equal("aerospike"))
		})

		It("should have created the hello-world pod", func() {
			pod, err := clientset.CoreV1().Pods("aerospike").Get(ctx, "hello-world", metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pod.ObjectMeta.Name).To(Equal("hello-world"))
			Expect(pod.Spec.Containers[0].Image).To(Equal("hello-world"))
		})
	})
})
