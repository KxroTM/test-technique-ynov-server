package handlers

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// ConfigureValidation aligne les noms de champs des erreurs de validation sur les noms utilisés dans le JSON
func ConfigureValidation() {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonName := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if jsonName == "-" {
			return field.Name
		}
		return jsonName
	})
}
