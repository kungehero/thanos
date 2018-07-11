package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/go-kit/kit/log"
)

func registerReceiver(m map[string]setupFunc, app *kingpin.Application, name string) {

	cmd := app.Command(name, "receiver node exposing URL For  Receive Collector Push Metric")
	grpcBindAddr, httpBindAddr, newPeerFn := regCommonServerFlags(cmd)
	httpReceiverAddr := cmd.Flag("http-receiver-address", "Explicit (external) host:port address to receiver for HTTP Post in gossip cluster.").
		String()

	dataDir := cmd.Flag("tsdb.path", "Data directory of TSDB.").
		Default("./data").String()



	m[name] = func(g *run.Group, logger log.Logger, reg *prometheus.Registry, tracer opentracing.Tracer, debugLogging bool) error {
		peer, err := newPeerFn(logger, reg, false, "", false)
		if err != nil {
			return errors.Wrap(err, "new cluster peer")
		}
		return runReceiver(g,
			logger,
			reg,
			tracer,
			*dataDir,
			*grpcBindAddr,
			*httpBindAddr,
			*httpReceiverAddr,
			peer,
			name,
			debugLogging,
		)
	}






}
