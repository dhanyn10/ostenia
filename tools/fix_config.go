package main

import (
	"fmt"
	"ostenia/internal/service"
)

func main() {
	// Ganti path ini sesuai dengan lokasi Apache kamu yang ingin diperbaiki
	apachePath := `D:\koding\ostenia\build\bin\bin\apache\httpd-2.4.66-260223\Apache24`
	wwwRoot := `D:\koding\ostenia\www`
	phpDll := `D:\koding\ostenia\bin\php\current\php8apache2_4.dll`
	phpIni := `D:\koding\ostenia\bin\php\current`

	fmt.Printf("Mencoba memperbaiki config di: %s\n", apachePath)

	err := service.UpdateApacheConfig(apachePath, phpDll, phpIni, "", 80, wwwRoot)
	if err != nil {
		fmt.Printf("Gagal: %v\n", err)
	} else {
		fmt.Println("Berhasil! Silakan cek httpd.conf kamu sekarang.")
	}
}
