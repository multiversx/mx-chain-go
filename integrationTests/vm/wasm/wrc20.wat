(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i32)))
  (type (;2;) (func (param i32 i32 i32)))
  (type (;3;) (func (param i32)))
  (type (;4;) (func))
  (type (;5;) (func (param i32 i32) (result i32)))
  (import "env" "getCallDataSize" (func $getCallDataSize (type 0)))
  (import "env" "revert" (func $revert (type 1)))
  (import "env" "callDataCopy" (func $callDataCopy (type 2)))
  (import "env" "finish" (func $finish (type 1)))
  (import "env" "storageLoad" (func $storageLoad (type 1)))
  (import "env" "getCaller" (func $getCaller (type 3)))
  (import "env" "storageStore" (func $storageStore (type 1)))
  (func $__wasm_call_ctors (type 4))
  (func $main (type 5) (param i32 i32) (result i32)
    i32.const 0)
  (func $main.1 (type 4)
    (local i32 i32)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 0
    global.set 0
    block  ;; label = @1
      block  ;; label = @2
        call $getCallDataSize
        i32.const 3
        i32.le_s
        br_if 0 (;@2;)
        local.get 0
        i32.const 0
        i32.store offset=12
        local.get 0
        i32.const 12
        i32.add
        call $callDataCopy_2O1SXmzMBKQV9cWGnElsimg
        block  ;; label = @3
          local.get 0
          i32.load offset=12
          local.tee 1
          i32.const 1563795389
          i32.ne
          br_if 0 (;@3;)
          call $do_transfer_E82kPpU5OcEfVOGiDsEd5g_2
          local.get 0
          i32.const 16
          i32.add
          global.set 0
          return
        end
        local.get 1
        i32.const -1718418918
        i32.ne
        br_if 1 (;@1;)
        call $do_balance_E82kPpU5OcEfVOGiDsEd5g
        unreachable
      end
      i32.const 0
      i32.const 0
      call $revert
      unreachable
    end
    i32.const 0
    i32.const 0
    call $revert
    unreachable)
  (func $callDataCopy_2O1SXmzMBKQV9cWGnElsimg (type 3) (param i32)
    local.get 0
    i32.const 0
    i32.const 4
    call $callDataCopy)
  (func $do_transfer_E82kPpU5OcEfVOGiDsEd5g_2 (type 4)
    (local i32 i32 i32 i64 i64)
    global.get 0
    i32.const 144
    i32.sub
    local.tee 0
    global.set 0
    block  ;; label = @1
      block  ;; label = @2
        call $getCallDataSize
        i32.const 32
        i32.ne
        br_if 0 (;@2;)
        local.get 0
        i32.const 120
        i32.add
        i32.const 16
        i32.add
        i32.const 0
        i32.store
        local.get 0
        i32.const 120
        i32.add
        i32.const 8
        i32.add
        i64.const 0
        i64.store
        local.get 0
        i64.const 0
        i64.store offset=120
        local.get 0
        i32.const 120
        i32.add
        call $getCaller
        local.get 0
        i32.const 96
        i32.add
        i32.const 16
        i32.add
        i32.const 0
        i32.store
        local.get 0
        i32.const 96
        i32.add
        i32.const 8
        i32.add
        i64.const 0
        i64.store
        local.get 0
        i64.const 0
        i64.store offset=96
        local.get 0
        i32.const 96
        i32.add
        call $callDataCopy_AkbZvPjJcvpe3TuWowJZAQ
        local.get 0
        i64.const 0
        i64.store offset=88
        local.get 0
        i32.const 88
        i32.add
        call $callDataCopy_Bbl6Oevm89cvsaRHbd6mdjQ
        local.get 0
        i32.const 56
        i32.add
        i32.const 24
        i32.add
        local.tee 1
        i64.const 0
        i64.store
        local.get 0
        i32.const 56
        i32.add
        i32.const 16
        i32.add
        i64.const 0
        i64.store
        local.get 0
        i32.const 56
        i32.add
        i32.const 8
        i32.add
        i64.const 0
        i64.store
        local.get 0
        i64.const 0
        i64.store offset=56
        local.get 0
        i32.const 120
        i32.add
        local.get 0
        i32.const 56
        i32.add
        call $storageLoad_WhAluFFocP9b5GuMCan9b06w
        local.get 0
        i32.const 24
        i32.add
        i32.const 24
        i32.add
        local.tee 2
        i64.const 0
        i64.store
        local.get 0
        i32.const 24
        i32.add
        i32.const 16
        i32.add
        i64.const 0
        i64.store
        local.get 0
        i32.const 24
        i32.add
        i32.const 8
        i32.add
        i64.const 0
        i64.store
        local.get 0
        i64.const 0
        i64.store offset=24
        local.get 0
        i32.const 96
        i32.add
        local.get 0
        i32.const 24
        i32.add
        call $storageLoad_WhAluFFocP9b5GuMCan9b06w
        local.get 0
        i64.const 0
        i64.store offset=8
        local.get 0
        i64.const 0
        i64.store offset=16
        local.get 0
        i64.const 0
        i64.store
        local.get 0
        local.get 0
        i32.const 88
        i32.add
        call $bigEndian64_IzdisrH4sYnsItUtxSkomA
        local.get 0
        i32.const 16
        i32.add
        local.get 1
        call $bigEndian64_IzdisrH4sYnsItUtxSkomA
        local.get 0
        i64.load offset=16
        local.tee 3
        local.get 0
        i64.load
        local.tee 4
        i64.lt_u
        br_if 1 (;@1;)
        local.get 0
        i32.const 8
        i32.add
        local.get 2
        call $bigEndian64_IzdisrH4sYnsItUtxSkomA
        local.get 0
        local.get 3
        local.get 4
        i64.sub
        i64.store offset=16
        local.get 0
        local.get 0
        i64.load offset=8
        local.get 4
        i64.add
        i64.store offset=8
        local.get 1
        local.get 0
        i32.const 16
        i32.add
        call $bigEndian64_IzdisrH4sYnsItUtxSkomA
        local.get 2
        local.get 0
        i32.const 8
        i32.add
        call $bigEndian64_IzdisrH4sYnsItUtxSkomA
        local.get 0
        i32.const 120
        i32.add
        local.get 0
        i32.const 56
        i32.add
        call $storageStore_WhAluFFocP9b5GuMCan9b06w_2
        local.get 0
        i32.const 96
        i32.add
        local.get 0
        i32.const 24
        i32.add
        call $storageStore_WhAluFFocP9b5GuMCan9b06w_2
        local.get 0
        i32.const 144
        i32.add
        global.set 0
        return
      end
      i32.const 0
      i32.const 0
      call $revert
      unreachable
    end
    i32.const 0
    i32.const 0
    call $revert
    unreachable)
  (func $do_balance_E82kPpU5OcEfVOGiDsEd5g (type 4)
    (local i32)
    global.get 0
    i32.const 64
    i32.sub
    local.tee 0
    global.set 0
    block  ;; label = @1
      call $getCallDataSize
      i32.const 24
      i32.ne
      br_if 0 (;@1;)
      local.get 0
      i32.const 40
      i32.add
      i32.const 16
      i32.add
      i32.const 0
      i32.store
      local.get 0
      i32.const 40
      i32.add
      i32.const 8
      i32.add
      i64.const 0
      i64.store
      local.get 0
      i64.const 0
      i64.store offset=40
      local.get 0
      i32.const 40
      i32.add
      call $callDataCopy_AkbZvPjJcvpe3TuWowJZAQ
      local.get 0
      i32.const 8
      i32.add
      i32.const 24
      i32.add
      i64.const 0
      i64.store
      local.get 0
      i32.const 8
      i32.add
      i32.const 16
      i32.add
      i64.const 0
      i64.store
      local.get 0
      i32.const 8
      i32.add
      i32.const 8
      i32.add
      i64.const 0
      i64.store
      local.get 0
      i64.const 0
      i64.store offset=8
      local.get 0
      i32.const 40
      i32.add
      local.get 0
      i32.const 8
      i32.add
      call $storageLoad_WhAluFFocP9b5GuMCan9b06w
      local.get 0
      i32.const 8
      i32.add
      i32.const 32
      call $finish
      unreachable
    end
    i32.const 0
    i32.const 0
    call $revert
    unreachable)
  (func $callDataCopy_AkbZvPjJcvpe3TuWowJZAQ (type 3) (param i32)
    local.get 0
    i32.const 4
    i32.const 20
    call $callDataCopy)
  (func $storageLoad_WhAluFFocP9b5GuMCan9b06w (type 1) (param i32 i32)
    (local i32)
    global.get 0
    i32.const 32
    i32.sub
    local.tee 2
    global.set 0
    local.get 2
    i32.const 24
    i32.add
    i64.const 0
    i64.store
    local.get 2
    i32.const 16
    i32.add
    i64.const 0
    i64.store
    local.get 2
    i32.const 8
    i32.add
    i64.const 0
    i64.store
    local.get 2
    i32.const 20
    i32.add
    local.get 0
    i32.const 8
    i32.add
    i64.load align=1
    i64.store align=4
    local.get 2
    i32.const 28
    i32.add
    local.get 0
    i32.const 16
    i32.add
    i32.load align=1
    i32.store
    local.get 2
    i64.const 0
    i64.store
    local.get 2
    local.get 0
    i64.load align=1
    i64.store offset=12 align=4
    local.get 2
    local.get 1
    call $storageLoad
    local.get 2
    i32.const 32
    i32.add
    global.set 0)
  (func $callDataCopy_Bbl6Oevm89cvsaRHbd6mdjQ (type 3) (param i32)
    local.get 0
    i32.const 24
    i32.const 8
    call $callDataCopy)
  (func $bigEndian64_IzdisrH4sYnsItUtxSkomA (type 1) (param i32 i32)
    local.get 0
    local.get 1
    call $swapEndian64_IzdisrH4sYnsItUtxSkomA_2)
  (func $storageStore_WhAluFFocP9b5GuMCan9b06w_2 (type 1) (param i32 i32)
    (local i32)
    global.get 0
    i32.const 32
    i32.sub
    local.tee 2
    global.set 0
    local.get 2
    i32.const 24
    i32.add
    i64.const 0
    i64.store
    local.get 2
    i32.const 16
    i32.add
    i64.const 0
    i64.store
    local.get 2
    i32.const 8
    i32.add
    i64.const 0
    i64.store
    local.get 2
    i32.const 20
    i32.add
    local.get 0
    i32.const 8
    i32.add
    i64.load align=1
    i64.store align=4
    local.get 2
    i32.const 28
    i32.add
    local.get 0
    i32.const 16
    i32.add
    i32.load align=1
    i32.store
    local.get 2
    i64.const 0
    i64.store
    local.get 2
    local.get 0
    i64.load align=1
    i64.store offset=12 align=4
    local.get 2
    local.get 1
    call $storageStore
    local.get 2
    i32.const 32
    i32.add
    global.set 0)
  (func $swapEndian64_IzdisrH4sYnsItUtxSkomA_2 (type 1) (param i32 i32)
    (local i32 i64)
    global.get 0
    i32.const 16
    i32.sub
    local.tee 2
    global.set 0
    local.get 2
    i64.const 0
    i64.store offset=8
    local.get 2
    i32.const 8
    i32.add
    local.get 1
    call $copyMem_fPlwH3r9agN9aEHB6yCPMh0w
    local.get 2
    local.get 2
    i64.load offset=8
    local.tee 3
    i64.const 56
    i64.shl
    local.get 3
    i64.const 40
    i64.shl
    i64.const 71776119061217280
    i64.and
    i64.or
    local.get 3
    i64.const 24
    i64.shl
    i64.const 280375465082880
    i64.and
    local.get 3
    i64.const 8
    i64.shl
    i64.const 1095216660480
    i64.and
    i64.or
    i64.or
    local.get 3
    i64.const 8
    i64.shr_u
    i64.const 4278190080
    i64.and
    local.get 3
    i64.const 24
    i64.shr_u
    i64.const 16711680
    i64.and
    i64.or
    local.get 3
    i64.const 40
    i64.shr_u
    i64.const 65280
    i64.and
    local.get 3
    i64.const 56
    i64.shr_u
    i64.or
    i64.or
    i64.or
    i64.store offset=8
    local.get 0
    local.get 2
    i32.const 8
    i32.add
    call $copyMem_fPlwH3r9agN9aEHB6yCPMh0w
    local.get 2
    i32.const 16
    i32.add
    global.set 0)
  (func $copyMem_fPlwH3r9agN9aEHB6yCPMh0w (type 1) (param i32 i32)
    local.get 0
    local.get 1
    i64.load align=1
    i64.store align=1)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 2)
  (global (;0;) (mut i32) (i32.const 66576))
  (global (;1;) i32 (i32.const 66576))
  (global (;2;) i32 (i32.const 1028))
  (global (;3;) i32 (i32.const 1024))
  (export "memory" (memory 0))
  (export "__heap_base" (global 1))
  (export "__data_end" (global 2))
  (export "main" (func $main))
  (export "main.1" (func $main.1))
  (export "nim_program_result" (global 3))
  (data (;0;) (i32.const 1024) "\00\00\00\00"))