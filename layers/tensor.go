//Package layers contains shared things between layers.  It also contains functions that will be supplimental to cudnn.
package layers

import (

	//"math/rand"
	"fmt"
	"sync"

	"github.com/dereklstinson/gocunets/devices/gpu/nvidia/cudnn"
	"github.com/dereklstinson/gocunets/devices/gpu/nvidia/cudnn/tensor"
	"github.com/dereklstinson/gocunets/utils"
	gocudnn "github.com/dereklstinson/gocudnn"
	"github.com/dereklstinson/gocudnn/gocu"
)

//Tensor is a tensor with volume
type Tensor struct {
	*tensor.Volume
	goptr                            *gocu.Wrapper
	minx, maxx, avgx, norm1x, norm2x *reduceop
	mux                              sync.Mutex
}

//ZeroClone returns a zeroed out clone of the tensor passed
func ZeroClone(h *cudnn.Handler, t *Tensor) (c *Tensor, err error) {
	c = new(Tensor)
	c.Volume, err = tensor.ZeroClone(h, t.Volume)
	return c, err
}

//MinX returns the minx value per batch in the tensor or if it is used for the filter it would be the minx value per neuron
func (t *Tensor) MinX(handle *cudnn.Handler) (x float32, err error) {

	if t.minx == nil {
		t.minx, err = buildminreduce(handle, t.Volume)
		if err != nil {
			return 0, err
		}
		err = handle.Sync()
		if err != nil {
			return 0, err
		}
	}
	t.mux.Lock()
	x, err = t.minx.Reduce(handle, t.Volume)
	t.mux.Unlock()
	return x, err

}

//MaxX returns the MaxX per batch value in the tensor or if it is used for the filter it would be the MaxX value per neuron
func (t *Tensor) MaxX(handle *cudnn.Handler) (x float32, err error) {
	if t.maxx == nil {
		t.maxx, err = buildmaxreduce(handle, t.Volume)
		if err != nil {
			return 0, err
		}
		err = handle.Sync()
		if err != nil {
			return 0, err
		}
	}
	t.mux.Lock()
	x, err = t.maxx.Reduce(handle, t.Volume)
	t.mux.Unlock()
	return x, err

}

//AvgX returns the Avg X value for the IO
func (t *Tensor) AvgX(handle *cudnn.Handler) (x float32, err error) {
	if t.avgx == nil {
		t.avgx, err = buildavgreduce(handle, t.Volume)
		if err != nil {
			return 0, err
		}
		err = handle.Sync()
		if err != nil {
			return 0, err
		}
	}
	t.mux.Lock()
	x, err = t.avgx.Reduce(handle, t.Volume)
	t.mux.Unlock()
	return x, err

}

//Norm1X returns Norm1 X value for IO
func (t *Tensor) Norm1X(handle *cudnn.Handler) (x float32, err error) {
	if t.norm1x == nil {
		t.norm1x, err = buildnorm1reduce(handle, t.Volume)
		if err != nil {
			return 0, err
		}
		err = handle.Sync()
		if err != nil {
			return 0, err
		}
	}
	t.mux.Lock()
	x, err = t.norm1x.Reduce(handle, t.Volume)
	t.mux.Unlock()
	return x, err

}

//Norm2X returns Norm2 X value for IO
func (t *Tensor) Norm2X(handle *cudnn.Handler) (x float32, err error) {
	if t.norm2x == nil {
		t.norm2x, err = buildnorm2reduce(handle, t.Volume)
		if err != nil {
			return 0, err
		}
		err = handle.Sync()
		if err != nil {
			return 0, err
		}
	}
	t.mux.Lock()
	x, err = t.norm2x.Reduce(handle, t.Volume)
	t.mux.Unlock()
	return x, err

}

//CreateTensor creates a tensor.
func CreateTensor(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (t *Tensor, err error) {
	t = new(Tensor)
	t.Volume, err = tensor.Build(handle, frmt, dtype, dims)

	return t, err

}

//LoadValuesFromSLice takes a go slice and fills it into the tensor sitting in the gpu.  If the length of goslice doesn't fit the input it will return an error
func (t *Tensor) LoadValuesFromSLice(handle *cudnn.Handler, input interface{}, length int32) error {
	if utils.FindVolumeInt32(t.Dims(), nil) != length {

		return fmt.Errorf("InputCurrent length (%v) not matching IO dims volume (%v)", length, utils.FindVolumeInt32(t.Dims(), nil))
	}
	var err error
	t.goptr, err = gocu.MakeGoMem(input)
	if err != nil {
		return err
	}
	err = t.LoadMem(handle, t.goptr, (t.goptr.TotalBytes()))
	if err != nil {
		return err
	}
	return handle.Sync()
}

//BuildRandomTensor builds a fix sized input
func BuildRandomTensor(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, mean, std float32) (*Tensor, error) {
	return buildRandIO(handle, frmt, dtype, dims, mean, std)

}
func buildRandIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, mean, std float32) (*Tensor, error) {

	x, err := tensor.BuildRandNorm(handle, frmt, dtype, dims, mean, std)
	if err != nil {
		return nil, err
	}

	return &Tensor{
		Volume: x,
	}, nil

}

