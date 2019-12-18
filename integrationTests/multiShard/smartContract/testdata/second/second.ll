; ModuleID = 'second/second.c'
source_filename = "second/second.c"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@zero = global [32 x i8] zeroinitializer, align 16
@firstScAddress = global [32 x i8] c"\00\00\00\00\00\00\00\00\05\00]=S\B5\D0\FC\F0}\22!p\97\892\16n\E9\F3\97-00", align 16
@.str = private unnamed_addr constant [10 x i8] c"callMe@01\00", align 1

; Function Attrs: noinline nounwind optnone
define void @doSomething() #0 {
entry:
  %call = call i32 @transferValue(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @firstScAddress, i32 0, i32 0), i8* getelementptr inbounds ([32 x i8], [32 x i8]* @zero, i32 0, i32 0), i8* getelementptr inbounds ([10 x i8], [10 x i8]* @.str, i32 0, i32 0), i32 9)
  ret void
}

declare i32 @transferValue(i8*, i8*, i8*, i32) #1

; Function Attrs: noinline nounwind optnone
define void @callBack() #0 {
entry:
  ret void
}

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
