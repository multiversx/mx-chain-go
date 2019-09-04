(module
  (type $t0 (func (param i32 i32 i32)))
  (type $t1 (func (result i32)))
  (type $t2 (func (param i32)))
  (type $t3 (func (param i32 i32)))
  (type $t4 (func))
  (import "ethereum" "callDataCopy" (func $env.callDataCopy (type $t0)))
  (import "ethereum" "getCallDataSize" (func $env.getCallDataSize (type $t1)))
  (import "ethereum" "getCaller" (func $env.getCaller (type $t2)))
  (import "ethereum" "revert" (func $env.revert (type $t3)))
  (import "ethereum" "storageStore" (func $env.storageStore (type $t3)))
  (table $T0 0 anyfunc)
  (func $main (export "main") (type $t4)
    (local $l0 i32) (local $l1 i32) (local $l2 i32) (local $l3 i32)
    (i32.store offset=4
      (i32.const 0)
      (tee_local $l3
        (i32.sub
          (i32.load offset=4
            (i32.const 0))
          (i32.const 192))))
    (i32.store align=1
      (get_local $l3)
      (i32.const 436376473))
    (i32.store offset=4 align=1
      (get_local $l3)
      (i32.const -1113639587))
    (block $B0
      (block $B1
        (br_if $B1
          (i32.gt_u
            (tee_local $l0
              (call $env.getCallDataSize))
            (i32.const 3)))
        (i32.store offset=160
          (get_local $l3)
          (i32.const 0))
        (call $env.revert
          (i32.add
            (get_local $l3)
            (i32.const 160))
          (i32.const 0))
        (br $B0))
      (i32.store offset=8
        (get_local $l3)
        (i32.const 0))
      (call $env.callDataCopy
        (i32.add
          (get_local $l3)
          (i32.const 8))
        (i32.const 0)
        (i32.const 4))
      (i64.store align=4
        (i32.add
          (get_local $l3)
          (i32.const 36))
        (i64.const 0))
      (i64.store align=4
        (i32.add
          (get_local $l3)
          (i32.const 28))
        (i64.const 0))
      (i64.store align=4
        (i32.add
          (get_local $l3)
          (i32.const 20))
        (i64.const 0))
      (i64.store offset=12 align=4
        (get_local $l3)
        (i64.const 0))
      (set_local $l2
        (i32.const 93))
      (set_local $l1
        (i32.const 1))
      (block $B2
        (block $B3
          (loop $L4
            (br_if $B3
              (i32.ne
                (i32.load8_u
                  (i32.add
                    (i32.add
                      (i32.add
                        (get_local $l3)
                        (i32.const 8))
                      (get_local $l1))
                    (i32.const -1)))
                (i32.and
                  (get_local $l2)
                  (i32.const 255))))
            (block $B5
              (br_if $B5
                (i32.gt_u
                  (get_local $l1)
                  (i32.const 3)))
              (set_local $l2
                (i32.load8_u
                  (i32.add
                    (i32.add
                      (get_local $l3)
                      (i32.const 4))
                    (get_local $l1))))
              (set_local $l1
                (i32.add
                  (get_local $l1)
                  (i32.const 1)))
              (br $L4)))
          (br_if $B2
            (i32.ne
              (get_local $l0)
              (i32.const 32)))
          (i32.store
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 44))
              (i32.const 16))
            (i32.const 0))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 44))
              (i32.const 8))
            (i64.const 0))
          (i64.store offset=44 align=4
            (get_local $l3)
            (i64.const 0))
          (call $env.getCaller
            (i32.add
              (get_local $l3)
              (i32.const 44)))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 64))
              (i32.const 24))
            (i64.const 0))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 64))
              (i32.const 16))
            (i64.const 0))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 64))
              (i32.const 8))
            (i64.const 0))
          (i64.store offset=64 align=4
            (get_local $l3)
            (i64.const 0))
          (call $env.callDataCopy
            (i32.add
              (get_local $l3)
              (i32.const 64))
            (i32.const 4)
            (i32.const 20))
          (i64.store offset=96 align=4
            (get_local $l3)
            (i64.const 1090921693438))
          (i64.store offset=104 align=4
            (get_local $l3)
            (i64.const 1090921693438))
          (i64.store offset=112 align=4
            (get_local $l3)
            (i64.const 0))
          (i64.store offset=120 align=4
            (get_local $l3)
            (i64.const 0))
          (call $env.storageStore
            (i32.add
              (get_local $l3)
              (i32.const 96))
            (i32.add
              (get_local $l3)
              (i32.const 64)))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 128))
              (i32.const 24))
            (i64.const 0))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 128))
              (i32.const 16))
            (i64.const 0))
          (i64.store align=4
            (i32.add
              (i32.add
                (get_local $l3)
                (i32.const 128))
              (i32.const 8))
            (i64.const 0))
          (i64.store offset=128 align=4
            (get_local $l3)
            (i64.const 0))
          (call $env.callDataCopy
            (i32.add
              (get_local $l3)
              (i32.const 128))
            (i32.const 24)
            (i32.const 8))
          (i64.store offset=160 align=4
            (get_local $l3)
            (i64.const 1095216660735))
          (i64.store offset=168 align=4
            (get_local $l3)
            (i64.const 1095216660735))
          (i64.store offset=176 align=4
            (get_local $l3)
            (i64.const 0))
          (i64.store offset=184 align=4
            (get_local $l3)
            (i64.const 0))
          (call $env.storageStore
            (i32.add
              (get_local $l3)
              (i32.const 160))
            (i32.add
              (get_local $l3)
              (i32.const 128)))
          (br $B0))
        (i64.store offset=160 align=1
          (get_local $l3)
          (i64.const 651061555525847048))
        (set_local $l2
          (i32.const 153))
        (set_local $l1
          (i32.const 1))
        (loop $L6
          (br_if $B0
            (i32.ne
              (i32.load8_u
                (i32.add
                  (i32.add
                    (i32.add
                      (get_local $l3)
                      (i32.const 8))
                    (get_local $l1))
                  (i32.const -1)))
              (i32.and
                (get_local $l2)
                (i32.const 255))))
          (block $B7
            (br_if $B7
              (i32.gt_u
                (get_local $l1)
                (i32.const 3)))
            (set_local $l2
              (i32.load8_u
                (i32.add
                  (get_local $l3)
                  (get_local $l1))))
            (set_local $l1
              (i32.add
                (get_local $l1)
                (i32.const 1)))
            (br $L6)))
        (call $env.storageStore
          (i32.add
            (get_local $l3)
            (i32.const 12))
          (i32.add
            (get_local $l3)
            (i32.const 160)))
        (br $B0))
      (i32.store offset=160
        (get_local $l3)
        (i32.const 0))
      (call $env.revert
        (i32.add
          (get_local $l3)
          (i32.const 160))
        (i32.const 0)))
    (i32.store offset=4
      (i32.const 0)
      (i32.add
        (get_local $l3)
        (i32.const 192))))
  (memory $memory (export "memory") 17)
  (data (i32.const 4) "\10\00\10\00"))
