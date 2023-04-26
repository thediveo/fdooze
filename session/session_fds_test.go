// Copyright 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

//go:build linux

package session

import (
	"os/exec"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/thediveo/fdooze"
	"github.com/thediveo/fdooze/filedesc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func HaveFdWithPath(matcher gomega.OmegaMatcher) gomega.OmegaMatcher {
	return WithTransform(
		func(fd filedesc.FileDescriptor) string {
			file, ok := fd.(*filedesc.PathFd)
			if !ok {
				return ""
			}
			return file.Path()
		},
		matcher,
	)
}

var _ = Describe("session fd leak detection", func() {

	Context("invalid sessions", func() {

		It("rejects nil sessions and commands", func() {
			Expect(FiledescriptorsFor(nil)).Error().To(HaveOccurred())
			Expect(FiledescriptorsFor(&gexec.Session{})).Error().To(HaveOccurred())
		})

		It("rejects session without a process", func() {
			session := &gexec.Session{Command: exec.Command("foobar")}
			Expect(FiledescriptorsFor(session)).Error().To(HaveOccurred())
		})

		It("returns an error when the session already has terminated", func() {
			cmd := exec.Command("go", "version")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			Expect(FiledescriptorsFor(session)).Error().To(MatchError("session has already ended"))
		})

	})

	It("finds leaks without false positives", func() {
		leakyPath, err := gexec.Build("./test/leaky")
		Expect(err).NotTo(HaveOccurred())

		cmd := exec.Command(leakyPath)
		in, err := cmd.StdinPipe()
		Expect(err).NotTo(HaveOccurred())
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		defer session.Terminate()

		sessionFds := func() ([]filedesc.FileDescriptor, error) {
			return FiledescriptorsFor(session)
		}

		By("getting initial reference")
		Eventually(session.Out).Should(gbytes.Say("READY"))

		goodfds, err := sessionFds()
		Expect(err).NotTo(HaveOccurred())
		Expect(goodfds[0]).NotTo(Equal(fdooze.Filedescriptors()[0]), "malfunction: got fds of myself")

		By("triggering a leak")
		_, _ = in.Write([]byte("\n"))
		Eventually(session.Out).Should(gbytes.Say("LEAK"))
		Eventually(sessionFds).Should(ContainElement(HaveFdWithPath(HaveSuffix("test/leaky/main.go"))))
		Eventually(sessionFds).Should(fdooze.HaveLeakedFds(goodfds), "should have leaked")

		By("plumbing the leak")
		_, _ = in.Write([]byte("\n"))
		Eventually(session.Out).Should(gbytes.Say("PLUMBED"))
		Eventually(sessionFds).ShouldNot(ContainElement(HaveFdWithPath(HaveSuffix("test/leaky/main.go"))))
		Eventually(sessionFds).ShouldNot(fdooze.HaveLeakedFds(goodfds), "leak should be gone")

		_, _ = in.Write([]byte("\n"))
		Eventually(session).Should(gexec.Exit())
	})

})
