package frontend

import (
	"fmt"
	"language/backend"
	"language/syntax"
	"strconv"
)

type compiler struct {
	program *backend.Program
	block   *backend.Block
	scope   *scope
	errors  []error
}

func Compile(span syntax.Span) (*backend.Program, *backend.Block, *backend.Block, Type, []error) {
	program := backend.NewProgram()
	entry := program.NewBlock()
	entry.SetName("_start")
	compiler := &compiler{program, entry, newScope(), []error{}}
	ty := compiler.compile(span)
	return program, entry, compiler.block, ty, compiler.errors
}

func (comp *compiler) throw(err error) {
	comp.errors = append(comp.errors, err)
}

func (comp *compiler) compile(span syntax.Span) Type {
	switch expr := span.GetExpr().(type) {

	case syntax.IntegerLiteral:
		value, _ := strconv.Atoi(expr.Value)
		return Integer{comp.block.Constant(value)}

	case syntax.BooleanLiteral:
		if expr.Value == "true" {
			return Boolean{comp.block.Constant(1)}
		}
		return Boolean{comp.block.Constant(0)}

	case syntax.Identifier:
		ty := comp.scope.get(expr.Ident)
		if ty == nil {
			comp.throw(fmt.Errorf("undefined variable '%s'", expr.Ident))
			return nil
		}
		return ty

	case syntax.Tuple:
		items := make([]Type, len(expr.Items))
		for i, item := range expr.Items {
			ty := comp.compile(item)
			if ty == nil {
				return nil
			}
			items[i] = ty
		}
		return Tuple{items}

	case syntax.Struct:
		comp.scope = comp.scope.newScope()
		if comp.compile(expr.Block) == nil {
			return nil
		}
		structure := Struct{comp.scope.dict}
		comp.scope = comp.scope.previous
		return structure

	case syntax.Block:
		for _, line := range expr.Statements[:len(expr.Statements)-1] {
			if comp.compile(line) == nil {
				return nil
			}
		}
		return comp.compile(expr.Statements[len(expr.Statements)-1])

	case syntax.Unary:
		switch expr.Op {
		default:
			comp.throw(fmt.Errorf("invalid binary expression '%s'", expr.Op))
			return nil
		}

	case syntax.Binary:
		switch expr.Op {
		case syntax.Dot:
			left := comp.compile(expr.Left)
			if left == nil {
				return nil
			}
			structure, ok := left.(Struct)
			if !ok {
				comp.throw(fmt.Errorf("cannot use '.' operator on non-structure '%s'", left.Type()))
			}
			comp.scope = comp.scope.newScope()
			for name, ty := range structure.dict {
				comp.scope.dict[name] = ty
			}
			right := comp.compile(expr.Right)
			if right == nil {
				return nil
			}
			newStruct := Struct{comp.scope.dict}
			comp.scope = comp.scope.previous
			if !comp.match(expr.Left, newStruct) {
				return nil
			}
			return right

		case syntax.LessThan:
			entryBlock := comp.block
			condBlock := comp.program.NewBlock()
			exitBlock := comp.program.NewBlock()
			if !comp.compileBoolExpr(span, condBlock, exitBlock) {
				return nil
			}
			comp.block = exitBlock
			dest := comp.program.NewValue()
			comp.block.Copy(condBlock.Constant(1), dest)
			comp.block.Copy(entryBlock.Constant(0), dest)
			entryBlock.Jump(exitBlock)
			condBlock.Jump(exitBlock)
			return Boolean{dest}

		case syntax.Add:
			left := comp.compile(expr.Left)
			right := comp.compile(expr.Right)
			if left == nil || right == nil {
				return nil
			}
			leftInteger, leftIs := left.(Integer)
			rightInteger, rightIs := right.(Integer)
			if leftIs && rightIs {
				return Integer{comp.block.Add(leftInteger.val, rightInteger.val)}
			}
			comp.throw(fmt.Errorf("incompatiable types for addition"))
			return nil

		case syntax.SingleEquals:
			right := comp.compile(expr.Right)
			if right == nil {
				return nil
			}
			if !comp.match(expr.Left, right) {
				return nil
			}
			return right

		case syntax.If:
			condBlock := comp.program.NewBlock()
			elseBlock := comp.program.NewBlock()
			exitBlock := comp.program.NewBlock()
			if !comp.compileBoolExpr(expr.Left, condBlock, elseBlock) {
				return nil
			}

			comp.block = condBlock
			comp.scope = comp.scope.newScope()
			ty := comp.compile(expr.Right)
			if ty == nil {
				return nil
			}

			comp.block = exitBlock
			if !comp.mergeScopes(elseBlock, condBlock) {
				return nil
			}
			ty = comp.newMaybe(ty, condBlock, elseBlock)
			condBlock.Jump(exitBlock)
			elseBlock.Jump(exitBlock)
			comp.block = exitBlock
			return ty

		case syntax.Else:
			comp.scope = comp.scope.newScope()
			left := comp.compile(expr.Left)
			if left == nil {
				return nil
			}
			maybe, ok := left.(Maybe)
			if !ok {
				comp.throw(fmt.Errorf("expected a maybe in else condition"))
				return nil
			}
			condBlock := comp.program.NewBlock()
			elseBlock := comp.program.NewBlock()
			exitBlock := comp.program.NewBlock()
			comp.block.JumpIfEqual(maybe.val, comp.block.Constant(0), condBlock, exitBlock)
			comp.block = condBlock
			ty := comp.compile(expr.Right)
			if ty == nil {
				return nil
			}
			comp.block = exitBlock
			if !comp.mergeScopes(elseBlock, condBlock) {
				return nil
			}
			ty = comp.mergeTypes(maybe.ty, ty, elseBlock, condBlock)
			condBlock.Jump(exitBlock)
			elseBlock.Jump(exitBlock)
			comp.block = exitBlock
			return ty

		case syntax.While:
			entryBlock := comp.block
			startBlock := comp.program.NewBlock()
			condBlock := comp.program.NewBlock()
			bodyBlock := comp.program.NewBlock()
			finalBlock := comp.program.NewBlock()
			exitBlock := comp.program.NewBlock()

			if !comp.compileBoolExpr(expr.Left, startBlock, exitBlock) {
				return nil
			}
			comp.scope = comp.scope.newScope()
			comp.block = startBlock
			if comp.compile(expr.Right) == nil {
				return nil
			}

			comp.block = condBlock
			comp.scope = comp.scope.newScope()

			comp.block = condBlock
			if !comp.compileBoolExpr(expr.Left, bodyBlock, finalBlock) {
				return nil
			}

			comp.block = bodyBlock
			ty := comp.compile(expr.Right)
			if ty == nil {
				return nil
			}

			order := []string{}
			for name, ty := range comp.scope.dict {
				if ty.Type() != comp.scope.previous.dict[name].Type() {
					comp.throw(fmt.Errorf("recursive type definition"))
					return nil
				}
				dest := comp.Duplicate(ty)
				comp.Copy(ty, dest)
				comp.scope.dict[name] = dest
				order = append(order, name)
			}
			for _, name := range order {
				comp.Copy(comp.scope.dict[name], comp.scope.previous.dict[name])
			}

			comp.block = exitBlock
			comp.scope = comp.scope.previous
			if !comp.mergeScopes(entryBlock, finalBlock) {
				return nil
			}
			comp.block = exitBlock

			bodyBlock.Jump(condBlock)
			startBlock.Jump(condBlock)
			finalBlock.Jump(exitBlock)

			return ty

		case syntax.Func:
			return &Func{expr.Left, expr.Right, comp.scope.newScope(), []Impl{}}

		case syntax.Call:
			ty := comp.compile(expr.Left)
			if ty == nil {
				return nil
			}
			fn, ok := ty.(*Func)
			if !ok {
				comp.throw(fmt.Errorf("expected a function in call"))
				return nil
			}
			args := comp.compile(expr.Right)
			if args == nil {
				return nil
			}
			for _, impl := range fn.impls {
				if impl.params.Type() == args.Type() {
					return comp.callFunction(impl, args)
				}
			}
			entryBlock := comp.block
			params := comp.Duplicate(args)
			functionBlock := comp.program.NewBlock()
			comp.block = functionBlock
			oldScope := comp.scope
			comp.scope = fn.scope
			if !comp.match(fn.param, params) {
				return nil
			}
			returns := comp.compile(fn.body)
			comp.block.Return()
			impl := Impl{params, returns, comp.scope, functionBlock}
			fn.impls = append(fn.impls, impl)
			comp.scope = oldScope
			comp.block = entryBlock
			return comp.callFunction(impl, args)

		default:
			comp.throw(fmt.Errorf("invalid binary expression '%s'", expr.Op))
			return nil
		}
	}
	comp.throw(fmt.Errorf("invalid expression '%T' to compile", span.GetExpr()))
	return nil
}

