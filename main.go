package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

type Appointment struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Date  string `json:"date"`
	Time  string `json:"time"`
}

var (
	appointments []Appointment
	mu           sync.Mutex
)

func bookAppointment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var a Appointment

	// Check content type to support HTMX form submission
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		err := json.NewDecoder(r.Body).Decode(&a)
		if err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid form input", http.StatusBadRequest)
			return
		}
		a = Appointment{
			Name:  r.FormValue("name"),
			Phone: r.FormValue("phone"),
			Date:  r.FormValue("date"),
			Time:  r.FormValue("time"),
		}
	}

	mu.Lock()
	appointments = append(appointments, a)
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Appointment booked successfully"))
}

func getAppointments(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appointments)
}

func main() {
	// Get the port from the environment variable (Render sets it)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set
	}

	// Serve static files from the 'static' directory
	fs := http.FileServer(http.Dir("./Static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "Static/index.html")
	})

	// API routes for booking and getting appointments
	http.HandleFunc("/book", bookAppointment)
	http.HandleFunc("/appointments", getAppointments)

	// Start the server on the dynamic port
	log.Printf("Server running on http://0.0.0.0:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
