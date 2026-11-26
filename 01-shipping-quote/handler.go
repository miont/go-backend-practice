package main

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"slices"
	"strings"
)

type quoteRequest struct {
	Zone        string `json:"zone"`
	WeightGrams int    `json:"weight_grams"`
}

type quoteResponse struct {
	PriceKopecks int `json:"price_kopecks"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type tariff struct {
	baseCost    int
	costPerKilo int
}

var zoneToTariff = map[string]tariff{
	"local":  {baseCost: 15_000, costPerKilo: 3_000},
	"remote": {baseCost: 30_000, costPerKilo: 6_000},
}

const jsonContentType = "application/json"

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Check the method and media type, decode and validate the body,
	// then calculate the price and write the response.

	// Check method
	if r.Method != "POST" {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method_not_allowed"})
		w.Header().Set("Allow", "POST")
		return
	}

	// Check media type
	if !hasContentType(r, jsonContentType) {
		writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{Error: "unsupported_media_type"})
		return
	}

	// Parse and validate request body
	var input quoteRequest
	if err := parseBody(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}

	if !isRequestDataValid(input) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}

	// Calculate price
	price := calcPrice(input.Zone, input.WeightGrams)

	// Return result
	writeJSON(w, http.StatusOK, quoteResponse{PriceKopecks: price})
}

func parseBody(request *http.Request, v interface{}) error {
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err != nil {
		return err
	}
	if dec.More() {
		return errors.New("multiple values")
	}
	return nil
}

func hasContentType(request *http.Request, mimetype string) bool {
	contentType := request.Header.Get("Content-Type")
	for _, v := range strings.Split(contentType, ";") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}

	return false
}

func isRequestDataValid(data quoteRequest) bool {
	if !slices.Contains([]string{"local", "remote"}, data.Zone) {
		return false
	}
	if data.WeightGrams < 1 || data.WeightGrams > 20_000 {
		return false
	}
	return true
}

func calcPrice(zone string, weight int) int {
	tariff := zoneToTariff[zone]
	kilograms := (weight + 999) / 1000
	return tariff.baseCost + kilograms*tariff.costPerKilo
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
