; ModuleID = 'first/first.c'
source_filename = "first/first.c"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@counterKey = global [32 x i8] c"*\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00*", align 16

; Function Attrs: noinline nounwind optnone
define void @init() #0 {
entry:
  %call = call i32 @int64storageStore(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @counterKey, i32 0, i32 0), i64 0)
  ret void
}

declare i32 @int64storageStore(i8*, i64) #1

; Function Attrs: noinline nounwind optnone
define void @callBack() #0 {
entry:
  ret void
}

; Function Attrs: noinline nounwind optnone
define void @callMe() #0 {
entry:
  %counter = alloca i64, align 8
  %call = call i64 @int64storageLoad(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @counterKey, i32 0, i32 0))
  store i64 %call, i64* %counter, align 8
  %0 = load i64, i64* %counter, align 8
  %inc = add i64 %0, 1
  store i64 %inc, i64* %counter, align 8
  %1 = load i64, i64* %counter, align 8
  %call1 = call i32 @int64storageStore(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @counterKey, i32 0, i32 0), i64 %1)
  ret void
}

declare i64 @int64storageLoad(i8*) #1

; Function Attrs: noinline nounwind optnone
define void @numCalled() #0 {
entry:
  %counter = alloca i64, align 8
  %call = call i64 @int64storageLoad(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @counterKey, i32 0, i32 0))
  store i64 %call, i64* %counter, align 8
  %0 = load i64, i64* %counter, align 8
  call void @int64finish(i64 %0)
  ret void
}

declare void @int64finish(i64) #1

; Function Attrs: noinline nounwind optnone
define void @_main() #0 {
entry:
  ret void
}

attributes #0 = { noinline nounwind optnone "correctly-rounded-divide-sqrt-fp-math"="false" "disable-tail-calls"="false" "less-precise-fpmad"="false" "min-legal-vector-width"="0" "no-frame-pointer-elim"="false" "no-infs-fp-math"="false" "no-jump-tables"="false" "no-nans-fp-math"="false" "no-signed-zeros-fp-math"="false" "no-trapping-math"="false" "stack-protector-buffer-size"="8" "unsafe-fp-math"="false" "use-soft-float"="false" }
attributes #1 = { "correctly-rounded-divide-sqrt-fp-math"="false" "disable-tail-calls"="false" "less-precise-fpmad"="false" "no-frame-pointer-elim"="false" "no-infs-fp-math"="false" "no-nans-fp-math"="false" "no-signed-zeros-fp-math"="false" "no-trapping-math"="false" "stack-protector-buffer-size"="8" "unsafe-fp-math"="false" "use-soft-float"="false" }

!llvm.module.flags = !{!0}
!llvm.ident = !{!1}

!0 = !{i32 1, !"wchar_size", i32 4}
!1 = !{!"clang version 8.0.0 (Fedora 8.0.0-3.fc30)"}
