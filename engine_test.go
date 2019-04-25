package engine

import (
	"go/ast"
	"go/parser"
	"log"
	"testing"
)

var dataSource map[string]interface{}

func TestMain(t *testing.M) {
	dataSource = map[string]interface{}{
		"str":    "abc123",
		"it":     123432423,
		"ft":     1.2345,
		"it2":    2,
		"bl":     true,
		"arrary": []string{"0a", "1b"},
		"mp": map[string]interface{}{
			"ma": -12232,
			"mb": -0.121231,
			"mc": "xxx",
			"md": []int{1, 2, 3},
			"ms": "12abc789_==+and m",
		},
	}

	t.Run()
}

func createExpr(expr string) ast.Expr {
	exprAst, err := parser.ParseExpr(expr)
	if err != nil {
		log.Fatalf("createExpr error: %v", err)
		return nil
	}

	return exprAst
}

func TestEvalBasicLit(t *testing.T) {
	fStr := func(_expr, expect string) {
		expr := createExpr(_expr)
		rst, err := evalBasicLit(expr.(*ast.BasicLit))
		if err != nil || rst.(string) != expect {
			t.Fatalf("string error: %v. expr: %+v", err, expr)
		}
	}

	fStr(`"abc123"`, "abc123")
	fStr(`"-a中国_+="`, "-a中国_+=")
	fStr(`"123"`, "123")

	fInt := func(_expr string, expect int64) {
		expr := createExpr(_expr)
		rst, err := evalBasicLit(expr.(*ast.BasicLit))
		if err != nil || rst.(int64) != expect {
			t.Fatalf("int64 error: %v. expr: %+v", err, expr)
		}
	}
	fInt("1", int64(1))
	fInt("01230", int64(1230))
	fInt("12304567", int64(12304567))

	fFloat := func(_expr string, expect float64) {
		expr := createExpr(_expr)
		rst, err := evalBasicLit(expr.(*ast.BasicLit))
		if err != nil || rst.(float64) != expect {
			t.Fatalf("float64 error: %v. expr: %+v", err, expr)
		}
	}
	fFloat("1.01", 1.01)
	fFloat("1.01", 1.01)
	fFloat("0.1234567", 0.1234567)
}

func TestEvalIdent(t *testing.T) {
	for k, v := range dataSource {
		expr := createExpr(k)
		r := evalIdent(expr.(*ast.Ident), dataSource)
		switch r := r.(type) {
		case []string:
			if r[1] != v.([]string)[1] {
				t.Fatalf("Ident error. expr: %+v", expr)
			}

		case map[string]interface{}:
			if r["mc"] != v.(map[string]interface{})["mc"] {
				t.Fatalf("Ident error. expr: %+v", expr)
			}

		default:
			if r != v {
				t.Fatalf("Ident error. expr: %+v", expr)
			}
		}
	}

	expr := createExpr("noexit")
	r := evalIdent(expr.(*ast.Ident), dataSource)
	if r != nil {
		t.Fatalf("Ident error. expr: %+v", expr)
	}

	expr = createExpr("true")
	r = evalIdent(expr.(*ast.Ident), dataSource)
	if r != true {
		t.Fatalf("Ident error. expr: %+v", expr)
	}
}

func TestEvalIndexExpr(t *testing.T) {
	expr := createExpr(`arrary[0]`)
	rst, err := evalIndexExpr(expr.(*ast.IndexExpr), dataSource)
	if err != nil || rst.(string) != dataSource["arrary"].([]string)[0] {
		t.Fatalf("error: %v. expr: %+v", err, expr)
	}

	expr = createExpr(`mp["md"][1]`)
	rst, err = evalIndexExpr(expr.(*ast.IndexExpr), dataSource)
	if err != nil || rst.(int) != dataSource["mp"].(map[string]interface{})["md"].([]int)[1] {
		t.Fatalf("error: %v. expr: %+v", err, expr)
	}
}

func TestEvalBinaryExpr(t *testing.T) {
	f := func(_expr string, expect interface{}) {
		expr := createExpr(_expr)
		rst, err := evalBinaryExpr(expr.(*ast.BinaryExpr), dataSource)
		if err != nil || rst != expect {
			t.Fatalf("error: %v. expr: %+v, expect: %v, rst: %v", err, expr, expect, rst)
		}
	}

	f("1+2-2", int64(1))
	f("(1-1)*100", int64(0))
	f("1.223400012-1.2", 0.023400012)
	f("1.000001*0.2", 0.2000002)
	f("1.0/2", 0.5)
	f("1.1*it2", 2.2)
	f("5-4.5", 0.5)
	f("7%3", int64(1))
	f(`"abc"+"123"+"X"`, "abc123X")

	f(`3>5`, false)
	f(`5>=5`, true)
	f(`5<=5`, true)
	f(`1.2344<1.23441`, true)
	f(`"abc123"=="abc123"`, true)
	f(`100>99&&true&&!false`, true)
}

func TestEvalUnaryExpr(t *testing.T) {
	f := func(_expr string, expect interface{}) {
		expr := createExpr(_expr)
		rst, err := evalUnaryExpr(expr.(*ast.UnaryExpr), dataSource)
		if err != nil || rst != expect {
			t.Fatalf("error: %v. expr: %+v, expect: %v, rst: %v", err, expr, expect, rst)
		}
	}

	f(`-123`, int64(-123))
	f(`-123`, int64(-123))
	f(`-0.99`, -0.99)
	f(`!true`, false)
	f(`!false`, true)
}

func TestEvalCallExpr(t *testing.T) {
	expr := createExpr(`newSlice("a",str)`)
	rst, err := evalCallExpr(expr.(*ast.CallExpr), dataSource)
	if err != nil || rst.([]string)[1] != dataSource["str"] {
		t.Fatalf("call error: %v, rst: %v", err, rst)
	}
}

func TestRun(t *testing.T) {
	f := func(_expr string, expect interface{}) {
		rst, err := Run(_expr, dataSource)
		if err != nil {
			t.Fatalf("run expr error: %v", err)
		}

		if rst != expect {
			t.Fatalf("unexpect. rst: %v, expect: %v", rst, expect)
		}

	}

	f(`(2+1-2)*(-2)/(mp["md"][1])`, int64(-1))
	f("((5%it2)+2-3)/100", int64(0))
	f(`bl&&(mp["md"][it2]-1==2)`, true)
	f(`contains(arrary,"1b")&&it>1.0&&it2<=3`, true)
	f(`(!matchString("abc[0-9]+.*m",mp["ms"])||!bl)==false`, true)
	f(`((!matchString("abc[0-9]+.*m",mp["ms"])||!bl)==false)&&(555%2+0.1234*10>(it2/5-1.0123)||contains(newSlice("abc","cfg","123"),"cfg"))`, true)

	//panic
	rst, err := Run(`"xx"+100`, dataSource)
	if err != nil && rst == nil {
		t.Logf("expected result. error: %v", err)
	} else {
		t.Fatalf("should return error.")
	}
}

func BenchmarkRun(b *testing.B) {
	expr := `(!matchString("abc[0-9]+.*m",mp["ms"])||!bl)==false)&&(555%2+0.1234*10>(it2/5-1.0123)||contains(newSlice("abc","cfg","123"),"cfg"))`

	for i := 0; i < b.N; i++ {
		Run(expr, dataSource)
	}
}
