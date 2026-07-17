#!/usr/bin/env bash
# Télécharge le modèle BERT-NER quantisé (Xenova/bert-base-NER, MIT) et la
# bibliothèque onnxruntime pour Linux. Usage : ./scripts/download-model.sh
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
model_dir="$root/models/bert-ner"
ort_dir="$root/models/onnxruntime"
mkdir -p "$model_dir" "$ort_dir"

hf="https://huggingface.co/Xenova/bert-base-NER/resolve/main"
[ -f "$model_dir/model_quantized.onnx" ] || curl -L "$hf/onnx/model_quantized.onnx" -o "$model_dir/model_quantized.onnx"
[ -f "$model_dir/vocab.txt" ] || curl -L "$hf/vocab.txt" -o "$model_dir/vocab.txt"
[ -f "$model_dir/config.json" ] || curl -L "$hf/config.json" -o "$model_dir/config.json"

ort_version="1.26.0"
if [ ! -f "$ort_dir/libonnxruntime.so" ]; then
  curl -L "https://github.com/microsoft/onnxruntime/releases/download/v${ort_version}/onnxruntime-linux-x64-${ort_version}.tgz" | tar xz -C "$ort_dir"
  cp "$ort_dir/onnxruntime-linux-x64-${ort_version}/lib/"libonnxruntime.so* "$ort_dir/"
fi
echo "OK — modèle dans $model_dir, runtime dans $ort_dir"
echo 'Lancer : ONNXRUNTIME_LIB=models/onnxruntime/libonnxruntime.so go run -tags onnx ./examples/ner'
