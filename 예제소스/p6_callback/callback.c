#include "_cgo_export.h"

void run_with_callback(void) {
    for (int i = 0; i < 3; i++) {
        goCallback(i);
    }
}
