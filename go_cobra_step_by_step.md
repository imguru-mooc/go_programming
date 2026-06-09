# Go Cobra 활용 방법 단계별 가이드

## 1. Cobra란?

Cobra는 Go에서 CLI(Command Line Interface) 프로그램을 만들 때 많이 사용하는 라이브러리입니다.

예를 들어 다음과 같은 명령어 구조를 만들 수 있습니다.

```bash
mycli hello
mycli add 10 20
mycli version
mycli config set name kim
```

Cobra는 다음 기능을 쉽게 만들 수 있게 도와줍니다.

- 루트 명령어
- 서브커맨드
- 인자 처리
- 플래그 처리
- 도움말 자동 생성
- 에러 처리

---

## 2. Cobra 없이 만든 기본 CLI

일반적인 Go 프로그램은 다음처럼 작성합니다.

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello CLI")
}
```

실행:

```bash
go run main.go
```

하지만 다음처럼 여러 명령을 가진 CLI를 만들려면 코드가 복잡해질 수 있습니다.

```bash
mycli hello
mycli add 10 20
mycli version
```

이럴 때 Cobra를 사용합니다.

---

## 3. 프로젝트 생성

```bash
mkdir mycli
cd mycli
go mod init mycli
```

Cobra 설치:

```bash
go get github.com/spf13/cobra
```

프로젝트 구조는 다음과 같이 만듭니다.

```text
mycli/
├── go.mod
├── main.go
└── cmd/
    ├── root.go
    ├── hello.go
    └── add.go
```

---

## 4. main.go 작성

`main.go`

```go
package main

import "mycli/cmd"

func main() {
	cmd.Execute()
}
```

`main.go`의 역할은 단순합니다.

```text
프로그램 시작
    ↓
cmd.Execute() 호출
    ↓
Cobra 명령어 실행
```

---

## 5. Root Command 만들기

`cmd/root.go`

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "mycli는 Cobra 예제 CLI입니다",
	Long:  "mycli는 Go Cobra를 단계적으로 배우기 위한 예제 프로그램입니다.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mycli 실행됨")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

핵심 구조:

```go
var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "짧은 설명",
	Long:  "긴 설명",
	Run: func(cmd *cobra.Command, args []string) {
		// 실제 실행 코드
	},
}
```

| 항목 | 의미 |
|---|---|
| `Use` | 명령어 이름 |
| `Short` | help에 나오는 짧은 설명 |
| `Long` | 자세한 설명 |
| `Run` | 명령어가 실행될 때 수행할 코드 |

실행:

```bash
go run .
```

출력:

```text
mycli 실행됨
```

도움말 확인:

```bash
go run . --help
```

---

## 6. hello 서브커맨드 추가

다음 명령을 만들겠습니다.

```bash
go run . hello
```

`cmd/hello.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "인사 메시지를 출력합니다",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello Cobra!")
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
}
```

중요한 부분:

```go
func init() {
	rootCmd.AddCommand(helloCmd)
}
```

의미:

```text
rootCmd 아래에 helloCmd를 등록한다
```

실행:

```bash
go run . hello
```

출력:

```text
Hello Cobra!
```

명령어 구조:

```text
mycli
└── hello
```

---

## 7. 인자를 받는 명령 만들기

이번에는 이름을 인자로 받겠습니다.

```bash
go run . hello kim
```

`cmd/hello.go` 수정:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "인사 메시지를 출력합니다",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Println("Hello,", name)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
}
```

추가된 부분:

```go
Args: cobra.ExactArgs(1),
```

의미:

```text
인자를 정확히 1개 받아야 한다
```

실행:

```bash
go run . hello kim
```

출력:

```text
Hello, kim
```

인자 없이 실행하면 Cobra가 인자 개수가 맞지 않는다는 에러를 출력합니다.

```bash
go run . hello
```

---

## 8. add 명령 만들기

두 숫자를 더하는 명령을 만들겠습니다.

```bash
go run . add 10 20
```

`cmd/add.go`

