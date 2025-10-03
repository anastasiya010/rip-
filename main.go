package main

import (
	"log"
	"net/http"
)

func main() {
	seedData()

	// Статика: css и локальные картинки (вместо Minio при разработке)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/medicine-catalog", http.StatusFound)
	})
	http.HandleFunc("/medicine-catalog", handleMedicineCatalog)
	http.HandleFunc("/medicine", handleMedicine)
	http.HandleFunc("/prescription", handlePrescription)

	// Корзина (рецепт)
	http.HandleFunc("/cart/add", handleCartAdd)
	http.HandleFunc("/cart/remove", handleCartRemove)
	http.HandleFunc("/cart/clear", handleCartClear)

	log.Println("Server started on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
