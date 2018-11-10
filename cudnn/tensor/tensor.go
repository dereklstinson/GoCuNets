//Package tensor is used to make tensors by using gocudnn.  It is currently not supporting what I call the "EX" functions.
//because the Tensor struct is also going to be carrying a filter descripter.  Also I call it "EX" functions loosly, because I think
//there is a miss labeling of the function names in cudnn. Basicly it is the set tensor fuctions that don't include the format and include
//the strides asdf
package tensor

//import "C"
import (
	"errors"
	"fmt"

	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/utils"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//Volume holds both a gocudnn.TensorD and gocudnn.FilterD and the allocated memory associated with it
type Volume struct {
	freed     bool
	tD        *gocudnn.TensorD
	tDstrided *gocudnn.TensorD
	fD        *gocudnn.FilterD
	dtype     gocudnn.DataType
	propnan   gocudnn.PropagationNAN
	memgpu    *gocudnn.Malloced
	frmt      gocudnn.TensorFormat
	thelp     gocudnn.Tensor
	fhelp     gocudnn.Filter
	ophelp    gocudnn.OpTensor
	dims      []int32
	strides   []int32
	managed   bool
	//scalar gocudnn.CScalar
}

//Info struct contains the info that is needed to build a volume
type Info struct {
	Format   gocudnn.TensorFormat   `json:"Format"`
	DataType gocudnn.DataType       `json:"DataType"`
	Nan      gocudnn.PropagationNAN `json:"Nan"`
	Dims     []int32                `json:"Dims"`
	Unified  bool                   `json:"Unified"`
	Values   interface{}            `json:"Values"`
}

//MakeInfo makes an info struct
func MakeInfo(frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, unified bool) Info {
	return Info{
		Format:   frmt,
		DataType: dtype,
		Dims:     dims,
		Unified:  unified,
	}
}

//DeleteMem will free the mem the tensor has for the gpu. if the mem is already freed it will return nil
func (t *Volume) DeleteMem() error {
	if t.freed != true {
		return t.memgpu.Free()
	}
	return nil
}

//ReBuildMem will rebuild the gpu mem if ConncervedGPUmem was used. If mem wasn't freed then it will do nothing and return nil
func (t *Volume) ReBuildMem() error {
	if t.freed == true {
		return nil
	}
	sizeT, err := t.tD.GetSizeInBytes()
	if err != nil {
		return err
	}
	if t.managed == true {
		t.memgpu, err = gocudnn.MallocManaged(sizeT, gocudnn.ManagedMemFlag{}.Global())
		return err
	}
	t.memgpu, err = gocudnn.Malloc(sizeT)
	return err
}

//SetPropNan will change the default nan propigation flag from PropNanNon to PropNaN
func (t *Volume) SetPropNan() {
	t.propnan = t.thelp.Flgs.NaN.PropagateNan()
}

//SetNotPropNan will set the nan propigation flag to NotPropigationNan (NotPropigationNan is default)
func (t *Volume) SetNotPropNan() {
	t.propnan = t.thelp.Flgs.NaN.NotPropagateNan()

}

//Flags returns a struct that passes gocudnn flags through methods used in building the tensor
func Flags() gocudnn.TensorFlags {
	return gocudnn.TensorFlags{}
}

//Info returns an Info struct that is used for saving info. If an error is returned then the values of Info will be set to default golang's default
func (t *Volume) Info() (Info, error) {
	frmt, dtype, dims, err := t.Properties()

	if err != nil {
		return Info{}, err
	}
	dflgs := t.thelp.Flgs.Data
	var values interface{}
	size := utils.FindVolumeInt32(dims)

	//I don't like this switch type stuff.  I am probably going to make something in the gocudnn package to get rid of this. I just haven't thought of a really easy way to implement this.
	switch dtype.Cu() {
	case dflgs.Double():
		values = make([]float64, size)
	case dflgs.Float():
		values = make([]float32, size)
	case dflgs.Int32():
		values = make([]int32, size)
	case dflgs.Int8():
		values = make([]float64, size)

	default:
		return Info{}, errors.New("Unsupported Format : Most likely internal error. Contact Code Writer")
	}
	err = t.memgpu.FillSlice(values)
	if err != nil {
		return Info{}, err
	}
	return Info{
		Format:   frmt.Cu(),
		DataType: dtype.Cu(),
		Dims:     dims,
		Unified:  t.managed,
		Values:   values,
	}, nil
}

//Build is a method for Info that will retrun a volume type. If Weights is nil the memory will still be malloced on the cuda side.  So make sure to add values if needed.
func (i Info) Build() (*Volume, error) {
	var thelper gocudnn.Tensor
	var fhelper gocudnn.Filter
	if len(i.Dims) < 4 {
		return nil, errors.New("Dims less than 4. Create A 4 dim Tensor and set dims not needed to 1")
	}
	var newmemer *gocudnn.Malloced
	var tens *gocudnn.TensorD
	var filts *gocudnn.FilterD
	var err error
	if len(i.Dims) > 4 {
		tens, err = thelper.NewTensorNdDescriptorEx(i.Format, i.DataType, i.Dims)
		if err != nil {
			return nil, err
		}
		filts, err = fhelper.NewFilterNdDescriptor(i.DataType, i.Format, i.Dims)
		if err != nil {
			tens.DestroyDescriptor()
			return nil, err
		}
		size, err := tens.GetSizeInBytes()

		if err != nil {
			tens.DestroyDescriptor()
			filts.DestroyDescriptor()
			return nil, err
		}
		if i.Unified == true {
			newmemer, err = gocudnn.MallocManaged(size, gocudnn.ManagedMemFlag{}.Global())
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err
			}
			newmemer.Set(0)

		} else {

			newmemer, err = gocudnn.Malloc(size)
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err

			}
			newmemer.Set(0)

		}

	} else {

		tens, err = thelper.NewTensor4dDescriptor(i.DataType, i.Format, i.Dims)
		if err != nil {
			return nil, err
		}
		filts, err = fhelper.NewFilter4dDescriptor(i.DataType, i.Format, i.Dims)
		if err != nil {
			tens.DestroyDescriptor()
			return nil, err
		}
		size, err := tens.GetSizeInBytes()
		if err != nil {
			tens.DestroyDescriptor()
			filts.DestroyDescriptor()
			return nil, err
		}
		if i.Unified == true {

			newmemer, err = gocudnn.MallocManaged(size, gocudnn.ManagedMemFlag{}.Global())
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err

			}
			newmemer.Set(0)
		} else {
			newmemer, err = gocudnn.Malloc(size)
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err

			}
			newmemer.Set(0)

		}

	}

	vol := &Volume{
		tD:      tens,
		fD:      filts,
		memgpu:  newmemer,
		frmt:    i.Format,
		dtype:   i.DataType,
		dims:    i.Dims,
		strides: utils.FindStridesInt32(i.Dims),
	}
	if i.Values == nil {
		return vol, nil
	}
	goptr, err := gocudnn.MakeGoPointer(i.Values)
	if err != nil {
		vol.Destroy()
		return nil, err
	}
	err = vol.LoadMem(goptr)
	if err != nil {
		vol.Destroy()
		return nil, err
	}
	return vol, nil
}

