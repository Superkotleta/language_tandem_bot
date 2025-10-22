package base

import (
	"fmt"
	"time"
)

// GenerateRequestID создает уникальный request ID для логирования.
func GenerateRequestID(operation string) string {
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), operation)
}
