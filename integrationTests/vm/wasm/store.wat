(module
  (type (;0;) (func (param i32 i32)))
  (type (;1;) (func))
  (import "ethereum" "storageStore" (func (;0;) (type 0)))
  (func (;1;) (type 1)
    i32.const 0
    i32.const 32
    call 0)
  (memory (;0;) 1)
  (export "memory" (memory 0))
  (export "main" (func 1))
  (data (i32.const 32) "\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\01"))