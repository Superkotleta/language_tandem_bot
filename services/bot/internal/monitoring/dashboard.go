// Package monitoring provides performance monitoring and dashboard functionality.
package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/logging"
)

// Dashboard представляет веб-дашборд для мониторинга.
type Dashboard struct {
	performanceMonitor *logging.PerformanceMonitor
	errorHandler       *errors.CentralizedErrorHandler
	server             *http.Server
	logger             *log.Logger
}

// NewDashboard создает новый дашборд мониторинга.
func NewDashboard(performanceMonitor *logging.PerformanceMonitor, errorHandler *errors.CentralizedErrorHandler) *Dashboard {
	return &Dashboard{
		performanceMonitor: performanceMonitor,
		errorHandler:       errorHandler,
		logger:             log.New(os.Stdout, "[DASHBOARD] ", log.LstdFlags),
	}
}

// Start запускает веб-сервер дашборда.
func (d *Dashboard) Start(port int) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/metrics", d.handleMetrics)
	mux.HandleFunc("/api/errors", d.handleErrors)
	mux.HandleFunc("/api/alerts", d.handleAlerts)
	mux.HandleFunc("/api/performance", d.handlePerformance)
	mux.HandleFunc("/api/health", d.handleHealth)

	// Web interface
	mux.HandleFunc("/", d.handleIndex)
	mux.HandleFunc("/metrics", d.handleMetricsPage)
	mux.HandleFunc("/errors", d.handleErrorsPage)
	mux.HandleFunc("/alerts", d.handleAlertsPage)

	d.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	d.logger.Printf("Starting monitoring dashboard on port %d", port)
	return d.server.ListenAndServe()
}

// Stop останавливает веб-сервер дашборда.
func (d *Dashboard) Stop(ctx context.Context) error {
	if d.server != nil {
		return d.server.Shutdown(ctx)
	}
	return nil
}

// handleIndex обрабатывает главную страницу дашборда.
func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Language Exchange Bot - Monitoring Dashboard</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .nav { display: flex; gap: 20px; margin-bottom: 30px; }
        .nav a { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
        .nav a:hover { background: #0056b3; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .stat-card { background: #f8f9fa; padding: 20px; border-radius: 4px; text-align: center; }
        .stat-value { font-size: 2em; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 Language Exchange Bot</h1>
            <h2>Monitoring Dashboard</h2>
        </div>
        
        <div class="nav">
            <a href="/">🏠 Dashboard</a>
            <a href="/metrics">📊 Metrics</a>
            <a href="/errors">🚨 Errors</a>
            <a href="/alerts">⚠️ Alerts</a>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value" id="active-traces">-</div>
                <div class="stat-label">Active Traces</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="total-errors">-</div>
                <div class="stat-label">Total Errors</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="active-alerts">-</div>
                <div class="stat-label">Active Alerts</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="avg-response-time">-</div>
                <div class="stat-label">Avg Response Time (ms)</div>
            </div>
        </div>
        
        <div id="status" style="text-align: center; padding: 20px; background: #d4edda; border-radius: 4px; color: #155724;">
            ✅ System Status: Healthy
        </div>
    </div>
    
    <script>
        // Auto-refresh every 5 seconds
        setInterval(function() {
            fetch('/api/performance')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('active-traces').textContent = data.active_traces || 0;
                    document.getElementById('total-errors').textContent = data.total_errors || 0;
                    document.getElementById('active-alerts').textContent = data.active_alerts || 0;
                    document.getElementById('avg-response-time').textContent = data.avg_response_time || 0;
                })
                .catch(error => console.error('Error fetching data:', error));
        }, 5000);
    </script>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprint(w, tmpl); err != nil {
		log.Printf("Failed to write dashboard template: %v", err)
		return
	}
}

// handleMetrics обрабатывает API метрик.
func (d *Dashboard) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := d.performanceMonitor.GetPerformanceReport()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

// handleErrors обрабатывает API ошибок.
func (d *Dashboard) handleErrors(w http.ResponseWriter, r *http.Request) {
	alerts := d.errorHandler.GetAlerts()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "Failed to encode alerts", http.StatusInternalServerError)
		return
	}
}

// handleAlerts обрабатывает API алертов.
func (d *Dashboard) handleAlerts(w http.ResponseWriter, r *http.Request) {
	activeAlerts := d.errorHandler.GetActiveAlerts()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(activeAlerts); err != nil {
		http.Error(w, "Failed to encode active alerts", http.StatusInternalServerError)
		return
	}
}