//Unified returns if the memory is under the unified memory system
func (t *Volume) Unified() bool {
	return t.managed
}

//BuildFromTensorD will take a TensorD and the volume
func BuildFromTensorD(desc *gocudnn.TensorD, managed bool) (*Volume, error) {
	dtype, dims, _, err := desc.GetDescrptor()
	if err != nil {
		return nil, err
	}
	frmt, err := desc.GetFormat()
	if err != nil {
		return nil, err
	}

	desc.DestroyDescriptor()

	return build(frmt, dtype, dims, managed)
}

//Build creates a tensor and mallocs the memory for the tensor
func Build(frmt cudnn.TensorFormat, dtype cudnn.DataType, dims []int32, managed bool) (*Volume, error) {
	return build(frmt.Cu(), dtype.Cu(), dims, managed)
}

//Build creates a tensor and mallocs the memory for the tensor
func build(frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, managed bool) (*Volume, error) {
	var thelper gocudnn.Tensor
	var fhelper gocudnn.Filter
	if len(dims) < 4 {
		return nil, errors.New("Dims less than 4. Create A 4 dim Tensor and set dims not needed to 1")
	}
	var newmemer *gocudnn.Malloced
	var tens *gocudnn.TensorD
	var filts *gocudnn.FilterD
	var tensstrided *gocudnn.TensorD
	var err error
	if len(dims) > 4 {
		tens, err = thelper.NewTensorNdDescriptorEx(frmt, dtype, dims)
		if err != nil {
			return nil, err
		}
		filts, err = fhelper.NewFilterNdDescriptor(dtype, frmt, dims)
		if err != nil {
			tens.DestroyDescriptor()
			return nil, err
		}

		tensstrided, err = thelper.NewTensorNdDescriptor(dtype, dims, utils.FindStridesInt32(dims))
		if err != nil {
			tens.DestroyDescriptor()
			return nil, err
		}
		size, err := tens.GetSizeInBytes()
		if err != nil {
			tens.DestroyDescriptor()
			filts.DestroyDescriptor()
			return nil, err
		}
		if managed == true {
			newmemer, err = gocudnn.MallocManaged(size, gocudnn.ManagedMemFlag{}.Global())
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err
			}

		} else {

			newmemer, err = gocudnn.Malloc(size)
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err

			}

		}

	} else {

		tens, err = thelper.NewTensor4dDescriptor(dtype, frmt, dims)
		if err != nil {
			return nil, err
		}
		tensstrided, err = thelper.NewTensor4dDescriptorEx(dtype, dims, utils.FindStridesInt32(dims))
		if err != nil {
			tens.DestroyDescriptor()
			return nil, err
		}
		filts, err = fhelper.NewFilter4dDescriptor(dtype, frmt, dims)
		if err != nil {
			tens.DestroyDescriptor()
			return nil, err
		}
		size, err := tens.GetSizeInBytes()
		if err != nil {
			tens.DestroyDescriptor()
			filts.DestroyDescriptor()
			return nil, err
		}
		if managed == true {

			newmemer, err = gocudnn.MallocManaged(size, gocudnn.ManagedMemFlag{}.Global())
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err

			}

		} else {
			newmemer, err = gocudnn.Malloc(size)
			if err != nil {

				tens.DestroyDescriptor()
				filts.DestroyDescriptor()
				return nil, err

			}

		}

	}
	err = newmemer.Set(0)
	if err != nil {
		newmemer.Free()
		return nil, err
	}
	return &Volume{
		tD:        tens,
		tDstrided: tensstrided,
		fD:        filts,
		memgpu:    newmemer,
		frmt:      frmt,
		dtype:     dtype,
		dims:      dims,
		strides:   utils.FindStridesInt32(dims),
	}, nil

}