func (comp *compiler) callFunction(impl Impl, args Type) Type {
	exitBlock := comp.program.NewBlock()
	comp.Copy(args, impl.params)
	comp.block.Call(impl.block, exitBlock)
	comp.block = exitBlock
	returns := comp.Duplicate(impl.returns)
	comp.Copy(impl.returns, returns)
	return returns
}

func (comp *compiler) mergeScopes(block, condBlock *backend.Block) bool {
	for name, newTy := range comp.scope.dict {
		oldTy := comp.scope.previous.get(name)
		if oldTy == nil {
			comp.scope.previous.assign(name, comp.newMaybe(newTy, condBlock, block))
		} else {
			mergeTy := comp.mergeTypes(oldTy, newTy, block, condBlock)
			if mergeTy == nil {
				return false
			}
			comp.scope.previous.assign(name, mergeTy)
		}
	}
	comp.scope = comp.scope.previous
	return true
}

func (comp *compiler) match(span syntax.Span, ty Type) bool {
	switch expr := span.GetExpr().(type) {

	case syntax.Identifier:
		comp.scope.assign(expr.Ident, ty)
		return true

	case syntax.Binary:
		switch expr.Op {
		case syntax.Dot:
			left := comp.compile(expr.Left)
			if left == nil {
				return false
			}
			structure, ok := left.(Struct)
			if !ok {
				comp.throw(fmt.Errorf("cannot use '.' operator on non-structure '%s'", left.Type()))
				return false
			}
			comp.scope = comp.scope.newScope()
			for name, ty := range structure.dict {
				comp.scope.dict[name] = ty
			}
			if !comp.match(expr.Right, ty) {
				return false
			}
			newStruct := Struct{comp.scope.dict}
			comp.scope = comp.scope.previous
			if !comp.match(expr.Left, newStruct) {
				return false
			}
			return true

		}
		comp.throw(fmt.Errorf("invalid binary expression '%s' to match against", expr.Op))
		return false

	case syntax.Tuple:
		tuple, ok := ty.(Tuple)
		if !ok {
			comp.throw(fmt.Errorf("cannot destructure type '%s' as it is not a tuple", ty.Type()))
			return false
		}
		if len(tuple.items) != len(expr.Items) {
			comp.throw(fmt.Errorf("cannot destructure tuple '%s' with '%s' different number of items", tuple.Type(), span))
			return false
		}
		for i := range tuple.items {
			if !comp.match(expr.Items[i], tuple.items[i]) {
				return false
			}
		}
		return true
	}
	comp.throw(fmt.Errorf("invalid expression '%T' to match against", span.GetExpr()))
	return false
}

