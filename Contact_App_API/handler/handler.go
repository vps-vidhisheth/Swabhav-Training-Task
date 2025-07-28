package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"Contact_App/auth"
	"Contact_App/service"

	"github.com/gorilla/mux"
)

// ---------- Authentication ----------

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	u, err := service.AuthenticateUser(creds.Email, creds.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	token, err := service.GenerateToken(u)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// ---------- User Handlers ----------

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	users, err := service.GetAllUsers(claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var input struct {
		FName    string `json:"fname"`
		LName    string `json:"lname"`
		IsAdmin  bool   `json:"isadmin"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if input.FName == "" || input.LName == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "All fields (fname, lname, email, password) are required", http.StatusBadRequest)
		return
	}
	createdUser, err := service.CreateUser(claims.UserID, input.FName, input.LName, input.IsAdmin, input.Email, input.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	u, err := service.GetUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	for field, value := range updates {
		if field != "fname" && field != "lname" {
			http.Error(w, "Only 'fname' and 'lname' can be updated", http.StatusBadRequest)
			return
		}
		err := service.UpdateUserByID(claims.UserID, id, field, value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	updated, _ := service.GetUserByID(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	err = service.DeleteUserByID(claims.UserID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Contact Handlers ----------

func CreateContactForUserHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	targetUserID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if claims.UserID != targetUserID {
		http.Error(w, "Forbidden: You can only add contacts to your own account", http.StatusForbidden)
		return
	}
	var input map[string]string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fname := strings.TrimSpace(input["fname"])
	lname := strings.TrimSpace(input["lname"])
	if fname == "" {
		http.Error(w, "Missing required field: fname", http.StatusBadRequest)
		return
	}
	for key := range input {
		if key != "fname" && key != "lname" {
			http.Error(w, "Invalid field: "+key, http.StatusBadRequest)
			return
		}
	}
	_, err = service.GetUserByID(targetUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	_, err = service.CreateContact(targetUserID, fname, lname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	contacts, _ := service.GetContacts(targetUserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

func GetContactsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	requestedID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if claims.UserID != requestedID {
		http.Error(w, "Forbidden: Cannot access another user's contacts", http.StatusForbidden)
		return
	}
	contacts, err := service.GetContacts(requestedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

func DeleteContactHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	contactID, err := strconv.Atoi(params["contact_id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}
	if claims.UserID != userID {
		http.Error(w, "Forbidden: Cannot delete another user's contact", http.StatusForbidden)
		return
	}
	err = service.DeleteContactByID(userID, contactID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Contact Detail Handler ----------

func AddContactDetailHandler(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserClaims(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	contactID, err := strconv.Atoi(params["contact_id"])
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}
	if _, err := service.GetContactByID(claims.UserID, contactID); err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}
	var input struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if input.Type == "" || input.Value == "" {
		http.Error(w, "Both 'type' and 'value' are required", http.StatusBadRequest)
		return
	}
	detail, err := service.AddDetailToContact(claims.UserID, contactID, input.Type, input.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}
