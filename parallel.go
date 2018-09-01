package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"runtime/trace"
	"strings"
	"time"
)

// readLines() reads each line from the supplied io.Reader and sends each
// to the returned read-only channel.
//
// once EOF is reached, it closes the channel.
func readLines(r io.Reader, j int) <-chan string {
	rc := make(chan string, j)
	go func() {
		line := bufio.NewScanner(r)
		line.Split(bufio.ScanLines)
		for line.Scan() {
			rc <- line.Text()
		}
		close(rc)
	}()
	return rc
}

func startWorkers(jobs int, timeout time.Duration, command []string, c <-chan string) chan struct{} {

	newContext := func() func() (context.Context, context.CancelFunc) {
		if timeout > 0 {
			return func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), timeout*time.Second)
			}
		}
		return func() (context.Context, context.CancelFunc) { return context.Background(), func() { /* do nothing */ } }
	}()

	executionMessage := func() func(j []string, t time.Duration) {
		if timeout > 0 {
			return func(j []string, t time.Duration) { log.Printf("executing %v with a timeout of %d", j, t) }
		}
		return func(j []string, t time.Duration) { log.Printf("executing %v", j) }
	}()

	done := make(chan struct{}, jobs)

	work := func(worker int, c <-chan string) {
		for payload := range c {
			log.Printf("worker %d received work to do!", worker)
			log.Printf("%s %s", strings.Join(command, " "), payload)

			job := append(command, payload)

			executionMessage(job, timeout)

			func() {
				ctx, cancel := newContext()
				defer cancel()

				out, err := exec.CommandContext(ctx, job[0], job[1:]...).Output()

				if err != nil {
					log.Print(err)
					return
				}

				fmt.Printf("%s", string(out))
			}()
		}
		done <- struct{}{}
	}

	go func() {
		for i := 0; i < jobs; i++ {
			go work(i, c)
		}
	}()

	return done
}

func main() {
	jobs := flag.Int("j", runtime.NumCPU(), "Run n jobs in parallel. Default is the number of logical cpus.")
	jt := flag.Int("t", 0, "Job timeout.  Maximum execution time for each job")
	traceFile := flag.String("trace", "", "Output trace info to a specified file.")
	debug := flag.Bool("d", false, "Enable debug output.")
	flag.Parse()

	if *debug == false {
		log.SetOutput(ioutil.Discard)
	}

	if *traceFile != "" {
		outf, err := os.Create(*traceFile)

		if err != nil {
			log.Fatal(err)
		}

		trace.Start(outf)

		defer trace.Stop()
	}

	log.Printf("executing %d jobs.", *jobs)
	log.Printf("%v", flag.Args())

	done := startWorkers(*jobs, time.Duration(*jt), flag.Args(), readLines(os.Stdin, *jobs))

	for i := 0; i < *jobs; i++ {
		<-done
	}
}
