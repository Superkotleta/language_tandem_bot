package handlers

import (
	"fmt"
	"time"
)

// generateRequestID создает уникальный request ID для логирования.
func generateRequestID(operation string) string {
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), operation)
}
