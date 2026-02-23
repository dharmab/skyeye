//go:build windows

package parakeet

//go:generate sh -c "cd \"$SHERPA_LIB\" && gendef sherpa-onnx-c-api.dll && dlltool -d sherpa-onnx-c-api.def -l libsherpa-onnx-c-api.dll.a && gendef onnxruntime.dll && dlltool -d onnxruntime.def -l libonnxruntime.dll.a && gendef sherpa-onnx-cxx-api.dll && dlltool -d sherpa-onnx-cxx-api.def -l libsherpa-onnx-cxx-api.dll.a"
