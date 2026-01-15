// Distributed tracing support

package monitoring

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is the Tortoise tracer
var Tracer = otel.Tracer("tortoise")

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(serviceName, otlpEndpoint string) (func(), error) {
	var exporter sdktrace.SpanExporter
	var err error

	if otlpEndpoint != "" {
		exporter, err = otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(otlpEndpoint),
			otlptracegrpc.WithInsecure(),
		)
	} else {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tp.Shutdown(ctx)
	}, nil
}

// Span wraps an OpenTelemetry span
type Span struct {
	ctx    context.Context
	span   trace.Span
	name   string
	start  time.Time
	attrs  []attribute.KeyValue
}

func (s *Span) End(opts ...trace.SpanEndOption) {
	s.span.End(opts...)
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, *Span) {
	ctx, span := Tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	return ctx, &Span{
		ctx:   ctx,
		span:  span,
		name:  name,
		start: time.Now(),
		attrs: attrs,
	}
}

// AddEvent adds an event to the span
func (s *Span) AddEvent(name string, attrs ...attribute.KeyValue) {
	s.span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the span
func (s *Span) SetAttributes(attrs ...attribute.KeyValue) {
	s.span.SetAttributes(attrs...)
}

// RecordError records an error on the span
func (s *Span) RecordError(err error) {
	s.span.RecordError(err)
}

// HTTP middleware for tracing
func TracingMiddleware(serviceName string) func(ctx context.Context, method, path string) context.Context {
	return func(ctx context.Context, method, path string) context.Context {
		ctx, span := Tracer.Start(ctx, fmt.Sprintf("%s %s", method, path),
			trace.WithAttributes(
				semconv.HTTPMethod(method),
				semconv.HTTPTarget(path),
				semconv.HTTPRoute(path),
				attribute.String("http.service", serviceName),
			),
		)
		return ctx
	}
}

// Traceable function wrapper
type TraceableFunc[T any] func(ctx context.Context) (T, error)

func Trace[T any](name string, fn TraceableFunc[T], attrs ...attribute.KeyValue) TraceableFunc[T] {
	return func(ctx context.Context) (T, error) {
		ctx, span := Tracer.Start(ctx, name, trace.WithAttributes(attrs...))
		defer span.End()

		result, err := fn(ctx)
		if err != nil {
			span.RecordError(err)
		}

		return result, err
	}
}

// TraceAsync wraps an async function with tracing
func TraceAsync[T any](name string, fn func(context.Context) <-chan T, attrs ...attribute.KeyValue) func(context.Context) <-chan T {
	return func(ctx context.Context) <-chan T {
		ctx, span := Tracer.Start(ctx, name, trace.WithAttributes(attrs...))
		
		result := make(chan T)
		
		go func() {
			defer close(result)
			defer span.End()
			
			for v := range fn(ctx) {
				select {
				case result <- v:
				case <-ctx.Done():
					return
				}
			}
		}()
		
		return result
	}
}
