package dynamic

import (
	"fmt"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric definitions
// Ensure that this follows best practices for naming: https://prometheus.io/docs/practices/naming/
var (
	metricNamePrefix      = "d4l_dynamic"
	MetricBootstrapStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamePrefix,
		Name:      "bootstrap_status",
		Help:      "A gauge with value 1 if bootstrapping was successful and 0 otherwise",
	},
		[]string{"vcName", "svcName"},
	)
	MetricConfigHotReloads = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamePrefix,
		Name:      "config_hot_reloads_total",
		Help:      "Number of viper-config hot-reloads",
	},
		[]string{"vcName", "svcName", "secret", "result"},
	)
	MetricSecretsLoaded = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricNamePrefix,
		Name:      "secrets_loaded",
		Help:      "Number of secrets loaded in the last viper-config hot-reload",
	},
		[]string{"vcName", "svcName", "secret"},
	)
)

// AddDynamicPkgMetrics adds a static metric with the build information
func AddDynamicPkgMetrics(svcName, vcFile string, autoBootstrap bool) {
	err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: metricNamePrefix,
			Name:      "dynamic_pkg_usage_info",
			Help:      "A metric with a constant '1' to show that the service uses go-svc/pkg/dynamic",
			ConstLabels: prometheus.Labels{
				"service":       svcName,
				"autoBootstrap": fmt.Sprintf("%t", autoBootstrap),
				"configFile":    vcFile,
			},
		},
		func() float64 { return 1 },
	))
	if err != nil {
		logging.LogErrorf(err, "Error registering dynamic pkg metric")
	}
}
