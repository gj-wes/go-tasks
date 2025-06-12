package main

// TODO: Add extras
// Change the IsComplete property of the Task data model to use a timestamp instead, which gives further information.
// Change from CSV to JSON, JSONL or SQLite
// Add in an optional due date to the tasks
// TODO: Improve errors if no tasks/CSV not present

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"
)

type Task struct {
	ID          int
	Description string
	CreatedAt   time.Time
	IsComplete  bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./task <command>")
		return
	}

	cmd, args, err := parseCommand()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if err := executeCommand(cmd, args); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

func parseCommand() (string, []string, error) {
	if len(os.Args) < 2 {
		return "", nil, fmt.Errorf("no command provided")
	}

	command := os.Args[1]
	anyarguments := []string{}

	if len(os.Args) > 2 {
		anyarguments = os.Args[2:]
	}

	return command, anyarguments, nil
}

func executeCommand(cmd string, args []string) error {
	switch cmd {
	case "add":
		return handleAddCommand(args)
	case "list":
		return handleListCommand(args)
	case "complete":
		return handleCompleteCommand(args)
	case "delete":
		return handleDeleteCommand(args)
	default:
		return fmt.Errorf("'%s' command not found. Available commands: 'add', 'list', 'complete' or 'delete'", cmd)
	}
}

func handleAddCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Add command requires task description")
	}

	description := args[0]
	return addTask(description)
}
func handleListCommand(args []string) error {
	showAll := false
	if len(args) > 0 && args[0] == "-all" {
		showAll = true
	}

	return listTasks(showAll)
}
func handleCompleteCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Complete command requires task ID")
	}

	ID := args[0]
	return markComplete(ID)
}
func handleDeleteCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Delete command requires task ID")
	}

	ID := args[0]
	return deleteTask(ID)
}

func addTask(desc string) error {
	file, err := os.OpenFile("tasks.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("Error opening file")
	}
	defer file.Close()

	readfile, err := os.Open("tasks.csv")
	r := csv.NewReader(readfile)
	tasks, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("Error reading CSV")
	}
	defer readfile.Close()

	newtask := &Task{
		ID:          getNextID(tasks),
		Description: desc,
		CreatedAt:   time.Now(),
		IsComplete:  false,
	}

	record := newtask.ToStringSlice()

	// TODO: Fix/Improve this new line when adding
	w := csv.NewWriter(file)
	file.WriteString("\n")
	if err := w.Write(record); err != nil {
		return fmt.Errorf("Error writing csv")
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("Error flushing csv")
	}

	return nil
}

func listTasks(showAll bool) error {
	file, err := os.Open("tasks.csv")
	if err != nil {
		return fmt.Errorf("Error opening file")
	}
	defer file.Close()

	r := csv.NewReader(file)

	tasks, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("Error reading CSV")
	}

	// TODO: Formatting for Time output, '1 day ago'
	t := new(tabwriter.Writer)
	t.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(t, tasks[0][0]+"\t"+tasks[0][1]+"\t"+tasks[0][2]+"\t"+tasks[0][3])
	for _, task := range tasks[1:] {
		if showAll || task[3] == "false" {
			line := task[0] + "\t" + task[1] + "\t" + task[2] + "\t" + task[3]
			fmt.Fprintln(t, line)
		}
	}
	t.Flush()

	return nil
}

func markComplete(ID string) error {
	readfile, err := os.Open("tasks.csv")
	r := csv.NewReader(readfile)
	tasks, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("Error reading CSV")
	}
	defer readfile.Close()

	for i, task := range tasks[1:] {
		id, err := strconv.Atoi(task[0])
		if err != nil {
			continue
		}
		idToUpdate, err := strconv.Atoi(ID)
		if err != nil {
			continue
		}
		if id == idToUpdate {
			tasks[i+1][3] = "true"
			break
		}
	}

	file, err := os.Create("tasks.csv")
	if err != nil {
		return fmt.Errorf("Error opening CSV")
	}
	defer file.Close()
	w := csv.NewWriter(file)
	if err := w.WriteAll(tasks); err != nil {
		return fmt.Errorf("Error writing csv")
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("Error flushing csv")
	}

	return nil
}

func deleteTask(ID string) error {
	readfile, err := os.Open("tasks.csv")
	r := csv.NewReader(readfile)
	tasks, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("Error reading CSV")
	}
	defer readfile.Close()

	var filteredTasks [][]string

	for i, task := range tasks {
		// for the headers
		if i == 0 {
			filteredTasks = append(filteredTasks, task)
			continue
		}

		id, err := strconv.Atoi(task[0])
		if err != nil {
			continue
		}
		idToUpdate, err := strconv.Atoi(ID)
		if err != nil {
			continue
		}
		if id == idToUpdate {
			continue
		}
		filteredTasks = append(filteredTasks, task)
	}

	file, err := os.Create("tasks.csv")
	if err != nil {
		return fmt.Errorf("Error opening CSV")
	}
	defer file.Close()
	w := csv.NewWriter(file)
	if err := w.WriteAll(filteredTasks); err != nil {
		return fmt.Errorf("Error writing CSV")
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("Error flushing CSV")
	}

	return nil
}

func getNextID(records [][]string) int {
	if len(records) <= 1 {
		return 1
	}

	lastRow := records[len(records)-1]
	if len(lastRow) == 0 {
		return 1
	}

	lastID, err := strconv.Atoi(lastRow[0])
	if err != nil {
		return 1
	}

	return lastID + 1
}

func (t *Task) ToStringSlice() []string {
	return []string{
		strconv.Itoa(t.ID),
		t.Description,
		t.CreatedAt.Format("2006-01-02 15:04:05"),
		strconv.FormatBool(t.IsComplete),
	}
}
