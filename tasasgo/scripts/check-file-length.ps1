<#
    Audita el límite de líneas por archivo del proyecto TasasGo.

    Regla: los archivos de producción no pueden salir del rango 30-60 líneas.
    Los *_test.go tienen un techo propio de 100 líneas para poder alojar tablas
    de casos sin fragmentarse de forma ilegible.

    Cuenta TODAS las líneas del archivo, incluidas las vacías, los comentarios y
    las llaves de cierre, porque la regla aplica al archivo completo y no al
    "código útil".

    Uso:  .\scripts\check-file-length.ps1
    Salida: tabla de archivos y código de salida 1 si alguno incumple.
#>
param(
    [string]$Root = "$PSScriptRoot/.."
)

$failures = @()

$rows = foreach ($file in Get-ChildItem -Path $Root -Recurse -Filter *.go) {
    $isTest = $file.Name -like '*_test.go'
    $min = 30
    $max = 60
    if ($isTest) {
        $min = 0
        $max = 100
    }

    $lines = @(Get-Content -LiteralPath $file.FullName).Count
    $status = 'OK'
    if ($lines -lt $min -or $lines -gt $max) {
        $status = 'FUERA DE RANGO'
        $failures += ('{0}: {1} lineas (permitido {2}-{3})' -f $file.Name, $lines, $min, $max)
    }

    [pscustomobject]@{
        Archivo = $file.Name
        Lineas  = $lines
        Min     = $min
        Max     = $max
        Estado  = $status
    }
}

$rows | Sort-Object Lineas | Format-Table -AutoSize

if ($failures.Count -gt 0) {
    Write-Host "`nIncumplimientos detectados:" -ForegroundColor Red
    $failures | ForEach-Object { Write-Host "  - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "`nTodos los archivos cumplen el limite de lineas." -ForegroundColor Green
exit 0
