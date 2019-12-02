package main

import "sync"

type Scheduler struct {
	group      sync.WaitGroup
	mutex      sync.Mutex
	running    int
	maxThreads int

	scanner *Scanner
}

func (scheduler *Scheduler) Schedule(dir string) {
	scheduler.mutex.Lock()

	// exclude current thread
	running := scheduler.running

	async := running+1 <= scheduler.maxThreads

	if async {
		running += 1

		scheduler.running = running
	}

	scheduler.mutex.Unlock()

	if async {
		//log.Infof(nil, "%d async: %s", running, dir)
		scheduler.group.Add(1)
		go func() {
			scheduler.scanner.Scan(dir)

			//log.Infof(nil, "async %s decrease", dir)
			scheduler.decrease()
			scheduler.group.Done()
		}()
	} else {
		//log.Infof(nil, "%d SYNC: [%s]", running, dir)
		scheduler.scanner.Scan(dir)
	}
}

func (scheduler *Scheduler) decrease() {
	scheduler.mutex.Lock()
	scheduler.running -= 1
	scheduler.mutex.Unlock()
}

func (scheduler *Scheduler) Wait() {
	scheduler.group.Wait()
}
