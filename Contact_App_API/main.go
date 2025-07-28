package main

import (
	"Contact_App/auth"
	"Contact_App/handler"
	"Contact_App/user"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	admin := user.CreateInitialAdminUser()
	log.Println("Initial Admin User Created:")
	log.Printf("Email: %s", admin.Email)
	log.Printf("Password: %s", admin.Password)
	log.Printf("Role: Admin")

	r := mux.NewRouter()

	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(auth.AttachUserToContextIfTokenValid)

	adminRoutes := api.PathPrefix("/admin").Subrouter()
	adminRoutes.Use(auth.MiddlewareAdminActive)
	adminRoutes.HandleFunc("/users", handler.CreateUserHandler).Methods("POST")
	adminRoutes.HandleFunc("/users/{id}", handler.DeleteUserHandler).Methods("DELETE")

	api.HandleFunc("/users", handler.GetAllUsersHandler).Methods("GET")
	api.HandleFunc("/users/{id}", handler.GetUserHandler).Methods("GET")
	api.HandleFunc("/users/{id}", handler.UpdateUserHandler).Methods("PUT")

	selfRoutes := api.PathPrefix("/user/{id}").Subrouter()
	selfRoutes.Use(auth.MiddlewareUserActive)
	selfRoutes.HandleFunc("/contacts", handler.GetContactsByUserIDHandler).Methods("GET")
	selfRoutes.HandleFunc("/contacts", handler.CreateContactForUserHandler).Methods("POST")
	selfRoutes.HandleFunc("/contacts/{contact_id}/details", handler.AddContactDetailHandler).Methods("POST")
	selfRoutes.HandleFunc("/contacts/{contact_id}", handler.DeleteContactHandler).Methods("DELETE")

	log.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
