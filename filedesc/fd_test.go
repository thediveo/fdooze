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

package filedesc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing/iotest"
	"time"

	"golang.org/x/sys/unix"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/fdooze/filedesc/test/canary/cage"
	. "github.com/thediveo/success"
)

var _ = Describe("file descriptors", func() {

	const procFdBase = "/proc/self/fd"

	When("dealing with a single file descriptor", func() {

		It("returns error when reading errors", func() {
			Expect(fdFromReader(42, iotest.ErrReader(errors.New("foobar")))).Error().To(
				MatchError("foobar"))
		})

		It("returns error when reading incomplete information", func() {
			r := strings.NewReader("pos:\t0\nflags:\t042\n")
			Expect(fdFromReader(42, r)).Error().To(
				MatchError(ContainSubstring("incomplete fdinfo data")))
		})

		It("returns error when reading out-of-range information", func() {
			r := strings.NewReader(fmt.Sprintf(
				"pos:\t0\nflags:\t%o\nmnt_id:\t123\n", uint64(math.MaxInt)+1))
			Expect(fdFromReader(42, r)).Error().To(
				MatchError(ContainSubstring("flags outside range:")))
			r = strings.NewReader("pos:\t0\nflags:\t042\nmnt_id:\t-1\n")
			Expect(fdFromReader(42, r)).Error().To(
				MatchError(ContainSubstring("mnt_id outside range:")))
		})

		It("reports invalid base", func() {
			Expect(newWithBase(-1, "/foobar")).Error().To(HaveOccurred())
		})

		It("reads and returns common fd information", func() {
			r := strings.NewReader("pos:\t0\nflags:\t042\nmnt_id:\t123\n")
			fdesc := Successful(fdFromReader(42, r))
			Expect(fdesc.FdNo()).To(Equal(42))
			Expect(fdesc.Flags()).To(Equal(Flags(042)))
			Expect(fdesc.MountId()).To(Equal(123))
		})

		It("returns a correct description", func() {
			fdesc := filedesc{
				fdNo:  42,
				flags: Flags(os.O_APPEND),
				mntId: 123,
			}
			Expect(fdesc.Description(0)).To(Equal(
				fmt.Sprintf("fd 42, flags 0x%x (O_RDONLY,O_APPEND)", os.O_APPEND)))
		})

		It("doesn't fail to read information about fd 0", func() {
			fdesc := Successful(newFiledesc(0, procFdBase))
			Expect(fdesc.fdNo).To(Equal(0))
			Expect(fdesc.mntId).NotTo(BeZero())
		})

		It("fails correctly to read invalid fd information", func() {
			r := strings.NewReader("pos:\t0\nflags:\t099\nmnt_id:\t123\n")
			Expect(fdFromReader(0, r)).Error().To(MatchError(MatchRegexp("invalid syntax")))

			r = strings.NewReader("pos:\t0\nflags:\t042\nmnt_id:\tabc\n")
			Expect(fdFromReader(0, r)).Error().To(MatchError(MatchRegexp("invalid syntax")))
		})

		It("fails correctly to read from fd -1", func() {
			Expect(newFiledesc(-1, procFdBase)).Error().To(MatchError(MatchRegexp("open.*/proc/self/fdinfo/-1")))
		})

	})

	When("discovering fds from our own process", Serial, func() {

		It("returns error or nothing for missing or invalid procfs", func() {
			Expect(filedescriptors("./test/missing-proc/fd")).Error().To(HaveOccurred())
			Expect(filedescriptors("./test/not-an-fd-directory")).Error().To(HaveOccurred())
			Expect(filedescriptors("./test/fake-proc/fd")).To(BeEmpty())
		})

		It("finds this process's file descriptors", func() {
			fd := Successful(unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0))
			defer unix.Close(fd)

			fdescs := Filedescriptors()
			Expect(fdescs).NotTo(BeEmpty())
			Expect(fdescs).To(ContainElement(HaveField("FdNo()", fd)))
		})

		It("doesn't include its own fd directory fd", func() {
			const dirPath = "/proc/self/fd"

			fdentries := Successful(os.ReadDir(dirPath))
			fdNoDict := map[int]struct{}{}
			for _, fdentry := range fdentries {
				fdno, err := strconv.Atoi(fdentry.Name())
				if err != nil {
					continue
				}
				// ReadDir has closed the directory fd by now, so we won't be
				// able to stat it and thus automatically exclude it :)
				if _, err := os.Readlink(dirPath + "/" + fdentry.Name()); err != nil {
					continue
				}
				fdNoDict[fdno] = struct{}{}
			}
			Expect(len(fdNoDict)).To(BeNumerically(">=", 3))
			fds := Successful(filedescriptors(dirPath))
			Expect(len(fds)).To(BeNumerically(">=", 3))
			Expect(fds).To(HaveLen(len(fdNoDict)))
			Expect(fds).To(HaveEach(
				HaveField("FdNo()", BeKeyOf(fdNoDict))))
		})

	})

	It("discovers fds from another process", func() {
		canaryPath := Successful(
			gexec.Build("github.com/thediveo/fdooze/filedesc/test/canary"))
		DeferCleanup(gexec.CleanupBuildArtifacts)
		canaryCmd := exec.Command(canaryPath)
		session := Successful(
			gexec.Start(canaryCmd, GinkgoWriter, GinkgoWriter))
		defer session.Terminate()

		Eventually(func() []FileDescriptor {
			fds, _ := ProcessFiledescriptors(session.Command.Process.Pid)
			return fds
		}).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).Should(
			ContainElement(SatisfyAll(
				BeAssignableToTypeOf(&SocketFd{}),
				HaveField("Type()", unix.SOCK_DGRAM),
				HaveField("Peer()", fmt.Sprintf("%s:%d", cage.IP, cage.Port)),
			)))
	})

})
