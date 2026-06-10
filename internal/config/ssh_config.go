package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type SSHSession struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	AuthMethod   string `json:"authMethod"` // "password" or "key"
	Password     string `json:"password,omitempty"`
	KeyPath      string `json:"keyPath,omitempty"`
	Passphrase   string `json:"passphrase,omitempty"`
	LastPath     string `json:"lastPath,omitempty"`
	CreatedAt    int64  `json:"createdAt"`
}

var (
	sshSessionsMu sync.RWMutex
)

func getSSHSessionsPath() string {
	return filepath.Join(GetBaseDir(), "ssh_sessions.json")
}

func LoadSSHSessions() ([]SSHSession, error) {
	sshSessionsMu.Lock()
	defer sshSessionsMu.Unlock()

	path := getSSHSessionsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []SSHSession{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sessions []SSHSession
	err = json.Unmarshal(data, &sessions)
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields after loading
	for i := range sessions {
		sessions[i].Password, _ = Decrypt(sessions[i].Password)
		sessions[i].Passphrase, _ = Decrypt(sessions[i].Passphrase)
	}

	return sessions, nil
}

func SaveSSHSessions(sessions []SSHSession) error {
	sshSessionsMu.Lock()
	defer sshSessionsMu.Unlock()

	// Create a copy to encrypt without side effects on the passed slice
	encryptedSessions := make([]SSHSession, len(sessions))
	for i, s := range sessions {
		encryptedSessions[i] = s
		encryptedSessions[i].Password, _ = Encrypt(s.Password)
		encryptedSessions[i].Passphrase, _ = Encrypt(s.Passphrase)
	}

	path := getSSHSessionsPath()
	data, err := json.MarshalIndent(encryptedSessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func AddSSHSession(session SSHSession) error {
	sessions, err := LoadSSHSessions()
	if err != nil {
		return err
	}
	sessions = append(sessions, session)
	return SaveSSHSessions(sessions)
}

func UpdateSSHSession(session SSHSession) error {
	sessions, err := LoadSSHSessions()
	if err != nil {
		return err
	}
	for i, s := range sessions {
		if s.ID == session.ID {
			sessions[i] = session
			break
		}
	}
	return SaveSSHSessions(sessions)
}

func DeleteSSHSession(id string) error {
	sessions, err := LoadSSHSessions()
	if err != nil {
		return err
	}
	newSessions := []SSHSession{}
	for _, s := range sessions {
		if s.ID != id {
			newSessions = append(newSessions, s)
		}
	}
	return SaveSSHSessions(newSessions)
}
