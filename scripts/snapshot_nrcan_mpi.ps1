param(
    [string]$OutputPath = "data/fixtures/nrcan_mpi_2025.json"
)

$ErrorActionPreference = "Stop"
$endpoint = "https://maps-cartes.services.geo.ca/server_serveur/rest/services/NRCan/major_projects_inventory_en/MapServer/0/query?where=1%3D1&outFields=*&returnGeometry=true&outSR=4326&resultRecordCount=1000&f=json"
$datasetPage = "https://open.canada.ca/data/en/dataset/f5f2db55-31e4-42fb-8c73-23e1c44de9b2"

$uri = [System.Uri]$endpoint
if ($uri.Scheme -ne "https" -or $uri.Host -ne "maps-cartes.services.geo.ca") {
    throw "Refusing to retrieve an unapproved NRCan endpoint."
}

$response = Invoke-RestMethod -Uri $uri -Method Get -TimeoutSec 45 -MaximumRedirection 2
if ($null -ne $response.error) {
    throw "NRCan ArcGIS returned error $($response.error.code): $($response.error.message)"
}
if ($null -eq $response.features -or $response.features.Count -lt 250 -or $response.features.Count -gt 1000) {
    throw "Unexpected NRCan feature count: $($response.features.Count). Snapshot was not replaced."
}

$requiredFields = @("id", "company", "project_name", "province", "capital_cost", "sector", "status")
$first = $response.features[0].attributes
foreach ($field in $requiredFields) {
    if ($null -eq $first.$field) {
        throw "NRCan schema no longer contains required field '$field'. Snapshot was not replaced."
    }
}

$envelope = [ordered]@{
    source = "Natural Resources Canada Major Projects Inventory"
    source_url = $datasetPage
    dataset_vintage = "2025-2035"
    effective_at = "2026-04-20T00:00:00Z"
    retrieved_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    license = "Open Government Licence - Canada"
    feature_count = $response.features.Count
    features = $response.features
}

$resolvedOutput = [System.IO.Path]::GetFullPath((Join-Path (Get-Location) $OutputPath))
$workspace = [System.IO.Path]::GetFullPath((Get-Location).Path) + [System.IO.Path]::DirectorySeparatorChar
if (-not $resolvedOutput.StartsWith($workspace, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Output path must remain inside the repository workspace."
}

$parent = [System.IO.Path]::GetDirectoryName($resolvedOutput)
[System.IO.Directory]::CreateDirectory($parent) | Out-Null
$json = $envelope | ConvertTo-Json -Depth 12
[System.IO.File]::WriteAllText($resolvedOutput, $json + [Environment]::NewLine, [System.Text.UTF8Encoding]::new($false))
Write-Host "Wrote $($response.features.Count) reviewed-source records to $resolvedOutput"
