package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/auth"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var benchmarkWatchers = 100

// benchmarkCmd represents the benchmark command
var benchmarkCmd = &cobra.Command{
	Use:   "benchmark <command> [args...]",
	Short: "A small command to benchmark e2e",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := foremanClient.StartJob(cmd.Context(), &foremanpb.StartJobRequest{Command: args[0], Args: args[1:]})
		if err != nil {
			st := status.Convert(err)
			slog.Error("Failed to start job", slog.Any("code", st.Code()), slog.String("error", st.Message()))
		}

		pros := make([]*progress, benchmarkWatchers)

		start := time.Now()
		var eg errgroup.Group
		for i := range benchmarkWatchers {
			i := i
			eg.Go(func() error {
				fp, err := os.OpenFile(fmt.Sprintf("out-%02d", i), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
				if err != nil {
					panic(err)
				}
				defer fp.Close()

				tlsConfig, err := auth.LoadClientTLS(clientCertPath, clientKeyPath, caCertPath)
				if err != nil {
					return fmt.Errorf("failed to load TLS config: %w", err)
				}

				grpcConn, err := grpc.NewClient(teleServer, grpc.WithTransportCredentials(tlsConfig))
				if err != nil {
					return fmt.Errorf("failed to connect to server: %w", err)
				}
				defer grpcConn.Close()

				foremanClient := foremanpb.NewForemanServiceClient(grpcConn)

				s, err := foremanClient.WatchJobOutput(cmd.Context(), &foremanpb.WatchJobOutputRequest{Id: resp.Id})
				if err != nil {
					return err
				}

				pro := &progress{id: i, last: time.Now()}
				pro.PrintProgress()
				pros[i] = pro

				br := bufio.NewWriter(fp)
				mr := io.MultiWriter(br, pro)
				for {
					out, err := s.Recv()
					if err != nil {
						return err
					}
					if data := out.GetData(); len(data) > 0 {
						for n, err := mr.Write(data); n < len(data) && err == nil; n, err = mr.Write(data) {
						}
					}
				}
				return nil
			})
		}

		if err := eg.Wait(); err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		end := time.Now()

		total := uint64(0)
		for _, pro := range pros {
			total += pro.total
		}

		fmt.Println("Avg:", formatBps(float64(total)/end.Sub(start).Seconds()))

		return nil
	},
}

type progress struct {
	id    int
	total uint64
	last  time.Time
}

func (p *progress) PrintProgress() {
	go func() {
		ticker := time.NewTicker(time.Second)
		lastTotal := uint64(0)
		for range ticker.C {
			last := p.last
			p.last = time.Now()

			since := time.Since(last)
			delta := p.total - lastTotal
			lastTotal = p.total
			fmt.Printf("[%d]: %s\n", p.id, formatBps(float64(delta)/since.Seconds()))
		}
	}()
}

func formatBps(v float64) string {
	suffix := "B/s"
	if v > 1024 {
		suffix = "KiB/s"
		v /= 1024
	}
	if v > 1024 {
		suffix = "MiB/s"
		v /= 1024
	}
	if v > 1024 {
		suffix = "GiB/s"
		v /= 1024
	}
	if v > 1024 {
		suffix = "TiB/s"
		v /= 1024
	}

	return fmt.Sprintf("%.02f %s", v, suffix)
}

func (p *progress) Write(b []byte) (int, error) {
	p.total += uint64(len(b))
	return len(b), nil
}

func init() {
	clientCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.Flags().IntVar(&benchmarkWatchers, "watchers", benchmarkWatchers, "the number of watchers per run")
}
