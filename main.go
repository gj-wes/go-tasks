package main

// TODO: Add extras
// Change the IsComplete property of the Task data model to use a timestamp instead, which gives further information.
// Change from CSV to JSON, JSONL or SQLite
// Add in an optional due date to the tasks
// TODO: Improve errors if no tasks/CSV not present

import (
	"encoding/csv"
	"flag"
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
	argsNoProg := os.Args[1:]

	if argsNoProg[0] != "task" {
		fmt.Println("start command with 'task'")
		return
	}

	command := argsNoProg[1]

	switch command {
	case "add":
		if len(argsNoProg) < 3 {
			fmt.Println("Task missing description")
			return
		}
		desc := argsNoProg[2]
		addTask(desc)
	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		showAll := listCmd.Bool("all", false, "list all tasks including completed")

		listCmd.Parse(argsNoProg[2:])

		listTasks(*showAll)
	case "complete":
		if len(argsNoProg) < 3 {
			fmt.Println("Provide Task ID")
			return
		}
		ID := argsNoProg[2]
		markComplete(ID)
	case "delete":
		if len(argsNoProg) < 3 {
			fmt.Println("Provide Task ID")
			return
		}
		ID := argsNoProg[2]
		deleteTask(ID)
	default:
		fmt.Println("command not found, please use 'add', 'list', 'complete' or 'delete")
	}
}

func addTask(desc string) {
	file, err := os.OpenFile("tasks.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	readfile, err := os.Open("tasks.csv")
	r := csv.NewReader(readfile)
	tasks, err := r.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
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
		fmt.Println("Error writing csv:", err)
	}

	w.Flush()

	if err := w.Error(); err != nil {
		fmt.Println("Error flushing csv:", err)
	}
}

func listTasks(showAll bool) {
	file, err := os.Open("tasks.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	r := csv.NewReader(file)

	tasks, err := r.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
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
}

func markComplete(ID string) {
	readfile, err := os.Open("tasks.csv")
	r := csv.NewReader(readfile)
	tasks, err := r.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
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
		fmt.Println("Error opening CSV:", err)
		return
	}
	defer file.Close()
	w := csv.NewWriter(file)
	if err := w.WriteAll(tasks); err != nil {
		fmt.Println("Error writing csv:", err)
	}

	w.Flush()

	if err := w.Error(); err != nil {
		fmt.Println("Error flushing csv:", err)
	}
}

func deleteTask(ID string) {
	readfile, err := os.Open("tasks.csv")
	r := csv.NewReader(readfile)
	tasks, err := r.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV:", err)
		return
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
		fmt.Println("Error opening CSV:", err)
		return
	}
	defer file.Close()
	w := csv.NewWriter(file)
	if err := w.WriteAll(filteredTasks); err != nil {
		fmt.Println("Error writing csv:", err)
	}

	w.Flush()

	if err := w.Error(); err != nil {
		fmt.Println("Error flushing csv:", err)
	}
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
