package sqlbuilder

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"testing"
)

type fuzzState struct {
	data                    []byte
	dataIndex               int
	callchainRepresentation string
	currentBuilder          reflect.Value
	usedMethods             map[string]bool
}

func (fs *fuzzState) consumeData(size int) []byte {
	if len(fs.data) <= fs.dataIndex+size {
		return []byte{}
	}
	result := make([]byte, size)
	copy(result, fs.data[fs.dataIndex:fs.dataIndex+size])
	fs.dataIndex += size
	return result
}

func (fs *fuzzState) updateCallchain(method string, args []reflect.Value) {
	fs.callchainRepresentation += "." + method + "("
	for i, arg := range args {
		if i > 0 {
			fs.callchainRepresentation += ", "
		}
		fs.callchainRepresentation += fmt.Sprintf("%q", arg)
	}
	fs.callchainRepresentation += ")"
}

func getSelectBuilderMethods() (map[string]reflect.Type, []string) {
	sb := NewSelectBuilder()
	sbType := reflect.TypeOf(sb)
	// Skip methods that are likely to cause issues or don't return builders
	skipMethods := []string{
		"Build", "String", "BuildWithFlavor", "Flavor",
		"NumCol", "NumValue", "NumAssignment", "TableNames", "Var",
	}

	methodList := make(map[string]reflect.Type)
	methodNames := make([]string, 0, sbType.NumMethod())

	for i := 0; i < sbType.NumMethod(); i++ {
		method := sbType.Method(i)
		if slices.Contains(skipMethods, method.Name) {
			continue
		}

		methodList[method.Name] = method.Type
		methodNames = append(methodNames, method.Name)
	}

	return methodList, methodNames
}

func generateMethodArgs(methodType reflect.Type, state *fuzzState) ([]reflect.Value, bool) {
	numArgs := methodType.NumIn() - 1 // Skip receiver
	isVariadic := methodType.IsVariadic()

	if isVariadic {
		return generateVariadicArgs(methodType, numArgs, state)
	}
	return generateFixedArgs(methodType, numArgs, state)
}

func generateFixedArgs(methodType reflect.Type, numArgs int, state *fuzzState) ([]reflect.Value, bool) {
	args := make([]reflect.Value, numArgs)
	for i := 0; i < numArgs; i++ {
		argType := methodType.In(i + 1) // Skip receiver
		argData := state.consumeData(16)
		args[i] = generateArgumentForType(argType, argData)

		if !args[i].IsValid() {
			return nil, false
		}

		// Additional type compatibility check for complex types
		if argType.Kind() == reflect.Ptr && args[i].Kind() == reflect.Ptr {
			if argType != args[i].Type() {
				return nil, false
			}
		}
	}
	return args, true
}

func generateVariadicArgs(methodType reflect.Type, numArgs int, state *fuzzState) ([]reflect.Value, bool) {
	numFixedArgs := numArgs - 1 // Last parameter is the variadic slice

	// Generate fixed arguments first
	args := make([]reflect.Value, numFixedArgs)
	for i := 0; i < numFixedArgs; i++ {
		argType := methodType.In(i + 1) // Skip receiver
		argData := state.consumeData(16)
		args[i] = generateArgumentForType(argType, argData)
		if !args[i].IsValid() {
			return nil, false
		}
	}

	// Generate variadic arguments (0-3 arguments to keep it reasonable)
	if numFixedArgs < numArgs {
		variadicType := methodType.In(numArgs).Elem() // Get the element type of the slice
		numVariadicArgs := 0
		if len(state.data) > state.dataIndex {
			// 0-3 variadic args, keep the number small while still exercising multiple values
			// TODO: Possible fuzz improvement to allow for more variadic args. Not sure it's worth it.
			numVariadicArgs = int(state.data[state.dataIndex] % 4)
			state.dataIndex++
		}

		for j := 0; j < numVariadicArgs; j++ {
			argData := state.consumeData(16)
			varArg := generateArgumentForType(variadicType, argData)
			if !varArg.IsValid() {
				return nil, false
			}
			args = append(args, varArg)
		}
	}

	return args, true
}

