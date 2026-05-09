package auth

// AuthRequest is the payload sent to the evaluation server to obtain a token.
type AuthRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	RollNo       string `json:"rollNo"`
	AccessCode   string `json:"accessCode"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

// AuthResponse is the expected response from the evaluation server.
// Adjust fields based on the actual response schema of the evaluation server.
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // assuming seconds, adjust if needed
}
