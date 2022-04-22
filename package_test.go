package fdooze

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFdoozePackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fdooze package")
}
