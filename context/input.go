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
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cooleo/goweb/session"
)

// Regexes for checking the accept headers
// TODO make sure these are correct
var (
	acceptsHTMLRegex = regexp.MustCompile(`(text/html|application/xhtml\+xml)(?:,|$)`)
	acceptsXMLRegex  = regexp.MustCompile(`(application/xml|text/xml)(?:,|$)`)
	acceptsJSONRegex = regexp.MustCompile(`(application/json)(?:,|$)`)
	maxParam         = 50
)

// gowebInput operates the http request header, data, cookie and body.
// it also contains router params and current session.
type gowebInput struct {
	Context     *Context
	CruSession  session.Store
	pnames      []string
	pvalues     []string
	data        map[interface{}]interface{} // store some values in this context when calling context in filter or controller.
	RequestBody []byte
}

// NewInput return gowebInput generated by Context.
func NewInput() *gowebInput {
	return &gowebInput{
		pnames:  make([]string, 0, maxParam),
		pvalues: make([]string, 0, maxParam),
		data:    make(map[interface{}]interface{}),
	}
}

// Reset init the gowebInput
func (input *gowebInput) Reset(ctx *Context) {
	input.Context = ctx
	input.CruSession = nil
	input.pnames = input.pnames[:0]
	input.pvalues = input.pvalues[:0]
	input.data = nil
	input.RequestBody = []byte{}
}

// Protocol returns request protocol name, such as HTTP/1.1 .
func (input *gowebInput) Protocol() string {
	return input.Context.Request.Proto
}

// URI returns full request url with query string, fragment.
func (input *gowebInput) URI() string {
	return input.Context.Request.RequestURI
}

// URL returns request url path (without query string, fragment).
func (input *gowebInput) URL() string {
	return input.Context.Request.URL.Path
}

// Site returns base site url as scheme://domain type.
func (input *gowebInput) Site() string {
	return input.Scheme() + "://" + input.Domain()
}

// Scheme returns request scheme as "http" or "https".
func (input *gowebInput) Scheme() string {
	if input.Context.Request.URL.Scheme != "" {
		return input.Context.Request.URL.Scheme
	}
	if input.Context.Request.TLS == nil {
		return "http"
	}
	return "https"
}

// Domain returns host name.
// Alias of Host method.
func (input *gowebInput) Domain() string {
	return input.Host()
}

// Host returns host name.
// if no host info in request, return localhost.
func (input *gowebInput) Host() string {
	if input.Context.Request.Host != "" {
		hostParts := strings.Split(input.Context.Request.Host, ":")
		if len(hostParts) > 0 {
			return hostParts[0]
		}
		return input.Context.Request.Host
	}
	return "localhost"
}

// Method returns http request method.
func (input *gowebInput) Method() string {
	return input.Context.Request.Method
}

// Is returns boolean of this request is on given method, such as Is("POST").
func (input *gowebInput) Is(method string) bool {
	return input.Method() == method
}

// IsGet Is this a GET method request?
func (input *gowebInput) IsGet() bool {
	return input.Is("GET")
}

// IsPost Is this a POST method request?
func (input *gowebInput) IsPost() bool {
	return input.Is("POST")
}

// IsHead Is this a Head method request?
func (input *gowebInput) IsHead() bool {
	return input.Is("HEAD")
}

// IsOptions Is this a OPTIONS method request?
func (input *gowebInput) IsOptions() bool {
	return input.Is("OPTIONS")
}

// IsPut Is this a PUT method request?
func (input *gowebInput) IsPut() bool {
	return input.Is("PUT")
}

// IsDelete Is this a DELETE method request?
func (input *gowebInput) IsDelete() bool {
	return input.Is("DELETE")
}

// IsPatch Is this a PATCH method request?
func (input *gowebInput) IsPatch() bool {
	return input.Is("PATCH")
}

// IsAjax returns boolean of this request is generated by ajax.
func (input *gowebInput) IsAjax() bool {
	return input.Header("X-Requested-With") == "XMLHttpRequest"
}

// IsSecure returns boolean of this request is in https.
func (input *gowebInput) IsSecure() bool {
	return input.Scheme() == "https"
}

// IsWebsocket returns boolean of this request is in webSocket.
func (input *gowebInput) IsWebsocket() bool {
	return input.Header("Upgrade") == "websocket"
}

