package appender

import (
	"github.com/mayadata-io/kiali/config"
	"github.com/mayadata-io/kiali/prometheus"
	"github.com/mayadata-io/kiali/prometheus/prometheustest"
	"github.com/prometheus/common/model"

	"github.com/stretchr/testify/mock"
)

// package-private test util functions (used by multiple test files)

// Setup mocks

func setupMocked() (*prometheus.Client, *prometheustest.PromAPIMock, error) {
	config.Set(config.NewConfig())
	api := new(prometheustest.PromAPIMock)
	client, err := prometheus.NewClient()
	if err != nil {
		return nil, nil, err
	}
	client.Inject(api)
	return client, api, nil
}

func mockQuery(api *prometheustest.PromAPIMock, query string, ret *model.Vector) {
	api.On(
		"Query",
		mock.AnythingOfType("*context.emptyCtx"),
		query,
		mock.AnythingOfType("time.Time"),
	).Return(*ret, nil)
	api.On(
		"Query",
		mock.AnythingOfType("*context.cancelCtx"),
		query,
		mock.AnythingOfType("time.Time"),
	).Return(*ret, nil)
}
