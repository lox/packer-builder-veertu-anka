package client

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type RunParams struct {
	VMName         string
	VolumesFrom    string
	Command        []string
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Debug          bool
}

type Runner struct {
	wg             sync.WaitGroup
	params         RunParams
	cmd            *exec.Cmd
	started        time.Time
	stdin          io.WriteCloser
	stdout, stderr io.ReadCloser
}

func NewRunner(params RunParams) *Runner {
	args := []string{}

	if params.Debug {
		args = append(args, "--debug")
	}

	args = append(args, "run")

	if params.VolumesFrom != "" {
		args = append(args, "--volumes-from", params.VolumesFrom)
	}

	args = append(args, params.VMName)
	args = append(args, params.Command...)

	return &Runner{
		params: params,
		cmd:    exec.Command("anka", args...),
	}
}

func (r *Runner) Start() error {
	var err error

	r.stdin, err = r.cmd.StdinPipe()
	if err != nil {
		return err
	}

	r.stderr, err = r.cmd.StderrPipe()
	if err != nil {
		return err
	}

	r.stdout, err = r.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	r.started = time.Now()
	if err := r.cmd.Start(); err != nil {
		return err
	}

	return r.readStreams()
}

func (r *Runner) readStreams() error {
	repeat := func(w io.Writer, rd io.ReadCloser, note string) {
		log.Printf("Copying %s", note)
		n, _ := io.Copy(w, rd)
		log.Printf("Copied %d bytes from %s", n, note)
		log.Printf("Closing %s", note)
		rd.Close()
		r.wg.Done()
	}

	for range time.Tick(20 * time.Second) {
		fmt.Printf("%#v", r.cmd.Process)
	}

	// for now just close stdin
	r.stdin.Close()

	if r.stdout != nil {
		r.wg.Add(1)
		go repeat(r.params.Stdout, r.stdout, "stdout")
	}

	if r.stderr != nil {
		r.wg.Add(1)
		go repeat(r.params.Stderr, r.stderr, "stderr")
	}

	return nil
}

func (r *Runner) Wait() error {
	log.Printf("Waiting for streams to finish")
	r.wg.Wait()

	log.Printf("Waiting for command to finish")
	err := r.cmd.Wait()

	log.Printf("Command finished in %s", time.Now().Sub(r.started))
	if err != nil {
		log.Printf("Command failed: %v", err)
	}
	return err
}

func (r *Runner) ExitStatus() int {
	err := r.Wait()
	if err == nil {
		return 0
	}

	exitStatus := 1
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitStatus = 1

		// There is no process-independent way to get the REAL
		// exit status so we just try to go deeper.
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
		}
	}

	log.Printf("Command exited with %d", exitStatus)
	return exitStatus
}