// IsUpload returns boolean of whether file uploads in this request or not..
func (input *gowebInput) IsUpload() bool {
	return strings.Contains(input.Header("Content-Type"), "multipart/form-data")
}

// AcceptsHTML Checks if request accepts html response
func (input *gowebInput) AcceptsHTML() bool {
	return acceptsHTMLRegex.MatchString(input.Header("Accept"))
}

// AcceptsXML Checks if request accepts xml response
func (input *gowebInput) AcceptsXML() bool {
	return acceptsXMLRegex.MatchString(input.Header("Accept"))
}

// AcceptsJSON Checks if request accepts json response
func (input *gowebInput) AcceptsJSON() bool {
	return acceptsJSONRegex.MatchString(input.Header("Accept"))
}

// IP returns request client ip.
// if in proxy, return first proxy id.
// if error, return 127.0.0.1.
func (input *gowebInput) IP() string {
	ips := input.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		rip := strings.Split(ips[0], ":")
		return rip[0]
	}
	ip := strings.Split(input.Context.Request.RemoteAddr, ":")
	if len(ip) > 0 {
		if ip[0] != "[" {
			return ip[0]
		}
	}
	return "127.0.0.1"
}

// Proxy returns proxy client ips slice.
func (input *gowebInput) Proxy() []string {
	if ips := input.Header("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

// Referer returns http referer header.
func (input *gowebInput) Referer() string {
	return input.Header("Referer")
}

// Refer returns http referer header.
func (input *gowebInput) Refer() string {
	return input.Referer()
}

// SubDomains returns sub domain string.
// if aa.bb.domain.com, returns aa.bb .
func (input *gowebInput) SubDomains() string {
	parts := strings.Split(input.Host(), ".")
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-2], ".")
	}
	return ""
}

// Port returns request client port.
// when error or empty, return 80.
func (input *gowebInput) Port() int {
	parts := strings.Split(input.Context.Request.Host, ":")
	if len(parts) == 2 {
		port, _ := strconv.Atoi(parts[1])
		return port
	}
	return 80
}

// UserAgent returns request client user agent string.
func (input *gowebInput) UserAgent() string {
	return input.Header("User-Agent")
}

// ParamsLen return the length of the params
func (input *gowebInput) ParamsLen() int {
	return len(input.pnames)
}

// Param returns router param by a given key.
func (input *gowebInput) Param(key string) string {
	for i, v := range input.pnames {
		if v == key && i <= len(input.pvalues) {
			return input.pvalues[i]
		}
	}
	return ""
}

// Params returns the map[key]value.
func (input *gowebInput) Params() map[string]string {
	m := make(map[string]string)
	for i, v := range input.pnames {
		if i <= len(input.pvalues) {
			m[v] = input.pvalues[i]
		}
	}
	return m
}

// SetParam will set the param with key and value
func (input *gowebInput) SetParam(key, val string) {
	// check if already exists
	for i, v := range input.pnames {
		if v == key && i <= len(input.pvalues) {
			input.pvalues[i] = val
			return
		}
	}
	input.pvalues = append(input.pvalues, val)
	input.pnames = append(input.pnames, key)
}

// Query returns input data item string by a given string.
func (input *gowebInput) Query(key string) string {
	if val := input.Param(key); val != "" {
		return val
	}
	if input.Context.Request.Form == nil {
		input.Context.Request.ParseForm()
	}
	return input.Context.Request.Form.Get(key)
}

// Header returns request header item string by a given string.
// if non-existed, return empty string.
func (input *gowebInput) Header(key string) string {
	return input.Context.Request.Header.Get(key)
}

// Cookie returns request cookie item string by a given key.
// if non-existed, return empty string.
func (input *gowebInput) Cookie(key string) string {
	ck, err := input.Context.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return ck.Value
}

// Session returns current session item value by a given key.
// if non-existed, return empty string.
func (input *gowebInput) Session(key interface{}) interface{} {
	return input.CruSession.Get(key)
}

// CopyBody returns the raw request body data as bytes.
func (input *gowebInput) CopyBody(MaxMemory int64) []byte {
	safe := &io.LimitedReader{R: input.Context.Request.Body, N: MaxMemory}
	requestbody, _ := ioutil.ReadAll(safe)
	input.Context.Request.Body.Close()
	bf := bytes.NewBuffer(requestbody)
	input.Context.Request.Body = ioutil.NopCloser(bf)
	input.RequestBody = requestbody
	return requestbody
}

