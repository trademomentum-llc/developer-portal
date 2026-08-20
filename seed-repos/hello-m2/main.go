package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// acceptanceMarker records the Wave-0 security-gate live CI acceptance trigger.
const acceptanceMarker = "wave0-security-gates-acceptance-2026-08-20"

// predictRuntimeNamespace mirrors the deterministic OpenChoreo data-plane
// namespace rule used by tools/namespace-predictor and the cluster.
func predictRuntimeNamespace(controlPlaneNs, projectName, environmentName string) string {
	const maxLen = 63
	input := fmt.Sprintf("%s-%s-%s", controlPlaneNs, projectName, environmentName)
	hash := sha256.Sum256([]byte(input))
	short := hex.EncodeToString(hash[:])[:8]

	name := fmt.Sprintf("dp-%s-%s-%s-%s", controlPlaneNs, projectName, environmentName, short)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")

	if len(name) > maxLen {
		name = name[:maxLen]
	}
	return name
}

func runtimeNamespace() string {
	if ns := os.Getenv("OPENCHOREO_RUNTIME_NAMESPACE"); ns != "" {
		return ns
	}
	project := os.Getenv("OPENCHOREO_PROJECT")
	component := os.Getenv("OPENCHOREO_COMPONENT")
	environment := os.Getenv("OPENCHOREO_ENVIRONMENT")
	if project != "" && component != "" && environment != "" {
		return predictRuntimeNamespace("default", project, environment)
	}
	return ""
}

func cleanEndpoint(raw string) string {
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err == nil && u.Host != "" {
			return u.Host
		}
	}
	return raw
}

func initTracer(ctx context.Context) (func(context.Context) error, error) {
	rawEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if rawEndpoint == "" {
		log.Println("OTEL_EXPORTER_OTLP_ENDPOINT not set -- running without telemetry")
		return func(context.Context) error { return nil }, nil
	}

	endpoint := cleanEndpoint(rawEndpoint)
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	env := os.Getenv("OPENCHOREO_ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	version := os.Getenv("GIT_SHA")
	if version == "" {
		version = os.Getenv("VERSION")
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceName("hello-m2"),
		attribute.String("deployment.environment", env),
		attribute.String("service.version", version),
		attribute.String("openchoreo.project", os.Getenv("OPENCHOREO_PROJECT")),
		attribute.String("openchoreo.component", os.Getenv("OPENCHOREO_COMPONENT")),
		attribute.String("openchoreo.environment", env),
		attribute.String("openchoreo.runtime_namespace", runtimeNamespace()),
		attribute.String("git.commit.sha", os.Getenv("GIT_SHA")),
		attribute.String("git.repository", "openchoreo/hello-m2"),
	}

	res, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("hello-m2")

	log.Printf("OpenTelemetry initialized, exporting to %s", rawEndpoint)
	return tp.Shutdown, nil
}

func main() {
	ctx := context.Background()
	shutdown, err := initTracer(ctx)
	if err != nil {
		log.Printf("Tracer init failed: %v -- continuing without telemetry", err)
	}
	if shutdown != nil {
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}()
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if tracer != nil {
			ctx, span := tracer.Start(r.Context(), "http.request",
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.target", r.URL.Path),
				),
			)
			defer span.End()
			r = r.WithContext(ctx)
		}

		// nosemgrep: go.lang.security.audit.xss.no-fprintf-to-responsewriter.no-fprintf-to-responsewriter -- demo endpoint prints only server-side env values (secret redacted) as plaintext; no request-controlled input, no HTML context; cluster-internal demo behind OpenChoreo
		fmt.Fprintf(w, "hello from hello-m2 env=%s secret=%s\n",
			os.Getenv("ENVIRONMENT"),
			redact(os.Getenv("EXAMPLE_SECRET")),
		)
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls -- cluster-internal demo service behind OpenChoreo; no external exposure, TLS terminates at the Envoy gateway per M4 networking
	log.Fatal(http.ListenAndServe(addr, nil))
}

func redact(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