func tryCallMethod(methodName string, methodType reflect.Type, state *fuzzState, t *testing.T) bool {
	// Check if method exists on current builder
	callableMethod := state.currentBuilder.MethodByName(methodName)
	if !callableMethod.IsValid() {
		return false
	}

	// Generate arguments
	args, canCall := generateMethodArgs(methodType, state)
	if !canCall {
		return false
	}

	// Update call chain representation and log it
	state.updateCallchain(methodName, args)
	t.Log("callchain:", state.callchainRepresentation)

	// Mark this method as used
	state.usedMethods[methodName] = true

	// Call method and capture result for chaining
	result := callableMethod.Call(args)

	// Only chain if method returns the same builder type (SelectBuilder)
	if len(result) > 0 && result[0].IsValid() {
		resultType := result[0].Type()
		if resultType.Kind() == reflect.Ptr &&
			resultType.String() == "*sqlbuilder.SelectBuilder" &&
			!result[0].IsNil() {
			state.currentBuilder = result[0]
		}
	}

	return true
}

func executeMethodChain(methodList map[string]reflect.Type, methodNames []string, state *fuzzState, maxChains uint8, t *testing.T) {
	for nbFunc := uint8(0); nbFunc < maxChains; nbFunc++ {
		methodCalled := false

		// Try to find a method we haven't used yet to create more varied chains
		for _, method := range methodNames {
			// Skip methods we've already used to create more variety
			if state.usedMethods[method] && len(state.usedMethods) < len(methodNames) {
				continue
			}

			if tryCallMethod(method, methodList[method], state, t) {
				methodCalled = true
				break // Move to next chain iteration
			}
		}

		// If no method could be called, break the chain
		if !methodCalled {
			break
		}
	}
}

func finalizeBuild(state *fuzzState) {
	// Always try to build the final SQL to ensure it doesn't panic
	if state.currentBuilder.IsValid() {
		buildMethod := state.currentBuilder.MethodByName("Build")
		if buildMethod.IsValid() {
			buildMethod.Call([]reflect.Value{})
		}
	}
}

func FuzzSelect(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte, seed int64, numberOfChainedFunction uint8) {
		if len(data) == 0 {
			return
		}

		// Get all available methods for SelectBuilder
		methodList, methodNames := getSelectBuilderMethods()

		// Randomize method order deterministically based on seed
		r := rand.New(rand.NewSource(seed))
		r.Shuffle(len(methodNames), func(i, j int) {
			methodNames[i], methodNames[j] = methodNames[j], methodNames[i]
		})

		// Initialize fuzzing state
		state := &fuzzState{
			data:                    data,
			dataIndex:               0,
			callchainRepresentation: "NewSelectBuilder()",
			currentBuilder:          reflect.ValueOf(NewSelectBuilder()),
			usedMethods:             make(map[string]bool),
		}

		// Limit the number of chained functions to prevent infinite loops
		maxChains := numberOfChainedFunction
		if maxChains > 10 {
			maxChains = 10
		}

		// Execute method chain
		executeMethodChain(methodList, methodNames, state, maxChains, t)

		t.Logf("Final callchain: %s", state.callchainRepresentation)
		// Try to build the final result
		finalizeBuild(state)
	})
}

