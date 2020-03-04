typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

void signalError(char *msg, int length);
void int64finish(i64 value);
i64 int64getArgument(int argumentIndex);

void testFunc() {
  i64 arg = int64getArgument(0);

  if (arg == 1) {
    int64finish(98);
    byte msg[] = "abort here";
    signalError(msg, 10);
    int64finish(99);
  } else {
    int64finish(100);
  }
}

void init() {
}

void _main() {
}
