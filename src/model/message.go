package model

// Define our message object
type Message struct {
	TypeMsg  string `json:"type"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	From	 string `json:"from"`
	Target	 string `json:"to_address"`
}