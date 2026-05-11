package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/lroolle/speedtestcli/internal/cmdutil"
	"github.com/lroolle/speedtestcli/internal/version"
	"github.com/lroolle/speedtestcli/pkg/speedtest"
	"github.com/lroolle/speedtestcli/pkg/speedtest/cfbackend"
	"github.com/lroolle/speedtestcli/pkg/speedtest/ooklabackend"
)

type runOpts struct {
	backend    string
	format     string
	quick      bool
	thorough   bool
	timeout    time.Duration
	baseURL    string
	noUpload   bool
	noDownload bool
	verbose    bool
}

var rootCmd = &cobra.Command{
	Use:           "speedtest",
	Short:         "Agentic-native internet speed test",
	Long:          "Test internet speed via Cloudflare and Ookla/speedtest.net.\nOutputs structured JSON for agent consumption.\nDefault runs both backends in parallel.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func newCmdRun() *cobra.Command {
	var opts runOpts
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a speed test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSpeedtest(cmd.Context(), opts)
		},
	}
	bindFlags(cmd, &opts)
	return cmd
}

func bindFlags(cmd *cobra.Command, opts *runOpts) {
	cmd.Flags().StringVar(&opts.backend, "backend", "all", "Backend: all (default), cloudflare, ookla")
	cmd.Flags().StringVar(&opts.format, "format", "json", "Output format: json, ndjson, text")
	cmd.Flags().BoolVar(&opts.quick, "quick", false, "Quick test (~5s, latency + small download)")
	cmd.Flags().BoolVar(&opts.thorough, "thorough", false, "Thorough test (~2min, large payloads)")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 0, "Override test timeout")
	cmd.Flags().StringVar(&opts.baseURL, "base-url", "", "Override Cloudflare base URL")
	cmd.Flags().BoolVar(&opts.noUpload, "no-upload", false, "Skip upload tests")
	cmd.Flags().BoolVar(&opts.noDownload, "no-download", false, "Skip download tests")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "Print log lines to stderr")
}

var runOpt runOpts

func init() {
	rootCmd.Version = version.Full()

	runCmd := newCmdRun()
	rootCmd.AddCommand(runCmd)

	bindFlags(rootCmd, &runOpt)
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runSpeedtest(cmd.Context(), runOpt)
	}
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rootCmd.SetContext(ctx)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(cmdutil.ExitCode(err))
	}
}

func runSpeedtest(ctx context.Context, opts runOpts) error {
	plan := selectPlan(opts)
	if opts.timeout > 0 {
		plan.Timeout = opts.timeout
	}
	if opts.noUpload {
		plan.Steps = filterSteps(plan.Steps, "download")
	}
	if opts.noDownload {
		plan.Steps = filterSteps(plan.Steps, "upload")
	}

	backends := selectBackends(opts)
	multi := len(backends) > 1

	var sink speedtest.EventSink
	if opts.format == "ndjson" {
		var mu sync.Mutex
		enc := json.NewEncoder(os.Stdout)
		sink = func(e speedtest.Event) {
			mu.Lock()
			enc.Encode(e)
			mu.Unlock()
		}
	}

	if opts.verbose {
		names := make([]string, len(backends))
		for i, b := range backends {
			names[i] = b.Name()
		}
		logStderr("starting speed test backends=[%s] preset=%s", strings.Join(names, ","), plan.Name)
	}

	sequential := opts.thorough
	if multi {
		return runMulti(ctx, opts, backends, plan, sink, sequential)
	}
	return runSingle(ctx, opts, backends[0], plan, sink)
}

func runSingle(ctx context.Context, opts runOpts, backend speedtest.Backend, plan speedtest.TestPlan, sink speedtest.EventSink) error {
	runner := speedtest.NewRunner(
		speedtest.WithBackend(backend),
		speedtest.WithPlan(plan),
		speedtest.WithSink(sink),
	)

	result, err := runner.Run(ctx)
	if err != nil {
		if opts.verbose {
			logStderr("error: %v", err)
		}
		if result != nil {
			outputSingle(opts.format, result)
		} else {
			outputError(opts.format, err)
		}
		return err
	}

	outputSingle(opts.format, result)
	if result.Status != "ok" {
		return fmt.Errorf("test completed with status: %s", result.Status)
	}
	return nil
}

