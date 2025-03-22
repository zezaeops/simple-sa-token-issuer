package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	Port           string
	AuthToken      string
	ReadOnlySA     string
	ReadOnlyNS     string
	AdminSA        string
	AdminNS        string
	TokenExpiryMin int64
}

type TokenRequest struct {
	Role string `json:"role"`
}

type TokenResponse struct {
	Kind                string      `json:"kind"`
	APIVersion          string      `json:"apiVersion"`
	Token               string      `json:"token"`
	ExpirationTimestamp metav1.Time `json:"expirationTimestamp,omitempty"`
}

func loadConfig() *Config {
	port := getEnvOrDefault("PORT", "8080")
	authToken := getEnvOrDefault("AUTH_TOKEN", "")
	readOnlySA := getEnvOrDefault("READONLY_SA", "readonly")
	readOnlyNS := getEnvOrDefault("READONLY_NS", "default")
	adminSA := getEnvOrDefault("ADMIN_SA", "admin")
	adminNS := getEnvOrDefault("ADMIN_NS", "default")
	tokenExpiryMin := int64(60) // Default 60 minutes

	return &Config{
		Port:           port,
		AuthToken:      authToken,
		ReadOnlySA:     readOnlySA,
		ReadOnlyNS:     readOnlyNS,
		AdminSA:        adminSA,
		AdminNS:        adminNS,
		TokenExpiryMin: tokenExpiryMin,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	config := loadConfig()

	// Create Kubernetes client
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	// Set up HTTP server
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/token", authMiddleware(config.AuthToken, tokenHandler(clientset, config)))

	log.Printf("Starting server on port %s", config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func authMiddleware(authToken string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth check if no token is configured
		if authToken == "" {
			next(w, r)
			return
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from Token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Token" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != authToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func tokenHandler(clientset *kubernetes.Clientset, config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Determine which service account to use based on role
		var saName, saNamespace string
		switch req.Role {
		case "read-only":
			saName = config.ReadOnlySA
			saNamespace = config.ReadOnlyNS
		case "admin":
			saName = config.AdminSA
			saNamespace = config.AdminNS
		default:
			http.Error(w, "Invalid role specified", http.StatusBadRequest)
			return
		}

		// Create token request
		expSeconds := config.TokenExpiryMin * 60
		tokenRequest := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				ExpirationSeconds: &expSeconds,
			},
		}

		// Request token from Kubernetes API
		tr, err := clientset.CoreV1().ServiceAccounts(saNamespace).CreateToken(r.Context(), saName, tokenRequest, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Failed to create token: %v", err)
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}

		// Prepare response in client.authentication.k8s.io/v1beta1 format
		resp := TokenResponse{
			Kind:                "ExecCredential",
			APIVersion:          "client.authentication.k8s.io/v1beta1",
			Token:               tr.Status.Token,
			ExpirationTimestamp: tr.Status.ExpirationTimestamp,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
