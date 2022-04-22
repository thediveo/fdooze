package filedesc

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFiledescPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "filedesc package")
}
