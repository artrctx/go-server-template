// singletons collection
package inst

import (
	v "github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var (
	Validate      *v.Validate     = v.New()
	SchemaDecoder *schema.Decoder = schema.NewDecoder()
)