// generateArgumentForType generates a reflect.Value for the given type based on the provided data.
// It will consume the data slice to create a value of the specified type.
// It handles specific custom types like JoinOption and Flavor, and Go will consider them disntinct types than their aliases.
func generateArgumentForType(argType reflect.Type, data []byte) reflect.Value {
	switch argType.Kind() {
	case reflect.String:
		// Handle specific custom string types first
		if argType.String() == "sqlbuilder.JoinOption" {
			joinOptions := []JoinOption{
				FullJoin, FullOuterJoin, InnerJoin,
				LeftJoin, LeftOuterJoin, RightJoin, RightOuterJoin,
			}
			if len(data) > 0 {
				return reflect.ValueOf(joinOptions[int(data[0])%len(joinOptions)])
			}
			return reflect.ValueOf(InnerJoin)
		}
		// Use remaining data as string
		return reflect.ValueOf(string(data))
	case reflect.Int:
		// Handle specific custom int types first
		if argType.String() == "sqlbuilder.Flavor" {
			return reflect.ValueOf(DefaultFlavor)
		}
		if len(data) > 0 {
			return reflect.ValueOf(int(data[0]))
		}
		return reflect.ValueOf(0)
	case reflect.Bool:
		if len(data) > 0 {
			return reflect.ValueOf(data[0]%2 == 0)
		}
		return reflect.ValueOf(false)
	case reflect.Int8:
		if len(data) > 0 {
			return reflect.ValueOf(int8(data[0]))
		}
		return reflect.ValueOf(int8(0))
	case reflect.Int16:
		if len(data) >= 2 {
			return reflect.ValueOf(int16(data[0])<<8 | int16(data[1]))
		}
		return reflect.ValueOf(int16(0))
	case reflect.Int32:
		if len(data) >= 4 {
			return reflect.ValueOf(int32(data[0])<<24 | int32(data[1])<<16 | int32(data[2])<<8 | int32(data[3]))
		}
		return reflect.ValueOf(int32(0))
	case reflect.Int64:
		if len(data) >= 8 {
			val := int64(data[0])<<56 | int64(data[1])<<48 | int64(data[2])<<40 | int64(data[3])<<32 |
				int64(data[4])<<24 | int64(data[5])<<16 | int64(data[6])<<8 | int64(data[7])
			return reflect.ValueOf(val)
		}
		return reflect.ValueOf(int64(0))
	case reflect.Uint:
		if len(data) > 0 {
			return reflect.ValueOf(uint(data[0]))
		}
		return reflect.ValueOf(uint(0))
	case reflect.Uint8:
		if len(data) > 0 {
			return reflect.ValueOf(uint8(data[0]))
		}
		return reflect.ValueOf(uint8(0))
	case reflect.Uint16:
		if len(data) >= 2 {
			return reflect.ValueOf(uint16(data[0])<<8 | uint16(data[1]))
		}
		return reflect.ValueOf(uint16(0))
	case reflect.Uint32:
		if len(data) >= 4 {
			return reflect.ValueOf(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
		}
		return reflect.ValueOf(uint32(0))
	case reflect.Uint64:
		if len(data) >= 8 {
			val := uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 |
				uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])
			return reflect.ValueOf(val)
		}
		return reflect.ValueOf(uint64(0))
	case reflect.Float32:
		if len(data) >= 4 {
			bits := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
			return reflect.ValueOf(float32(bits))
		}
		return reflect.ValueOf(float32(0))
	case reflect.Float64:
		if len(data) >= 8 {
			bits := uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 |
				uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])
			return reflect.ValueOf(float64(bits))
		}
		return reflect.ValueOf(float64(0))
	case reflect.Slice:
		if argType.Elem().Kind() == reflect.String {
			return reflect.ValueOf([]string{string(data)})
		}
		if argType.Elem().Kind() == reflect.Interface {
			return reflect.ValueOf([]interface{}{string(data)})
		}
		return reflect.ValueOf([]interface{}{string(data)})
	case reflect.Ptr:
		// Handle pointer types by creating a pointer to the underlying type
		// Handle specific pointer types
		if argType == reflect.TypeOf((*WhereClause)(nil)) {
			return reflect.ValueOf(NewWhereClause())
		}
		if argType == reflect.TypeOf((*SelectBuilder)(nil)) {
			return reflect.ValueOf(NewSelectBuilder())
		}
		if argType == reflect.TypeOf((*Args)(nil)) {
			return reflect.ValueOf(&Args{})
		}
		if argType == reflect.TypeOf((*CTEBuilder)(nil)) {
			return reflect.ValueOf(DefaultFlavor.NewCTEBuilder())
		}
		if argType == reflect.TypeOf((*InsertBuilder)(nil)) {
			return reflect.ValueOf(DefaultFlavor.NewInsertBuilder())
		}
		if argType == reflect.TypeOf((*UpdateBuilder)(nil)) {
			return reflect.ValueOf(DefaultFlavor.NewUpdateBuilder())
		}
		if argType == reflect.TypeOf((*DeleteBuilder)(nil)) {
			return reflect.ValueOf(DefaultFlavor.NewDeleteBuilder())
		}
		// For other pointer types, create a pointer to the underlying type
		str := string(data)
		return reflect.ValueOf(&str)
	case reflect.Interface:
		// Handle specific interface types
		if argType.String() == "sqlbuilder.Builder" {
			// Create a simple SelectBuilder for Builder interface
			return reflect.ValueOf(NewSelectBuilder())
		}
		return reflect.ValueOf(string(data))
	default:
		// For other types, use zero value
		return reflect.Zero(argType)
	}
}
