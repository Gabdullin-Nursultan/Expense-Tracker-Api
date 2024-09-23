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

// Загружает расходы из файла
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

// Сохраняет расходы в файл
func saveExpenses(expenses []Expense) error {
	data, err := json.MarshalIndent(expenses, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(expenseFile, data, 0644)
}

// Генерация уникального ID для нового расхода
func generateUniqueID(expenses []Expense) int {
	maxID := 0
	for _, expense := range expenses {
		if expense.ID > maxID {
			maxID = expense.ID
		}
	}
	return maxID + 1
}

// Handler для добавления нового расхода
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

	// Генерация уникального ID
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

// Handler для получения списка расходов
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

// Handler для получения сводки расходов
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

// Handler для обновления существующего расхода
func updateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT requests allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL, например, /expenses/update/1
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
			// Обновляем поля
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

// Handler для удаления расхода
func deleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE requests allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из URL, например, /expenses/delete/1
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
			// Удаляем расход из списка
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

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// Запуск сервера
func main() {
	// Маршруты для CRUD операций
	http.HandleFunc("/expenses", listExpensesHandler)            // GET: Получение списка расходов
	http.HandleFunc("/expenses/add", addExpenseHandler)          // POST: Добавление расхода
	http.HandleFunc("/expenses/summary", summaryExpensesHandler) // GET: Сводка расходов
	http.HandleFunc("/expenses/update/", updateExpenseHandler)   // PUT: Обновление расхода (URL: /expenses/update/{id})
	http.HandleFunc("/expenses/delete/", deleteExpenseHandler)   // DELETE: Удаление расхода (URL: /expenses/delete/{id})

	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
