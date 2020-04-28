package evaluator

import (
	"github.com/avinassh/monkey/ast"
	"github.com/avinassh/monkey/object"
	"github.com/avinassh/monkey/token"
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		return evalFnLiteral(node, env)
	case *ast.ArrayLiteral:
		items := evalExpressions(node.Elements, env)
		if len(items) == 1 && isError(items[0]) {
			return items[0]
		}
		return &object.Array{Elements: items}

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.CallExpression:
		// we will evaluate call expressions, first we will eval the
		// func part. This will have the relevant body of the function
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		// and this will have all the parameters evaluated
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		array, ok := left.(*object.Array)
		if !ok {
			return nil
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		idx, ok := index.(*object.Integer)
		if !ok {
			return NULL
		}
		if int(idx.Value) >= len(array.Elements) || idx.Value < 0 {
			return NULL
		}
		return array.Elements[idx.Value]
	}
	return nil
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		// if one of the statements had a return statement, then we don't
		// need to run next statements and we could do an early return
		// since Eval returns ReturnObj for return statements, we will check
		// if the `result` is ReturnObj
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		// from book:
		//
		// Here we explicitly don’t unwrap the return value and only check the Type() of each evaluation result.
		// If it’s object.RETURN_VALUE_OBJ we simply return the *object.ReturnValue, without unwrapping its .Value,
		// so it stops execution in a possible outer block statement and bubbles up to evalProgram, where it finally
		// get’s unwrapped.
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return evalIntegerInfixExpression(operator, left, right)
	}
	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return evalStringInfixExpression(operator, left, right)
	}
	if operator == token.EQ {
		return nativeBoolToBooleanObject(left == right)
	}
	if operator == token.NOT_EQ {
		return nativeBoolToBooleanObject(left != right)
	}
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	}
	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}
	return NULL
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case token.PLUS:
		return &object.Integer{Value: leftVal + rightVal}
	case token.MINUS:
		return &object.Integer{Value: leftVal - rightVal}
	case token.ASTERISK:
		return &object.Integer{Value: leftVal * rightVal}
	case token.SLASH:
		return &object.Integer{Value: leftVal / rightVal}
	case token.LT:
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case token.GT:
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case token.EQ:
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	if operator == token.PLUS {
		return &object.String{Value: leftVal + rightVal}
	}
	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	obj := right.(*object.Integer)
	return &object.Integer{Value: -obj.Value}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if val, ok := builtins[node.Value]; ok {
		return val
	}
	return newError("identifier not found: " + node.Value)
}

func evalFnLiteral(node *ast.FunctionLiteral, env *object.Environment) object.Object {
	return &object.Function{
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env,
	}
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// from book:
//
// The extendFunctionEnv function creates a new *object.Environment that’s enclosed by the function’s environment.
// In this new, enclosed environment it binds the arguments of the function call to the function’s parameter names.
//
// And this newly enclosed and updated environment is then the environment in which the function’s body is evaluated.
// The result of this evaluation is unwrapped if it’s an *object.ReturnValue. That’s necessary, because otherwise a
// return statement would bubble up through several functions and stop the evaluation in all of them. But we only want
// to stop the evaluation of the last called function’s body. That’s why we need unwrap it, so that evalBlockStatement
// won’t stop evaluating statements in “outer” functions.
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	// create a new env from the fn's env
	// from book:
	// Instead we’ll use the environment our *object.Function carries around. Remember that one? That’s the environment
	// our function was defined in.
	env := object.NewEnclosedEnvironment(fn.Env)

	// args basically contains values, either as raw values or as identifiers
	// lets say the call argument is:
	//
	// fnCall(a, 10);
	//
	// fn.Parameters contain the list of parameters and args are the same, but of
	// values. So we will set in the extended environment, taking the name from
	// params and value from args
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}
