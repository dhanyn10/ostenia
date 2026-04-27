# Ostenia

Ostenia is a lightweight, modern local development environment manager for Windows. It provides a seamless way to manage web servers, databases, and PHP/Node.js versions with built-in SSL automation.
  
<img width="1263" height="952" alt="ostenia (2)" src="https://github.com/user-attachments/assets/08c0e894-9df1-42b4-88ff-7e0d27ec760f" />

## Key Features

- **Multi-Service Stack**: Manage Apache, Nginx, MySQL, PHP, Node.js, and HeidiSQL from a single dashboard.
- **Automated SSL/HTTPS**: Generate local SSL certificates and toggle HTTPS for web servers with one click.
- **Dynamic PHP Management**: 
  - Switch PHP versions instantly.
  - Automatic Windows User PATH updates.
  - GUI-based PHP Extension manager (php.ini editor).
- **Node.js Version Management**: 
  - Switch Node.js versions instantly.
  - Automatic Windows System PATH updates (requires Administrator privileges).
- **Smart Monitoring**: Real-time PID and multi-port detection (e.g., seeing 80 and 443 active simultaneously).
- **Integrated Terminal**: Open CMD or PowerShell directly in the context of each service's directory.

## Getting Started

1. **Set Server Root**: Select your `www` folder in the Activity Center.
2. **Setup SSL**: Toggle **OpenSSL** to generate your local Root CA.
3. **Install Plugins**: Go to Plugin Management to download and install your preferred tools.
4. **Go Live**: Toggle services on/off as needed.

## Tech Stack

- **Backend**: Go (Wails)
- **Frontend**: React + Tailwind CSS
- **Icons**: Lucide React
