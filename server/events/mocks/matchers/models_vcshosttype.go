package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	models "github.com/runatlantis/atlantis/server/events/models"
)

func AnyModelsVCSHostType() models.VCSHostType {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.VCSHostType))(nil)).Elem()))
	var nullValue models.VCSHostType
	return nullValue
}

func EqModelsVCSHostType(value models.VCSHostType) models.VCSHostType {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.VCSHostType
	return nullValue
}