func (comp *compiler) newMaybe(ty Type, someBlock, noneBlock *backend.Block) Type {
	dest := comp.program.NewValue()
	noneBlock.Copy(noneBlock.Constant(0), dest)
	if maybeTy, ok := ty.(Maybe); ok {
		someBlock.Copy(maybeTy.val, dest)
		ty = maybeTy.ty
	} else {
		someBlock.Copy(someBlock.Constant(1), dest)
	}
	return Maybe{dest, ty}
}

func (comp *compiler) compileBoolExpr(span syntax.Span, ifTrue, ifFalse *backend.Block) bool {
	switch expr := span.GetExpr().(type) {

	case syntax.BooleanLiteral:
		if expr.Value == "true" {
			comp.block.Jump(ifTrue)
		} else {
			comp.block.Jump(ifFalse)
		}

	case syntax.Identifier:
		ty := comp.scope.get(expr.Ident)
		if ty == nil {
			comp.throw(fmt.Errorf("undefined variable '%s'", expr.Ident))
			return false
		}
		boolean, ok := ty.(Boolean)
		if !ok {
			comp.throw(fmt.Errorf("expected a boolean in variable '%s'", expr.Ident))
			return false
		}
		comp.block.JumpIfEqual(boolean.val, comp.block.Constant(1), ifTrue, ifFalse)

	case syntax.Unary:
		switch expr.Op {
		case syntax.Maybe:
			ty := comp.compile(expr.Expr)
			if ty == nil {
				return false
			}
			maybe, ok := ty.(Maybe)
			if !ok {
				comp.throw(fmt.Errorf("was expecting a maybe type before '?' instead of '%s'", ty.Type()))
				return false
			}
			comp.block.JumpIfEqual(maybe.val, comp.block.Constant(1), ifTrue, ifFalse)
			comp.match(expr.Expr, maybe.ty)
		}

	case syntax.Binary:
		switch expr.Op {
		case syntax.LessThan:
			left := comp.compile(expr.Left)
			right := comp.compile(expr.Right)
			if left == nil || right == nil {
				return false
			}
			leftInteger, leftIs := left.(Integer)
			rightInteger, rightIs := right.(Integer)
			if !leftIs || !rightIs {
				comp.throw(fmt.Errorf("incompatiable types for addition"))
				return false
			}
			comp.block.JumpIfGreater(rightInteger.val, leftInteger.val, ifTrue, ifFalse)

		}
	default:
		comp.throw(fmt.Errorf("invalid expression used as boolean %T", span.GetExpr()))
		return false
	}
	return true
}

