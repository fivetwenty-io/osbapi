package server

import "testing"

func TestCredentialsMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		gotUser  string
		gotPass  string
		wantUser string
		wantPass string
		want     bool
	}{
		{"exact match", "admin", "secret", "admin", "secret", true},
		{"wrong username", "admim", "secret", "admin", "secret", false},
		{"wrong password", "admin", "secrex", "admin", "secret", false},
		{"both wrong", "root", "hunter2", "admin", "secret", false},
		{"password prefix", "admin", "secre", "admin", "secret", false},
		{"password with suffix", "admin", "secret!", "admin", "secret", false},
		{"username prefix", "adm", "secret", "admin", "secret", false},
		{"empty supplied", "", "", "admin", "secret", false},
		{"empty password supplied", "admin", "", "admin", "secret", false},
		{"username and password swapped", "secret", "admin", "admin", "secret", false},
		{"empty configured password matches empty", "admin", "", "admin", "", true},
		{"case differs", "Admin", "Secret", "admin", "secret", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := credentialsMatch(tc.gotUser, tc.gotPass, tc.wantUser, tc.wantPass); got != tc.want {
				t.Errorf("credentialsMatch(%q, %q, %q, %q) = %v, want %v",
					tc.gotUser, tc.gotPass, tc.wantUser, tc.wantPass, got, tc.want)
			}
		})
	}
}
