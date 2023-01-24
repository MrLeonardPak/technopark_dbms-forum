package response

import (
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
)

type Data interface {
	MarshalJSON() ([]byte, error)
}

const (
	errMsg = "send data error"
)

func Send[T Data](code int, payload T, ctx *fasthttp.RequestCtx) {

	ctx.SetStatusCode(code)

	if code == http.StatusInternalServerError {
		fmt.Println("500 status code")
		ctx.SetBody([]byte(errMsg))
		return
	}
	data, err := payload.MarshalJSON()
	if err != nil {
		fmt.Println("Send: ", err)
		ctx.SetBody([]byte(errMsg))
		return
	}
	ctx.SetBody(data)
}
