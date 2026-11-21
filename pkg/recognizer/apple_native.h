#ifndef SKYEYE_APPLE_SPEECH_H
#define SKYEYE_APPLE_SPEECH_H
#include <stddef.h>
// Returned strings are malloc-owned; handles are released by apple_release.
char *apple_preflight(void);
void *apple_begin(const float *pcm, size_t count, double rate, const char *hints);
int apple_poll(void *handle, char **text, char **error);
void apple_release(void *handle);
#endif