```go
package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [a] [b]",
	Short: "두 정수를 더합니다",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("첫 번째 값이 정수가 아닙니다: %w", err)
		}

		b, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("두 번째 값이 정수가 아닙니다: %w", err)
		}

		fmt.Println(a + b)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
```

여기서는 `Run` 대신 `RunE`를 사용했습니다.

| 항목 | 의미 |
|---|---|
| `Run` | 에러를 반환하지 않는 실행 함수 |
| `RunE` | 에러를 반환할 수 있는 실행 함수 |

에러 처리가 필요한 명령은 `RunE`를 쓰면 좋습니다.

실행:

```bash
go run . add 10 20
```

출력:

```text
30
```

잘못된 값 입력:

```bash
go run . add 10 abc
```

출력 예:

```text
Error: 두 번째 값이 정수가 아닙니다: strconv.Atoi: parsing "abc": invalid syntax
```

---

## 9. Flag 사용하기

CLI에서는 다음과 같은 옵션을 자주 씁니다.

```bash
go run . hello kim --upper
```

또는 짧은 옵션:

```bash
go run . hello kim -u
```

`cmd/hello.go` 수정:

```go
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var upper bool

var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "인사 메시지를 출력합니다",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if upper {
			name = strings.ToUpper(name)
		}

		fmt.Println("Hello,", name)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)

	helloCmd.Flags().BoolVarP(
		&upper,
		"upper",
		"u",
		false,
		"이름을 대문자로 출력합니다",
	)
}
```

핵심 부분:

```go
helloCmd.Flags().BoolVarP(
	&upper,
	"upper",
	"u",
	false,
	"이름을 대문자로 출력합니다",
)
```

| 값 | 의미 |
|---|---|
| `&upper` | 옵션 값을 저장할 변수 |
| `"upper"` | 긴 옵션 이름 `--upper` |
| `"u"` | 짧은 옵션 이름 `-u` |
| `false` | 기본값 |
| 설명 | help에 표시될 설명 |

실행:

```bash
go run . hello kim --upper
```

또는:

```bash
go run . hello kim -u
```

출력:

```text
Hello, KIM
```

---

## 10. 전역 Flag 만들기

모든 명령에서 공통으로 쓰는 옵션은 `PersistentFlags()`를 사용합니다.

예를 들어 `--verbose` 옵션을 만들겠습니다.

`cmd/root.go`

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "mycli는 Cobra 예제 CLI입니다",
	Long:  "mycli는 Go Cobra를 단계적으로 배우기 위한 예제 프로그램입니다.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mycli 실행됨")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"자세한 로그를 출력합니다",
	)
}
```

`cmd/hello.go`에서 사용:

```go
Run: func(cmd *cobra.Command, args []string) {
	name := args[0]

	if verbose {
		fmt.Println("[debug] hello command 실행")
	}

	if upper {
		name = strings.ToUpper(name)
	}

	fmt.Println("Hello,", name)
},
```

실행:

```bash
go run . --verbose hello kim
```

또는:

```bash
go run . hello kim --verbose
```

출력:

```text
[debug] hello command 실행
Hello, kim
```

---

## 11. 완성된 파일 구조

```text
mycli/
├── go.mod
├── main.go
└── cmd/
    ├── root.go
    ├── hello.go
    └── add.go
```

---

## 12. 전체 코드

### main.go

```go
package main

import "mycli/cmd"

func main() {
	cmd.Execute()
}
```

### cmd/root.go

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "mycli는 Cobra 예제 CLI입니다",
	Long:  "mycli는 Go Cobra를 단계적으로 배우기 위한 예제 프로그램입니다.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mycli 실행됨")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"자세한 로그를 출력합니다",
	)
}
```

### cmd/hello.go

```go
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var upper bool

var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "인사 메시지를 출력합니다",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if verbose {
			fmt.Println("[debug] hello command 실행")
		}

		if upper {
			name = strings.ToUpper(name)
		}

		fmt.Println("Hello,", name)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)

	helloCmd.Flags().BoolVarP(
		&upper,
		"upper",
		"u",
		false,
		"이름을 대문자로 출력합니다",
	)
}
```