// handlePerformance обрабатывает API производительности.
func (d *Dashboard) handlePerformance(w http.ResponseWriter, r *http.Request) {
	report := d.performanceMonitor.GetPerformanceReport()

	// Извлекаем основные метрики
	summary, ok := report["summary"].(map[string]interface{})
	if !ok {
		summary = make(map[string]interface{})
	}

	response := map[string]interface{}{
		"active_traces":     0, // TODO: Implement GetActiveTraces method
		"total_errors":      summary["total_operations"],
		"active_alerts":     len(d.errorHandler.GetActiveAlerts()),
		"avg_response_time": summary["average_duration_ms"],
		"error_rate":        summary["error_rate_percent"],
		"cache_hit_rate":    summary["cache_hit_rate_percent"],
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleHealth обрабатывает health check.
func (d *Dashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(time.Now()).String(), // Placeholder
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleMetricsPage обрабатывает страницу метрик.
func (d *Dashboard) handleMetricsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics - Language Exchange Bot</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .nav { display: flex; gap: 20px; margin-bottom: 30px; }
        .nav a { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
        .nav a:hover { background: #0056b3; }
        .metrics-table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        .metrics-table th, .metrics-table td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        .metrics-table th { background-color: #f8f9fa; }
        .refresh-btn { background: #28a745; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #218838; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Performance Metrics</h1>
        </div>
        
        <div class="nav">
            <a href="/">🏠 Dashboard</a>
            <a href="/metrics">📊 Metrics</a>
            <a href="/errors">🚨 Errors</a>
            <a href="/alerts">⚠️ Alerts</a>
        </div>
        
        <button class="refresh-btn" onclick="loadMetrics()">🔄 Refresh</button>
        
        <table class="metrics-table" id="metrics-table">
            <thead>
                <tr>
                    <th>Metric</th>
                    <th>Value</th>
                    <th>Type</th>
                    <th>Timestamp</th>
                </tr>
            </thead>
            <tbody id="metrics-body">
                <tr><td colspan="4">Loading...</td></tr>
            </tbody>
        </table>
    </div>
    
    <script>
        function loadMetrics() {
            fetch('/api/metrics')
                .then(response => response.json())
                .then(data => {
                    const tbody = document.getElementById('metrics-body');
                    tbody.innerHTML = '';
                    
                    if (data.metrics) {
                        Object.entries(data.metrics).forEach(([key, metric]) => {
                            const row = document.createElement('tr');
                            row.innerHTML = 
                                '<td>' + key + '</td>' +
                                '<td>' + (metric.value || 'N/A') + '</td>' +
                                '<td>' + (metric.type || 'N/A') + '</td>' +
                                '<td>' + new Date(metric.timestamp).toLocaleString() + '</td>';
                            tbody.appendChild(row);
                        });
                    }
                })
                .catch(error => {
                    console.error('Error loading metrics:', error);
                    document.getElementById('metrics-body').innerHTML = '<tr><td colspan="4">Error loading metrics</td></tr>';
                });
        }
        
        // Load metrics on page load
        loadMetrics();
        
        // Auto-refresh every 10 seconds
        setInterval(loadMetrics, 10000);
    </script>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprint(w, tmpl); err != nil {
		log.Printf("Failed to write dashboard template: %v", err)
		return
	}
}

// handleErrorsPage обрабатывает страницу ошибок.
func (d *Dashboard) handleErrorsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Errors - Language Exchange Bot</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .nav { display: flex; gap: 20px; margin-bottom: 30px; }
        .nav a { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
        .nav a:hover { background: #0056b3; }
        .error-card { background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 4px; padding: 15px; margin-bottom: 15px; }
        .error-level { font-weight: bold; padding: 4px 8px; border-radius: 4px; color: white; }
        .error-level.critical { background: #dc3545; }
        .error-level.warning { background: #ffc107; color: #000; }
        .error-level.info { background: #17a2b8; }
        .refresh-btn { background: #28a745; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #218838; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚨 Error Logs</h1>
        </div>
        
        <div class="nav">
            <a href="/">🏠 Dashboard</a>
            <a href="/metrics">📊 Metrics</a>
            <a href="/errors">🚨 Errors</a>
            <a href="/alerts">⚠️ Alerts</a>
        </div>
        
        <button class="refresh-btn" onclick="loadErrors()">🔄 Refresh</button>
        
        <div id="errors-container">
            <div class="error-card">
                <p>Loading errors...</p>
            </div>
        </div>
    </div>
    
    <script>
        function loadErrors() {
            fetch('/api/errors')
                .then(response => response.json())
                .then(data => {
                    const container = document.getElementById('errors-container');
                    container.innerHTML = '';
                    
                    if (Object.keys(data).length === 0) {
                        container.innerHTML = '<div class="error-card"><p>No errors found</p></div>';
                        return;
                    }
                    
                    Object.entries(data).forEach(([id, error]) => {
                        const errorCard = document.createElement('div');
                        errorCard.className = 'error-card';
                        
                        const levelClass = error.level ? error.level.toLowerCase() : 'info';
                        const levelText = error.level || 'INFO';
                        
                        errorCard.innerHTML = 
                            '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">' +
                                '<h3>' + (error.title || 'Error') + '</h3>' +
                                '<span class="error-level ' + levelClass + '">' + levelText + '</span>' +
                            '</div>' +
                            '<p><strong>Message:</strong> ' + (error.message || 'N/A') + '</p>' +
                            '<p><strong>Timestamp:</strong> ' + new Date(error.timestamp).toLocaleString() + '</p>' +
                            '<p><strong>Resolved:</strong> ' + (error.resolved ? 'Yes' : 'No') + '</p>';
                        
                        container.appendChild(errorCard);
                    });
                })
                .catch(error => {
                    console.error('Error loading errors:', error);
                    document.getElementById('errors-container').innerHTML = '<div class="error-card"><p>Error loading errors</p></div>';
                });
        }
        
        // Load errors on page load
        loadErrors();
        
        // Auto-refresh every 15 seconds
        setInterval(loadErrors, 15000);
    </script>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprint(w, tmpl); err != nil {
		log.Printf("Failed to write dashboard template: %v", err)
		return
	}
}

// handleAlertsPage обрабатывает страницу алертов.
func (d *Dashboard) handleAlertsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Alerts - Language Exchange Bot</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .nav { display: flex; gap: 20px; margin-bottom: 30px; }
        .nav a { padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
        .nav a:hover { background: #0056b3; }
        .alert-card { background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 4px; padding: 15px; margin-bottom: 15px; }
        .alert-level { font-weight: bold; padding: 4px 8px; border-radius: 4px; color: white; }
        .alert-level.emergency { background: #dc3545; }
        .alert-level.critical { background: #fd7e14; }
        .alert-level.warning { background: #ffc107; color: #000; }
        .alert-level.info { background: #17a2b8; }
        .refresh-btn { background: #28a745; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #218838; }
        .resolve-btn { background: #6c757d; color: white; border: none; padding: 5px 10px; border-radius: 4px; cursor: pointer; }
        .resolve-btn:hover { background: #5a6268; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚠️ Active Alerts</h1>
        </div>
        
        <div class="nav">
            <a href="/">🏠 Dashboard</a>
            <a href="/metrics">📊 Metrics</a>
            <a href="/errors">🚨 Errors</a>
            <a href="/alerts">⚠️ Alerts</a>
        </div>
        
        <button class="refresh-btn" onclick="loadAlerts()">🔄 Refresh</button>
        
        <div id="alerts-container">
            <div class="alert-card">
                <p>Loading alerts...</p>
            </div>
        </div>
    </div>
    
    <script>
        function loadAlerts() {
            fetch('/api/alerts')
                .then(response => response.json())
                .then(data => {
                    const container = document.getElementById('alerts-container');
                    container.innerHTML = '';
                    
                    if (Object.keys(data).length === 0) {
                        container.innerHTML = '<div class="alert-card"><p>No active alerts</p></div>';
                        return;
                    }
                    
                    Object.entries(data).forEach(([id, alert]) => {
                        const alertCard = document.createElement('div');
                        alertCard.className = 'alert-card';
                        
                        const levelClass = alert.level ? alert.level.toLowerCase() : 'info';
                        const levelText = alert.level || 'INFO';
                        
                        alertCard.innerHTML = 
                            '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">' +
                                '<h3>' + (alert.title || 'Alert') + '</h3>' +
                                '<div>' +
                                    '<span class="alert-level ' + levelClass + '">' + levelText + '</span>' +
                                    (!alert.resolved ? '<button class="resolve-btn" onclick="resolveAlert(\'' + id + '\')">Resolve</button>' : '') +
                                '</div>' +
                            '</div>' +
                            '<p><strong>Message:</strong> ' + (alert.message || 'N/A') + '</p>' +
                            '<p><strong>Timestamp:</strong> ' + new Date(alert.timestamp).toLocaleString() + '</p>' +
                            '<p><strong>Resolved:</strong> ' + (alert.resolved ? 'Yes' : 'No') + '</p>';
                        
                        container.appendChild(alertCard);
                    });
                })
                .catch(error => {
                    console.error('Error loading alerts:', error);
                    document.getElementById('alerts-container').innerHTML = '<div class="alert-card"><p>Error loading alerts</p></div>';
                });
        }
        
        function resolveAlert(alertId) {
            // This would typically make a POST request to resolve the alert
            console.log('Resolving alert:', alertId);
            alert('Alert resolution not implemented in this demo');
        }
        
        // Load alerts on page load
        loadAlerts();
        
        // Auto-refresh every 20 seconds
        setInterval(loadAlerts, 20000);
    </script>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprint(w, tmpl); err != nil {
		log.Printf("Failed to write dashboard template: %v", err)
		return
	}
}
