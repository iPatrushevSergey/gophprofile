package prometheus

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

func TestMetrics_RecordUpload(t *testing.T) {
	metrics, err := NewMetrics()
	require.NoError(t, err)

	metrics.RecordUpload(t.Context(), pkgport.MetricStatusSuccess, 150*time.Millisecond)

	metric := findMetric(t, metrics, "avatars_uploads_total")
	label := findLabelPair(t, metric.GetLabel(), "status", pkgport.MetricStatusSuccess)
	assert.Equal(t, 1.0, metric.GetCounter().GetValue())
	require.NotNil(t, label)
}

func findMetric(t *testing.T, metrics *Metrics, name string) *dto.Metric {
	t.Helper()

	families, err := metrics.registry.Gather()
	require.NoError(t, err)

	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		require.NotEmpty(t, family.GetMetric())
		return family.GetMetric()[0]
	}

	t.Fatalf("metric %q not found", name)
	return nil
}

func findLabelPair(t *testing.T, pairs []*dto.LabelPair, key, value string) *dto.LabelPair {
	t.Helper()

	for _, pair := range pairs {
		if pair.GetName() == key && pair.GetValue() == value {
			return pair
		}
	}

	t.Fatalf("label %s=%s not found in %v", key, value, pairs)
	return nil
}
