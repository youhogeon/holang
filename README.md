# HOLang; 호랭
22세기를 선도하는 언어

## 실행
### Compile & Run (.hbc 생성 없이 바로 실행)
```sh
# Windows
./bin/holangc --run ./sample/tetris
# Linux
./bin/holangc --run ./sample/tetris
# Mac
./bin/holangc_darwin_amd64 --run ./sample/tetris
./bin/holangc_darwin_arm64 --run ./sample/tetris
```

### Compile (.hbc 생성)
```sh
# Windows
./bin/holangc ./sample/tetris
# Linux
./bin/holangc ./sample/tetris
# Mac
./bin/holangc_darwin_amd64 ./sample/tetris
./bin/holangc_darwin_arm64 ./sample/tetris
```

### Run (.hbc 실행)
```sh
# Windows
./bin/holang ./sample/tetris
# Linux
./bin/holang ./sample/tetris
# Mac
./bin/holang_darwin_amd64 ./sample/tetris
./bin/holang_darwin_arm64 ./sample/tetris
```

## 예제
```holang
class GuGuDan {
    init(dan) {
        print("GuGuDan");
        this.dan = dan;
    }

    show() {
        var dan = this.dan;

        for (var i = 1; i < 10; i = i + 1) {
            print(str(dan) + " * " + str(i) + " = " + str(dan * i));
        }
    }
}

class PrettyGuGuDan < GuGuDan {
    /**
    * GuGuDan을 상속받는 예쁜구구단 class
    */

    init(dan) {
        super.init(dan);
        print("PrettyGuGuDan");
    }

    show() {
        print("==========" + str(this.dan) + "단==========");
        super.show();
        print("=======================");
    }
}

fun readInt(message) {
    /**
    * 숫자 값을 입력받는 함수
    */

    var strVal = input(message);
    var intVal = int(strVal);

    return intVal;
}

// 스크립트 시작

var dan = readInt("좋아하는 숫자를 입력하세요: ");
var ggd = PrettyGuGuDan(dan);

ggd.show();
```

## 문서
### 내장 함수
* `print(message)`
* `input(message)`
* `clock()`
* `str(value)`
* `int(value)`
* `float(value)`
* `rand()` 0이상 1미만 난수(float)
* `randInt(n)` 0이상 n미만 정수 난수
* `sleep(ms)` ms 동안 대기
* `clear()` 터미널 화면 지움
* `strlen(s)` 문자열 길이
* `substring(s, start, end)` start~end-1 부분 문자열
* `getch()` 한 글자 입력