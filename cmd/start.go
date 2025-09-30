package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmostation/cvms/internal/app/exporter"
	"github.com/cosmostation/cvms/internal/app/indexer"
	"github.com/cosmostation/cvms/internal/helper/config"
	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ServerBuilder func(
	port string,
	logger *logrus.Logger,
	cfg *config.MonitoringConfig,
	supportChains *config.SupportChains,
) (*http.Server, error)

var flagSets []*pflag.FlagSet

func init() {
	flagSets = []*pflag.FlagSet{
		ConfigFlag(),
		LogFlags(),
		PortFlag(),
		FilterFlag(),
	}
}

var setFlags = func(cmd *cobra.Command) {
	for _, set := range flagSets {
		cmd.Flags().AddFlagSet(set)
	}
}

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "start",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("CVMS Start subcommands")
		},
	}
	cmd.AddCommand(StartIndexerCmd(), StartExporterCmd())
	return cmd
}

// runServer 함수는 서버 빌더 함수를 인자로 받아서 서버를 실행하는 공통 로직을 구현합니다
func runServer(cmd *cobra.Command, serverBuilder ServerBuilder) error {
	ctx := cmd.Context()
	logLevel := cmd.Flag(LogLevel).Value.String()
	logColorDisable := cmd.Flag(LogColorDisable).Value.String()
	configfile := cmd.Flag(Config).Value.String()
	port := cmd.Flag(Port).Value.String()

	cfg, err := config.GetConfig(configfile)
	if err != nil {
		return err
	}

	supportChains, err := config.GetSupportChainConfig()
	if err != nil {
		return err
	}

	lgr, err := logger.GetLogger(logColorDisable, logLevel)
	if err != nil {
		return err
	}

	server, err := serverBuilder(port, lgr, cfg, supportChains)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		<-sigs
		lgr.Println("Received interrupt signal, shutting down...")
		if err := server.Shutdown(ctx); err != nil {
			lgr.Fatalf("Server Shutdown Failed:%+v", err)
		}
		lgr.Println("Server Stopped")
		close(done)
	}()
	lgr.Debugf("the server is listening in :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		lgr.Fatalf("Listen error: %v", err)
	}

	<-done
	lgr.Println("Server Exited Properly")
	return nil
}

func StartExporterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exporter",
		Short: "Start CVMS Exporter!",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			packageFilter := cmd.Flag(PackageFilter).Value.String()
			exporter.PackageFilter = packageFilter
			return runServer(cmd, exporter.Build)
		},
	}
	setFlags(cmd)
	return cmd
}

func StartIndexerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indexer",
		Short: "Start CVMS Indexer!",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			packageFilter := cmd.Flag(PackageFilter).Value.String()
			indexer.PackageFilter = packageFilter
			return runServer(cmd, indexer.Build)
		},
	}
	setFlags(cmd)
	return cmd
}
