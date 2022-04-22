package fdooze

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IgnoringFds matcher", func() {

	It("fails for invalid actual", func() {
		m := IgnoringFiledescriptors(nil)
		Expect(m.Match(nil)).Error().To(HaveOccurred())
		Expect(m.Match(42)).Error().To(HaveOccurred())
	})

})
