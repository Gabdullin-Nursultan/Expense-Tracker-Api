package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Expense struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Amount      int       `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const expenseFile = "expense.json"

func loadExpenses() ([]Expense, error) {
	if _, err := os.Stat(expenseFile); os.IsNotExist(err) {
		return []Expense{}, nil
	}

	data, err := os.ReadFile(expenseFile)
	if err != nil {
		return nil, err
	}

	var expenses []Expense
	err = json.Unmarshal(data, &expenses)
	if err != nil {
		return nil, err
	}

	return expenses, nil
}

func saveExpenses(expenses []Expense) error {
	data, err := json.MarshalIndent(expenses, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(expenseFile, data, 0644)
}

func generateUniqueID(expenses []Expense) int {
	maxID := 0
	for _, expense := range expenses {
		if expense.ID > maxID {
			maxID = expense.ID
		}
	}
	return maxID + 1
}

func addExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
		return
	}

	var expense Expense
	err := json.NewDecoder(r.Body).Decode(&expense)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	expenses, err := loadExpenses()
	if err != nil {
		http.Error(w, "Error loading expenses", http.StatusInternalServerError)
		return
	}

	expense.ID = generateUniqueID(expenses)
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	expenses = append(expenses, expense)

	err = saveExpenses(expenses)
	if err != nil {
		http.Error(w, "Error saving expense", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

func listExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed", http.StatusMethodNotAllowed)
		return
	}

	expenses, err := loadExpenses()
	if err != nil {
		http.Error(w, "Error loading expenses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func summaryExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed", http.StatusMethodNotAllowed)
		return
	}

	expenses, err := loadExpenses()
	if err != nil {
		http.Error(w, "Error loading expenses", http.StatusInternalServerError)
		return
	}

	var totalSum int
	for _, expense := range expenses {
		totalSum += expense.Amount
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"total": totalSum})
}

func updateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT requests allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/expenses/update/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	var updatedExpense Expense
	err = json.NewDecoder(r.Body).Decode(&updatedExpense)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	expenses, err := loadExpenses()
	if err != nil {
		http.Error(w, "Error loading expenses", http.StatusInternalServerError)
		return
	}

	found := false
	for i, expense := range expenses {
		if expense.ID == id {
			expenses[i].Description = updatedExpense.Description
			expenses[i].Amount = updatedExpense.Amount
			expenses[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	err = saveExpenses(expenses)
	if err != nil {
		http.Error(w, "Error saving expenses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func deleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE requests allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/expenses/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	expenses, err := loadExpenses()
	if err != nil {
		http.Error(w, "Error loading expenses", http.StatusInternalServerError)
		return
	}

	found := false
	for i, expense := range expenses {
		if expense.ID == id {
			expenses = append(expenses[:i], expenses[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	err = saveExpenses(expenses)
	if err != nil {
		http.Error(w, "Error saving expenses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	http.HandleFunc("/expenses", listExpensesHandler)
	http.HandleFunc("/expenses/add", addExpenseHandler)
	http.HandleFunc("/expenses/summary", summaryExpensesHandler)
	http.HandleFunc("/expenses/delete/", deleteExpenseHandler)
	http.HandleFunc("/expenses/update/", updateExpenseHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
