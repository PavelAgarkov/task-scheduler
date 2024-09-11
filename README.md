# task-scheduler

Запуск параллельных процессов по расписанию

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
				Locator: ml3,
			},
		}},
	)))
   ```
   
2. Запуск
```
	life.Run(ctx)
```

3. Выключение
```
   life.Stop()
```