//DataType returns the datatype of the volume
func (t *Volume) DataType() cudnn.DataType {
	return cudnn.DataType(t.dtype)
}

//Format returns the format of the volume
func (t *Volume) Format() cudnn.TensorFormat {
	return cudnn.TensorFormat(t.frmt)
}

//TDStrided is a function that returns the strided tensor descriptor.
func (t *Volume) TDStrided() *gocudnn.TensorD {
	return t.tDstrided
}

//TD returns the tensor descriptor for Tensor
func (t *Volume) TD() *gocudnn.TensorD {
	return t.tD
}

//FD returns the filter descriptor for Tensor
func (t *Volume) FD() *gocudnn.FilterD {
	return t.fD
}

//Memer returns the Memer for Tensor
func (t *Volume) Memer() *gocudnn.Malloced {
	return t.memgpu
}

//Size returns the size in bytes in type gocudnn.SizeT
func (t *Volume) Size() (gocudnn.SizeT, error) {
	return t.tD.GetSizeInBytes()
}

//Properties returns the properties of the tensor
func (t *Volume) Properties() (cudnn.TensorFormat, cudnn.DataType, []int32, error) {
	a, b, _, err := t.tD.GetDescrptor()

	return cudnn.TensorFormat(t.frmt), cudnn.DataType(a), b, err

}

//ZeroClone returns a zero clone of the the memory
func (t *Volume) ZeroClone() (*Volume, error) {

	if t.tD == nil || t.fD == nil || t.memgpu == nil {
		return nil, errors.New("Tensor is nil")
	}
	dtype, dims, strides, err := t.tD.GetDescrptor()
	if err != nil {
		return nil, err
	}

	var filt *gocudnn.FilterD
	var tens *gocudnn.TensorD
	if len(strides) > 0 {
		if len(dims) > 4 {
			tens, err = t.thelp.NewTensorNdDescriptor(dtype, dims, strides)
		} else {
			tens, err = t.thelp.NewTensor4dDescriptorEx(dtype, dims, strides)
		}
	} else {
		if len(dims) > 4 {
			tens, err = t.thelp.NewTensorNdDescriptorEx(t.frmt, dtype, dims)
		} else {
			tens, err = t.thelp.NewTensor4dDescriptor(dtype, t.frmt, dims)
		}
	}
	if err != nil {
		return nil, err
	}
	if len(dims) > 4 {
		filt, err = t.fhelp.NewFilterNdDescriptor(dtype, t.frmt, dims)
	} else {
		filt, err = t.fhelp.NewFilter4dDescriptor(dtype, t.frmt, dims)
	}
	if err != nil {
		return nil, err
	}
	var newmem *gocudnn.Malloced
	if t.managed == true {
		newmem, err = gocudnn.MallocManaged(t.memgpu.ByteSize(), gocudnn.ManagedMemFlag{}.Global())
	} else {
		newmem, err = gocudnn.Malloc(t.memgpu.ByteSize())
	}

	if err != nil {
		return nil, err

	}
	err = newmem.Set(0)
	if err != nil {
		return nil, err
	}

	return &Volume{tD: tens, fD: filt, memgpu: newmem, frmt: t.frmt}, nil
}

