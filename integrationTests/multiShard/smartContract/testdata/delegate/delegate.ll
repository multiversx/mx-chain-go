; ModuleID = 'testdata/delegate/delegate.c'
source_filename = "testdata/delegate/delegate.c"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@totalStakeKey = global [32 x i8] c"*\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00*", align 16
@totalStakeBytes = global <{ i8, [31 x i8] }> <{ i8 42, [31 x i8] zeroinitializer }>, align 16
@stakingSc = global [32 x i8] c"\00\00\00\00\00\00\00\00\00\01\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\FF\FF", align 16
@callValue = global [32 x i8] zeroinitializer, align 16
@data = global [262 x i8] c"stake@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", align 16

; Function Attrs: noinline nounwind optnone
define void @delegate() #0 {
entry:
  %stake = alloca i64, align 8
  %totalStake = alloca i64, align 8
  %call = call i64 @int64getArgument(i32 0)
  store i64 %call, i64* %stake, align 8
  %call1 = call i64 @int64storageLoad(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @totalStakeKey, i32 0, i32 0))
  store i64 %call1, i64* %totalStake, align 8
  %0 = load i64, i64* %stake, align 8
  %1 = load i64, i64* %totalStake, align 8
  %add = add i64 %1, %0
  store i64 %add, i64* %totalStake, align 8
  %2 = load i64, i64* %totalStake, align 8
  %call2 = call i32 @int64storageStore(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @totalStakeKey, i32 0, i32 0), i64 %2)
  ret void
}

declare i64 @int64getArgument(i32) #1

declare i64 @int64storageLoad(i8*) #1

declare i32 @int64storageStore(i8*, i64) #1

; Function Attrs: noinline nounwind optnone
define void @sendToStaking() #0 {
entry:
  %call = call i32 @getCallValue(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @callValue, i32 0, i32 0))
  call void @asyncCall(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @stakingSc, i32 0, i32 0), i8* getelementptr inbounds ([32 x i8], [32 x i8]* @callValue, i32 0, i32 0), i8* getelementptr inbounds ([262 x i8], [262 x i8]* @data, i32 0, i32 0), i32 262)
  ret void
}

declare i32 @getCallValue(i8*) #1

declare void @asyncCall(i8*, i8*, i8*, i32) #1

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
!1 = !{!"clang version 9.0.0 (Fedora 9.0.0-1.fc31)"}