// Data return the implicit data in the input
func (input *gowebInput) Data() map[interface{}]interface{} {
	if input.data == nil {
		input.data = make(map[interface{}]interface{})
	}
	return input.data
}

// GetData returns the stored data in this context.
func (input *gowebInput) GetData(key interface{}) interface{} {
	if v, ok := input.data[key]; ok {
		return v
	}
	return nil
}

// SetData stores data with given key in this context.
// This data are only available in this context.
func (input *gowebInput) SetData(key, val interface{}) {
	if input.data == nil {
		input.data = make(map[interface{}]interface{})
	}
	input.data[key] = val
}

// ParseFormOrMulitForm parseForm or parseMultiForm based on Content-type
func (input *gowebInput) ParseFormOrMulitForm(maxMemory int64) error {
	// Parse the body depending on the content type.
	if strings.Contains(input.Header("Content-Type"), "multipart/form-data") {
		if err := input.Context.Request.ParseMultipartForm(maxMemory); err != nil {
			return errors.New("Error parsing request body:" + err.Error())
		}
	} else if err := input.Context.Request.ParseForm(); err != nil {
		return errors.New("Error parsing request body:" + err.Error())
	}
	return nil
}

// Bind data from request.Form[key] to dest
// like /?id=123&isok=true&ft=1.2&ol[0]=1&ol[1]=2&ul[]=str&ul[]=array&user.Name=cooleo
// var id int  gowebInput.Bind(&id, "id")  id ==123
// var isok bool  gowebInput.Bind(&isok, "isok")  isok ==true
// var ft float64  gowebInput.Bind(&ft, "ft")  ft ==1.2
// ol := make([]int, 0, 2)  gowebInput.Bind(&ol, "ol")  ol ==[1 2]
// ul := make([]string, 0, 2)  gowebInput.Bind(&ul, "ul")  ul ==[str array]
// user struct{Name}  gowebInput.Bind(&user, "user")  user == {Name:"cooleo"}
func (input *gowebInput) Bind(dest interface{}, key string) error {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return errors.New("goweb: non-pointer passed to Bind: " + key)
	}
	value = value.Elem()
	if !value.CanSet() {
		return errors.New("goweb: non-settable variable passed to Bind: " + key)
	}
	rv := input.bind(key, value.Type())
	if !rv.IsValid() {
		return errors.New("goweb: reflect value is empty")
	}
	value.Set(rv)
	return nil
}

func (input *gowebInput) bind(key string, typ reflect.Type) reflect.Value {
	rv := reflect.Zero(typ)
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindInt(val, typ)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindUint(val, typ)
	case reflect.Float32, reflect.Float64:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindFloat(val, typ)
	case reflect.String:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindString(val, typ)
	case reflect.Bool:
		val := input.Query(key)
		if len(val) == 0 {
			return rv
		}
		rv = input.bindBool(val, typ)
	case reflect.Slice:
		rv = input.bindSlice(&input.Context.Request.Form, key, typ)
	case reflect.Struct:
		rv = input.bindStruct(&input.Context.Request.Form, key, typ)
	case reflect.Ptr:
		rv = input.bindPoint(key, typ)
	case reflect.Map:
		rv = input.bindMap(&input.Context.Request.Form, key, typ)
	}
	return rv
}

func (input *gowebInput) bindValue(val string, typ reflect.Type) reflect.Value {
	rv := reflect.Zero(typ)
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv = input.bindInt(val, typ)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv = input.bindUint(val, typ)
	case reflect.Float32, reflect.Float64:
		rv = input.bindFloat(val, typ)
	case reflect.String:
		rv = input.bindString(val, typ)
	case reflect.Bool:
		rv = input.bindBool(val, typ)
	case reflect.Slice:
		rv = input.bindSlice(&url.Values{"": {val}}, "", typ)
	case reflect.Struct:
		rv = input.bindStruct(&url.Values{"": {val}}, "", typ)
	case reflect.Ptr:
		rv = input.bindPoint(val, typ)
	case reflect.Map:
		rv = input.bindMap(&url.Values{"": {val}}, "", typ)
	}
	return rv
}

