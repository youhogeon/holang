# HOLang; 호랭
22세기를 선도하는 언어

## 실행
```sh
./dist/holang
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