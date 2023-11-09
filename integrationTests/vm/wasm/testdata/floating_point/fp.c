typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

void int64finish(i64 value);

void doSomething() {
  i64 x = 6;
  float a = 1.0f;
  a = a + 0.3f;
  float q = x * a;
  
  i64 s = *(i64*)(&q);

  int64finish(s);
}

void init() {
}

void _main() {
}
