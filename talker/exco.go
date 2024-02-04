package talker

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Callback is a function that can be used as a step in a workflow.
type Callback func(context.Context) error

// Sequential runs all callbacks sequentially.
// Example:
//
//	err := talker.Sequential(
//		func(ctx context.Context) error {
//			// ... do something
//			return nil
//		},
//		func(ctx context.Context) error {
//			// ... do something
//			return nil
//		},
//	)(context.Background())
func Sequential(callbacks ...Callback) Callback {
	return func(ctx context.Context) error {
		for _, callback := range callbacks {
			if err := callback(ctx); err != nil {
				return err
			}
		}

		return nil
	}
}

// Parallel runs all callbacks in parallel.
// Example:
//
//	err := talker.Parallel(
//		func(ctx context.Context) error {
//			// ... do something
//			return nil
//		},
//		func(ctx context.Context) error {
//			// ... do something
//			return nil
//		},
//	)(context.Background())
func Parallel(callbacks ...Callback) Callback {
	return func(ctx context.Context) error {
		var wg sync.WaitGroup

		errChan := make(chan error, len(callbacks))

		for _, callback := range callbacks {
			wg.Add(1)

			go func(w *sync.WaitGroup, callback Callback) {
				defer w.Done()
				errChan <- callback(ctx)
			}(&wg, callback)
		}

		wg.Wait()
		errs := []error{}

		for i := 0; i < len(callbacks); i++ {
			if err := <-errChan; err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	}
}

// Timeout runs callback with timeout.
// Example:
//
//	err := talker.Timeout(
//		func(ctx context.Context) error {
//			// ... do something
//			return nil
//		},
//		5*time.Second,
//	)(context.Background())
func Timeout(callback Callback, timeout time.Duration) Callback {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return callback(ctx)
	}
}

// Retry runs callback with retries.
// Example:
//
//	err := talker.Retry(
//		func(ctx context.Context) error {
//			// ... do something
//			return nil
//		},
//		3,
//		1*time.Second,
//	)(context.Background())
func Retry(callback Callback, retries int, delay time.Duration) Callback {
	return func(ctx context.Context) error {
		var err error

		for i := 0; i < retries; i++ {
			err = callback(ctx)
			if err == nil {
				return nil
			}

			time.Sleep(delay)
		}

		return err
	}
}

// IgnoreError runs callback and ignore the error.
// Example:
//
//	err := talker.IgnoreError(
//		func(ctx context.Context) error {
//			// ... do something
//			return errors.New("something went wrong") // this will be ignored by IgnoreError
//		},
//	)(context.Background())
func IgnoreError(callback Callback) Callback {
	return func(ctx context.Context) error {
		_ = callback(ctx)
		return nil
	}
}

// Atomic runs commit and rollback in sequence.
// Example:
//
//	err := talker.Atomic(
//		func(ctx context.Context) error { // Commit function, it will be called first.
//			// ... do something
//			return nil
//		},
//		func(ctx context.Context) error { // Rollback function, it will be called if commit fails.
//			// ... rollback
//			return nil
//		},
//	)(context.Background())
func Atomic(commit Callback, rollback Callback) Callback {
	return func(ctx context.Context) error {
		err := commit(ctx)
		if err != nil {
			return rollback(ctx)
		}

		return nil
	}
}

// Process is a process that can be run.
// This struct is used by the Serve function (check out the example in the Serve function).
type Process struct {
	Start       Callback     // Start is a callback that runs when the process starts.
	Live        Callback     // Live is a callback that runs periodically to check if the process is still alive.
	Ready       Callback     // Ready is a callback that runs periodically to check if the process is ready to serve requests.
	Stop        Callback     // Stop is a callback that runs when the process stops.
	Logger      *slog.Logger // Logger is the logger used by the process.
	MonitorAddr string       // MonitorAddr is the address used by the process to serve health check requests.
}

func emptyCallback(ctx context.Context) error {
	return nil
}

func sanitizeProcess(proc Process) Process {
	if proc.Start == nil {
		proc.Start = emptyCallback
	}

	if proc.Live == nil {
		proc.Live = emptyCallback
	}

	if proc.Ready == nil {
		proc.Ready = emptyCallback
	}

	if proc.Stop == nil {
		proc.Stop = emptyCallback
	}

	if proc.Logger == nil {
		proc.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	if proc.MonitorAddr == "" {
		proc.MonitorAddr = ":0" // Random port
	}

	return proc
}

func callbackToHealthCheckHandler(cb Callback) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := cb(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// Serve runs the process.
// Example:
//
//	proc := talker.Process{
//		Start: func(ctx context.Context) error {
//			// ... do something, like starting a server
//			return nil
//		},
//		Live: func(ctx context.Context) error {
//			// ... do something, like checking if the server is still alive
//			return nil
//		},
//		Ready: func(ctx context.Context) error {
//			// ... do something, like checking if the server is ready to serve requests
//			return nil
//		},
//		Stop: func(ctx context.Context) error {
//			// ... do something, like stopping the server
//			return nil
//		},
//		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
//		MonitorAddr: ":8086", // Monitor address, default is ":0" (random port)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM) // Listen to SIGINT and SIGTERM (Ctrl+C and docker stop)
//	talker.Serve(proc, sig) // Start the process, and stop it when the signal is received
func Serve(proc Process, stopSignal chan os.Signal) {
	proc = sanitizeProcess(proc)

	proc.Logger.Info("Start process")

	mainCtx, mainCancel := context.WithCancel(context.Background())

	// Health check server
	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/live", callbackToHealthCheckHandler(proc.Live))
		mux.HandleFunc("/ready", callbackToHealthCheckHandler(proc.Ready))

		server := http.Server{
			Addr:    proc.MonitorAddr,
			Handler: mux,
		}

		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			proc.Logger.Error(err.Error())
			return
		}

		defer listener.Close() // Ensure listener is closed after Serve() returns

		proc.Logger.Info("Monitor address: " + listener.Addr().String())

		go func() {
			<-stopSignal

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err := server.Shutdown(ctx)
			if err != nil {
				proc.Logger.Error(err.Error())
			}
		}()

		err = server.Serve(listener)
		if err != nil {
			proc.Logger.Error(err.Error())
		}
	}()

	// Stop process when stopSignal is received
	go func() {
		<-stopSignal

		proc.Logger.Info("Stop process")

		stopCtx, stopCancel := context.WithCancel(context.Background())

		err := proc.Stop(stopCtx)
		if err != nil {
			proc.Logger.Error(err.Error())
		}

		stopCancel()
		mainCancel()
	}()

	// Start process
	err := proc.Start(mainCtx)
	if err != nil {
		proc.Logger.Error(err.Error())
	}

	// Block until mainCtx is canceled
	<-mainCtx.Done()
}
