package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
)

type MXRecord struct {
	Host string `json:"host"`
	Pref uint16 `json:"pref"`
}

func hostFromPath(r *http.Request) (string, bool) {
	parts := strings.SplitN(r.URL.Path, "/", 4)
	var host string
	switch len(parts) {
	case 3:
		// for A record alias handler path
		host = parts[2] // /resolve/{host}
	case 4:
		host = parts[3] // /resolve/{type}/{host}
	default:
		return "", false
	}
	host = strings.Trim(host, "/")
	return host, host != ""
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func resolveHandler(lookup func(string) (map[string]any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, ok := hostFromPath(r)
		if !ok {
			http.Error(w, "host is required", http.StatusBadRequest)
			return
		}
		result, err := lookup(host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result["host"] = host
		writeJSON(w, result)
	}
}

func lookupA(host string) (map[string]any, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	return map[string]any{"addresses": addrs}, nil
}

func lookupCNAME(host string) (map[string]any, error) {
	cname, err := net.LookupCNAME(host)
	if err != nil {
		return nil, err
	}
	return map[string]any{"cname": strings.TrimSuffix(cname, ".")}, nil
}

func lookupMX(host string) (map[string]any, error) {
	mxRecords, err := net.LookupMX(host)
	if err != nil {
		return nil, err
	}
	records := make([]MXRecord, 0, len(mxRecords))
	for _, mx := range mxRecords {
		records = append(records, MXRecord{
			Host: strings.TrimSuffix(mx.Host, "."),
			Pref: mx.Pref,
		})
	}
	return map[string]any{"mx": records}, nil
}

func lookupNS(host string) (map[string]any, error) {
	nsRecords, err := net.LookupNS(host)
	if err != nil {
		return nil, err
	}
	ns := make([]string, 0, len(nsRecords))
	for _, n := range nsRecords {
		ns = append(ns, strings.TrimSuffix(n.Host, "."))
	}
	return map[string]any{"ns": ns}, nil
}

func lookupTXT(host string) (map[string]any, error) {
	txtRecords, err := net.LookupTXT(host)
	if err != nil {
		return nil, err
	}
	return map[string]any{"txt": txtRecords}, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	mux := http.NewServeMux()
	// alias for the A record
	mux.HandleFunc("/resolve/", resolveHandler(lookupA))
	mux.HandleFunc("/resolve/a/", resolveHandler(lookupA))
	mux.HandleFunc("/resolve/cname/", resolveHandler(lookupCNAME))
	mux.HandleFunc("/resolve/mx/", resolveHandler(lookupMX))
	mux.HandleFunc("/resolve/ns/", resolveHandler(lookupNS))
	mux.HandleFunc("/resolve/txt/", resolveHandler(lookupTXT))
	mux.HandleFunc("/healthz", healthHandler)

	addr := ":8080"
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
