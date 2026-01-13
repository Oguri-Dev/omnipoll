package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/omnipoll/backend/internal/mongo"
)

// handleEventsGet returns a list of events with filtering and pagination
func (s *Server) handleEventsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if s.worker == nil {
		WriteError(w, http.StatusInternalServerError, "Worker not initialized")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page := 1
	pageSize := 50
	var startDate, endDate *time.Time
	source := query.Get("source")
	unitName := query.Get("unitName")
	centro := query.Get("centro")
	jaula := query.Get("jaula")
	sortBy := query.Get("sortBy")
	sortOrder := -1

	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := query.Get("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	if ps := query.Get("limit"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	if start := query.Get("startDate"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startDate = &t
		}
	}

	if end := query.Get("endDate"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endDate = &t
		}
	}

	if order := query.Get("sortOrder"); order != "" {
		if parsed, err := strconv.Atoi(order); err == nil {
			sortOrder = parsed
		}
	}

	// Query events
	queryOpts := mongo.QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		StartDate: startDate,
		EndDate:   endDate,
		Source:    source,
		UnitName:  unitName,
		Centro:    centro,
		Jaula:     jaula,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	result, err := s.worker.QueryEvents(r.Context(), queryOpts)
	if err != nil {
		log.Printf("Error querying events: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to query events: "+err.Error())
		return
	}

	WritePaginated(w, http.StatusOK, result.Data, result.Page, result.TotalPages, result.Total, result.PageSize)
}

// handleEventByID handles GET, PUT, DELETE for a single event
func (s *Server) handleEventByID(w http.ResponseWriter, r *http.Request) {
	if s.worker == nil {
		WriteError(w, http.StatusInternalServerError, "Worker not initialized")
		return
	}

	// Extract event ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		WriteError(w, http.StatusBadRequest, "Invalid event ID")
		return
	}
	eventID := parts[3]

	switch r.Method {
	case http.MethodGet:
		s.handleEventGetByID(w, r, eventID)
	case http.MethodPut:
		s.handleEventUpdate(w, r, eventID)
	case http.MethodDelete:
		s.handleEventDelete(w, r, eventID)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleEventGetByID retrieves a single event by ID
func (s *Server) handleEventGetByID(w http.ResponseWriter, r *http.Request, eventID string) {
	event, err := s.worker.GetEventByID(r.Context(), eventID)
	if err != nil {
		log.Printf("Error fetching event: %v", err)
		WriteError(w, http.StatusNotFound, "Event not found")
		return
	}

	WriteSuccess(w, http.StatusOK, event)
}

// handleEventUpdate updates an event
func (s *Server) handleEventUpdate(w http.ResponseWriter, r *http.Request, eventID string) {
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.worker.UpdateEvent(r.Context(), eventID, updateData); err != nil {
		log.Printf("Error updating event: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to update event: "+err.Error())
		return
	}

	// Fetch updated event
	event, err := s.worker.GetEventByID(r.Context(), eventID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Event updated but failed to retrieve")
		return
	}

	WriteSuccess(w, http.StatusOK, event)
}

// handleEventDelete deletes an event
func (s *Server) handleEventDelete(w http.ResponseWriter, r *http.Request, eventID string) {
	if err := s.worker.DeleteEvent(r.Context(), eventID); err != nil {
		log.Printf("Error deleting event: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to delete event: "+err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"message": "Event deleted successfully"})
}

// handleEventsBatch handles batch operations (delete multiple)
func (s *Server) handleEventsBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, http.StatusMethodNotAllowed, "Only DELETE is allowed for batch operations")
		return
	}

	if s.worker == nil {
		WriteError(w, http.StatusInternalServerError, "Worker not initialized")
		return
	}

	var batchOpts struct {
		Source    string `json:"source"`
		BeforeDate string `json:"beforeDate"` // Delete events before this date
	}

	if err := json.NewDecoder(r.Body).Decode(&batchOpts); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var beforeDate *time.Time
	if batchOpts.BeforeDate != "" {
		if t, err := time.Parse(time.RFC3339, batchOpts.BeforeDate); err == nil {
			beforeDate = &t
		}
	}

	deleted, err := s.worker.DeleteEventsBatch(r.Context(), batchOpts.Source, beforeDate)
	if err != nil {
		log.Printf("Error deleting events batch: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to delete events: "+err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Batch delete completed",
		"deleted": deleted,
	})
}
