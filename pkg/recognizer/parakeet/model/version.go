package model

// DirName is the subdirectory name used for the Parakeet model within a models directory.
const DirName = "parakeet"

const modelURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-nemo-parakeet-tdt-0.6b-v2-int8.tar.bz2"

// archiveHash is the expected SHA256 hash of the downloaded tar.bz2 archive.
const archiveHash = "157c157bc51155e03e37d2466522a3a737dd9c72bb25f36eb18912964161e1ad"

// Filenames lists the filenames required for the Parakeet TDT model.
var Filenames = []string{
	"encoder.int8.onnx",
	"decoder.int8.onnx",
	"joiner.int8.onnx",
	"tokens.txt",
}

// fileHashes maps each model filename to its expected SHA256 hash.
var fileHashes = map[string]string{ //nolint:gosec // these are file integrity hashes, not credentials
	"encoder.int8.onnx": "a32b12d17bbbc309d0686fbbcc2987b5e9b8333a7da83fa6b089f0a2acd651ab",
	"decoder.int8.onnx": "b6bb64963457237b900e496ee9994b59294526439fbcc1fecf705b31a15c6b4e",
	"joiner.int8.onnx":  "7946164367946e7f9f29a122407c3252b680dbae9a51343eb2488d057c3c43d2",
	"tokens.txt":        "ec182b70dd42113aff6c5372c75cac58c952443eb22322f57bbd7f53977d497d",
}