func runMulti(ctx context.Context, opts runOpts, backends []speedtest.Backend, plan speedtest.TestPlan, sink speedtest.EventSink, sequential bool) error {
	runner := speedtest.NewRunner(
		speedtest.WithBackends(backends),
		speedtest.WithPlan(plan),
		speedtest.WithSink(sink),
		speedtest.WithSequential(sequential),
	)

	report, err := runner.RunAll(ctx)

	if report != nil && len(report.Results) > 0 {
		outputReport(opts.format, report)
		hasPartial := false
		for _, r := range report.Results {
			if r.Status == "partial" || r.Status == "failed" {
				hasPartial = true
				break
			}
		}
		if hasPartial {
			return fmt.Errorf("one or more backends returned non-ok status")
		}
		return nil
	}

	if err != nil {
		outputError(opts.format, err)
		return err
	}
	return nil
}

func selectPlan(opts runOpts) speedtest.TestPlan {
	switch {
	case opts.quick:
		return speedtest.QuickPlan
	case opts.thorough:
		return speedtest.ThoroughPlan
	default:
		return speedtest.DefaultPlan
	}
}

func selectBackends(opts runOpts) []speedtest.Backend {
	var cfOpts []cfbackend.Option
	if opts.baseURL != "" {
		cfOpts = append(cfOpts, cfbackend.WithBaseURL(opts.baseURL))
	}

	switch strings.ToLower(opts.backend) {
	case "cloudflare", "cf":
		return []speedtest.Backend{cfbackend.New(cfOpts...)}
	case "ookla", "speedtest.net":
		return []speedtest.Backend{ooklabackend.New()}
	default:
		return []speedtest.Backend{
			cfbackend.New(cfOpts...),
			ooklabackend.New(),
		}
	}
}

func filterSteps(steps []speedtest.TestStep, keep string) []speedtest.TestStep {
	var filtered []speedtest.TestStep
	for _, s := range steps {
		if s.Direction == keep {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func outputSingle(format string, result *speedtest.Result) {
	switch format {
	case "text":
		printResult(result)
	case "ndjson":
		// already streamed
	default:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
	}
}

func outputReport(format string, report *speedtest.Report) {
	switch format {
	case "text":
		for i, r := range report.Results {
			if i > 0 {
				fmt.Println()
			}
			printResult(&r)
		}
		fmt.Printf("\nTotal duration: %.1fs\n", report.DurationS)
	case "ndjson":
		// already streamed
	default:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(report)
	}
}

func outputError(format string, err error) {
	switch format {
	case "text":
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	default:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(map[string]string{"error": err.Error()})
	}
}

func printResult(r *speedtest.Result) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "[%s]\t%s preset\n", r.Backend, r.Preset)

	server := r.Connection.Colo.IATA
	if server == "" {
		server = "-"
	}
	fmt.Fprintf(w, "Server\t%s\n", server)
	fmt.Fprintf(w, "Location\t%s, %s (%s)\n", r.Connection.City, r.Connection.Country, r.Connection.ASOrg)
	fmt.Fprintf(w, "IP\t%s\n", r.Connection.ClientIP)
	fmt.Fprintf(w, "Latency\t%s median, %s jitter\n",
		cmdutil.FormatMs(r.Latency.Stats.MedianMs),
		cmdutil.FormatMs(r.Latency.JitterMs))
	fmt.Fprintf(w, "Download\t%s\n", cmdutil.FormatBitsPerSec(r.Download.BitsPerSec))
	if r.Upload.BitsPerSec > 0 {
		fmt.Fprintf(w, "Upload\t%s\n", cmdutil.FormatBitsPerSec(r.Upload.BitsPerSec))
	}
	fmt.Fprintf(w, "Duration\t%.1fs\n", r.DurationS)
	w.Flush()
}

func logStderr(format string, args ...any) {
	ts := time.Now().Format(time.RFC3339)
	fmt.Fprintf(os.Stderr, "ts=%s %s\n", ts, fmt.Sprintf(format, args...))
}
