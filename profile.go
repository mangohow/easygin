//go:build linux || darwin
// +build linux darwin

package easygin

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync/atomic"
	"syscall"
	"time"
)

const timeFormat = "0102150405"

const DefaultMemProfileRate = 4096

// 记录启动的次数
var started uint32

type Stopper interface {
	Stop()
}

type fakeStopper struct{}

func (s fakeStopper) Stop() {}

type Profile struct {
	closers []func()

	stopped uint32
}

func (p *Profile) close() {
	for _, fn := range p.closers {
		fn()
	}
}

func (p *Profile) startBlockProfile() {
	ppf := "block"
	name := createDumpFile(ppf)
	f, err := os.Create(name)
	if err != nil {
		elog.Printf("profile: could not create %s profile, err:%v", ppf, err)
		return
	}

	runtime.SetBlockProfileRate(1)
	elog.Printf("profile: %s profiling enabled, %s", ppf, name)

	p.closers = append(p.closers, func() {
		_ = pprof.Lookup(ppf).WriteTo(f, 0)
		_ = f.Close()
		runtime.SetBlockProfileRate(0)
		elog.Printf("profile: %s profiling disabled, %s", ppf, name)
	})
}

func (p *Profile) startCpuProfile() {
	ppf := "cpu"
	name := createDumpFile(ppf)
	f, err := os.Create(name)
	if err != nil {
		elog.Printf("profile: could not create %s profile, err:%v", ppf, err)
		return
	}

	_ = pprof.StartCPUProfile(f)
	elog.Printf("profile: %s profiling enabled, %s", ppf, name)

	p.closers = append(p.closers, func() {
		pprof.StopCPUProfile()
		_ = f.Close()
		elog.Printf("profile: %s profiling disabled, %s", ppf, name)
	})
}

func (p *Profile) startMemProfile() {
	ppf := "mem"
	name := createDumpFile(ppf)
	f, err := os.Create(name)
	if err != nil {
		elog.Printf("profile: could not create %s profile, err:%v", ppf, err)
		return
	}

	old := runtime.MemProfileRate
	runtime.MemProfileRate = DefaultMemProfileRate
	elog.Printf("profile: %s profiling enabled (rate %d), %s", ppf, DefaultMemProfileRate, name)

	p.closers = append(p.closers, func() {
		pprof.Lookup("heap").WriteTo(f, 0)
		_ = f.Close()
		runtime.MemProfileRate = old
		elog.Printf("profile: %s profiling disabled, %s", ppf, name)
	})
}

func (p *Profile) startMutexProfile() {
	ppf := "mutex"
	name := createDumpFile(ppf)
	f, err := os.Create(name)
	if err != nil {
		elog.Printf("profile: could not create %s profile, err:%v", ppf, err)
		return
	}

	runtime.SetMutexProfileFraction(1)
	elog.Printf("profile: %s profiling enabled, %s", ppf, name)

	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup(ppf); mp != nil {
			_ = mp.WriteTo(f, 0)
		}
		_ = f.Close()
		runtime.SetMutexProfileFraction(0)
		elog.Printf("profile: %s profiling disabled, %s", ppf, name)
	})
}

func (p *Profile) startThreadCreateProfile() {
	ppf := "threadcreate"
	name := createDumpFile(ppf)
	f, err := os.Create(name)
	if err != nil {
		elog.Printf("profile: could not create %s profile, err:%v", ppf, err)
		return
	}

	elog.Printf("profile: %s profiling enabled, %s", ppf, name)

	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup(ppf); mp != nil {
			_ = mp.WriteTo(f, 0)
		}
		_ = f.Close()
		elog.Printf("profile: %s profiling disabled, %s", ppf, name)
	})
}

func (p *Profile) startTraceProfile() {
	ppf := "trace"
	name := createDumpFile(ppf)
	f, err := os.Create(name)
	if err != nil {
		elog.Printf("profile: could not create %s profile, err:%v", ppf, err)
		return
	}

	if err = trace.Start(f); err != nil {
		elog.Printf("profile: could not start trace: %v", err)
		return
	}

	elog.Printf("profile: %s profiling enabled, %s", ppf, name)

	p.closers = append(p.closers, func() {
		trace.Stop()
		elog.Printf("profile: %s profiling disabled, %s", ppf, name)
	})
}

func (p *Profile) Stop() {
	if !atomic.CompareAndSwapUint32(&p.stopped, 0, 1) {
		return
	}
	p.close()
	atomic.StoreUint32(&started, 0)
}

func StartProfile() Stopper {
	if !atomic.CompareAndSwapUint32(&started, 0, 1) {
		elog.Printf("profile: Start() already called")
		return fakeStopper{}
	}

	var prof Profile
	prof.startBlockProfile()
	prof.startCpuProfile()
	prof.startMemProfile()
	prof.startMutexProfile()
	prof.startTraceProfile()
	prof.startThreadCreateProfile()

	return &prof
}

func createDumpFile(kind string) string {
	command := path.Base(os.Args[0])
	pid := syscall.Getpid()

	p := path.Join(os.TempDir(), fmt.Sprintf("%s-%d-%s-%s.pprof",
		command, pid, kind, time.Now().Format(timeFormat)))

	return p
}
