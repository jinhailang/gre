package engine

import (
	"fmt"
	"github.com/shopspring/decimal"
	"go/ast"
	"regexp"
	"strings"
)

type fn func(args []ast.Expr, dataSource map[string]interface{}) (interface{}, error)
type fn1 func(args []ast.Expr, dataSource map[string]interface{}) error

var fns map[string]interface{}

func init() {
	fns = map[string]interface{}{
		"contains":      fn(contains),
		"matchString":   fn(matchString),
		"findAllString": fn(findAllString),
		"newSlice":      fn(newSlice),
		"isEmpty":       fn(isEmpty),
	}
}

func isEmpty(args []ast.Expr, dataSource map[string]interface{}) (interface{}, error) {
	err := fmt.Errorf("invalid args. args: %+v", args)
	if len(args) != 1 {
		return false, err
	}

	v, err := eval(args[0], dataSource)
	if err != nil {
		return false, err
	}

	switch v := v.(type) {
	case nil:
		return true, nil
	case string:
		return v == "", nil
	}

	return false, nil
}

func newSlice(args []ast.Expr, dataSource map[string]interface{}) (interface{}, error) {
	var sl []string
	for _, arg := range args {
		v, err := eval(arg, dataSource)
		if err != nil {
			return nil, err
		}

		sl = append(sl, v.(string))
	}

	return sl, nil
}

// contains(s, substr string) bool
// contains(as []string, s string) bool
func contains(args []ast.Expr, dataSource map[string]interface{}) (interface{}, error) {
	err := fmt.Errorf("invalid args. args: %+v", args)

	if len(args) == 2 {
		arg0, err := eval(args[0], dataSource)
		if err != nil {
			return false, err
		}
		arg1, err := eval(args[1], dataSource)
		if err != nil {
			return false, err
		}

		switch arg0 := arg0.(type) {
		case string:
			v := arg1.(string)
			return strings.Contains(arg0, v), nil
		case []string:
			v := arg1.(string)
			for _, av := range arg0 {
				if av == v {
					return true, nil
				}
			}
			return false, nil
		}

		return false, err
	}

	return false, err
}

// matchString(pattern string, s string) (matched bool, err error)
func matchString(args []ast.Expr, dataSource map[string]interface{}) (interface{}, error) {
	err := fmt.Errorf("invalid args. args: %+v", args)

	if len(args) == 2 {
		arg0, err := eval(args[0], dataSource)
		if err != nil {
			return false, err
		}
		arg1, err := eval(args[1], dataSource)
		if err != nil {
			return false, err
		}

		return regexp.MatchString(arg0.(string), arg1.(string))
	}

	return false, err
}

// findAllString(s string, n int) []string
func findAllString(args []ast.Expr, dataSource map[string]interface{}) (interface{}, error) {
	err := fmt.Errorf("invalid args. args: %+v", args)

	if len(args) == 3 {
		arg0, err := eval(args[0], dataSource)
		if err != nil {
			return nil, err
		}
		arg1, err := eval(args[1], dataSource)
		if err != nil {
			return nil, err
		}

		arg2, err := eval(args[2], dataSource)
		if err != nil {
			return nil, err
		}

		re := regexp.MustCompile(arg0.(string))
		return re.FindAllString(arg1.(string), arg2.(int)), nil
	}

	return false, err
}

func floatCal(l, r float64, op byte) float64 {
	var rst float64
	ld := decimal.NewFromFloat(l)
	rd := decimal.NewFromFloat(r)

	switch op {
	case '+':
		rst, _ = ld.Add(rd).Float64()
	case '-':
		rst, _ = ld.Sub(rd).Float64()
	case '*':
		rst, _ = ld.Mul(rd).Float64()
	case '/':
		return l / r
	default:
		return 0
	}

	return rst
}
