package middlewares

import (
	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const prometheusPath = "/metrics"

var rpsCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "rps",
		Help: "Request Per Second",
	},
)

func WrapperRPS(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		rpsCounter.Inc()
		handler(ctx)
	}
}

func InitPrometheus(r *router.Router) {
	prometheus.MustRegister(rpsCounter)
	r.GET(prometheusPath, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))
}
