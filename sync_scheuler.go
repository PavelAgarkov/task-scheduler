package task_scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"runtime"
	"sync"
	"time"
)

type jobProperties struct {
	goName string
}

type syncScheduler struct {
	aliveGo int64
	config  BackgroundConfiguration

	aliveMapMu sync.RWMutex
	aliveMap   map[string]*jobProperties
}

type ServiceLocator interface {
	ServiceName() string
}

type BackgroundJob func(ctx context.Context, locator ServiceLocator) error

type BackgroundConfiguration struct {
	BackgroundJobFunc         BackgroundJob
	AppName                   string
	BackgroundJobName         string
	BackgroundJobWaitDuration time.Duration
	LifeCheckDuration         time.Duration
	Locator                   ServiceLocator
	DependsOf                 map[string]struct{}
	DependsDuration           time.Duration
}

func newScheduler(config BackgroundConfiguration) *syncScheduler {
	return &syncScheduler{
		config:   config,
		aliveMap: make(map[string]*jobProperties),
	}
}

func (ss *syncScheduler) getAliveGo() int {
	ss.aliveMapMu.RLock()
	n := len(ss.aliveMap)
	//log.Println("key from goName", ss.aliveMap)
	ss.aliveMapMu.RUnlock()
	//return atomic.LoadInt64(&ss.aliveGo)
	return n
}

func (ss *syncScheduler) incrementAliveGo() {
	uid, _ := uuid.NewUUID()
	key := ss.config.AppName + "." + ss.config.BackgroundJobName
	goName := key + "." + uid.String()
	ss.aliveMapMu.Lock()
	ss.aliveMap[key] = &jobProperties{goName: goName}
	ss.aliveMapMu.Unlock()
	//atomic.AddInt64(&ss.aliveGo, 1)
}

func (ss *syncScheduler) resetAliveGo() {
	ss.aliveMapMu.Lock()
	if len(ss.aliveMap) > 0 {
		ss.aliveMap = make(map[string]*jobProperties)
	} else {
		log.Println("really len", len(ss.aliveMap))
	}
	ss.aliveMapMu.Unlock()
	//atomic.StoreInt64(&ss.aliveGo, 0)
}

func (ss *syncScheduler) decrementAliveGo() {
	key := ss.config.AppName + "." + ss.config.BackgroundJobName
	ss.aliveMapMu.Lock()
	delete(ss.aliveMap, key)
	ss.aliveMapMu.Unlock()
	//atomic.AddInt64(&ss.aliveGo, -1)
}

func (ss *syncScheduler) runSchedule(importCtx context.Context, scheduleLife *ScheduleLife, locator ServiceLocator) {
	ctx, cancel := context.WithCancel(importCtx)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Println(fmt.Sprintf("syncScheduler.runSchedule() %s %s was Recovered. Error: %s",
				ss.config.AppName,
				ss.config.BackgroundJobName,
				r.(string)),
			)
		}
	}()

	for {
		select {
		case <-importCtx.Done():
			log.Println("context DONE controller")
			return

		case <-time.After(ss.config.LifeCheckDuration):
			if ss.getAliveGo() == 0 {
				log.Println(runtime.NumGoroutine(), "in_check")

				go ss.runJob(ctx, scheduleLife, locator)
			}
		}
	}
}

func (ss *syncScheduler) runJob(ctx context.Context, scheduleLife *ScheduleLife, locator ServiceLocator) {
	ss.incrementAliveGo()

	//выполняется после recover
	defer func() {
		ss.decrementAliveGo()
	}()
	defer func() {
		if r := recover(); r != nil {
			log.Println(fmt.Sprintf("syncScheduler.runJob() %s %s Has been Recovered. Error: %s", ss.config.AppName, ss.config.BackgroundJobName, r))
		}
	}()

	dependsDuration := ss.config.DependsDuration
	if dependsDuration == 0 {
		dependsDuration = 100 * time.Millisecond
	}

	// выполняет поочереденый доступ к запуску зависиостей на старте программы
dependsLoop:
	for {
		if len(ss.config.DependsOf) == 0 {
			break dependsLoop
		}
		select {
		case <-ctx.Done():
			log.Println("job context exit")
			return
		case <-time.After(dependsDuration):
			doneExpectedGoroutines := 0
			for k := range ss.config.DependsOf {
				logHit := scheduleLife.getScheduleLogTime(k)
				if logHit != nil {
					if logHit.lastEndOfExecution != nil {
						doneExpectedGoroutines++
					}
				}
			}
			if doneExpectedGoroutines == len(ss.config.DependsOf) {
				log.Println("the dependent task has been unblocked", fmt.Sprintf("%s.%s", ss.config.AppName, ss.config.BackgroundJobName))
				break dependsLoop
			}
		}
	}

	select {
	case <-ctx.Done():
		log.Println("job context exit")
		return
	case <-time.After(ss.config.BackgroundJobWaitDuration):
		key := ss.config.AppName + "." + ss.config.BackgroundJobName
		start := time.Now()
		err := ss.config.BackgroundJobFunc(ctx, locator)
		end := time.Now()
		scheduleLife.setScheduleLogTime(start, end, key)
		if err != nil {
			log.Println(fmt.Sprintf("can't background for %s %s", ss.config.AppName, ss.config.BackgroundJobName))
		}
		return
	}
}

func (ss *syncScheduler) toValidate() error {
	if ss.config.AppName == "" {
		return errors.New("app name must be filled in")
	}
	if ss.config.BackgroundJobName == "" {
		return errors.New("background job name must be filled in")
	}
	return nil
}
