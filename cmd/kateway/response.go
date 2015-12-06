package main

import (
	"encoding/json"
	"net/http"
)

func (this *Gateway) writeErrorResponse(w http.ResponseWriter, err string, code int) {
	var out = map[string]string{
		"errmsg": err,
	}
	b, _ := json.Marshal(out)

	w.Header().Set("Content-Type", "application/json")
	this.writeKatewayHeader(w)
	http.Error(w, string(b), code)
}

func (this *Gateway) writeKatewayHeader(w http.ResponseWriter) {
	w.Header().Set("Server", "kateway")
}

func (this *Gateway) writeAuthFailure(w http.ResponseWriter) {
	// close the suspicous http connection
	w.Header().Set("Connection", "close")

	this.writeErrorResponse(w, "invalid secret", http.StatusUnauthorized)
}

func (this *Gateway) writeBreakerOpen(w http.ResponseWriter) {
	this.writeErrorResponse(w, "backend busy", http.StatusBadGateway)
}

func (this *Gateway) writeBadRequest(w http.ResponseWriter, err error) {
	this.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
}