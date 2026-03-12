package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/electr1fy0/sorta/internal/bench"
	"github.com/electr1fy0/sorta/internal/core"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench <directory>",
	Short: "Run a duplicate-scan microbenchmark",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		var progressMu sync.Mutex
		progress := func(event core.ProgressEvent) {
			progressMu.Lock()
			defer progressMu.Unlock()
			fmt.Fprintf(os.Stderr, "\r[%s] %d/%d", event.Stage, event.Completed, event.Total)
		}

		report, err := bench.BenchmarkDuplicatesCtx(ctx, dir, progress)
		if err != nil {
			return fmt.Errorf("benchmark failed: %w", err)
		}
		fmt.Fprintln(os.Stderr)

		mbps := 0.0
		if report.Stats.DecideDuration > 0 {
			mbps = (float64(report.Stats.BytesHashed) / 1024.0 / 1024.0) / report.Stats.DecideDuration.Seconds()
		}

		fmt.Printf("Benchmark: duplicates (%s)\n", report.Directory)
		fmt.Printf("- files scanned: %d\n", report.Files)
		fmt.Printf("- planned operations: %d (%d dedupes)\n", report.Ops, report.Dedupes)
		fmt.Printf("- walk time: %s\n", report.Stats.WalkDuration)
		fmt.Printf("- plan time: %s\n", report.Stats.DecideDuration)
		fmt.Printf("- total time: %s\n", report.Stats.TotalDuration)
		fmt.Printf("- partial hashes: %d\n", report.Stats.PartialHashed)
		fmt.Printf("- full hashes: %d\n", report.Stats.FullHashed)
		fmt.Printf("- cache hits: %d\n", report.Stats.CacheHits)
		fmt.Printf("- cache misses: %d\n", report.Stats.CacheMisses)
		fmt.Printf("- bytes hashed: %s\n", core.HumanReadable(report.Stats.BytesHashed))
		fmt.Printf("- hash throughput: %.2f MiB/s\n", mbps)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(benchCmd)
}
