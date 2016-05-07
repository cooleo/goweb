# Captcha

an example for use captcha

```
package controllers

import (
	"github.com/cooleo/goweb"
	"github.com/cooleo/goweb/cache"
	"github.com/cooleo/goweb/utils/captcha"
)

var cpt *captcha.Captcha

func init() {
	// use goweb cache system store the captcha data
	store := cache.NewMemoryCache()
	cpt = captcha.NewWithFilter("/captcha/", store)
}

type MainController struct {
	goweb.Controller
}

func (this *MainController) Get() {
	this.TplName = "index.tpl"
}

func (this *MainController) Post() {
	this.TplName = "index.tpl"

	this.Data["Success"] = cpt.VerifyReq(this.Ctx.Request)
}
```

template usage

```
{{.Success}}
<form action="/" method="post">
	{{create_captcha}}
	<input name="captcha" type="text">
</form>
```