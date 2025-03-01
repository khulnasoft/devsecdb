// Package cmd implements the cobra CLI for Devsecdb server.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/khulnasoft/devsecdb/backend/common"
	"github.com/khulnasoft/devsecdb/backend/common/log"
	"github.com/khulnasoft/devsecdb/backend/server"
)

// -----------------------------------Global constant BEGIN----------------------------------------.
const (

	// greetingBanner is the greeting banner.
	// http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=Devsecdb
	greetingBanner = `
___________________________________________________________________________________________

██████╗ ██╗   ██╗████████╗███████╗██████╗  █████╗ ███████╗███████╗
██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝
██████╔╝ ╚████╔╝    ██║   █████╗  ██████╔╝███████║███████╗█████╗
██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██╔══██╗██╔══██║╚════██║██╔══╝
██████╔╝   ██║      ██║   ███████╗██████╔╝██║  ██║███████║███████╗
╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝

%s
___________________________________________________________________________________________

`
	// byeBanner is the bye banner.
	// http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=BYE
	byeBanner = `
██████╗ ██╗   ██╗███████╗
██╔══██╗╚██╗ ██╔╝██╔════╝
██████╔╝ ╚████╔╝ █████╗
██╔══██╗  ╚██╔╝  ██╔══╝
██████╔╝   ██║   ███████╗
╚═════╝    ╚═╝   ╚══════╝

`
)

// -----------------------------------Global constant END------------------------------------------