func (input *gowebInput) bindInt(val string, typ reflect.Type) reflect.Value {
	intValue, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetInt(intValue)
	return pValue.Elem()
}

func (input *gowebInput) bindUint(val string, typ reflect.Type) reflect.Value {
	uintValue, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetUint(uintValue)
	return pValue.Elem()
}

func (input *gowebInput) bindFloat(val string, typ reflect.Type) reflect.Value {
	floatValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return reflect.Zero(typ)
	}
	pValue := reflect.New(typ)
	pValue.Elem().SetFloat(floatValue)
	return pValue.Elem()
}

func (input *gowebInput) bindString(val string, typ reflect.Type) reflect.Value {
	return reflect.ValueOf(val)
}

func (input *gowebInput) bindBool(val string, typ reflect.Type) reflect.Value {
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "true", "on", "1":
		return reflect.ValueOf(true)
	}
	return reflect.ValueOf(false)
}

type sliceValue struct {
	index int           // Index extracted from brackets.  If -1, no index was provided.
	value reflect.Value // the bound value for this slice element.
}

func (input *gowebInput) bindSlice(params *url.Values, key string, typ reflect.Type) reflect.Value {
	maxIndex := -1
	numNoIndex := 0
	sliceValues := []sliceValue{}
	for reqKey, vals := range *params {
		if !strings.HasPrefix(reqKey, key+"[") {
			continue
		}
		// Extract the index, and the index where a sub-key starts. (e.g. field[0].subkey)
		index := -1
		leftBracket, rightBracket := len(key), strings.Index(reqKey[len(key):], "]")+len(key)
		if rightBracket > leftBracket+1 {
			index, _ = strconv.Atoi(reqKey[leftBracket+1 : rightBracket])
		}
		subKeyIndex := rightBracket + 1

		// Handle the indexed case.
		if index > -1 {
			if index > maxIndex {
				maxIndex = index
			}
			sliceValues = append(sliceValues, sliceValue{
				index: index,
				value: input.bind(reqKey[:subKeyIndex], typ.Elem()),
			})
			continue
		}

		// It's an un-indexed element.  (e.g. element[])
		numNoIndex += len(vals)
		for _, val := range vals {
			// Unindexed values can only be direct-bound.
			sliceValues = append(sliceValues, sliceValue{
				index: -1,
				value: input.bindValue(val, typ.Elem()),
			})
		}
	}
	resultArray := reflect.MakeSlice(typ, maxIndex+1, maxIndex+1+numNoIndex)
	for _, sv := range sliceValues {
		if sv.index != -1 {
			resultArray.Index(sv.index).Set(sv.value)
		} else {
			resultArray = reflect.Append(resultArray, sv.value)
		}
	}
	return resultArray
}

func (input *gowebInput) bindStruct(params *url.Values, key string, typ reflect.Type) reflect.Value {
	result := reflect.New(typ).Elem()
	fieldValues := make(map[string]reflect.Value)
	for reqKey, val := range *params {
		if !strings.HasPrefix(reqKey, key+".") {
			continue
		}

		fieldName := reqKey[len(key)+1:]

		if _, ok := fieldValues[fieldName]; !ok {
			// Time to bind this field.  Get it and make sure we can set it.
			fieldValue := result.FieldByName(fieldName)
			if !fieldValue.IsValid() {
				continue
			}
			if !fieldValue.CanSet() {
				continue
			}
			boundVal := input.bindValue(val[0], fieldValue.Type())
			fieldValue.Set(boundVal)
			fieldValues[fieldName] = boundVal
		}
	}

	return result
}

func (input *gowebInput) bindPoint(key string, typ reflect.Type) reflect.Value {
	return input.bind(key, typ.Elem()).Addr()
}

func (input *gowebInput) bindMap(params *url.Values, key string, typ reflect.Type) reflect.Value {
	var (
		result    = reflect.MakeMap(typ)
		keyType   = typ.Key()
		valueType = typ.Elem()
	)
	for paramName, values := range *params {
		if !strings.HasPrefix(paramName, key+"[") || paramName[len(paramName)-1] != ']' {
			continue
		}

		key := paramName[len(key)+1 : len(paramName)-1]
		result.SetMapIndex(input.bindValue(key, keyType), input.bindValue(values[0], valueType))
	}
	return result
}
