//go:build windows

package parakeet

//go:generate sh -c "SHERPA_LIB=$(go list -m -json github.com/k2-fsa/sherpa-onnx-go-windows 2>/dev/null | grep '\"Dir\"' | cut -d'\"' -f4)/lib/x86_64-pc-windows-gnu && cd \"$SHERPA_LIB\" && gendef sherpa-onnx-c-api.dll && dlltool -d sherpa-onnx-c-api.def -l libsherpa-onnx-c-api.dll.a && gendef onnxruntime.dll && dlltool -d onnxruntime.def -l libonnxruntime.dll.a && gendef sherpa-onnx-cxx-api.dll && dlltool -d sherpa-onnx-cxx-api.def -l libsherpa-onnx-cxx-api.dll.a"
