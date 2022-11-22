package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/karlib/simple_bank/util"
)

// It is custop validator which will be registred in GIN framework
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}

	return false
}