func destroy(t *Volume) error {
	var flag bool

	err1 := t.tD.DestroyDescriptor()
	if err1 != nil {
		flag = true
	}
	err2 := t.fD.DestroyDescriptor()
	if err2 != nil {
		flag = true
	}
	err3 := t.memgpu.Free()
	if err3 != nil {
		flag = true
	}
	err4 := t.tDstrided.DestroyDescriptor()

	if flag == true {
		return fmt.Errorf("error::TensorD: %s, FilterD: %s, TensorDstrided: %s, Memory: %s", err1, err2, err4, err3)
	}
	return nil
}

//arraysize will return the size of the array and will return 0 if unsupported type is used.
func arraysize(dtype gocudnn.DataType, size gocudnn.SizeT) int {
	var flg gocudnn.DataTypeFlag
	x := int(size)
	switch dtype {
	case flg.Double():
		return x / 8
	case flg.Float():
		return x / 4
	case flg.Int32():
		return x / 4
	case flg.UInt8():
		return x
	case flg.Int8():
		return x
	default:
		return 0
	}
}

//PrintUnifiedMem prints the unified Memory
func (t *Volume) PrintUnifiedMem(comment string) error {
	kind := gocudnn.MemcpyKindFlag{}.Default()
	return t.printmem(comment, kind)
}

func (t *Volume) printmem(comment string, kind gocudnn.MemcpyKind) error {
	var flg gocudnn.DataTypeFlag
	sib := t.memgpu.ByteSize()
	as := arraysize(t.dtype, sib)

	switch t.dtype {
	case flg.Double():

		array := make([]float64, as)
		ptr, err := gocudnn.MakeGoPointer(array)
		if err != nil {
			return err
		}
		err = gocudnn.CudaMemCopy(ptr, t.memgpu, sib, kind)
		if err != nil {
			return err
		}
		fmt.Printf("\n{")
		fmt.Println(comment, array)
		fmt.Printf("\n}")
	case flg.Float():

		array := make([]float32, as)
		ptr, err := gocudnn.MakeGoPointer(array)

		if err != nil {
			return err
		}
		err = gocudnn.CudaMemCopy(ptr, t.memgpu, sib, kind)
		if err != nil {
			return err
		}
		fmt.Printf("\n{")
		fmt.Println(comment, array)
		fmt.Printf("\n}")
	case flg.Int32():

		array := make([]int32, as)
		ptr, err := gocudnn.MakeGoPointer(array)

		if err != nil {
			return err
		}
		err = gocudnn.CudaMemCopy(ptr, t.memgpu, sib, kind)
		if err != nil {
			return err
		}
		fmt.Printf("\n{")
		fmt.Println(comment, array)
		fmt.Printf("\n}")
	case flg.UInt8():

		array := make([]byte, as)
		ptr, err := gocudnn.MakeGoPointer(array)

		if err != nil {
			return err
		}
		err = gocudnn.CudaMemCopy(ptr, t.memgpu, sib, kind)
		if err != nil {
			return err
		}
		fmt.Printf("\n{")
		fmt.Println(comment, array)
		fmt.Printf("\n}")
	case flg.Int8():

		array := make([]int8, as)
		ptr, err := gocudnn.MakeGoPointer(array)
		if err != nil {

			return err
		}
		err = gocudnn.CudaMemCopy(ptr, t.memgpu, sib, kind)
		if err != nil {

			return err
		}
		fmt.Printf("\n{")
		fmt.Println(comment, array)
		fmt.Printf("\n}")
	default:
		return errors.New("Unsupoorted Format")
	}

	return nil
}

//PrintDeviceMem Kind of a shortcut function. I would like to build a more extensive function in the future where it would just know what to do without much user input.
func (t *Volume) PrintDeviceMem(comment string) error {
	kind := gocudnn.MemcpyKindFlag{}.DeviceToHost()
	return t.printmem(comment, kind)
}

//Destroy will release the memory of the tensor
func (t *Volume) Destroy() error {
	err := destroy(t)
	if err != nil {
		return err
	}
	t = nil
	return nil
}