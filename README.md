# task-scheduler

Запуск параллельных процессов по расписанию

```go get github.com/PavelAgarkov/task-scheduler@rc-2```

1. Конфигурирование задач
   ```
	ml2 := &structs.ExampleLocator{Sn: "main2"}
	ml3 := &structs.ExampleLocator{Sn: "main3"}
	life, err := CreateSchedule(
		[[]BackgroundConfiguration{
			{
				BackgroundJobFunc:         Test1,
				AppName:                   "api",
				BackgroundJobName:         "UpdateFeatureFlags1",
				BackgroundJobWaitDuration: 5 * time.Second,
				LifeCheckDuration:         3 * time.Second,
			},
			{
				BackgroundJobFunc:         Test2,
				AppName:                   "api",
				BackgroundJobName:         "UpdateFeatureFlags2",
				BackgroundJobWaitDuration: 5 * time.Second,
				LifeCheckDuration:         3 * time.Second,
				DependsOf: map[string]struct{}{
					"api.UpdateFeatureFlags1": {},
				},
                DependsDuration: 100 * time.Millisecond,
				Locator: ml2,
			},
			{
				BackgroundJobFunc:         Test3,
				AppName:                   "non_api",
				BackgroundJobName:         "UpdateFeatureFlags3",
				BackgroundJobWaitDuration: 5 * time.Second,
				LifeCheckDuration:         3 * time.Second,
				DependsOf: map[string]struct{}{
					"api.UpdateFeatureFlags1": {},
					"api.UpdateFeatureFlags2": {},
				},
                DependsDuration: 100 * time.Millisecond,
				Locator: ml3,
			},
		}},
	)))
   ```
   
2. Запуск
```
	life.Run()
```

3. Выключение
```
   life.Stop()
```
