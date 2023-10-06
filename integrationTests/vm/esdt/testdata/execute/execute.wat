(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i64) (result i32)))
  (type (;2;) (func (param i32 i32) (result i32)))
  (type (;3;) (func (param i32 i32)))
  (type (;4;) (func (param i32 i32 i32 i32)))
  (type (;5;) (func (param i32) (result i64)))
  (type (;6;) (func (param i32 i32 i32) (result i32)))
  (type (;7;) (func (param i64)))
  (type (;8;) (func (param i32) (result i32)))
  (type (;9;) (func))
  (import "env" "mBufferNew" (func (;0;) (type 0)))
  (import "env" "bigIntNew" (func (;1;) (type 1)))
  (import "env" "mBufferGetArgument" (func (;2;) (type 2)))
  (import "env" "bigIntGetUnsignedArgument" (func (;3;) (type 3)))
  (import "env" "managedAsyncCall" (func (;4;) (type 4)))
  (import "env" "int64getArgument" (func (;5;) (type 5)))
  (import "env" "mBufferAppendBytes" (func (;6;) (type 6)))
  (import "env" "getNumArguments" (func (;7;) (type 0)))
  (import "env" "int64finish" (func (;8;) (type 7)))
  (import "env" "mBufferFinish" (func (;9;) (type 8)))
  (func (;10;) (type 9)
    (local i32 i32 i32 i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 0
    global.set 0
    call 0
    local.set 1
    i64.const 0
    call 1
    local.set 2
    call 0
    local.set 3
    i32.const 0
    local.get 1
    call 2
    drop
    i32.const 1
    local.get 2
    call 3
    local.get 0
    i32.const 3
    i32.store offset=12
    i32.const 2
    local.get 3
    call 2
    drop
    local.get 1
    local.get 2
    local.get 3
    local.get 0
    i32.const 12
    i32.add
    call 11
    call 4
    local.get 0
    i32.const 16
    i32.add
    global.set 0)
  (func (;11;) (type 8) (param i32) (result i32)
    (local i32 i32 i32 i32 i32 i32 i32 i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 1
    global.set 0
    call 0
    local.set 2
    local.get 0
    local.get 0
    i32.load
    local.tee 3
    i32.const 1
    i32.add
    i32.store
    block  ;; label = @1
      local.get 3
      call 5
      i32.wrap_i64
      local.tee 4
      i32.const 1
      i32.lt_s
      br_if 0 (;@1;)
      local.get 1
      i32.const 16
      i32.add
      local.set 5
      local.get 1
      i32.const 20
      i32.add
      local.set 6
      local.get 1
      i32.const 24
      i32.add
      local.set 7
      loop  ;; label = @2
        call 0
        local.set 3
        local.get 0
        local.get 0
        i32.load
        local.tee 8
        i32.const 1
        i32.add
        i32.store
        local.get 1
        local.get 3
        i32.store offset=12
        local.get 8
        local.get 3
        call 2
        drop
        local.get 2
        local.get 7
        i32.const 1
        call 6
        drop
        local.get 2
        local.get 6
        i32.const 1
        call 6
        drop
        local.get 2
        local.get 5
        i32.const 1
        call 6
        drop
        local.get 2
        local.get 1
        i32.const 12
        i32.add
        i32.const 1
        call 6
        drop
        local.get 4
        i32.const -1
        i32.add
        local.tee 4
        br_if 0 (;@2;)
      end
    end
    local.get 1
    i32.const 16
    i32.add
    global.set 0
    local.get 2)
  (func (;12;) (type 9)
    (local i32 i32 i32)
    call 7
    local.set 0
    i64.const 3390159555
    call 8
    block  ;; label = @1
      local.get 0
      i32.const 1
      i32.lt_s
      br_if 0 (;@1;)
      i32.const 0
      local.set 1
      loop  ;; label = @2
        local.get 1
        call 0
        local.tee 2
        call 2
        drop
        local.get 2
        call 9
        drop
        local.get 0
        local.get 1
        i32.const 1
        i32.add
        local.tee 1
        i32.ne
        br_if 0 (;@2;)
      end
    end
    i64.const 3390159555
    call 8)
  (func (;13;) (type 9)
    return)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 2)
  (global (;0;) (mut i32) (i32.const 66560))
  (export "memory" (memory 0))
  (export "doAsyncCall" (func 10))
  (export "callBack" (func 12))
  (export "init" (func 13)))
