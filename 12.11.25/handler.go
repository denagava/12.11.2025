package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Request struct {
	Links     []string `json:"links"`
	LinksList []int    `json:"links_list"`
}

type Response struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

type Server struct {
	storage      *Storage
	linkChecker  *LinkChecker
	pdfGenerator *PDFGenerator
	shutdown     bool
	shutdownMu   sync.RWMutex
	activeTasks  sync.WaitGroup
}

func NewServer(storageFile string) *Server {
	return &Server{
		storage:      NewStorage(storageFile),
		linkChecker:  NewLinkChecker(),
		pdfGenerator: NewPDFGenerator(),
		shutdown:     false,
	}
}

func (s *Server) checkLinksHandler(w http.ResponseWriter, r *http.Request) {
	s.shutdownMu.RLock()
	if s.shutdown {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		s.shutdownMu.RUnlock()
		return
	}
	s.activeTasks.Add(1)
	s.shutdownMu.RUnlock()
	defer s.activeTasks.Done()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Links) == 0 {
		http.Error(w, "No links provided", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	results := s.linkChecker.CheckLinks(ctx, req.Links)

	linksNum := s.storage.SaveLinkSet(results)

	resp := Response{
		Links:    results,
		LinksNum: linksNum,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) reportHandler(w http.ResponseWriter, r *http.Request) {
	s.shutdownMu.RLock()
	if s.shutdown {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		s.shutdownMu.RUnlock()
		return
	}
	s.activeTasks.Add(1)
	s.shutdownMu.RUnlock()
	defer s.activeTasks.Done()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(req.LinksList) == 0 {
		http.Error(w, "No links list provided", http.StatusBadRequest)
		return
	}
	linkSets := s.storage.GetLinkSets(req.LinksList)
	if len(linkSets) == 0 {
		http.Error(w, "No data found for provided links numbers", http.StatusNotFound)
		return
	}
	pdfData, err := s.pdfGenerator.GeneratePDF(linkSets)
	if err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=link_report.pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	w.Write(pdfData)
}

func (s *Server) Shutdown() {
	s.shutdownMu.Lock()
	s.shutdown = true
	s.shutdownMu.Unlock()
	done := make(chan struct{})
	go func() {
		s.activeTasks.Wait()
		close(done)
	}()
	select {
	case <-done:
		fmt.Println("All active tasks completed")
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout waiting for active tasks")
	}
}
