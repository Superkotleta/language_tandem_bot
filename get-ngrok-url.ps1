try {
    $response = Invoke-WebRequest -Uri 'http://localhost:4040/api/tunnels' -UseBasicParsing
    $data = $response.Content | ConvertFrom-Json
    $tunnel = $data.tunnels | Where-Object {$_.config.addr -eq 'http://localhost:8082'}
    if ($tunnel) {
        Write-Host $tunnel.public_url
    } else {
        Write-Host 'No tunnel found for port 8082'
    }
} catch {
    Write-Host 'ngrok not ready yet'
}
