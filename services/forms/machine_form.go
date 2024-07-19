package forms

import (
	"net/http"

	"code.gitea.io/gitea/modules/web/middleware"
	"code.gitea.io/gitea/services/context"
	"gitea.com/go-chi/binding"
)

type SettingMachineForm struct {
	Name   string
	User   string
	Host   string
	Port   int32
	SshKey int64
}

func (f *SettingMachineForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}
