## goweb



goweb is used for rapid development of RESTful APIs, web apps and backend services in Go.
It is inspired by Tornado, Sinatra and Flask. goweb has some Go-specific features such as interfaces and struct embedding.


##Quick Start
######Download and install

    go get github.com/cooleo/goweb

######Create file `hello.go`
```go
package main

import "github.com/cooleo/goweb"

func main(){
    goweb.Run()
}
```
######Build and run
```bash
    go build hello.go
    ./hello
```
######Congratulations! 
You just built your first goweb app.
Open your browser and visit `http://localhost:8000`.

## Features

* RESTful support
* MVC architecture
* Modularity
* Auto API documents
* Annotation router
* Namespace
* Powerful development tools
* Full stack for Web & API


## LICENSE

goweb source code is licensed under the Apache Licence, Version 2.0
(http://www.apache.org/licenses/LICENSE-2.0.html).
