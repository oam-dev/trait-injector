package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("Inside of a new namespace", func() {

	Describe("when creating Deployment", func() {
		It("should inject env", func() {
			Expect(1).NotTo(HaveOccurred())
		})
		It("should inject file", func() {
		})
	})
})
