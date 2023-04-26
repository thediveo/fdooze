// Copyright 2023 Harald Albrecht.
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

package main

import (
	"net"

	"github.com/thediveo/fdooze/filedesc/test/canary/cage"
	"golang.org/x/sys/unix"
)

func main() {
	canaryfd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(canaryfd)

	unix.Connect(canaryfd, &unix.SockaddrInet4{
		// https://tip.golang.org/ref/spec#Conversions, sub heading "Conversions
		// from slice to array or array pointer"
		Addr: *(*[4]byte)(net.ParseIP(cage.IP).To4()),
		Port: cage.Port,
	})

	select {}
}
