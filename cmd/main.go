package main

import (
	"fmt"
	"github.com/PavelAgarkov/task-scheduler"
	"github.com/PavelAgarkov/task-scheduler/structs"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	gcCoont := runtime.NumGoroutine()
	log.Println(gcCoont, "in_main_start")
	//_, cancel := context.WithCancel(context.Background())
	//defer cancel()

	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)

	gcCoont = runtime.NumGoroutine()
	log.Println(gcCoont, "in_main_middle")

	//ml1 := &structs.MainLocator{Sn: "main1"}
	ml2 := &structs.ExampleLocator{Sn: "main2"}
	ml3 := &structs.ExampleLocator{Sn: "main3"}
	life, err := task_scheduler.CreateSchedule(
		[]task_scheduler.BackgroundConfiguration{
			{
				BackgroundJobFunc:         task_scheduler.Test1,
				AppName:                   "api",
				BackgroundJobName:         "UpdateFeatureFlags1",
				BackgroundJobWaitDuration: 5 * time.Second,
				LifeCheckDuration:         3 * time.Second,
				//Locator:                   ml1,
			},
			{
				BackgroundJobFunc:         task_scheduler.Test2,
				AppName:                   "api",
				BackgroundJobName:         "UpdateFeatureFlags2",
				BackgroundJobWaitDuration: 5 * time.Second,
				LifeCheckDuration:         3 * time.Second,
				DependsOf: map[string]struct{}{
					"api.UpdateFeatureFlags1": {},
				},
				DependsDuration: 100 * time.Millisecond,
				Locator:         ml2,
			},
			{
				BackgroundJobFunc:         task_scheduler.Test3,
				AppName:                   "non_api",
				BackgroundJobName:         "UpdateFeatureFlags3",
				BackgroundJobWaitDuration: 5 * time.Second,
				LifeCheckDuration:         3 * time.Second,
				DependsOf: map[string]struct{}{
					"api.UpdateFeatureFlags1": {},
					"api.UpdateFeatureFlags2": {},
					//"e.UpdateFeatureFlags4":   {},
				},
				DependsDuration: 100 * time.Millisecond,
				Locator:         ml3,
			},
		},
	)
	if err != nil {
		log.Println(err)
		return
	}

	life.Run()

	<-sigCh
	//cancel()

	log.Println(fmt.Sprintf("Alive %v", life.Alive()))
	life.Stop()
	log.Println(fmt.Sprintf("Alive %v", life.Alive()))

	logs := life.GetScheduleLogTime(time.DateTime)
	log.Println(logs)

	log.Println(runtime.NumGoroutine(), "in_main_end", "first - main, second - signal, third - test")
	log.Println(fmt.Sprintf("Alive %v", life.Alive()))
}
