package zapgcp

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

type testClock struct {
	idx int64
}

func (t *testClock) Now() time.Time {
	t.idx++
	return time.Unix(t.idx, 0)
}

func (t *testClock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

func TestProductionConfig(t *testing.T) {
	var writer zaptest.Buffer
	cfg := &Config{
		Local: false,
	}

	zCfg, opts := cfg.ToZapConfig()

	opts = append(opts, zap.WithClock(&testClock{}))

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zCfg.EncoderConfig),
		&writer,
		zapcore.DebugLevel,
	), opts...)

	logger.Debug("testing debug", zap.Ints("the_data", []int{1, 2, 3}))
	logger.Info("testing info", zap.String("some_field", "abc123"), zap.Int("another_field", 321))
	logger.Warn("testing warn", zap.Bool("real_bad", false), zap.Error(errors.New("a minor inconvenience")))
	logger.Error("testing error", zap.Float64("boiler_temp", 98734.32), zap.Error(errors.New("a real concern")))

	got := writer.Lines()
	want := []string{
		`{"severity":"DEBUG","time":"1969-12-31T16:00:01-08:00","message":"testing debug","the_data":[1,2,3]}`,
		`{"severity":"INFO","time":"1969-12-31T16:00:02-08:00","message":"testing info","some_field":"abc123","another_field":321}`,
		`{"severity":"WARNING","time":"1969-12-31T16:00:03-08:00","message":"testing warn","real_bad":false,"error":"a minor inconvenience"}`,
		`{"severity":"ERROR","time":"1969-12-31T16:00:04-08:00","message":"testing error","boiler_temp":98734.32,"error":"a real concern"}`,
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected log lines (-want +got)\n%s", diff)
	}

	logger.DPanic("testing dpanic", zap.Duration("time_til_implosion", 3*time.Millisecond), zap.Error(errors.New("divided by zero successfully")))

	// DPanic should include a stacktrace, which will be machine-specific. So we just do a string match on the prefix.
	wantPrefix := `{"severity":"CRITICAL","time":"1969-12-31T16:00:05-08:00","message":"testing dpanic","time_til_implosion":3,"error":"divided by zero successfully","stacktrace":"github.com/Silicon-Ally/zapgcp.TestProductionConfig`
	lines := writer.Lines()
	gotLine := lines[len(lines)-1]

	if !strings.HasPrefix(gotLine, wantPrefix) {
		t.Fatalf("DPanic line did not match prefix, got:\n%s", gotLine)
	}
}
