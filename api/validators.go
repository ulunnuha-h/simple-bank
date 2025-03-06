package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/ulunnuha-h/simple_bank/util"
)

var currencyValidator validator.Func = func (fl validator.FieldLevel) bool {
	data, ok := fl.Field().Interface().(string)
	if ok {
		return util.IsValidCurrency(data)
	}

	return false
}