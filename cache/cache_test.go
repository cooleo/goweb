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

package cache

import (
	"os"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	bm, err := NewCache("memory", `{"interval":20}`)
	if err != nil {
		t.Error("init err")
	}
	timeoutDuration := 10 * time.Second
	if err = bm.Put("cooleo", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("cooleo") {
		t.Error("check err")
	}

	if v := bm.Get("cooleo"); v.(int) != 1 {
		t.Error("get err")
	}

	time.Sleep(30 * time.Second)

	if bm.IsExist("cooleo") {
		t.Error("check err")
	}

	if err = bm.Put("cooleo", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}

	if err = bm.Incr("cooleo"); err != nil {
		t.Error("Incr Error", err)
	}

	if v := bm.Get("cooleo"); v.(int) != 2 {
		t.Error("get err")
	}

	if err = bm.Decr("cooleo"); err != nil {
		t.Error("Decr Error", err)
	}

	if v := bm.Get("cooleo"); v.(int) != 1 {
		t.Error("get err")
	}
	bm.Delete("cooleo")
	if bm.IsExist("cooleo") {
		t.Error("delete err")
	}

	//test GetMulti
	if err = bm.Put("cooleo", "author", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("cooleo") {
		t.Error("check err")
	}
	if v := bm.Get("cooleo"); v.(string) != "author" {
		t.Error("get err")
	}

	if err = bm.Put("cooleo1", "author1", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("cooleo1") {
		t.Error("check err")
	}

	vv := bm.GetMulti([]string{"cooleo", "cooleo1"})
	if len(vv) != 2 {
		t.Error("GetMulti ERROR")
	}
	if vv[0].(string) != "author" {
		t.Error("GetMulti ERROR")
	}
	if vv[1].(string) != "author1" {
		t.Error("GetMulti ERROR")
	}
}

func TestFileCache(t *testing.T) {
	bm, err := NewCache("file", `{"CachePath":"cache","FileSuffix":".bin","DirectoryLevel":2,"EmbedExpiry":0}`)
	if err != nil {
		t.Error("init err")
	}
	timeoutDuration := 10 * time.Second
	if err = bm.Put("cooleo", 1, timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("cooleo") {
		t.Error("check err")
	}

	if v := bm.Get("cooleo"); v.(int) != 1 {
		t.Error("get err")
	}

	if err = bm.Incr("cooleo"); err != nil {
		t.Error("Incr Error", err)
	}

	if v := bm.Get("cooleo"); v.(int) != 2 {
		t.Error("get err")
	}

	if err = bm.Decr("cooleo"); err != nil {
		t.Error("Decr Error", err)
	}

	if v := bm.Get("cooleo"); v.(int) != 1 {
		t.Error("get err")
	}
	bm.Delete("cooleo")
	if bm.IsExist("cooleo") {
		t.Error("delete err")
	}

	//test string
	if err = bm.Put("cooleo", "author", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("cooleo") {
		t.Error("check err")
	}
	if v := bm.Get("cooleo"); v.(string) != "author" {
		t.Error("get err")
	}

	//test GetMulti
	if err = bm.Put("cooleo1", "author1", timeoutDuration); err != nil {
		t.Error("set Error", err)
	}
	if !bm.IsExist("cooleo1") {
		t.Error("check err")
	}

	vv := bm.GetMulti([]string{"cooleo", "cooleo1"})
	if len(vv) != 2 {
		t.Error("GetMulti ERROR")
	}
	if vv[0].(string) != "author" {
		t.Error("GetMulti ERROR")
	}
	if vv[1].(string) != "author1" {
		t.Error("GetMulti ERROR")
	}

	os.RemoveAll("cache")
}
