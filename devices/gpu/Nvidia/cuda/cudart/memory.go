package cudart

import (
	"errors"
	"fmt"

	"github.com/dereklstinson/half"

	"github.com/dereklstinson/GoCuNets/devices"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//CudaSlice is cuda memory.
type CudaSlice struct {
	mem       *gocudnn.Malloced
	dtype     devices.Type
	device    bool
	length    uint
	capacity  uint
	memcpyflg gocudnn.MemcpyKindFlag
}

//MallocDevice allocates memory to set device
func mallocdevice(dtype devices.Type, length, capacity uint) (*CudaSlice, error) {
	mem, err := gocudnn.Malloc(gocudnn.SizeT(capacity * uint(dtype)))
	if err != nil {
		return nil, err
	}
	return &CudaSlice{
		length:   length,
		capacity: capacity,
		mem:      mem,
		dtype:    dtype,
		device:   true,
	}, nil

}

//MallocCudaHost allocates paged host memory that is usable by cuda
func mallochost(dtype devices.Type, length, capacity uint) (*CudaSlice, error) {
	mem, err := gocudnn.MallocHost(gocudnn.SizeT(capacity * dtype.SizeOf().Uint()))
	if err != nil {
		return nil, err
	}
	return &CudaSlice{

		mem: mem,
	}, nil

}

func makesice() (*CudaSlice, error) {
	return Make([]int{}, 20)
}

//Append appends the slice
func (c *CudaSlice) Append(val interface{}) {
	if devices.Len(val) <= c.capacity-c.length {
		offset := c.length
		c.length += devices.Len(val)
		c.Set(val, offset)
		return
	}
	if devices.Len(val) > c.capacity-c.length {
		newsize := c.length + devices.Len(val)
		newcudaslice, err := mallocdevice(c.dtype, c.length, newsize)
		if err != nil {
			panic(err)
		}

		err = gocudnn.CudaMemCopyUnsafe(newcudaslice.mem.Ptr(), c.mem.Ptr(), gocudnn.SizeT(c.length*c.dtype.SizeOf().Uint()), c.memcpyflg.DeviceToDevice())
		if err != nil {
			panic(err)
		}
		c = newcudaslice
		c.Append(val)
		return
	}

}

//Get will get values of a cudaslice and fill the slice
func (c *CudaSlice) Get(val interface{}, offset uint) {
	length := devices.Len(val)
	if offset+length > c.length {

		panic(
			fmt.Sprintf("Illegal Access SliceLength: %d, Offset: %d, ValLength: %d", c.length, offset, offset+length),
		)
	}

	gptr, err := gocudnn.MakeGoPointer(val)
	if err != nil {
		panic(err)
	}
	destloc := c.mem.Offset(offset + c.dtype.SizeOf().Uint())
	err = gocudnn.CudaMemCopyUnsafe(gptr.Ptr(), destloc, gptr.ByteSize(), c.memcpyflg.HostToDevice())
	if err != nil {
		panic(err)
	}

}

//Type returns the devices.Type the cudaslice contains
func (c *CudaSlice) Type() devices.Type {
	return c.dtype
}

//Length returns the length of elements in the CudaSlice
func (c *CudaSlice) Length() uint {
	return c.length
}

//Set is an experimental function that will set a value or slice of values from host into cuda mem
func (c *CudaSlice) Set(val interface{}, offset uint) {
	length := devices.Len(val)
	if offset+length > c.length {

		panic(
			fmt.Sprintf("Illegal Access SliceLength: %d, Offset: %d, ValLength: %d", c.length, offset, offset+length),
		)
	}

	gptr, err := gocudnn.MakeGoPointer(val)
	if err != nil {
		panic(err)
	}

	destloc := c.mem.Offset(offset + c.dtype.SizeOf().Uint())
	err = gocudnn.CudaMemCopyUnsafe(destloc, gptr.Ptr(), gptr.ByteSize(), c.memcpyflg.HostToDevice())
	if err != nil {
		panic(err)
	}

}

//Make is like the make function of golang pass []type{}, gotypes with the exception of device.FLoat16
//Then you define the length and the capacity.
func Make(x interface{}, args ...uint) (*CudaSlice, error) {

	switch len(args) {
	case 1:
		return make1(x, args...)
	case 2:
		return make2(x, args...)

	}

	return nil, errors.New("Unsupported Length of args")
}

func make1(x interface{}, args ...uint) (*CudaSlice, error) {

	switch x.(type) {
	case []uint8:
		return mallocdevice(devices.Uint8, args[0], args[0])
	case []int8:
		return mallocdevice(devices.Int8, args[0], args[0])
	case []uint16:
		return mallocdevice(devices.Uint16, args[0], args[0])
	case []int16:
		return mallocdevice(devices.Int16, args[0], args[0])
	case []uint32:
		return mallocdevice(devices.Uint32, args[0], args[0])
	case []int32:
		return mallocdevice(devices.Int32, args[0], args[0])
	case []uint64:
		return mallocdevice(devices.Uint64, args[0], args[0])
	case []int64:
		return mallocdevice(devices.Int64, args[0], args[0])
	case []uint:
		return mallocdevice(devices.Uint, args[0], args[0])
	case []int:
		return mallocdevice(devices.Int, args[0], args[0])
	case []devices.Float16:
		return mallocdevice(devices.Float16H, args[0], args[0])
	case []half.Float16:
		return mallocdevice(devices.Float16H, args[0], args[0])
	case []float32:
		return mallocdevice(devices.Float32, args[0], args[0])
	case []float64:
		return mallocdevice(devices.Float64, args[0], args[0])

	}

	return nil, errors.New("Unsupported Type")
}

func make2(x interface{}, args ...uint) (*CudaSlice, error) {
	if args[1] < args[0] {
		return nil, errors.New("Length has to be less than Capacity")
	}
	switch x.(type) {
	case []uint8:
		return mallocdevice(devices.Uint8, args[0], args[1])
	case []int8:
		return mallocdevice(devices.Int8, args[0], args[1])
	case []uint16:
		return mallocdevice(devices.Uint16, args[0], args[1])
	case []int16:
		return mallocdevice(devices.Int16, args[0], args[1])
	case []uint32:
		return mallocdevice(devices.Uint32, args[0], args[1])
	case []int32:
		return mallocdevice(devices.Int32, args[0], args[1])
	case []uint64:
		return mallocdevice(devices.Uint64, args[0], args[1])
	case []int64:
		return mallocdevice(devices.Int64, args[0], args[1])
	case []uint:
		return mallocdevice(devices.Uint, args[0], args[1])
	case []int:
		return mallocdevice(devices.Int, args[0], args[1])
	case []devices.Float16:
		return mallocdevice(devices.Float16H, args[0], args[1])
	case []half.Float16:
		return mallocdevice(devices.Float16H, args[0], args[1])
	case []float32:
		return mallocdevice(devices.Float32, args[0], args[1])
	case []float64:
		return mallocdevice(devices.Float64, args[0], args[1])

	}

	return nil, errors.New("Unsupported Type")
}
