package taskranker

import (
	"github.com/mesos/mesos-go/api/v0/examples/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	tRanker, err := New(WithConfigFile("config.yml"))
	assert.NoError(t, err)
	assert.NotNil(t, tRanker)
	assert.Equal(t, "http://localhost:9090", tRanker.PrometheusEndpoint)
	assert.Equal(t, "cpushare_static", tRanker.Strategy)
	assert.Len(t, tRanker.FilterLabels, 2)
}
