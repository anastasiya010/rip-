package main

import (
	"log"
	"net/http"
)

func main() {
	seedMedicationDataset()

	// Статика: css и локальные картинки (вместо Minio при разработке)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/catalog", http.StatusFound) })
	http.HandleFunc("/catalog", handleMedicationCatalog)
	http.HandleFunc("/service", handleMedicationCard)
	http.HandleFunc("/order", handlePrescriptionRequest)

	log.Println("Server started on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
