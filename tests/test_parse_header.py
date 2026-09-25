"""The cffi declarations exclude cgo runtime helpers that newer Go toolchains emit."""
from mlflow_go_backend.lib import _parse_header

HEADER = """
extern size_t _GoStringLen(_GoString_ s);
extern const char *_GoStringPtr(_GoString_ s);
extern __declspec(dllexport) int64_t CreateTrackingService(void* configData, int configSize);
extern void DestroyTrackingService(GoInt64 id);
"""


def test_parse_header_keeps_exports_and_drops_cgo_helpers(tmp_path):
    path = tmp_path / "libmlflow-go-backend.h"
    path.write_text(HEADER)
    parsed = _parse_header(path)
    assert "_GoString" not in parsed
    assert "extern int64_t CreateTrackingService(void* configData, int configSize);" in parsed
    assert "extern void DestroyTrackingService(int64_t id);" in parsed
