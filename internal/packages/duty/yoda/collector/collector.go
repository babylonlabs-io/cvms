package collector

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/processor"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/router"
	"github.com/cosmostation/cvms/internal/packages/duty/yoda/types"
	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ common.CollectorStart = Start
	_ common.CollectorLoop  = loop
)

// Due to the nature of how request misses are tracked and counted,
// the list of all misses collected in the previous iteration is kept in memory
var yodaRequestMisses = make([]types.ValidatorStatus, 0)

const (
	Subsystem      = "yoda"
	SubsystemSleep = 10 * time.Second
	UnHealthSleep  = 10 * time.Second

	YodaStatusMetricName = "status"
	YodaStatusHelp       = "Collects the status of the yoda oracle; 1.0: Oracle is active, 0.0: Oracle is inactive"

	// Collects the maximum number of current misses from all requests the validator
	// has to respond to and which are not yet expired
	YodaMaxMissesName = "max_miss_counter"
	YodaMaxMissesHelp = "Collects the maximum number of misses " +
		"from all requests the validator has currently to respond to " +
		"and which are not yet expired"

	// Collects the total number of current misses from all requests the validator
	YodaValidatorMissSummaryName = "validator_miss_summary"
	YodaValidatorMissSummaryHelp = "Collects the total number of current misses " +
		"from all requests the validator has to respond " +
		"to and which are not yet expired, " +
		"grouped by validator address. Only " +
		"enabled for validators selected in config "

	// Collect the total number of current misses from all requests the validator
	// and calculates responde percentiles in blocks
	YodaMissSummaryName = "miss_summary"
	YodaMissSummaryHelp = "Collects the total number of current " +
		"misses from all requests the validator and " +
		"calculates respond percentiles in blocks"

	// Shows the maximum number of blocks a yoda oracle has time to respond to a request
	YodaRequestSlashWindow     = "slash_window"
	YodaRequestSlashWindowHelp = "Maximum number of blocks a yoda oracle has " +
		"time to respond if a request was assigned to it"

	yodaRequestCountName = "request_count"
	yodaRequestCountHelp = "Total number of all requests for yoda so far"
)

func Start(p common.Packager) error {
	if ok := helper.Contains(types.SupportedChains, p.ChainName); ok {
		exporter := common.NewExporter(p)
		for _, rpc := range p.RPCs {
			exporter.SetRPCEndPoint(rpc)
			break
		}
		for _, api := range p.APIs {
			exporter.SetAPIEndPoint(api)
			break
		}
		go loop(exporter, p)
		return nil
	}
	return errors.Errorf("unsupported chain: %s", p.ChainName)
}

func loop(c *common.Exporter, p common.Packager) {
	rootLabels := common.BuildRootLabels(p)
	packageLabels := common.BuildPackageLabels(p)

	// each validators
	yodaStatusMetrics := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaStatusMetricName,
		Help:        YodaStatusHelp,
	}, []string{
		common.ValidatorAddressLabel,
		common.MonikerLabel,
	})

	YodaMaxMisses := p.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaMaxMissesName,
		Help:        YodaMaxMissesHelp,
	}, []string{
		common.ValidatorAddressLabel,
		common.MonikerLabel,
	})

	YodaMissSummary := p.Factory.NewSummary(prometheus.SummaryOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaMissSummaryName,
		Help:        YodaMissSummaryHelp,
		Objectives:  map[float64]float64{0.25: 0.05, 0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.99: 0.001},
	})

	YodaValidatorMissSummary := p.Factory.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaValidatorMissSummaryName,
		Help:        YodaValidatorMissSummaryHelp,
		Objectives:  map[float64]float64{0.25: 0.05, 0.5: 0.05, 0.75: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{
		common.ValidatorAddressLabel,
		common.MonikerLabel,
	})

	// each chain
	yodaSlashWindow := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        YodaRequestSlashWindow,
		Help:        YodaRequestSlashWindowHelp,
	})

	yodaRequestCount := p.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		ConstLabels: packageLabels,
		Name:        yodaRequestCountName,
		Help:        yodaRequestCountHelp,
	})

	isUnhealth := false
	for {
		// node health check
		if isUnhealth {
			healthEndpoints := healthcheck.FilterHealthEndpoints(p.APIs, p.ProtocolType)
			for _, endpoint := range healthEndpoints {
				c.SetAPIEndPoint(endpoint)
				c.Infoln("client endpoint will be changed with health endpoint for this package")
				isUnhealth = false
				break
			}
			if len(healthEndpoints) == 0 {
				c.Errorln("failed to get any health endpoints from healthcheck filter, retry sleep 10s")
				time.Sleep(UnHealthSleep)
				continue
			}
		}

		// start collect status
		status, err := router.GetStatus(c, p.ChainName)
		if err != nil {
			common.Health.With(rootLabels).Set(0)
			common.Ops.With(rootLabels).Inc()
			isUnhealth = true

			c.Logger.Errorf("failed to update metrics: %s", err.Error())
			time.Sleep(SubsystemSleep)

			continue
		}

		reqsFinished := processor.ProcessYodaMisses(yodaRequestMisses, status.Validators)

		if p.Mode == common.NETWORK {
			// update metrics for each validators
			for _, item := range status.Validators {
				yodaStatusMetrics.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(item.IsActive)

				YodaMaxMisses.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
						common.MonikerLabel:          item.Moniker,
					}).
					Set(item.MaxMisses)
			}

			for _, req := range reqsFinished {
				YodaValidatorMissSummary.
					With(prometheus.Labels{
						common.ValidatorAddressLabel: req.Validator.ValidatorOperatorAddress,
						common.MonikerLabel:          req.Validator.Moniker,
					}).Observe(float64(req.Request.BlocksPassed))
			}

		} else {
			// filter metrics for only specific validator
			for _, item := range status.Validators {
				if ok := helper.Contains(p.Monikers, item.Moniker); ok {
					yodaStatusMetrics.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(item.IsActive)

					YodaMaxMisses.
						With(prometheus.Labels{
							common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
							common.MonikerLabel:          item.Moniker,
						}).
						Set(item.MaxMisses)

					for _, req := range reqsFinished {
						if req.Validator.Moniker == item.Moniker {
							YodaValidatorMissSummary.
								With(prometheus.Labels{
									common.ValidatorAddressLabel: item.ValidatorOperatorAddress,
									common.MonikerLabel:          item.Moniker,
								}).Observe(float64(req.Request.BlocksPassed))
						}
					}
				}
			}
		}

		// update metrics for each chain

		for _, item := range reqsFinished {
			YodaMissSummary.Observe(float64(item.Request.BlocksPassed))
		}

		yodaSlashWindow.Set(status.SlashWindow)
		yodaRequestCount.Set(status.RequestCount)

		c.Infof("updated %s metrics successfully and going to sleep %s ...", Subsystem, SubsystemSleep.String())

		// update health and ops
		common.Health.With(rootLabels).Set(1)
		common.Ops.With(rootLabels).Inc()

		// update yodaRequestMisses
		yodaRequestMisses = status.Validators

		// sleep
		time.Sleep(SubsystemSleep)
	}
}
