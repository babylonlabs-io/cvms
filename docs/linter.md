# golangci-lint description

## EN

| Linter                        | Description                                                                              |
| :---------------------------- | :--------------------------------------------------------------------------------------- |
| **asasalint**                 | Checks issues when passing []any to variadic func(...any).                               |
| **asciicheck**                | Warns if the code contains non-ASCII characters (like Korean, Japanese, etc.).           |
| **bidichk**                   | Detects dangerous Unicode BiDi (Bidirectional) sequences that can cause security issues. |
| **bodyclose**                 | Ensures that HTTP response bodies are properly closed.                                   |
| **canonicalheader**           | Checks if net/http.Header keys use canonical formatting (capitalization rules).          |
| **copyloopvar**               | Detects mistakes where loop variables are copied incorrectly (Go 1.22+).                 |
| **cyclop**                    | Measures function/package cyclomatic complexity and warns if it exceeds the threshold.   |
| **depguard**                  | Enforces rules to restrict allowed imported packages.                                    |
| **dupl**                      | Detects duplicated (copied-pasted) code.                                                 |
| **durationcheck**             | Detects invalid operations multiplying two durations together.                           |
| **errcheck**                  | Warns about unchecked errors, which can lead to critical bugs.                           |
| **errname**                   | Ensures sentinel errors are prefixed with Err and error types suffixed with Error.       |
| **errorlint**                 | Detects issues with error wrapping introduced in Go 1.13+.                               |
| **exhaustive**                | Ensures enum switch statements cover all cases.                                          |
| **exptostd**                  | Detects usage of golang.org/x/exp functions that can be replaced by stdlib.              |
| **fatcontext**                | Finds improper context.Context creation inside loops.                                    |
| **forbidigo**                 | Forbids specific identifiers from being used.                                            |
| **funcorder**                 | Enforces order of functions, methods, and constructors in source code.                   |
| **funlen**                    | Warns if functions are too long (for readability).                                       |
| **gocheckcompilerdirectives** | Validates Go compiler directive comments (//go:).                                        |
| **gochecknoglobals**          | Warns about usage of global variables.                                                   |
| **gochecknoinits**            | Warns against using init() functions that hide logic.                                    |
| **gochecksumtype**            | Ensures exhaustiveness when handling sum types in Go.                                    |
| **gocognit**                  | Computes cognitive complexity of functions.                                              |
| **goconst**                   | Suggests extracting repeated string literals into constants.                             |
| **gocritic**                  | Provides various diagnostics for bugs, performance, and style issues.                    |
| **gocyclo**                   | Measures cyclomatic complexity and warns about overly complex functions.                 |
| **godot**                     | Ensures comments end with a period.                                                      |
| **gomoddirectives**           | Checks for proper usage of "replace", "retract", and "exclude" directives in go.mod.     |
| **goprintffuncname**          | Checks that printf-like functions are named with a trailing "f".                         |
| **gosec**                     | Detects potential security vulnerabilities in Go code.                                   |
| **govet**                     | Runs Go's standard vet tool to detect suspicious code.                                   |
| **iface**                     | Helps avoid unnecessary or improper interface usage.                                     |
| **ineffassign**               | Detects ineffective assignments (assigned values not used).                              |
| **intrange**                  | Suggests better integer range usage in for loops.                                        |
| **loggercheck**               | Verifies key-value pairs in logger libraries (kitlog, klog, zap, etc.).                  |
| **makezero**                  | Warns about non-zero-length slice declarations.                                          |
| **mirror**                    | Detects incorrect byte/string mirror patterns.                                           |
| **mnd**                       | Warns against magic numbers in code.                                                     |
| **musttag**                   | Enforces field tags (e.g., json) for structs being marshaled/unmarshaled.                |
| **nakedret**                  | Warns about naked returns in functions longer than a specified length.                   |
| **nestif**                    | Warns about deeply nested if statements.                                                 |
| **nilerr**                    | Finds code returning nil after checking err != nil.                                      |
| **nilnesserr**                | Detects logical inconsistencies with nil error handling.                                 |
| **nilnil**                    | Ensures no simultaneous return of nil error and invalid value.                           |
| **noctx**                     | Detects missing context.Context when making HTTP requests.                               |
| **nolintlint**                | Checks for improperly formatted nolint directives.                                       |
| **nonamedreturns**            | Reports all usage of named return values.                                                |
| **nosprintfhostport**         | Detects misuse of fmt.Sprintf when constructing host:port strings.                       |
| **perfsprint**                | Suggests faster alternatives to fmt.Sprintf when applicable.                             |
| **predeclared**               | Prevents shadowing of Go predeclared identifiers (e.g., int, true).                      |
| **promlinter**                | Checks Prometheus metric naming conventions using promlint.                              |
| **protogetter**               | Warns against direct reads of proto message fields without getters.                      |
| **reassign**                  | Prevents reassignment of package-level variables.                                        |
| **recvcheck**                 | Ensures method receiver type consistency.                                                |
| **revive**                    | Modern, configurable replacement for golint with style checking.                         |
| **rowserrcheck**              | Ensures that SQL rows errors are properly checked.                                       |
| **sloglint**                  | Enforces consistent style when using the log/slog package.                               |
| **spancheck**                 | Detects mistakes with OpenTelemetry/Census span usage.                                   |
| **sqlclosecheck**             | Ensures that sql.Rows and sql.Stmt are properly closed.                                  |
| **staticcheck**               | High-powered static analysis for Go (better than govet).                                 |
| **testableexamples**          | Verifies if example functions are properly testable (with output).                       |
| **testifylint**               | Checks proper usage of the github.com/stretchr/testify library.                          |
| **testpackage**               | Forces test code to use a separate \_test package.                                       |
| **tparallel**                 | Detects improper usage of t.Parallel() in tests.                                         |
| **unconvert**                 | Removes unnecessary type conversions.                                                    |
| **unparam**                   | Reports unused function parameters.                                                      |
| **unused**                    | Detects unused constants, variables, functions, and types.                               |
| **usestdlibvars**             | Suggests using variables/constants from Go standard library when possible.               |
| **usetesting**                | Warns against improper replacements of functions in the testing package.                 |
| **wastedassign**              | Finds wasted assignments that are never used.                                            |
| **whitespace**                | Detects leading and trailing whitespace issues.                                          |

## KR

| Linter                        | 설명                                                                                   |
| :---------------------------- | :------------------------------------------------------------------------------------- |
| **asasalint**                 | `[]any` 타입을 `...any` 함수에 넘길 때 생기는 타입 문제를 체크합니다.                  |
| **asciicheck**                | 코드에 **비ASCII 문자(한글, 일본어 등)** 가 있으면 경고합니다.                         |
| **bidichk**                   | **유니코드 BiDi(Bidirectional) 시퀀스** 악용(보안 문제)을 찾아냅니다.                  |
| **bodyclose**                 | HTTP 요청을 보낸 후 **response body를 닫지 않는 실수**를 체크합니다.                   |
| **canonicalheader**           | HTTP 헤더 키가 **표준 형식(대문자/소문자 규칙)** 을 따르는지 확인합니다.               |
| **copyloopvar**               | 루프 안에서 **루프 변수를 복사하는 실수**를 잡아냅니다 (Go 1.22+).                     |
| **cyclop**                    | **함수/패키지의 복잡도(cyclomatic complexity)** 를 측정하고 기준 초과를 경고합니다.    |
| **depguard**                  | **허용된 패키지만 import** 하도록 강제합니다.                                          |
| **dupl**                      | **복붙(duplicate) 코드**를 찾아줍니다.                                                 |
| **durationcheck**             | **duration끼리 곱하는 잘못된 계산**을 잡아냅니다.                                      |
| **errcheck**                  | 반환된 **에러를 체크하지 않은 코드**를 경고합니다.                                     |
| **errname**                   | 에러 이름 규칙 (`ErrXXX`, `XXXError`) 을 맞추지 않은 경우 잡아냅니다.                  |
| **errorlint**                 | Go 1.13+ 에서 추가된 **에러 래핑(wrapping)** 관련 잘못된 코드를 잡습니다.              |
| **exhaustive**                | enum switch 문에서 **모든 케이스를 다루었는지** 검사합니다.                            |
| **exptostd**                  | `golang.org/x/exp`를 사용 중인데 **표준 라이브러리로 바꿀 수 있으면** 경고합니다.      |
| **fatcontext**                | 루프 안에서 **context.Context를 잘못 중첩 생성한 경우**를 찾아냅니다.                  |
| **forbidigo**                 | **금지된 식별자(변수명, 함수명)** 사용을 막습니다.                                     |
| **funcorder**                 | **함수/메서드/생성자 순서**를 표준 규칙대로 정렬했는지 체크합니다.                     |
| **funlen**                    | **함수 길이**가 너무 길면 경고합니다. (가독성 이슈)                                    |
| **gocheckcompilerdirectives** | `//go:` 같은 **컴파일러 지시문이 잘못 사용된 경우**를 체크합니다.                      |
| **gochecknoglobals**          | **전역 변수 사용을 금지**하거나 경고합니다.                                            |
| **gochecknoinits**            | **init() 함수**를 금지하거나 경고합니다. (초기화 로직 숨기는 문제 방지)                |
| **gochecksumtype**            | Go에서 **sum type**을 완전히 exhaustive하게 다루었는지 검사합니다.                     |
| **gocognit**                  | **함수의 인지 복잡도(Cognitive Complexity)** 를 체크합니다.                            |
| **goconst**                   | **반복되는 문자열/상수**를 상수로 뽑을 수 있는지 추천합니다.                           |
| **gocritic**                  | **버그/성능/스타일 문제**를 전반적으로 잡아주는 고급 종합 검사기입니다.                |
| **gocyclo**                   | **함수의 cyclomatic complexity**를 측정해서 기준 초과를 경고합니다.                    |
| **godot**                     | **주석**이 마침표로 끝나지 않으면 경고합니다.                                          |
| **gomoddirectives**           | `go.mod` 파일의 `replace`, `retract`, `exclude` 지시어 사용을 점검합니다.              |
| **goprintffuncname**          | `Printf` 류 함수를 만들 때 **함수 이름 끝에 f를 붙이는 규칙**을 체크합니다.            |
| **gosec**                     | **보안 취약점(Security issues)** 을 잡는 전용 static analyzer입니다.                   |
| **govet**                     | Go 기본 제공 **버그 탐지 도구(vet)** 를 실행합니다.                                    |
| **iface**                     | 인터페이스 사용이 **오염(interface pollution)** 되지 않았는지 확인합니다.              |
| **ineffassign**               | **효과 없는(assign 후 사용 안 하는) 할당**을 경고합니다.                               |
| **intrange**                  | `for` 루프를 **더 깔끔한 integer range**로 바꿀 수 있는 경우 추천합니다.               |
| **loggercheck**               | **로거 라이브러리** 사용시 key-value 짝 맞춤 오류를 체크합니다. (kitlog, klog, zap 등) |
| **makezero**                  | 초기 길이가 0이 아닌 **슬라이스(slice) 선언**을 경고합니다.                            |
| **mirror**                    | byte/string mirror 연산이 잘못된 경우를 찾아냅니다.                                    |
| **mnd**                       | 코드 안에 **매직 넘버(magic number)** 가 있으면 경고합니다.                            |
| **musttag**                   | (un)marshal되는 struct는 **필드 태그(json 등)** 를 반드시 가져야 합니다.               |
| **nakedret**                  | **길다란 함수에서 naked return** (값 없이 return) 을 쓰면 경고합니다.                  |
| **nestif**                    | **if문이 너무 깊게 중첩**되면 경고합니다.                                              |
| **nilerr**                    | `if err != nil` 체크했는데도 **nil 반환하는 코드**를 잡습니다.                         |
| **nilnesserr**                | nil error 관련 논리적 모순을 체크합니다 (nilness + nilerr 합친 것).                    |
| **nilnil**                    | **nil 에러와 nil 값**을 동시에 리턴하는 코드 문제를 잡습니다.                          |
| **noctx**                     | **http 요청에 context.Context가 빠진 경우**를 잡습니다.                                |
| **nolintlint**                | **잘못된 nolint 주석 사용**을 체크합니다.                                              |
| **nonamedreturns**            | **named return** (이름 붙은 반환값)을 모두 리포트합니다.                               |
| **nosprintfhostport**         | `Sprintf`로 host:port를 만들 때 실수하는 경우를 잡습니다.                              |
| **perfsprint**                | `fmt.Sprintf` 대신 더 빠른 방법이 있을 때 추천합니다.                                  |
| **predeclared**               | Go 내장 타입/상수(`int`, `true` 등)를 **잘못 shadowing** 하는 걸 방지합니다.           |
| **promlinter**                | Prometheus 메트릭 이름을 **promlint** 규칙에 맞는지 체크합니다.                        |
| **protogetter**               | proto message를 **getter 없이 직접 읽는 실수**를 경고합니다.                           |
| **reassign**                  | **패키지 레벨 변수 재할당**을 막습니다.                                                |
| **recvcheck**                 | 메서드 리시버 타입 일관성을 체크합니다.                                                |
| **revive**                    | 빠르고 유연한 스타일 linter입니다. (`golint`의 후계자격)                               |
| **rowserrcheck**              | `sql.Rows` 결과에서 **에러 체크를 빼먹은 코드**를 경고합니다.                          |
| **sloglint**                  | `log/slog` 패키지를 일관된 스타일로 쓰는지 검사합니다.                                 |
| **spancheck**                 | OpenTelemetry/Census **Span 사용 실수**를 찾아줍니다.                                  |
| **sqlclosecheck**             | **sql.Rows, sql.Stmt를 close 안 한 경우**를 경고합니다.                                |
| **staticcheck**               | Go vet보다 강력한 고급 정적 분석기입니다.                                              |
| **testableexamples**          | **예제 함수**가 제대로 테스트 가능한지 체크합니다.                                     |
| **testifylint**               | `stretchr/testify` 라이브러리의 잘못된 사용을 체크합니다.                              |
| **testpackage**               | 테스트를 `_test` 별도 패키지에서 작성하도록 강제합니다.                                |
| **tparallel**                 | `t.Parallel()` 잘못 사용한 경우를 잡습니다.                                            |
| **unconvert**                 | 불필요한 타입 변환을 제거할 수 있는 경우를 찾습니다.                                   |
| **unparam**                   | 사용되지 않는 함수 파라미터를 찾아냅니다.                                              |
| **unused**                    | 사용하지 않는 변수/상수/함수/타입을 찾습니다.                                          |
| **usestdlibvars**             | Go 표준 라이브러리의 상수/변수를 쓸 수 있는 경우 추천합니다.                           |
| **usetesting**                | testing 패키지 내부 함수 대신 외부 대체 함수를 잘못 쓰는 경우를 잡습니다.              |
| **wastedassign**              | 아무 쓸모 없는(assign 후 사용 안하는) 할당을 잡습니다.                                 |
| **whitespace**                | 앞뒤에 이상한 공백(whitespace)이 있는지 찾아냅니다.                                    |
