package main

// TODO: Add extras
// Change the IsComplete property of the Task data model to use a timestamp instead, which gives further information.
// Change from CSV to JSON, JSONL or SQLite
// Add in an optional due date to the tasks

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const TasksFile = "tasks.csv"

type Task struct {
	ID          int
	Description string
	CreatedAt   time.Time
	IsComplete  bool
}

// TaskManager to hold tasks in memory and manage file operations
// Reduce repeated opens/reads had before
type TaskManager struct {
	tasks    []Task
	filename string
	loaded   bool
}

func NewTaskManager(filename string) *TaskManager {
	return &TaskManager{
		tasks:    []Task{},
		filename: filename,
		loaded:   false,
	}
}

func (tm *TaskManager) LoadTasks() error {
	if tm.loaded {
		return nil
	}

	file, err := os.Open(tm.filename)
	if err != nil {
		if os.IsNotExist(err) {
			tm.loaded = true
			return nil
		}
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %w", err)
	}

	// Skip header row if it exists
	if len(records) > 0 {
		for _, record := range records[1:] {
			task, err := tm.parseTaskFromRecord(record)
			if err != nil {
				continue
			}
			tm.tasks = append(tm.tasks, task)
		}
	}

	tm.loaded = true
	return nil
}

func (tm *TaskManager) SaveTasks() error {
	file, err := os.Create(tm.filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Description", "CreatedAt", "IsComplete"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write all tasks
	for _, task := range tm.tasks {
		record := task.ToStringSlice()
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing task: %w", err)
		}
	}

	return nil
}

func (tm *TaskManager) AddTask(description string) error {
	if err := tm.LoadTasks(); err != nil {
		return err
	}

	newTask := Task{
		ID:          tm.getNextID(),
		Description: description,
		CreatedAt:   time.Now(),
		IsComplete:  false,
	}

	tm.tasks = append(tm.tasks, newTask)
	if err := tm.SaveTasks(); err != nil {
		return err
	}

	fmt.Printf("Task added: %s (ID: %d)\n", description, newTask.ID)
	return nil
}

func (tm *TaskManager) ListTasks(showAll bool) error {
	if err := tm.LoadTasks(); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "ID\tDescription\tCreated\tComplete")

	for _, task := range tm.tasks {
		if showAll || !task.IsComplete {
			fmt.Fprintf(w, "%d\t%s\t%s\t%t\n",
				task.ID,
				task.Description,
				task.CreatedAt.Format("2006-01-02 15:04"),
				task.IsComplete)
		}
	}

	return w.Flush()
}

func (tm *TaskManager) MarkComplete(taskID int) error {
	if err := tm.LoadTasks(); err != nil {
		return err
	}

	for i := range tm.tasks {
		if tm.tasks[i].ID == taskID {
			tm.tasks[i].IsComplete = true
			if err := tm.SaveTasks(); err != nil {
				return err
			}
			fmt.Println("Task completed!")
			return nil
		}
	}

	return fmt.Errorf("task with ID %d not found", taskID)
}

func (tm *TaskManager) DeleteTask(taskID int) error {
	if err := tm.LoadTasks(); err != nil {
		return err
	}

	for i, task := range tm.tasks {
		if task.ID == taskID {
			tm.tasks = append(tm.tasks[:i], tm.tasks[i+1:]...)
			if err := tm.SaveTasks(); err != nil {
				return err
			}
			fmt.Println("Task deleted!")
			return nil
		}
	}

	return fmt.Errorf("task with ID %d not found", taskID)
}

func (tm *TaskManager) getNextID() int {
	if len(tm.tasks) == 0 {
		return 1
	}

	maxID := 0
	for _, task := range tm.tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	return maxID + 1
}

func (tm *TaskManager) parseTaskFromRecord(record []string) (Task, error) {
	if len(record) != 4 {
		return Task{}, fmt.Errorf("invalid record length")
	}

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return Task{}, fmt.Errorf("invalid ID: %w", err)
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", record[2])
	if err != nil {
		return Task{}, fmt.Errorf("invalid date: %w", err)
	}

	isComplete, err := strconv.ParseBool(record[3])
	if err != nil {
		return Task{}, fmt.Errorf("invalid completion status: %w", err)
	}

	return Task{
		ID:          id,
		Description: record[1],
		CreatedAt:   createdAt,
		IsComplete:  isComplete,
	}, nil
}

func (t *Task) ToStringSlice() []string {
	return []string{
		strconv.Itoa(t.ID),
		t.Description,
		t.CreatedAt.Format("2006-01-02 15:04:05"),
		strconv.FormatBool(t.IsComplete),
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./task <command>")
		return
	}

	tm := NewTaskManager(TasksFile)

	cmd, args, err := parseCommand(os.Args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if err := executeCommand(cmd, tm, args); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

func parseCommand(args []string) (string, []string, error) {
	if len(args) < 2 {
		return "", nil, fmt.Errorf("no command provided")
	}

	command := args[1]
	anyarguments := []string{}

	if len(args) > 2 {
		anyarguments = args[2:]
	}

	return command, anyarguments, nil
}

func executeCommand(cmd string, tm *TaskManager, args []string) error {
	switch cmd {
	case "add":
		return handleAddCommand(tm, args)
	case "list":
		return handleListCommand(tm, args)
	case "complete":
		return handleCompleteCommand(tm, args)
	case "delete":
		return handleDeleteCommand(tm, args)
	default:
		return fmt.Errorf("'%s' command not found. Available commands: 'add', 'list', 'complete' or 'delete'", cmd)
	}
}

func handleAddCommand(tm *TaskManager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Add command requires task description")
	}

	description := strings.Join(args, " ")
	return tm.AddTask(description)
}
func handleListCommand(tm *TaskManager, args []string) error {
	showAll := false
	if len(args) > 0 && args[0] == "-all" {
		showAll = true
	}

	return tm.ListTasks(showAll)
}
func handleCompleteCommand(tm *TaskManager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Complete command requires task ID")
	}

	ID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: must be a number")
	}
	return tm.MarkComplete(ID)
}
func handleDeleteCommand(tm *TaskManager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Delete command requires task ID")
	}

	ID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID: must be a number")
	}
	return tm.DeleteTask(ID)
}