// -----------------------------------Command Line Config BEGIN------------------------------------.
var (
	flags struct {
		// Used for Devsecdb command line config
		port        int
		externalURL string
		// pgURL must follow PostgreSQL connection URIs pattern.
		// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
		pgURL   string
		dataDir string
		// When we are running in readonly mode:
		// - The data file will be opened in readonly mode, no applicable migration or seeding will be applied.
		// - Requests other than GET will be rejected
		// - Any operations involving mutation will not start (e.g. Background schema syncer, task scheduler)
		readonly bool
		// saas means the Devsecdb is running in SaaS mode, several features is only controlled by us instead of users under this mode.
		saas bool
		// output logs in json format
		enableJSONLogging bool
		// demoName is the name of the demo and should be one of the subpath name in the ../migrator/demo directory.
		// empty means no demo.
		demoName string
		debug    bool
		// disableMetric is the flag to disable the metric collector.
		disableMetric bool
		// disableSample is the flag to disable the sample instance.
		disableSample bool
		lsp           bool

		developmentVersioned bool
	}

	rootCmd = &cobra.Command{
		Use:   "devsecdb",
		Short: "Devsecdb is a database schema change and version control tool",
		Run: func(_ *cobra.Command, _ []string) {
			start()

			fmt.Printf("%s", byeBanner)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// In the release build, Devsecdb bundles frontend and backend together and runs on a single port as a mono server.
	// During development, Devsecdb frontend runs on a separate port.
	rootCmd.PersistentFlags().IntVar(&flags.port, "port", 8080, "port where Devsecdb server runs. Default to 80")
	// When running the release build in production, most of the time, users would not expose Devsecdb directly to the public.
	// Instead they would configure a gateway to forward the traffic to Devsecdb. Users need to set --external-url to the address
	// exposed on that gateway accordingly.
	//
	// It's important to set the correct --external-url. This is used for:
	// 1. Constructing the correct callback URL when configuring the VCS provider. The callback URL points to the frontend.
	// 2. Creating the correct webhook endpoint when configuring the project GitOps workflow. The webhook endpoint points to the backend.
	// Since frontend and backend are bundled and run on the same address in the release build, thus we just need to specify a single external URL.
	rootCmd.PersistentFlags().StringVar(&flags.externalURL, "external-url", "", "the external URL where user visits Devsecdb, must start with http:// or https://")
	// Support environment variable for deploying to render.com using its blueprint file.
	// Render blueprint allows to specify a postgres database along with a service.
	// It allows to pass the postgres connection string as an ENV to the service.
	rootCmd.PersistentFlags().StringVar(&flags.pgURL, "pg", os.Getenv("PG_URL"), "optional external PostgreSQL instance connection url (must provide dbname); for example postgresql://user:secret@masterhost:5432/dbname?sslrootcert=cert")
	rootCmd.PersistentFlags().StringVar(&flags.dataDir, "data", ".", "not recommended for production. Directory where Devsecdb stores data if --pg is not specified. If relative path is supplied, then the path is relative to the directory where Devsecdb is under")
	rootCmd.PersistentFlags().BoolVar(&flags.readonly, "readonly", false, "whether to run in read-only mode")
	rootCmd.PersistentFlags().BoolVar(&flags.saas, "saas", false, "whether to run in SaaS mode")
	rootCmd.PersistentFlags().BoolVar(&flags.enableJSONLogging, "enable-json-logging", false, "enable output logs in devsecdb in json format")
	// Must be one of the subpath name in the ../migrator/demo directory
	rootCmd.PersistentFlags().StringVar(&flags.demoName, "demo", "", "name of the demo to use. Empty means not running in demo mode.")
	rootCmd.PersistentFlags().BoolVar(&flags.debug, "debug", false, "whether to enable debug level logging")
	rootCmd.PersistentFlags().BoolVar(&flags.lsp, "lsp", true, "whether to enable lsp in SQL Editor")
	rootCmd.PersistentFlags().BoolVar(&flags.disableMetric, "disable-metric", false, "disable the metric collector")
	rootCmd.PersistentFlags().BoolVar(&flags.disableSample, "disable-sample", false, "disable the sample instance")

	rootCmd.PersistentFlags().BoolVar(&flags.developmentVersioned, "development-versioned", false, "(WIP) versioned workflow")
}

// -----------------------------------Command Line Config END--------------------------------------

func checkDataDir() error {
	// Clean data directory path.
	flags.dataDir = filepath.Clean(flags.dataDir)

	// Convert to absolute path if relative path is supplied.
	if !filepath.IsAbs(flags.dataDir) {
		absDir, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/" + flags.dataDir)
		if err != nil {
			return err
		}
		flags.dataDir = absDir
	}

	if _, err := os.Stat(flags.dataDir); err != nil {
		return errors.Wrapf(err, "unable to access --data directory %s", flags.dataDir)
	}

	return nil
}

// Check the port availability by trying to bind and immediately release it.
func checkPort(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}
	return l.Close()
}

func start() {
	if flags.debug {
		log.LogLevel.Set(slog.LevelDebug)
	}
	if flags.saas || flags.enableJSONLogging {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: log.LogLevel, ReplaceAttr: log.Replace})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: log.LogLevel, ReplaceAttr: log.Replace})))
	}

	var err error

	if flags.externalURL != "" {
		flags.externalURL, err = common.NormalizeExternalURL(flags.externalURL)
		if err != nil {
			slog.Error("invalid --external-url", log.BBError(err))
			return
		}
	}

	if err := checkDataDir(); err != nil {
		slog.Error(err.Error())
		return
	}

	// A safety measure to prevent accidentally resetting user's actual data with demo data.
	// For emebeded mode, we control where data is stored and we put demo data in a separate directory
	// from the non-demo data.
	if flags.demoName != "" && flags.pgURL != "" {
		slog.Error("demo mode is disallowed when storing metadata in external PostgreSQL instance")
		return
	}

	profile := activeProfile(flags.dataDir)

	// The ideal bootstrap order is:
	// 1. Connect to the metadb
	// 2. Start echo server
	// 3. Start various background runners
	//
	// Strangely, when the port is unavailable, echo server would return OK response for /healthz
	// and then complain unable to bind port. Thus we cannot rely on checking /healthz. As a
	// workaround, we check whether the port is available here.
	if err := checkPort(flags.port); err != nil {
		slog.Error(fmt.Sprintf("server port %d is not available", flags.port), log.BBError(err))
		return
	}
	if profile.UseEmbedDB() {
		if err := checkPort(profile.DatastorePort); err != nil {
			slog.Error(fmt.Sprintf("database port %d is not available", profile.DatastorePort), log.BBError(err))
			return
		}
	}

	var s *server.Server
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	// Trigger graceful shutdown on SIGINT or SIGTERM.
	// The default signal sent by the `kill` command is SIGTERM,
	// which is taken as the graceful shutdown signal for many systems, eg., Kubernetes, Gunicorn.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		slog.Info(fmt.Sprintf("%s received.", sig.String()))
		if s != nil {
			_ = s.Shutdown(ctx)
		}
		cancel()
	}()

	s, err = server.NewServer(ctx, profile)
	if err != nil {
		slog.Error("Cannot new server", log.BBError(err))
		return
	}

	fmt.Printf(greetingBanner, fmt.Sprintf("Version %s(%s) has started on port %d 🚀", profile.Version, profile.GitCommit, flags.port))

	// Execute program.
	if err := s.Run(ctx, flags.port); err != nil {
		if err != http.ErrServerClosed {
			slog.Error(err.Error())
			_ = s.Shutdown(ctx)
			cancel()
		}
	}

	// Wait for CTRL-C.
	<-ctx.Done()
}
