package goja

import (
	"fmt"
	"github.com/gogap/config"
	"time"

	crand "crypto/rand"
	"encoding/binary"
	"io/ioutil"
	"math/rand"

	"github.com/gogap/flow"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gogap/context"
)

func init() {
	flow.RegisterHandler("lang.javascript.goja", ExecuteJS)
}

func ExecuteJS(ctx context.Context, conf config.Configuration) (err error) {

	src := conf.GetString("src")
	if len(src) == 0 {
		err = fmt.Errorf("src not set")
		return
	}

	timeout := conf.GetTimeDuration("timeout", 10*time.Minute)

	vm := newVM()
	vm.Set("ctx", ctx)
	vm.Set("conf", conf)

	var data []byte
	data, err = ioutil.ReadFile(src)
	if err != nil {
		return
	}

	var prg *goja.Program

	prg, err = goja.Compile(src, string(data), false)
	if err != nil {
		return
	}

	if timeout > 0 {
		time.AfterFunc(timeout, func() {
			vm.Interrupt("timeout")
		})
	}

	_, err = vm.RunProgram(prg)

	if err != nil {
		err = fmt.Errorf("execute goja script failure: %s\n%s\n", err.Error(), src)
		return
	}

	return
}

func newRandSource() goja.RandSource {
	var seed int64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &seed); err != nil {
		panic(fmt.Errorf("Could not read random bytes: %v", err))
	}
	return rand.New(rand.NewSource(seed)).Float64
}

func newVM() *goja.Runtime {
	vm := goja.New()
	vm.SetRandSource(newRandSource())
	new(require.Registry).Enable(vm)
	console.Enable(vm)

	return vm
}