### cmd/add.go

```go
package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [a] [b]",
	Short: "두 정수를 더합니다",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("첫 번째 값이 정수가 아닙니다: %w", err)
		}

		b, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("두 번째 값이 정수가 아닙니다: %w", err)
		}

		fmt.Println(a + b)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
```

---

## 13. 실행 테스트

### 기본 실행

```bash
go run .
```

출력:

```text
mycli 실행됨
```

### 도움말

```bash
go run . --help
```

출력 예:

```text
mycli는 Cobra 예제 CLI입니다

Usage:
  mycli [flags]
  mycli [command]

Available Commands:
  add         두 정수를 더합니다
  hello       인사 메시지를 출력합니다
  help        Help about any command

Flags:
  -h, --help      help for mycli
  -v, --verbose   자세한 로그를 출력합니다
```

### hello 명령

```bash
go run . hello kim
```

출력:

```text
Hello, kim
```

### hello + flag

```bash
go run . hello kim --upper
```

출력:

```text
Hello, KIM
```

### add 명령

```bash
go run . add 10 20
```

출력:

```text
30
```

---

## 14. 빌드해서 실행하기

```bash
go build -o mycli
```

Linux/macOS:

```bash
./mycli hello kim
./mycli add 10 20
```

Windows PowerShell:

```powershell
.\mycli.exe hello kim
.\mycli.exe add 10 20
```

---

## 15. Cobra CLI 생성기 사용 방법

직접 파일을 만들어도 되지만, Cobra에는 프로젝트와 명령어 파일을 생성해주는 `cobra-cli` 도구도 있습니다.

설치:

```bash
go install github.com/spf13/cobra-cli@latest
```

설치 후:

```bash
cobra-cli init
cobra-cli add hello
cobra-cli add add
```

이렇게 하면 기본 구조를 자동으로 만들어줍니다.

처음 배울 때는 직접 작성하는 것이 이해하기 좋고, 익숙해지면 `cobra-cli`를 쓰면 됩니다.

---

## 16. 핵심 정리

Cobra의 기본 구조:

```go
var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "설명",
	Run: func(cmd *cobra.Command, args []string) {
		// 실행 코드
	},
}
```

서브커맨드 등록:

```go
rootCmd.AddCommand(helloCmd)
```

인자 개수 검사:

```go
Args: cobra.ExactArgs(1)
```

에러를 반환해야 할 때:

```go
RunE: func(cmd *cobra.Command, args []string) error {
	return nil
}
```

플래그 추가:

```go
helloCmd.Flags().BoolVarP(&upper, "upper", "u", false, "대문자로 출력")
```

Cobra는 다음과 같은 구조의 프로그램을 만들 때 유용합니다.

```text
mycli
├── hello
├── add
├── version
└── config
    ├── set
    └── get
```

---

## 17. 자주 쓰는 명령 요약

| 명령 | 의미 |
|---|---|
| `go mod init mycli` | Go 모듈 생성 |
| `go get github.com/spf13/cobra` | Cobra 설치 |
| `go run .` | 현재 프로젝트 실행 |
| `go run . --help` | 도움말 확인 |
| `go run . hello kim` | hello 명령 실행 |
| `go run . add 10 20` | add 명령 실행 |
| `go build -o mycli` | 실행 파일 빌드 |
| `go install github.com/spf13/cobra-cli@latest` | cobra-cli 설치 |

---

## 18. 학습 포인트

Cobra를 배울 때는 다음 순서로 이해하면 좋습니다.

1. `rootCmd`가 전체 CLI의 시작점이다.
2. `AddCommand()`로 서브커맨드를 붙인다.
3. `Args`로 인자 개수를 검사한다.
4. `Run` 또는 `RunE`에 실제 실행 코드를 작성한다.
5. `Flags()`로 특정 명령 전용 옵션을 만든다.
6. `PersistentFlags()`로 모든 명령에서 쓰는 전역 옵션을 만든다.
7. `go run . --help`로 자동 생성된 도움말을 확인한다.
