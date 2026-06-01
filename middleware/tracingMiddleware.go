package middleware

import (
	"net/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func TracingMiddleware(next http.Handler) http.Handler{
	tracer := otel.Tracer("config-service")


	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.client_ip", r.RemoteAddr),
		)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}