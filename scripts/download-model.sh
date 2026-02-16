#!/usr/bin/env bash
set -euo pipefail

# Downloads and verifies the Parakeet TDT speech recognition model files.
# Called by the Makefile download-model target.

MODEL_DIR="pkg/recognizer/model"
MODEL_URL="https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-nemo-parakeet-tdt-0.6b-v2-int8.tar.bz2"
ARCHIVE="sherpa-onnx-nemo-parakeet-tdt-0.6b-v2-int8.tar.bz2"
EXTRACTED_DIR="sherpa-onnx-nemo-parakeet-tdt-0.6b-v2-int8"

MODEL_FILES=(
    "encoder.int8.onnx"
    "decoder.int8.onnx"
    "joiner.int8.onnx"
    "tokens.txt"
)

# SHA256 hashes of the extracted model files.
declare -A EXPECTED_HASHES
EXPECTED_HASHES=(
    ["encoder.int8.onnx"]="a32b12d17bbbc309d0686fbbcc2987b5e9b8333a7da83fa6b089f0a2acd651ab"
    ["decoder.int8.onnx"]="b6bb64963457237b900e496ee9994b59294526439fbcc1fecf705b31a15c6b4e"
    ["joiner.int8.onnx"]="7946164367946e7f9f29a122407c3252b680dbae9a51343eb2488d057c3c43d2"
    ["tokens.txt"]="ec182b70dd42113aff6c5372c75cac58c952443eb22322f57bbd7f53977d497d"
)

sha256() {
    if command -v sha256sum &>/dev/null; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum &>/dev/null; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        echo "ERROR: No SHA256 tool found (need sha256sum or shasum)" >&2
        exit 1
    fi
}

# Check if all model files already exist with correct hashes.
all_valid=true
for file in "${MODEL_FILES[@]}"; do
    path="${MODEL_DIR}/${file}"
    if [[ ! -f "$path" ]]; then
        all_valid=false
        break
    fi
    actual=$(sha256 "$path")
    if [[ "$actual" != "${EXPECTED_HASHES[$file]}" ]]; then
        echo "Hash mismatch for ${file}: expected ${EXPECTED_HASHES[$file]}, got ${actual}"
        all_valid=false
        break
    fi
done

if [[ "$all_valid" == "true" ]]; then
    echo "Model files already present and verified."
    exit 0
fi

echo "Downloading Parakeet TDT model..."
curl -fSL -o "$ARCHIVE" "$MODEL_URL"

echo "Extracting model files..."
tar -xjf "$ARCHIVE"

mkdir -p "$MODEL_DIR"
for file in "${MODEL_FILES[@]}"; do
    cp "${EXTRACTED_DIR}/${file}" "${MODEL_DIR}/"
done

echo "Verifying file hashes..."
for file in "${MODEL_FILES[@]}"; do
    path="${MODEL_DIR}/${file}"
    actual=$(sha256 "$path")
    expected="${EXPECTED_HASHES[$file]}"
    if [[ "$actual" != "$expected" ]]; then
        echo "ERROR: Hash mismatch for ${file}" >&2
        echo "  Expected: ${expected}" >&2
        echo "  Actual:   ${actual}" >&2
        rm -rf "$ARCHIVE" "$EXTRACTED_DIR"
        exit 1
    fi
    echo "  ${file}: OK"
done

rm -rf "$ARCHIVE" "$EXTRACTED_DIR"
echo "Model download complete."
