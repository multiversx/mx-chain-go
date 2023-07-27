#include "../mxvm/context.h"

byte counterKey[32] = {'m','y','c','o','u','n','t','e','r',0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};

void init() {
    int64storageStore(counterKey, 32, 1);
}

void upgrade() {
	init();
}

void increment() {
    i64 counter = int64storageLoad(counterKey, 32);
    counter++;
    int64storageStore(counterKey, 32, counter);
    int64finish(counter);
}

void decrement() {
    i64 counter = int64storageLoad(counterKey, 32);
    counter--;
    int64storageStore(counterKey, 32, counter);
    int64finish(counter);
}

void get() {
    i64 counter = int64storageLoad(counterKey, 32);
    int64finish(counter);
}
