#!/usr/bin/env bash
# Télécharge le modèle BERT-NER multilingue quantisé (Xenova/bert-base-multilingual-cased-ner-hrl,
# base Davlan/bert-base-multilingual-cased-ner-hrl, afl-3.0) et la bibliothèque onnxruntime
# pour macOS arm64. Usage : ./scripts/download-model-macos.sh
#
# Multilingue plutôt que Xenova/bert-base-NER (anglais seul, cf. download-model.sh) : les
# noms propres à détecter dans les dossiers de courtage sont en français. Même architecture
# BERT/WordPiece (vocab.txt), donc compatible tel quel avec nlp/bertner sans changement de
# tokenizer — seul le modèle change, pas le code.
#
# Version onnxruntime 1.28.0 : la 1.26.0 (utilisée par le script Linux) répond une version
# d'API trop ancienne pour github.com/yalue/onnxruntime_go v1.31.0 ("API versions [1, 20]
# is not available, only [26]" — trouvé en vérifiant en réel sur cette machine).
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
model_dir="$root/models/bert-ner-multilingual"
ort_dir="$root/models/onnxruntime"
mkdir -p "$model_dir" "$ort_dir"

hf="https://huggingface.co/Xenova/bert-base-multilingual-cased-ner-hrl/resolve/main"
[ -f "$model_dir/model_quantized.onnx" ] || curl -L "$hf/onnx/model_quantized.onnx" -o "$model_dir/model_quantized.onnx"
[ -f "$model_dir/vocab.txt" ] || curl -L "$hf/vocab.txt" -o "$model_dir/vocab.txt"
[ -f "$model_dir/config.json" ] || curl -L "$hf/config.json" -o "$model_dir/config.json"
[ -f "$model_dir/tokenizer_config.json" ] || curl -L "$hf/tokenizer_config.json" -o "$model_dir/tokenizer_config.json"

ort_version="1.28.0"
if [ ! -f "$ort_dir/libonnxruntime.dylib" ]; then
  curl -L "https://github.com/microsoft/onnxruntime/releases/download/v${ort_version}/onnxruntime-osx-arm64-${ort_version}.tgz" | tar xz -C "$ort_dir"
  cp "$ort_dir/onnxruntime-osx-arm64-${ort_version}/lib/libonnxruntime.dylib" "$ort_dir/"
fi
echo "OK — modèle dans $model_dir, runtime dans $ort_dir"
echo 'Lancer : ONNXRUNTIME_LIB=models/onnxruntime/libonnxruntime.dylib go run -tags onnx ./examples/ner-fr'