func (comp *compiler) mergeTypes(a, b Type, aBlock, bBlock *backend.Block) Type {

	if a.Type() == b.Type() {
		dest := comp.Duplicate(a)
		comp.block = aBlock
		comp.Copy(a, dest)
		comp.block = bBlock
		comp.Copy(b, dest)
		return dest
	}

	aStruct, aIs := a.(Struct)
	bStruct, bIs := b.(Struct)

	if aIs && bIs {
		dict := make(map[string]Type)
		for name, ty := range aStruct.dict {
			_, collision := bStruct.dict[name]
			if collision {
				dict[name] = ty
			} else {
				dict[name] = comp.newMaybe(ty, aBlock, bBlock)
			}
		}
		for name, ty := range bStruct.dict {
			other, collision := dict[name]
			if collision {
				dict[name] = comp.mergeTypes(other, ty, aBlock, bBlock)
				if dict[name] == nil {
					return nil
				}
			} else {
				dict[name] = comp.newMaybe(ty, bBlock, aBlock)
			}
		}
		return Struct{dict}
	}

	aMaybe, aIs := a.(Maybe)
	bMaybe, bIs := b.(Maybe)

	if aIs || bIs {
		aVal := aMaybe.val
		if aIs {
			a = aMaybe.ty
		} else {
			aVal = aBlock.Constant(1)
		}
		bVal := bMaybe.val
		if aIs {
			b = bMaybe.ty
		} else {
			bVal = bBlock.Constant(1)
		}
		dest := comp.program.NewValue()
		aBlock.Copy(aVal, dest)
		bBlock.Copy(bVal, dest)
		merged := comp.mergeTypes(a, b, aBlock, bBlock)
		if merged == nil {
			return nil
		}
		return Maybe{dest, merged}
	}

	comp.throw(fmt.Errorf("incompatiable types, '%s' and '%s'", a.Type(), b.Type()))
	return nil
}

func (comp *compiler) Copy(src, dest Type) {
	srcVals := ToValues(src)
	destVals := ToValues(dest)
	for i := range srcVals {
		comp.block.Copy(srcVals[i], destVals[i])
	}
}

func (comp *compiler) Duplicate(src Type) Type {
	vals := make([]*backend.Value, len(ToValues(src)))
	for i := range vals {
		vals[i] = comp.program.NewValue()
	}
	return FromValues(vals, src)
}
