(module
  (type (;0;) (func (param i32 i32 i64) (result i32)))
  (type (;1;) (func (param i32 i32) (result i64)))
  (type (;2;) (func (param i64)))
  (type (;3;) (func))
  (import "env" "int64storageStore" (func (;0;) (type 0)))
  (import "env" "int64storageLoad" (func (;1;) (type 1)))
  (import "env" "int64finish" (func (;2;) (type 2)))
  (func (;3;) (type 3)
    i32.const 1024
    i32.const 32
    i64.const 1
    call 0
    drop)
  (func (;4;) (type 3)
    i32.const 1024
    i32.const 32
    i64.const 1
    call 0
    drop)
  (func (;5;) (type 3)
    (local i64)
    i32.const 1024
    i32.const 32
    i32.const 1024
    i32.const 32
    call 1
    i64.const 1
    i64.add
    local.tee 0
    call 0
    drop
    local.get 0
    call 2)
  (func (;6;) (type 3)
    (local i64)
    i32.const 1024
    i32.const 32
    i32.const 1024
    i32.const 32
    call 1
    i64.const -1
    i64.add
    local.tee 0
    call 0
    drop
    local.get 0
    call 2)
  (func (;7;) (type 3)
    i32.const 1024
    i32.const 32
    call 1
    call 2)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 2)
  (global (;0;) (mut i32) (i32.const 66592))
  (export "memory" (memory 0))
  (export "init" (func 3))
  (export "upgrade" (func 4))
  (export "increment" (func 5))
  (export "decrement" (func 6))
  (export "get" (func 7))
  (data (;0;) (i32.const 1024) "mycounter\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00"))
