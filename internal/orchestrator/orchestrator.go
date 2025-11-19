package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Orchestrator struct {
	Config  *Config
	Results *ExecutionResult
}

type ExecutionResult struct {
	Success      bool
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	TasksRun     int
	TasksSuccess int
	TasksFailed  int
	Errors       []error
	mu           sync.Mutex
}

func New(config *Config) *Orchestrator {
	return &Orchestrator{
		Config: config,
		Results: &ExecutionResult{
			Success:   true,
			StartTime: time.Now(),
		},
	}
}

func (o *Orchestrator) ExecutePipeline(ctx context.Context, pipelineName string) error {
	pipeline, ok := o.Config.GetPipeline(pipelineName)
	if !ok {
		return fmt.Errorf("pipeline '%s' not found", pipelineName)
	}

	if !pipeline.Enabled {
		return fmt.Errorf("pipeline '%s' is disabled", pipelineName)
	}

	fmt.Printf("🚀 Starting pipeline: %s\n", pipeline.Name)
	fmt.Printf("   %s\n\n", pipeline.Description)

	for _, stage := range pipeline.Stages {
		if err := o.executeStage(ctx, stage); err != nil {
			o.Results.Success = false
			o.Results.Errors = append(o.Results.Errors, err)
			return fmt.Errorf("stage '%s' failed: %w", stage.Name, err)
		}
	}

	o.Results.EndTime = time.Now()
	o.Results.Duration = o.Results.EndTime.Sub(o.Results.StartTime)

	fmt.Printf("\n✅ Pipeline completed in %v\n", o.Results.Duration)
	return nil
}

func (o *Orchestrator) executeStage(ctx context.Context, stage Stage) error {
	fmt.Printf("📋 Stage: %s\n", stage.Name)

	if stage.Parallel {
		return o.executeTasksParallel(ctx, stage.Tasks)
	}
	return o.executeTasksSequential(ctx, stage.Tasks)
}

func (o *Orchestrator) executeTasksSequential(ctx context.Context, tasks []Task) error {
	for _, task := range tasks {
		if err := o.executeTask(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orchestrator) executeTasksParallel(ctx context.Context, tasks []Task) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := o.executeTask(ctx, t); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Orchestrator) executeTask(ctx context.Context, task Task) error {
	o.Results.mu.Lock()
	o.Results.TasksRun++
	o.Results.mu.Unlock()

	fmt.Printf("   → %s: %s", task.Type, task.Name)

	var err error
	for attempt := 0; attempt <= task.Retry; attempt++ {
		if attempt > 0 {
			fmt.Printf(" (retry %d/%d)", attempt, task.Retry)
			time.Sleep(time.Second * time.Duration(attempt))
		}

		err = o.runTask(ctx, task)
		if err == nil {
			fmt.Println(" ✓")
			o.Results.mu.Lock()
			o.Results.TasksSuccess++
			o.Results.mu.Unlock()
			return nil
		}
	}

	fmt.Printf(" ✗ - %v\n", err)
	o.Results.mu.Lock()
	o.Results.TasksFailed++
	o.Results.Errors = append(o.Results.Errors, err)
	o.Results.mu.Unlock()

	return fmt.Errorf("task '%s' failed after %d retries: %w", task.Name, task.Retry, err)
}

func (o *Orchestrator) runTask(ctx context.Context, task Task) error {
	switch task.Type {
	case "collector":
		return o.runCollector(ctx, task)
	case "exporter":
		return o.runExporter(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (o *Orchestrator) runCollector(ctx context.Context, task Task) error {
	var binaryName string
	var args []string

	switch task.Name {
	case "profile":
		binaryName = "user_collector"
		for _, user := range o.Config.GetUsers() {
			output := getConfigString(task.Config, "output", "data/profile.json")
			if user == o.Config.GetUsers()[0] {
				args = []string{"-user=" + user, "-out=" + output}
			} else {
				output = getConfigString(task.Config, "output_secondary", "data/profile-secondary.json")
				args = []string{"-user=" + user, "-out=" + output}
			}
			if err := o.runBinary(binaryName, args...); err != nil {
				return err
			}
		}
		return nil

	case "repos":
		binaryName = "repos_collector"
		users := ""
		for i, user := range o.Config.GetUsers() {
			if i > 0 {
				users += ","
			}
			users += user
		}
		args = []string{"-users=" + users}

	case "stats":
		binaryName = "stats_collector"
		for i, user := range o.Config.GetUsers() {
			var output string
			if i == 0 {
				output = getConfigString(task.Config, "output", "data/languages.json")
			} else {
				output = getConfigString(task.Config, "output_secondary", "data/languages-secondary.json")
			}
			args = []string{"-user=" + user, "-out=" + output}
			if err := o.runBinary(binaryName, args...); err != nil {
				return err
			}
		}
		return nil

	case "activity":
		binaryName = "activity_collector"
		days := getConfigInt(task.Config, "days", 365)
		for i, user := range o.Config.GetUsers() {
			args = []string{fmt.Sprintf("-user=%s", user), fmt.Sprintf("-days=%d", days)}
			if i > 0 {
				// Secondary user - usa arquivo secundário
				args = append(args, "-out=data/activity-daily-secondary.json")
			}
			if err := o.runBinary(binaryName, args...); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown collector: %s", task.Name)
	}

	return o.runBinary(binaryName, args...)
}

func (o *Orchestrator) runExporter(ctx context.Context, task Task) error {
	switch task.Name {
	case "dashboard":
		source := getConfigString(task.Config, "source", "data")
		dest := getConfigString(task.Config, "destination", "dashboard/public/data")
		return o.runBinary("cp", "-r", source+"/.", dest+"/")
	default:
		return fmt.Errorf("unknown exporter: %s", task.Name)
	}
}

func (o *Orchestrator) runBinary(name string, args ...string) error {
	// Check if it's a binary in bin/ or system command
	var cmdPath string
	if _, err := os.Stat("./bin/" + name); err == nil {
		cmdPath = "./bin/" + name
	} else {
		cmdPath = name
	}

	cmd := exec.Command(cmdPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

func (o *Orchestrator) PrintReport() {
	separator := strings.Repeat("=", 50)
	fmt.Println("\n" + separator)
	fmt.Println("📊 Execution Report")
	fmt.Println(separator)
	fmt.Printf("Status: ")
	if o.Results.Success {
		fmt.Println("✅ Success")
	} else {
		fmt.Println("❌ Failed")
	}
	fmt.Printf("Duration: %v\n", o.Results.Duration)
	fmt.Printf("Tasks Run: %d\n", o.Results.TasksRun)
	fmt.Printf("Tasks Success: %d\n", o.Results.TasksSuccess)
	fmt.Printf("Tasks Failed: %d\n", o.Results.TasksFailed)

	if len(o.Results.Errors) > 0 {
		fmt.Println("\n❌ Errors:")
		for i, err := range o.Results.Errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}
	fmt.Println(separator)
}

func getConfigString(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getConfigInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}
