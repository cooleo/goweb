// Copyright 2016 goweb Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goweb

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// VERSION represent goweb web framework version.
	VERSION = "1.6.1"

	// DEV is for develop
	DEV = "dev"
	// PROD is for production
	PROD = "prod"
)

//hook function to run
type hookfunc func() error

var (
	hooks = make([]hookfunc, 0) //hook function slice to store the hookfunc
)

// AddAPPStartHook is used to register the hookfunc
// The hookfuncs will run in goweb.Run()
// such as sessionInit, middlerware start, buildtemplate, admin start
func AddAPPStartHook(hf hookfunc) {
	hooks = append(hooks, hf)
}

// Run goweb application.
// goweb.Run() default run on HttpPort
// goweb.Run("localhost")
// goweb.Run(":8089")
// goweb.Run("127.0.0.1:8089")
func Run(params ...string) {
	initBeforeHTTPRun()

	if len(params) > 0 && params[0] != "" {
		strs := strings.Split(params[0], ":")
		if len(strs) > 0 && strs[0] != "" {
			BConfig.Listen.HTTPAddr = strs[0]
		}
		if len(strs) > 1 && strs[1] != "" {
			BConfig.Listen.HTTPPort, _ = strconv.Atoi(strs[1])
		}
	}

	BeeApp.Run()
}

func initBeforeHTTPRun() {
	//init hooks
	AddAPPStartHook(registerMime)
	AddAPPStartHook(registerDefaultErrorHandler)
	AddAPPStartHook(registerSession)
	AddAPPStartHook(registerDocs)
	AddAPPStartHook(registerTemplate)
	AddAPPStartHook(registerAdmin)

	for _, hk := range hooks {
		if err := hk(); err != nil {
			panic(err)
		}
	}
}

// TestgowebInit is for test package init
func TestgowebInit(ap string) {
	os.Setenv("goweb_RUNMODE", "test")
	appConfigPath = filepath.Join(ap, "conf", "app.conf")
	os.Chdir(ap)
	initBeforeHTTPRun()
}
