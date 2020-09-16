package frontend

import (
	"fmt"
	"language/backend"
	"language/syntax"
	"sort"
	"strings"
)

type Type interface {
	// Value(*backend.VirtualMachine) string
	Type() string
}

type Integer struct {
	val *backend.Value
}

type Boolean struct {
	val *backend.Value
}

type Maybe struct {
	val *backend.Value
	ty  Type
}

type Func struct {
	param syntax.Span
	body  syntax.Span
	scope *scope
	impls []Impl
}

type Impl struct {
	params  Type
	returns Type
	scope   *scope
	block   *backend.Block
}

type Struct struct {
	dict map[string]Type
}

type Tuple struct {
	items []Type
}

// func (ty Integer) Value(vm *backend.VirtualMachine) string {
// 	return fmt.Sprintf("%d", vm.GetValue(ty.val))
// }

// func (ty Boolean) Value(vm *backend.VirtualMachine) string {
// 	if vm.GetValue(ty.val) == 1 {
// 		return "true"
// 	}
// 	return "false"
// }

// func (ty Maybe) Value(vm *backend.VirtualMachine) string {
// 	if vm.GetValue(ty.val) == 1 {
// 		return ty.ty.Value(vm)
// 	}
// 	return "none"
// }

// func (ty Func) Value(vm *backend.VirtualMachine) string {
// 	return "fn"
// }

// func (ty Struct) Value(vm *backend.VirtualMachine) string {
// 	fields := make([]string, len(ty.dict))
// 	order := []string{}
// 	for name := range ty.dict {
// 		order = append(order, name)
// 	}
// 	sort.Strings(order)
// 	for i, name := range order {
// 		fields[i] = fmt.Sprintf("%s: %s", name, ty.dict[name].Value(vm))
// 	}
// 	return "{" + strings.Join(fields, ", ") + "}"
// }

// func (ty Tuple) Value(vm *backend.VirtualMachine) string {
// 	items := make([]string, len(ty.items))
// 	for i, ty := range ty.items {
// 		items[i] = ty.Value(vm)
// 	}
// 	return "(" + strings.Join(items, ", ") + ")"
// }

func (ty Integer) Type() string {
	return "int"
}

func (ty Boolean) Type() string {
	return "bool"
}

func (ty Maybe) Type() string {
	return fmt.Sprintf("%s?", ty.ty.Type())
}

func (ty Func) Type() string {
	return "fn"
}

func (ty Struct) Type() string {
	fields := make([]string, len(ty.dict))
	order := []string{}
	for name := range ty.dict {
		order = append(order, name)
	}
	sort.Strings(order)
	for i, name := range order {
		fields[i] = fmt.Sprintf("%s: %s", name, ty.dict[name].Type())
		i++
	}
	return "{" + strings.Join(fields, ", ") + "}"
}

func (ty Tuple) Type() string {
	items := make([]string, len(ty.items))
	for i, ty := range ty.items {
		items[i] = ty.Type()
	}
	return "(" + strings.Join(items, ", ") + ")"
}

func ToValues(ty Type) []*backend.Value {
	vals := []*backend.Value{}
	appendValues(ty, &vals)
	return vals
}

func appendValues(ty Type, vals *[]*backend.Value) {
	switch ty := ty.(type) {
	case Integer:
		*vals = append(*vals, ty.val)
	case Boolean:
		*vals = append(*vals, ty.val)
	case Maybe:
		*vals = append(*vals, ty.val)
		appendValues(ty.ty, vals)
	case Struct:
		order := []string{}
		for name := range ty.dict {
			order = append(order, name)
		}
		sort.Strings(order)
		for _, name := range order {
			appendValues(ty.dict[name], vals)
		}
	case Tuple:
		for _, item := range ty.items {
			appendValues(item, vals)
		}
	case Func:
		break
	default:
		panic(fmt.Sprintf("Undefined type %T", ty))
	}
}

func FromValues(vals []*backend.Value, ty Type) Type {
	newTy, _ := castValuesToType(vals, ty)
	return newTy
}

func castValuesToType(vals []*backend.Value, ty Type) (Type, []*backend.Value) {
	switch ty := ty.(type) {
	case Integer:
		return Integer{vals[0]}, vals[1:]
	case Boolean:
		return Boolean{vals[0]}, vals[1:]
	case Maybe:
		newTy, leftover := castValuesToType(vals[1:], ty.ty)
		return Maybe{vals[0], newTy}, leftover
	case Struct:
		newStruct := Struct{make(map[string]Type)}
		order := []string{}
		for name := range ty.dict {
			order = append(order, name)
		}
		sort.Strings(order)
		for _, name := range order {
			newTy, leftoverRegs := castValuesToType(vals, ty.dict[name])
			newStruct.dict[name] = newTy
			vals = leftoverRegs
		}
		return newStruct, vals
	case Tuple:
		items := make([]Type, len(ty.items))
		for i, item := range ty.items {
			newTy, leftoverRegs := castValuesToType(vals, item)
			items[i] = newTy
			vals = leftoverRegs
		}
		return Tuple{items}, vals
	case Func:
		return ty, vals
	default:
		panic("unreachable")
	}
}
