package engine

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func evalBasicLit(expr *ast.BasicLit) (interface{}, error) {
	value := expr.Value
	switch expr.Kind {
	case token.INT:
		return strconv.ParseInt(value, 10, 64)
	case token.FLOAT:
		return strconv.ParseFloat(value, 64)
	case token.STRING:
		return value[1 : len(value)-1], nil
	default:
		return nil, fmt.Errorf("unsupported basic lit types.")
	}
}

func evalIdent(expr *ast.Ident, dataSource map[string]interface{}) interface{} {
	key := expr.Name

	// bool type is Ident
	if key == "true" {
		return true
	} else if key == "false" {
		return false
	}

	value, ok := dataSource[key]
	if !ok {
		return nil
	} else {
		return value
	}
}

func evalIndexExpr(expr *ast.IndexExpr, dataSource map[string]interface{}) (interface{}, error) {
	data, err := eval(expr.X, dataSource)
	if err != nil {
		return nil, err
	}

	index, err := eval(expr.Index, dataSource)
	if err != nil {
		return nil, err
	}

	var index64 int64
	switch index := index.(type) {
	case int:
		index64 = int64(index)
	case int64:
		index64 = index
	case string:
	default:
		return nil, fmt.Errorf("index must be int or int64.")
	}

	switch data := data.(type) {
	case map[string]interface{}:
		return data[index.(string)], nil
	case []string:
		return data[index64], nil
	case []int64:
		return data[index64], nil
	case []int:
		return data[index64], nil
	default:
		return nil, fmt.Errorf("unsupported type. index: %v", index)
	}
}

func evalBinaryExpr(expr *ast.BinaryExpr, dataSource map[string]interface{}) (interface{}, error) {
	left, err := eval(expr.X, dataSource)
	if err != nil {
		return nil, err
	}

	right, err := eval(expr.Y, dataSource)
	if err != nil {
		return nil, err
	}

	op := expr.Op
	err = fmt.Errorf("invalid binary operation. left: %v, operation: %v, right: %v", left, op, right)
SW:
	switch lv := left.(type) {
	case int64:
		var rv int64

		switch right := right.(type) {
		case int:
			rv = int64(right)
		case int64:
			rv = right
		case float64, float32:
			left = float64(lv)
			goto SW
		}

		switch op {
		case token.ADD: // +
			return lv + rv, nil
		case token.SUB: // -
			return lv - rv, nil
		case token.MUL: // *
			return lv * rv, nil
		case token.QUO: // /
			return lv / rv, nil
		case token.REM: // %
			return lv % rv, nil
		case token.LSS: // <
			return lv < rv, nil
		case token.GTR: // >
			return lv > rv, nil
		case token.LEQ: // <=
			return lv <= rv, nil
		case token.GEQ: // >=
			return lv >= rv, nil
		}
	case int:
		left = int64(lv)
		goto SW
	case string:
		rv := right.(string)
		switch op {
		case token.ADD: // +
			return lv + rv, nil
		}
	case float64:
		var rv float64
		switch right := right.(type) {
		case int:
			rv = float64(right)
		case int64:
			rv = float64(right)
		case float32:
			rv = float64(right)
		default:
			rv = right.(float64)
		}

		switch op {
		case token.ADD: // +
			return floatCal(lv, rv, '+'), nil
		case token.SUB: // -
			return floatCal(lv, rv, '-'), nil
		case token.MUL: // *
			return floatCal(lv, rv, '*'), nil
		case token.QUO: // /
			return floatCal(lv, rv, '/'), nil
		case token.LSS: // <
			return lv < rv, nil
		case token.GTR: // >
			return lv > rv, nil
		case token.LEQ: // <=
			return lv <= rv, nil
		case token.GEQ: // >=
			return lv >= rv, nil
		}
	case float32:
		left = float64(lv)
		goto SW
	case bool:
		rv := right.(bool)
		switch op {
		case token.LAND: // &&
			return lv && rv, nil
		case token.LOR: // ||
			return lv || rv, nil
		}
	}

	switch op {
	case token.EQL: // ==
		return left == right, nil
	case token.NEQ: // !=
		return left != right, nil
	}

	return nil, err
}

func evalUnaryExpr(expr *ast.UnaryExpr, dataSource map[string]interface{}) (interface{}, error) {
	operand, err := eval(expr.X, dataSource)
	if err != nil {
		return nil, err
	}

	op := expr.Op
	err = fmt.Errorf("invalid unary operation. operation: %v, right: %v", op, operand)
SW:
	switch opd := operand.(type) {
	case bool:
		switch op {
		case token.NOT: // !
			return !opd, nil
		}
	case int64:
		switch op {
		case token.SUB: // -
			return (-1) * opd, nil
		}
	case int:
		operand = int64(opd)
		goto SW
	case float64:
		switch op {
		case token.SUB: // -
			return (-1.0) * opd, nil
		}
	}

	return err, nil
}

func evalCallExpr(expr *ast.CallExpr, dataSource map[string]interface{}) (interface{}, error) {
	f, err := eval(expr.Fun, fns)
	if err != nil {
		return nil, err
	}

	switch f := f.(type) {
	case fn:
		return f(expr.Args, dataSource)
	default:
		return nil, fmt.Errorf("unknow method type. fun: %+v", f)
	}
}

func eval(expr ast.Expr, dataSource map[string]interface{}) (interface{}, error) {
	switch expr := expr.(type) {
	case *ast.ParenExpr: // 括号内表达式
		return eval(expr.X, dataSource)
	case *ast.SelectorExpr:
		data, err := eval(expr.X, dataSource)
		if err != nil {
			return nil, err
		}

		return evalIdent(expr.Sel, data.(map[string]interface{})), nil
	case *ast.IndexExpr:
		return evalIndexExpr(expr, dataSource)
	case *ast.BinaryExpr: // 二元表达式
		return evalBinaryExpr(expr, dataSource)
	case *ast.CallExpr: // 方法表达式
		return evalCallExpr(expr, dataSource)
	case *ast.UnaryExpr: // 一元表达式
		return evalUnaryExpr(expr, dataSource)
	case *ast.BasicLit: // 基本类型文字（当作字符串存储）
		return evalBasicLit(expr)
	case *ast.Ident: // 标志符（已定义变量或常量（bool））
		return evalIdent(expr, dataSource), nil
	default:
		return nil, fmt.Errorf("unsupported expression.")
	}
}

func Run(expr string, dataSource map[string]interface{}) (rst interface{}, err error) {
	defer func() {
		if _err := recover(); _err != nil {
			err = fmt.Errorf("%+v", _err)
		}
	}()

	exprAst, err := parser.ParseExpr(expr)
	if err != nil {
		return nil, err
	}

	return eval(exprAst, dataSource)
}