/*
//IO is an all purpose struct that contains an x tensor and a dx tensor used for training
type IO struct {
	x                                     *tensor.Volume
	dx                                    *tensor.Volume
	minx, maxx, avgx, norm1x, norm2x      *reduceop
	mindx, maxdx, avgdx, norm1dx, norm2dx *reduceop
	input                                 bool
	dims                                  []int32
	weights                               bool
	mux                                   sync.Mutex
	gxptr                                 *gocu.Wrapper
	gdxptr                                *gocu.Wrapper
}

//Settings contains the info that is needed to build an IO
type Settings struct {
	Dims     []int32              `json:"dims,omitempty"`
	Managed  bool                 `json:"managed,omitempty"`
	Format   gocudnn.TensorFormat `json:"format,omitempty"`
	DataType gocudnn.DataType     `json:"data_type,omitempty"`
	NanProp  gocudnn.NANProp      `json:"nan_prop,omitempty"`
}

//Info is a struct that contains all the information to build an IO struct
type Info struct {
	NetworkInput bool        `json:"NetworkInput"`
	Dims         []int32     `json:"Dims"`
	Unified      bool        `json:"Unified"`
	X            tensor.Info `json:"X"`
	Dx           tensor.Info `json:"dX"`
}



//StoreDeltas will flip a flag to allow deltas to be stored on this IO.
//Useful when training gans when you don't want the errors when training the descriminator to propigate through this.
//You would want to switch it back when passing the errors for the generator.
func (i *IO) StoreDeltas(x bool) {
	i.input = x
}

//ClearDeltas allows the user to clear the deltas of the IO
func (i *IO) ClearDeltas() error {
	return i.dx.Memer().SetAll(0)

}

//MinX returns the minx value per batch in the tensor or if it is used for the filter it would be the minx value per neuron
func (i *IO) MinX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.minx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//MaxX returns the MaxX per batch value in the tensor or if it is used for the filter it would be the MaxX value per neuron
func (i *IO) MaxX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.maxx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//AvgX returns the Avg X value for the IO
func (i *IO) AvgX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.avgx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//Norm1X returns Norm1 X value for IO
func (i *IO) Norm1X(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.norm1x.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//Norm2X returns Norm2 X value for IO
func (i *IO) Norm2X(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.norm2x.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//MinDX returns the min dx value for th io.
func (i *IO) MinDX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.mindx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//MaxDX returns the Max DX  value for th io.
func (i *IO) MaxDX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.maxdx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//AvgDX returns the Avg DX value for the IO
func (i *IO) AvgDX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.avgdx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//Norm1DX returns Norm1 dX value for IO
func (i *IO) Norm1DX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.norm1dx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//Norm2DX returns Norm2 dX value for IO
func (i *IO) Norm2DX(handle *cudnn.Handler) (x float32, e error) {
	i.mux.Lock()
	x, e = i.norm2dx.Reduce(handle, i.T())
	i.mux.Unlock()
	return x, e

}

//Info returns info
func (i *IO) Info() (Info, error) {
	x, err := i.x.Info()
	if err != nil {
		return Info{}, err
	}
	dx, err := i.dx.Info()
	if err != nil {
		return Info{}, err
	}
	return Info{
		NetworkInput: i.input,
		Dims:         i.dims,
		//	Unified:      i.managed,
		X:  x,
		Dx: dx,
	}, nil
}

//IsInput returns if it is an input
func (i *IO) IsInput() bool {
	return i.input
}

func findslide(dims []int32) []int {
	multiplier := 1
	slide := make([]int, len(dims))
	for i := len(dims) - 1; i >= 0; i-- {
		slide[i] = multiplier
		multiplier *= int(dims[i])
	}
	return slide
}
func findvol(dims []int32) int {
	multiplier := 1

	for i := 0; i < len(dims); i++ {

		multiplier *= int(dims[i])
	}
	return multiplier
}

//Properties returns the tensorformat, datatype and a slice of dims that describe the tensor
func (i *IO) Properties() (gocudnn.TensorFormat, gocudnn.DataType, []int32, error) {
	return i.x.Properties()
}

//DeltaT returns d tensor
func (i *IO) DeltaT() *tensor.Volume {
	return i.dx
}

//T returns the tensor
func (i *IO) T() *tensor.Volume {
	return i.x
}
func addtoerror(addition string, current error) error {
	errorstring := current.Error()
	return errors.New(addition + ": " + errorstring)
}

//SetXStatReducers will build the reducers for the IO x min,max,avg,norm1, norm2
func (i *IO) SetXStatReducers(handle *cudnn.Handler) (err error) {

	i.minx, err = buildminreduce(handle, i.T())
	if err != nil {
		return err
	}

	i.maxx, err = buildmaxreduce(handle, i.T())
	if err != nil {
		return err
	}
	i.avgx, err = buildavgreduce(handle, i.T())
	if err != nil {
		return err
	}
	i.norm1x, err = buildnorm1reduce(handle, i.T())
	if err != nil {
		return err
	}
	i.norm2x, err = buildnorm2reduce(handle, i.T())
	if err != nil {
		return err
	}
	return err
}

//SetDXStatReducers will build the reducers for the IO dx min,max,avg,norm1, norm2
func (i *IO) SetDXStatReducers(handle *cudnn.Handler) (err error) {
	i.mindx, err = buildminreduce(handle, i.DeltaT())
	if err != nil {
		return err
	}
	i.maxdx, err = buildmaxreduce(handle, i.DeltaT())
	if err != nil {
		return err
	}
	i.avgdx, err = buildavgreduce(handle, i.DeltaT())
	if err != nil {
		return err
	}
	i.norm1dx, err = buildnorm1reduce(handle, i.DeltaT())
	if err != nil {
		return err
	}
	i.norm2dx, err = buildnorm2reduce(handle, i.DeltaT())
	if err != nil {
		return err
	}
	return err
}

//ZeroCloneInference does zeroclone inferenence
func (i *IO) ZeroCloneInference(handle *cudnn.Handler) (*IO, error) {
	frmt, dtype, dims, err := i.Properties()
	if err != nil {
		return nil, err
	}
	return BuildInferenceIO(handle, frmt, dtype, dims)
}

//ZeroClone Makes a zeroclone of the IO
func (i *IO) ZeroClone(handle *cudnn.Handler) (*IO, error) {
	frmt, dtype, dims, err := i.Properties()
	if err != nil {
		return nil, err
	}

	return buildIO(handle, frmt, dtype, dims, i.input, i.weights)
}

//BuildIOWeightsT builds The IOweights
func BuildIOWeightsT(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (*IO, error) {
	return buildIO(handle, frmt, dtype, dims, true, true)
}

//BuildIOWeights builds the weights for the IO
func BuildIOWeights(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (*IO, error) {
	return buildIO(handle, frmt, dtype, dims, false, true)
}

//BuildIO builds a regular IO with both a T tensor and a DeltaT tensor
func BuildIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (*IO, error) {

	return buildIO(handle, frmt, dtype, dims, false, false)
}

//BuildNormRandIO builds a regular IO with both a T tensor and a DeltaT tensor.  But the T tensor is randomized
func BuildNormRandIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, mean, std float32, seed uint64) (*IO, error) {
	return buildRandIO(handle, frmt, dtype, dims, mean, std, seed, false, false)

}



//BuildNormRandInputIO builds a regular IO but the input is set to nil
func BuildNormRandInputIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, mean, std float32, seed uint64) (*IO, error) {
	return buildRandIO(handle, frmt, dtype, dims, mean, std, seed, true, false)

}


//BuildNetworkInputHost build the input tensor to paged memory on host ram
func BuildNetworkInputHost(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (*IO, error) {
	x, err := tensor.BuildtoCudaHost(handle, frmt, dtype, dims)
	if err != nil {
		return nil, err
	}

	return &IO{
		x:     x,
		dx:    nil,
		input: true,
	}, nil
}

//BuildNetworkInputIO builds an input IO which is an IO with DeltaT() set to nil. This is used for the input or the output of a network.
//If it is the output of a network in training. Then DeltaT will Need to be loaded with the labeles between batches.
func BuildNetworkInputIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (*IO, error) {
	return buildIO(handle, frmt, dtype, dims, true, false)
}

//BuildInferenceIO builds an IO used for only inference.  It doesn't contain a tensor for the errors.
func BuildInferenceIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32) (*IO, error) {
	return buildIO(handle, frmt, dtype, dims, true, false)
}

//BuildNetworkOutputIOFromSlice will return IO with the slice put into the DeltaT() section of the IO
func BuildNetworkOutputIOFromSlice(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, slice []float32) (*IO, error) {

	chkr := int32(1)
	for i := 0; i < len(dims); i++ {
		chkr *= dims[i]
	}
	if chkr != int32(len(slice)) {
		return nil, errors.New("Slice passed length don't match dim volume")
	}

	slice2 := make([]float32, handle.FindVol(dims))
	copy(slice2, slice)

	newio, err := BuildIO(handle, frmt, dtype, dims)
	if err != nil {
		return nil, err
	}

	err = newio.LoadDeltaTValuesFromGoSlice(handle, slice, int32(len(slice)))
	if err != nil {
		return nil, err
	}
	return newio, err
}

func buildIO(handle *cudnn.Handler, frmt gocudnn.TensorFormat, dtype gocudnn.DataType, dims []int32, input, weights bool) (*IO, error) {

	if input {
		if weights {
			x, err := tensor.BuildWeights(handle, frmt, dtype, dims)
			if err != nil {

				return nil, err
			}

			return &IO{
				x: x,
			}, nil
		}
		x, err := tensor.Build(handle, frmt, dtype, dims)
		if err != nil {
			return nil, err
		}

		return &IO{

			x:     x,
			dx:    nil,
			input: true,
		}, nil

	}
	if weights {
		x, err := tensor.BuildWeights(handle, frmt, dtype, dims)
		if err != nil {

			return nil, err
		}
		dx, err := tensor.BuildWeights(handle, frmt, dtype, dims)
		if err != nil {

			return nil, err
		}
		return &IO{
			x:  x,
			dx: dx,
		}, nil
	}
	x, err := tensor.Build(handle, frmt, dtype, dims)
	if err != nil {

		return nil, err
	}
	dx, err := tensor.Build(handle, frmt, dtype, dims)
	if err != nil {

		return nil, err
	}
	return &IO{
		x:  x,
		dx: dx,
	}, nil
}



//LoadTValuesFromGoSlice takes a go slice and fills it into the tensor sitting in the gpu.  If the length of goslice doesn't fit the input it will return an error
func (i *IO) LoadTValuesFromGoSlice(handle *cudnn.Handler, input interface{}, length int32) error {
	if utils.FindVolumeInt32(i.x.Dims(), nil) != length {
		return errors.New("InputCurrent length not matching IO dims volume")
	}
	var err error
	i.gxptr, err = gocu.MakeGoMem(input)
	if err != nil {
		return err
	}
	err = i.x.LoadMem(handle, i.gxptr, (i.gxptr.TotalBytes()))
	if err != nil {
		return err
	}
	return handle.Sync()
}

//GetLength returns the length in int32
func (i *IO) GetLength() (int32, error) {
	_, _, dims, err := i.Properties()
	if err != nil {
		return 0, err
	}
	mult := int32(1)
	for i := 0; i < len(dims); i++ {
		mult *= dims[i]
	}
	return mult, nil
}

//LoadDeltaTValuesFromGoSlice takes a go slice and fills it into the tensor sitting in the gpu.  If the length of goslice doesn't fit the input it will return an error
func (i *IO) LoadDeltaTValuesFromGoSlice(handle *cudnn.Handler, input interface{}, length int32) error {
	if i.input == true {
		return errors.New("Can't Load any values into DeltaT because it is exclusivly an Input")
	}
	if utils.FindVolumeInt32(i.dx.Dims(), nil) != length {
		return errors.New("InputCurrent length not matching IO dims volume")
	}
	var err error
	i.gdxptr, err = gocu.MakeGoMem(input)
	if err != nil {
		return err
	}
	return i.dx.LoadMem(handle, i.gdxptr, (i.gdxptr.TotalBytes()))
}

//LoadDeltaTValues loads a piece of memory that was made in golang and loads into a previously created delta tensor volume in cuda.
func (i *IO) LoadDeltaTValues(handle *cudnn.Handler, input *tensor.Volume) error {
	if i.input == true {
		return errors.New("Can't Load any values into DeltaT because it is exclusivly an Input")
	}
	if utils.FindVolumeInt32(i.dx.Dims(), nil) != utils.FindVolumeInt32(input.Dims(), nil) {
		return errors.New("InputCurrent dims not matching IO current dims")
	}
	err := i.dx.LoadMem(handle, input.Memer(), input.SIB())
	if err != nil {
		return err
	}
	return handle.Sync()
}

*/
