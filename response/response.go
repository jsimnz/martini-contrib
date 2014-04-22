package response

import (
	"encoding/xml"
	"errors"
	"net/http"
	"reflect"

	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
)

type wrappedResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

func newWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	wr := &wrappedResponseWriter{ResponseWriter: w}
	return wr
}

func (wr *wrappedResponseWriter) WriteHeader(code int) {
	wr.ResponseWriter.WriteHeader(code)
	wr.statusCode = code
}

type errorResponse struct {
	XMLName xml.Name `json:"-" xml:"error"`
	Error   int      `json:"error" xml:"code"`
	Message string   `json:"message" xml:"message"`
}

func NewEncoder() martini.Handler {
	return func(c martini.Context, w http.ResponseWriter) {
		wrappedWriter := newWrappedResponseWriter(w)
		c.MapTo(wrappedWriter, (*http.ResponseWriter)(nil))

		var rtnHandler martini.ReturnHandler
		rtnHandler = func(ctx martini.Context, vals []reflect.Value) {
			rv := ctx.Get(inject.InterfaceOf((*http.ResponseWriter)(nil)))
			res := rv.Interface().(http.ResponseWriter)
			var responseVal reflect.Value
			if len(vals) > 1 && vals[0].Kind() == reflect.Int {
				res.WriteHeader(int(vals[0].Int()))
				responseVal = vals[1]
			} else if len(vals) > 0 {
				responseVal = vals[0]
			}
			if isNil(responseVal) {
				wrappedRes := res.(*wrappedResponseWriter)
				code := wrappedRes.statusCode
				if code == 0 {
					panic(errors.New("No return code set for error"))
				}
				responseVal = reflect.ValueOf(errorResponse{Error: code, Message: http.StatusText(code)})
			}
			if canDeref(responseVal) {
				responseVal = responseVal.Elem()
			}
			if isByteSlice(responseVal) {
				res.Write(responseVal.Bytes())
			} else if isStruct(responseVal) || isStructSlice(responseVal) {
				encv := ctx.Get(inject.InterfaceOf((*encoder.Encoder)(nil)))
				enc := encv.Interface().(encoder.Encoder)
				res.Write(encoder.Must(enc.Encode(responseVal.Interface())))
			} else {
				res.Write([]byte(responseVal.String()))
			}
		}
		c.Map(rtnHandler)
	}
}

func isByteSlice(val reflect.Value) bool {
	return val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8
}

func isStruct(val reflect.Value) bool {
	return val.Kind() == reflect.Struct
}

func isStructSlice(val reflect.Value) bool {
	return val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Struct
}

func isNil(val reflect.Value) bool {
	return val.Kind() == reflect.Invalid
}

func canDeref(val reflect.Value) bool {
	return val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr
}
