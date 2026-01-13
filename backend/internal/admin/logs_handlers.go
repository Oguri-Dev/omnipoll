package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/omnipoll/backend/internal/events"
)

// handleLogsImproved returns logs with filtering and pagination
func (s *Server) handleLogsImproved(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get worker logs
	var allLogs []events.LogEntry
	if s.worker != nil {
		allLogs = s.worker.GetLogs()
	}

	if allLogs == nil {
		allLogs = []events.LogEntry{}
	}

	// Parse query parameters for filtering
	query := r.URL.Query()
	level := strings.ToUpper(query.Get("level")) // Filter by level
	page := 1
	pageSize := 100

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

	// Filter by level if provided
	var filtered []events.LogEntry
	if level != "" {
		for _, log := range allLogs {
			if strings.ToUpper(log.Level) == level {
				filtered = append(filtered, log)
			}
		}
	} else {
		filtered = allLogs
	}

	// Calculate pagination
	total := int64(len(filtered))
	totalPages := (int(total) + pageSize - 1) / pageSize
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	// Reverse order to show most recent first
	var paginated []events.LogEntry
	for i := end - 1; i >= start; i-- {
		paginated = append(paginated, filtered[i])
	}

	WritePaginated(w, http.StatusOK, paginated, page, totalPages, total, pageSize)
}
