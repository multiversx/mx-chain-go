(module
  (type (;0;) (func (param i64) (result i32)))
  (type (;1;) (func (param i32)))
  (type (;2;) (func (param i32) (result i64)))
  (type (;3;) (func (param i64) (result i64)))
  (type (;4;) (func))
  (import "env" "bigIntNew" (func (;0;) (type 0)))
  (import "env" "bigIntGetCallValue" (func (;1;) (type 1)))
  (import "env" "bigIntGetInt64" (func (;2;) (type 2)))
  (import "env" "bigIntFinishUnsigned" (func (;3;) (type 1)))
  (func (;4;) (type 3) (param i64) (result i64)
    block  ;; label = @1
      block  ;; label = @2
        block  ;; label = @3
          local.get 0
          i64.const 1
          i64.gt_u
          br_if 0 (;@3;)
          local.get 0
          i32.wrap_i64
          br_table 2 (;@1;) 1 (;@2;) 2 (;@1;)
        end
        local.get 0
        i64.const -1
        i64.add
        call 4
        local.get 0
        i64.const -2
        i64.add
        call 4
        i64.add
        return
      end
      i64.const 1
      local.set 0
    end
    local.get 0)
  (func (;5;) (type 4)
    (local i32)
    i64.const 0
    call 0
    local.tee 0
    call 1
    local.get 0
    call 2
    call 4
    call 0
    call 3)
  (func (;6;) (type 4)
    return)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 2)
  (global (;0;) (mut i32) (i32.const 66560))
  (export "memory" (memory 0))
  (export "_main" (func 5))
  (export "init" (func 6)))
