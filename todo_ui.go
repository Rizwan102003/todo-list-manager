package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	_ "modernc.org/sqlite"
)

type Task struct {
	ID          int
	Description string
	Completed   bool
	Deadline    time.Time
	Priority    string
	Category    string
}

var db *sql.DB
var tasks []Task

func main() {
	initializeDatabase()
	defer db.Close()

	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize termui: %v", err)
	}
	defer ui.Close()

	loadTasks()
	showMainMenu()
}

func initializeDatabase() {
	var err error
	db, err = sql.Open("sqlite", "tasks.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT,
			completed BOOLEAN,
			deadline TEXT,
			priority TEXT,
			category TEXT
		);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create tasks table: %v\n", err)
	}
}

func loadTasks() {
	tasks = []Task{}
	rows, err := db.Query("SELECT id, description, completed, deadline, priority, category FROM tasks")
	if err != nil {
		log.Fatalf("Error fetching tasks: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		var deadlineStr string
		if err := rows.Scan(&task.ID, &task.Description, &task.Completed, &deadlineStr, &task.Priority, &task.Category); err != nil {
			log.Fatalf("Error scanning task: %v\n", err)
		}
		task.Deadline, _ = time.Parse("2006-01-02", deadlineStr)
		tasks = append(tasks, task)
	}
}

func showMainMenu() {
	// Define widgets
	menu := widgets.NewList()
	menu.Title = "To-Do List Manager"
	menu.Rows = []string{
		"1. Add Task",
		"2. Display Tasks",
		"3. Remove Task",
		"4. Mark Task as Completed",
		"5. Exit",
	}
	menu.SelectedRow = 0
	menu.SetRect(0, 0, 50, 10)
	ui.Render(menu)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "<Down>":
			menu.ScrollDown()
		case "<Up>":
			menu.ScrollUp()
		case "<Enter>":
			switch menu.SelectedRow {
			case 0:
				addTaskUI()
			case 1:
				displayTasksUI()
			case 2:
				removeTaskUI()
			case 3:
				markTaskCompletedUI()
			case 4:
				return
			}
		}
		ui.Render(menu)
	}
}

func addTaskUI() {
	clearScreen()
	inputs := []string{"Description", "Deadline (YYYY-MM-DD)", "Priority (Low, Medium, High)", "Category"}
	task := Task{}

	for i, label := range inputs {
		fmt.Printf("%s: ", label)
		var input string
		fmt.Scanln(&input)

		switch i {
		case 0:
			task.Description = input
		case 1:
			deadline, err := time.Parse("2006-01-02", input)
			if err != nil {
				fmt.Println("Invalid date format. Returning to menu.")
				time.Sleep(2 * time.Second)
				return
			}
			task.Deadline = deadline
		case 2:
			task.Priority = input
		case 3:
			task.Category = input
		}
	}

	insertQuery := `INSERT INTO tasks (description, completed, deadline, priority, category) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(insertQuery, task.Description, false, task.Deadline.Format("2006-01-02"), task.Priority, task.Category)
	if err != nil {
		fmt.Printf("Error adding task: %v\n", err)
		return
	}

	fmt.Println("Task added successfully! Returning to menu...")
	time.Sleep(2 * time.Second)
}

func displayTasksUI() {
	clearScreen()
	fmt.Println("To-Do List:")
	fmt.Println(strings.Repeat("-", 50))
	for _, task := range tasks {
		status := "Pending"
		if task.Completed {
			status = "Completed"
		}
		fmt.Printf("ID: %d\nDescription: %s\nDeadline: %s\nPriority: %s\nCategory: %s\nStatus: %s\n\n",
			task.ID, task.Description, task.Deadline.Format("2006-01-02"), task.Priority, task.Category, status)
	}
	fmt.Println("Press Enter to return to the menu...")
	fmt.Scanln()
}

func removeTaskUI() {
	displayTasksUI()
	fmt.Print("Enter the task ID to remove: ")
	var id int
	fmt.Scanln(&id)

	deleteQuery := `DELETE FROM tasks WHERE id = ?`
	_, err := db.Exec(deleteQuery, id)
	if err != nil {
		fmt.Printf("Error removing task: %v\n", err)
		return
	}

	fmt.Println("Task removed successfully! Returning to menu...")
	time.Sleep(2 * time.Second)
	loadTasks()
}

func markTaskCompletedUI() {
	displayTasksUI()
	fmt.Print("Enter the task ID to mark as completed: ")
	var id int
	fmt.Scanln(&id)

	updateQuery := `UPDATE tasks SET completed = ? WHERE id = ?`
	_, err := db.Exec(updateQuery, true, id)
	if err != nil {
		fmt.Printf("Error marking task as completed: %v\n", err)
		return
	}

	fmt.Println("Task marked as completed! Returning to menu...")
	time.Sleep(2 * time.Second)
	loadTasks()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
