#include "types.h"

void* memset(void *dest, int c, unsigned long n) {
	byte v = (byte)c;
	for (unsigned long i = 0; i < n; i++) {
		((byte*)dest)[i] = v;
	}
	return dest;
}

void *memcpy(void *dest, const void *src, unsigned long n)
{
    char *csrc = (char *)src;
    char *cdest = (char *)dest;

    for (unsigned long i = 0; i < n; i++)
    {
        cdest[i] = csrc[i];
    }

    return dest;
}