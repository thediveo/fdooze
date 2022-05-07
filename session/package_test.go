//go:build linux

package session

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSessionPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "session package")
}
