package tracing

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
)

type umbRecorderLogger struct {
	logger log.Logger
}

func (l *umbRecorderLogger) Infof(format string, args ...interface{}) {
	level.Info(l.logger).Log("msg", fmt.Sprintf(format, args...))
}

func (l *umbRecorderLogger) Errorf(format string, args ...interface{}) {
	level.Error(l.logger).Log("msg", fmt.Sprintf(format, args...))
}

// NewOptionalGCloudTracer returns GoogleCloudTracer Tracer. In case of error it log warning and returns noop tracer.
func NewOptionalUmbTracer(ctx context.Context, logger log.Logger, uMonibenchTraceProjectID string, sampleFactor uint64, debugName string) (opentracing.Tracer, func() error) {

	if uMonibenchTraceProjectID == "" {
		return &opentracing.NoopTracer{}, func() error { return nil }
	}

	tracer, closeFn, err := newUmbTracer(ctx, logger, uMonibenchTraceProjectID, sampleFactor, debugName)
	if err != nil {
		level.Warn(logger).Log("msg", "failed to init Google Cloud Tracer. Tracing will be disabled", "err", err)
		return &opentracing.NoopTracer{}, func() error { return nil }
	}

	return tracer, closeFn
}

func newUmbTracer(ctx context.Context, logger log.Logger, uMonibenchTraceProjectID string, sampleFactor uint64, debugName string) (opentracing.Tracer, func() error, error) {

	m := &basictracer.InMemorySpanRecorder{}
	r := &forceRecorder{wrapped: m}

	shouldSample := func(traceID uint64) bool {
		// Set the sampling rate.
		return traceID%sampleFactor == 0
	}
	if sampleFactor < 1 {
		level.Debug(logger).Log("msg", "Tracing is enabled, but sampling is 0 which means only spans with 'force tracing' baggage will enable tracing.")
		shouldSample = func(_ uint64) bool {
			return false
		}
	}
	return &tracer{
		debugName: debugName,
		wrapped: basictracer.NewWithOptions(basictracer.Options{
			ShouldSample:   shouldSample,
			Recorder:       &forceRecorder{wrapped: r},
			MaxLogsPerSpan: 100,
		}),
	}, nil, nil
}
