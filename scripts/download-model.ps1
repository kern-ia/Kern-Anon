# Telecharge le modele BERT-NER quantise (Xenova/bert-base-NER, MIT) et la
# bibliotheque onnxruntime pour Windows. Usage : ./scripts/download-model.ps1
$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
$modelDir = Join-Path $root "models\bert-ner"
$ortDir = Join-Path $root "models\onnxruntime"
New-Item -ItemType Directory -Force $modelDir | Out-Null
New-Item -ItemType Directory -Force $ortDir | Out-Null

$hf = "https://huggingface.co/Xenova/bert-base-NER/resolve/main"
$files = @{
    "$hf/onnx/model_quantized.onnx" = Join-Path $modelDir "model_quantized.onnx"
    "$hf/vocab.txt"                 = Join-Path $modelDir "vocab.txt"
    "$hf/config.json"               = Join-Path $modelDir "config.json"
}
foreach ($url in $files.Keys) {
    $dest = $files[$url]
    if (-not (Test-Path $dest)) {
        Write-Host "Telechargement $url"
        Invoke-WebRequest -Uri $url -OutFile $dest
    }
}

$ortVersion = "1.26.0"
$ortZip = Join-Path $ortDir "onnxruntime.zip"
$dll = Join-Path $ortDir "onnxruntime.dll"
if (-not (Test-Path $dll)) {
    $url = "https://github.com/microsoft/onnxruntime/releases/download/v$ortVersion/onnxruntime-win-x64-$ortVersion.zip"
    Write-Host "Telechargement $url"
    Invoke-WebRequest -Uri $url -OutFile $ortZip
    Expand-Archive $ortZip -DestinationPath $ortDir -Force
    Copy-Item (Join-Path $ortDir "onnxruntime-win-x64-$ortVersion\lib\*.dll") $ortDir
    Remove-Item $ortZip -Confirm:$false
}
Write-Host "OK - modele dans $modelDir, runtime dans $ortDir"
Write-Host 'Lancer : $env:ONNXRUNTIME_LIB="models\onnxruntime\onnxruntime.dll"; go run -tags onnx ./examples/ner'
