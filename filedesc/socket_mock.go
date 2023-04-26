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

import "golang.org/x/sys/unix"

// To 100+% coverage and beyond...!!!

// So, who is mocking whom?

var getsockoptInt func(int, int, int) (int, error) = unix.GetsockoptInt
var getsockname func(int) (unix.Sockaddr, error) = unix.Getsockname
var getpeername func(int) (unix.Sockaddr, error) = unix.Getpeername
