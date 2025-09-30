package collector

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/pkg/errors"

	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/healthcheck"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/router"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/types"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ common.CollectorStart = Start
	_ common.CollectorLoop  = loop
)

const (
	Subsystem      = "axelar_vald"
	SubsystemSleep = 60 * time.Second
	UnHealthSleep  = 10 * time.Second

	CountTotalMetricName = "heartbeats_count_total"
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
	return errors.Errorf("unsupported chain type: %s", p.ProtocolType)
}

func loop(c *common.Exporter, p common.Packager) {
	rootLabels := common.BuildRootLabels(p)
	packageLabels := common.BuildPackageLabels(p)

	heartbeatsMetric := p.Factory.NewCounterVec(prometheus.CounterOpts{
		Namespace:   common.Namespace,
		Subsystem:   Subsystem,
		Name:        CountTotalMetricName,
		ConstLabels: packageLabels,
	}, []string{
		common.MonikerLabel,
		common.ValidatorAddressLabel,
		common.BroadcastorAddressLabel,
		"status",
	})

	isUnhealth := false
	initMetricFlag := false
	var latestHeartbeatsHeight int64 = 0

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

		heartbeats, err := router.GetHeartbeats(c, p.ChainName, latestHeartbeatsHeight)
		if err != nil {
			common.Health.With(rootLabels).Set(0)
			common.Ops.With(rootLabels).Inc()

			c.Logger.Errorf("failed to update heartbeats metrics: %s", err.Error())
			time.Sleep(SubsystemSleep)

			continue
		}

		if !initMetricFlag {
			initHeartbeatsMetric(heartbeatsMetric, p.Monikers, heartbeats.Validators, p.Mode)
			initMetricFlag = true
		}

		if p.Mode == common.NETWORK {
			// update metrics by each validators
			for _, item := range heartbeats.Validators {
				heartbeatsMetric.
					With(prometheus.Labels{
						common.MonikerLabel:            item.Moniker,
						common.ValidatorAddressLabel:   item.ValidatorOperatorAddress,
						common.BroadcastorAddressLabel: item.BroadcastorAddress,
						"status":                       item.Status,
					}).Add(1)
			}
		} else {
			for _, item := range heartbeats.Validators {
				if ok := helper.Contains(p.Monikers, item.Moniker); ok {
					heartbeatsMetric.
						With(prometheus.Labels{
							common.MonikerLabel:            item.Moniker,
							common.ValidatorAddressLabel:   item.ValidatorOperatorAddress,
							common.BroadcastorAddressLabel: item.BroadcastorAddress,
							"status":                       item.Status,
						}).Add(1)
				}
			}
		}

		latestHeartbeatsHeight = heartbeats.LatestHeartBeatsHeight
		c.Infof("updated %s metrics successfully and going to sleep %s ...", Subsystem, SubsystemSleep.String())

		// update health and ops
		common.Health.With(rootLabels).Set(1)
		common.Ops.With(rootLabels).Inc()

		// sleep
		time.Sleep(SubsystemSleep)
	}
}

func initHeartbeatsMetric(metric *prometheus.CounterVec, monikers []string, validators []types.BroadcastorStatus, mode common.Mode) {
	if mode == common.NETWORK {
		for _, item := range validators {
			metric.
				With(prometheus.Labels{
					common.MonikerLabel:            item.Moniker,
					common.ValidatorAddressLabel:   item.ValidatorOperatorAddress,
					common.BroadcastorAddressLabel: item.BroadcastorAddress,
					"status":                       "success",
				}).Add(0)

			metric.
				With(prometheus.Labels{
					common.MonikerLabel:            item.Moniker,
					common.ValidatorAddressLabel:   item.ValidatorOperatorAddress,
					common.BroadcastorAddressLabel: item.BroadcastorAddress,
					"status":                       "missed",
				}).Add(0)
		}
	} else {
		for _, item := range validators {
			if ok := helper.Contains(monikers, item.Moniker); ok {
				metric.
					With(prometheus.Labels{
						common.MonikerLabel:            item.Moniker,
						common.ValidatorAddressLabel:   item.ValidatorOperatorAddress,
						common.BroadcastorAddressLabel: item.BroadcastorAddress,
						"status":                       "success",
					}).Add(0)

				metric.
					With(prometheus.Labels{
						common.MonikerLabel:            item.Moniker,
						common.ValidatorAddressLabel:   item.ValidatorOperatorAddress,
						common.BroadcastorAddressLabel: item.BroadcastorAddress,
						"status":                       "missed",
					}).Add(0)
			}
		}
	}
}
