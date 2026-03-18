package model

// DirName is the subdirectory name used for the Pocket TTS model within a models directory.
const DirName = "pocket"

const modelURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/tts-models/sherpa-onnx-pocket-tts-int8-2026-01-26.tar.bz2"

// archiveHash is the expected SHA256 hash of the downloaded tar.bz2 archive.
const archiveHash = "2f3b88823cbbb9bf0b2477ec8ae7b3fec417b3a87b6bb5f256dba66f2ad967cb"

// Model file names for Pocket TTS.
const (
	FilenameLmMain          = "lm_main.int8.onnx"
	FilenameLmFlow          = "lm_flow.int8.onnx"
	FilenameDecoder         = "decoder.int8.onnx"
	FilenameEncoder         = "encoder.onnx"
	FilenameTextConditioner = "text_conditioner.onnx"
	FilenameVocabJSON       = "vocab.json"
	FilenameTokenScoresJSON = "token_scores.json"
)

// Filenames lists the filenames required for the Pocket TTS model.
var Filenames = []string{
	FilenameLmMain,
	FilenameLmFlow,
	FilenameDecoder,
	FilenameEncoder,
	FilenameTextConditioner,
	FilenameVocabJSON,
	FilenameTokenScoresJSON,
}

// fileHashes maps each model filename to its expected SHA256 hash.
var fileHashes = map[string]string{ //nolint:gosec // SHA256 hashes for model verification, not credentials
	FilenameLmMain:          "bfc0c7e7e3d72864fa3bb2ee499f62f21ddc1474b885f5f3ca570f8be73e787e",
	FilenameLmFlow:          "8d627d235c44a597da908e1085ebe241cbbe358964c502c5a5063d18851a5529",
	FilenameDecoder:         "12b0857402d31aead94df19d6783b4350d1f740e811f3a3202c70ad89ae11eea",
	FilenameEncoder:         "e8f2f6d301ffb96e398b138a7dc6d3038622d236044636b73d920bab85890260",
	FilenameTextConditioner: "0b84e837d7bfaf2c896627b03e3f080320309f37f4fc7df7698c644f7ba5e6b1",
	FilenameVocabJSON:       "6fb646346cf931016f70c4921aab0900ce7a304b893cb02135c74e294abfea01",
	FilenameTokenScoresJSON: "5be2f278caf9b9800741f0fd82bff677f4943ec764c356f907213434b622d958",
}
