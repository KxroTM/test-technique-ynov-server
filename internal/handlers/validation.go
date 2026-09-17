package handlers

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// ConfigureValidation aligne les noms de champs des erreurs de validation sur
// les noms utilisés dans le JSON.
//
// Par défaut, le validateur remonte le nom du champ Go (« Email », « SpaceID »).
// Le client, lui, a envoyé « email » et « space_id ». Sans cette correspondance,
// il ne pourrait pas rattacher une erreur au champ de formulaire concerné.
//
// La fonction est appelée explicitement au démarrage plutôt que dans un init(),
// afin que l'ordre d'initialisation reste visible dans main.
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
