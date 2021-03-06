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

package context

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	r, _ := http.NewRequest("GET", "/?id=123&isok=true&ft=1.2&ol[0]=1&ol[1]=2&ul[]=str&ul[]=array&user.Name=cooleo", nil)
	gowebInput := NewInput()
	gowebInput.Context = NewContext()
	gowebInput.Context.Reset(httptest.NewRecorder(), r)
	gowebInput.ParseFormOrMulitForm(1 << 20)

	var id int
	err := gowebInput.Bind(&id, "id")
	if id != 123 || err != nil {
		t.Fatal("id should has int value")
	}
	fmt.Println(id)

	var isok bool
	err = gowebInput.Bind(&isok, "isok")
	if !isok || err != nil {
		t.Fatal("isok should be true")
	}
	fmt.Println(isok)

	var float float64
	err = gowebInput.Bind(&float, "ft")
	if float != 1.2 || err != nil {
		t.Fatal("float should be equal to 1.2")
	}
	fmt.Println(float)

	ol := make([]int, 0, 2)
	err = gowebInput.Bind(&ol, "ol")
	if len(ol) != 2 || err != nil || ol[0] != 1 || ol[1] != 2 {
		t.Fatal("ol should has two elements")
	}
	fmt.Println(ol)

	ul := make([]string, 0, 2)
	err = gowebInput.Bind(&ul, "ul")
	if len(ul) != 2 || err != nil || ul[0] != "str" || ul[1] != "array" {
		t.Fatal("ul should has two elements")
	}
	fmt.Println(ul)

	type User struct {
		Name string
	}
	user := User{}
	err = gowebInput.Bind(&user, "user")
	if err != nil || user.Name != "cooleo" {
		t.Fatal("user should has name")
	}
	fmt.Println(user)
}

func TestSubDomain(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://www.example.com/?id=123&isok=true&ft=1.2&ol[0]=1&ol[1]=2&ul[]=str&ul[]=array&user.Name=cooleo", nil)
	gowebInput := NewInput()
	gowebInput.Context = NewContext()
	gowebInput.Context.Reset(httptest.NewRecorder(), r)

	subdomain := gowebInput.SubDomains()
	if subdomain != "www" {
		t.Fatal("Subdomain parse error, got" + subdomain)
	}

	r, _ = http.NewRequest("GET", "http://localhost/", nil)
	gowebInput.Context.Request = r
	if gowebInput.SubDomains() != "" {
		t.Fatal("Subdomain parse error, should be empty, got " + gowebInput.SubDomains())
	}

	r, _ = http.NewRequest("GET", "http://aa.bb.example.com/", nil)
	gowebInput.Context.Request = r
	if gowebInput.SubDomains() != "aa.bb" {
		t.Fatal("Subdomain parse error, got " + gowebInput.SubDomains())
	}

	/* TODO Fix this
	r, _ = http.NewRequest("GET", "http://127.0.0.1/", nil)
	gowebInput.Request = r
	if gowebInput.SubDomains() != "" {
		t.Fatal("Subdomain parse error, got " + gowebInput.SubDomains())
	}
	*/

	r, _ = http.NewRequest("GET", "http://example.com/", nil)
	gowebInput.Context.Request = r
	if gowebInput.SubDomains() != "" {
		t.Fatal("Subdomain parse error, got " + gowebInput.SubDomains())
	}

	r, _ = http.NewRequest("GET", "http://aa.bb.cc.dd.example.com/", nil)
	gowebInput.Context.Request = r
	if gowebInput.SubDomains() != "aa.bb.cc.dd" {
		t.Fatal("Subdomain parse error, got " + gowebInput.SubDomains())
	}
}

func TestParams(t *testing.T) {
	inp := NewInput()

	inp.SetParam("p1", "val1_ver1")
	inp.SetParam("p2", "val2_ver1")
	inp.SetParam("p3", "val3_ver1")
	if l := inp.ParamsLen(); l != 3 {
		t.Fatalf("Input.ParamsLen wrong value: %d, expected %d", l, 3)
	}

	if val := inp.Param("p1"); val != "val1_ver1" {
		t.Fatalf("Input.Param wrong value: %s, expected %s", val, "val1_ver1")
	}
	if val := inp.Param("p3"); val != "val3_ver1" {
		t.Fatalf("Input.Param wrong value: %s, expected %s", val, "val3_ver1")
	}
	vals := inp.Params()
	expected := map[string]string{
		"p1": "val1_ver1",
		"p2": "val2_ver1",
		"p3": "val3_ver1",
	}
	if !reflect.DeepEqual(vals, expected) {
		t.Fatalf("Input.Params wrong value: %s, expected %s", vals, expected)
	}

	// overwriting existing params
	inp.SetParam("p1", "val1_ver2")
	inp.SetParam("p2", "val2_ver2")
	expected = map[string]string{
		"p1": "val1_ver2",
		"p2": "val2_ver2",
		"p3": "val3_ver1",
	}
	vals = inp.Params()
	if !reflect.DeepEqual(vals, expected) {
		t.Fatalf("Input.Params wrong value: %s, expected %s", vals, expected)
	}

	if l := inp.ParamsLen(); l != 3 {
		t.Fatalf("Input.ParamsLen wrong value: %d, expected %d", l, 3)
	}

	if val := inp.Param("p1"); val != "val1_ver2" {
		t.Fatalf("Input.Param wrong value: %s, expected %s", val, "val1_ver2")
	}

	if val := inp.Param("p2"); val != "val2_ver2" {
		t.Fatalf("Input.Param wrong value: %s, expected %s", val, "val1_ver2")
	}

}
