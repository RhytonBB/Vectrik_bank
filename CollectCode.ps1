[System.Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$outputFile = "FullProjectCode.txt"
$rootDir = "C:\Users\iosir\Desktop\Vektrik_bank"
$excludedDirs = @("bin", "obj", "vendor", "node_modules")
$allowedExtensions = @(".go", ".html")

# Чистим файл перед началом
"=== COLLECTED GO PROJECT CODE ===`n`n" | Out-File -FilePath $outputFile -Encoding UTF8

function Should-SkipFile($file) {
    $filename = $file.Name.ToLower()
    return ($filename -eq "go.mod" -or $filename -eq "go.sum")
}

function Process-File($file) {
    $relativePath = $file.FullName.Substring($rootDir.Length)
    "`n" + ("=" * 80) + "`n" | Out-File -FilePath $outputFile -Append -Encoding UTF8
    "=== FILE: $relativePath ===`n" | Out-File -FilePath $outputFile -Append -Encoding UTF8
    "=" * 80 + "`n" | Out-File -FilePath $outputFile -Append -Encoding UTF8
    Get-Content -Path $file.FullName -Encoding UTF8 | Out-File -FilePath $outputFile -Append -Encoding UTF8
}

Get-ChildItem -Path $rootDir -Recurse -File | Where-Object {
    $_.Extension -in $allowedExtensions -and
    $excludedDirs -notcontains $_.Directory.Name -and
    -not (Should-SkipFile $_)
} | ForEach-Object {
    Process-File $_
}

Write-Host "✅ All .go files collected in $outputFile (UTF-8 encoded)"
