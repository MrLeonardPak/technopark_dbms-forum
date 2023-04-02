package middlewares

import (
	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

const prometheusPath = "/metrics"

var requests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "forum_backend_requests_total",
		Help: "All requests counter",
	},
)

func WrapperRPS(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		requests.Inc()
		handler(ctx)
	}
}

func InitPrometheus(r *router.Router) {
	prometheus.MustRegister(requests)
	r.GET(prometheusPath, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))
}
