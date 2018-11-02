package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/ipfs/go-ipfs-api"
)

func main() {
	intrh, ctx := setupInterruptHandler(context.Background())
	defer intrh.Close()

	doneCh := make(chan struct{})
	done := func() {
		close(doneCh)
	}

	err := initNodeAndDaemon(ctx, done)
	if err != nil {
		panic(err)
	}

	select {
	case <-doneCh:
		fmt.Println("daemon is ready")
	}

	shell := api.NewShell("/ip4/127.0.0.1/tcp/5001")

	fileReader, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	defer fileReader.Close()

	hash, err := shell.Add(fileReader)
	if err != nil {
		panic(err)
	}

	fmt.Println(hash)
	select {}
}

func initNodeAndDaemon(ctx context.Context, done func()) error {
	cfg := api.NodeConfig{
		Root:       "./test",
		Logger:     &logger{},
		StorageMax: "100G",
	}

	node := api.NewNode(&cfg)
	if !node.IsInitialized() {
		err := node.Init()
		if err != nil {
			return err
		}
	}

	go func() {
		err := node.Daemon(ctx, done)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

type logger struct{}

func (l *logger) Info(msg string, ctx ...interface{}) {
	fmt.Fprintf(os.Stdout, msg, ctx...)
}

func (l *logger) Warn(msg string, ctx ...interface{}) {
	fmt.Fprintf(os.Stdout, msg, ctx...)
}

func (l *logger) Error(msg string, ctx ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, ctx...)
}

// IntrHandler helps set up an interrupt handler that can
// be cleanly shut down through the io.Closer interface.
type IntrHandler struct {
	sig chan os.Signal
	wg  sync.WaitGroup
}

func NewIntrHandler() *IntrHandler {
	ih := &IntrHandler{}
	ih.sig = make(chan os.Signal, 1)
	return ih
}

func (ih *IntrHandler) Close() error {
	close(ih.sig)
	ih.wg.Wait()
	return nil
}

// Handle starts handling the given signals, and will call the handler
// callback function each time a signal is catched. The function is passed
// the number of times the handler has been triggered in total, as
// well as the handler itself, so that the handling logic can use the
// handler's wait group to ensure clean shutdown when Close() is called.
func (ih *IntrHandler) Handle(handler func(count int, ih *IntrHandler), sigs ...os.Signal) {
	signal.Notify(ih.sig, sigs...)
	ih.wg.Add(1)
	go func() {
		defer ih.wg.Done()
		count := 0
		for range ih.sig {
			count++
			handler(count, ih)
		}
		signal.Stop(ih.sig)
	}()
}

func setupInterruptHandler(ctx context.Context) (io.Closer, context.Context) {
	intrh := NewIntrHandler()
	ctx, cancelFunc := context.WithCancel(ctx)

	handlerFunc := func(count int, ih *IntrHandler) {
		switch count {
		case 1:
			fmt.Println() // Prevent un-terminated ^C character in terminal

			ih.wg.Add(1)
			go func() {
				defer ih.wg.Done()
				cancelFunc()
				time.Sleep(time.Second)
				os.Exit(0)
			}()
		}
	}

	intrh.Handle(handlerFunc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	return intrh, ctx
}
